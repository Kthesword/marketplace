// Package order handles checkout and order queries.
package order

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kthesword/marketplace/backend/internal/db"
)

var ErrNotFound = errors.New("order not found")

// Order statuses.
const (
	StatusPending   = "Pending"
	StatusShipped   = "Shipped"
	StatusDelivered = "Delivered"
	StatusCancelled = "Cancelled"
)

// Item is an order line. Name and price are SNAPSHOTS taken at purchase time.
type Item struct {
	ID             int64  `json:"id"`
	ProductID      *int64 `json:"product_id"` // nil if the product was deleted later
	ProductName    string `json:"product_name"`
	UnitPriceCents int64  `json:"unit_price_cents"`
	Quantity       int    `json:"quantity"`
	LineTotalCents int64  `json:"line_total_cents"`
}

type Order struct {
	ID         int64     `json:"id"`
	CheckoutID string    `json:"checkout_id"`
	BuyerID    int64     `json:"buyer_id"`
	BuyerName  string    `json:"buyer_name"`
	SellerID   int64     `json:"seller_id"`
	StoreName  string    `json:"store_name"`
	Status     string    `json:"status"`
	TotalCents int64     `json:"total_cents"`
	Items      []Item    `json:"items"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

const orderSelect = `
	SELECT o.id, o.checkout_id::text, o.buyer_id, COALESCE(u.name, ''),
	       o.seller_id, COALESCE(s.name, ''),
	       o.status, o.total_cents, o.created_at, o.updated_at
	FROM orders o
	LEFT JOIN stores s ON s.seller_id = o.seller_id
	LEFT JOIN users u ON u.id = o.buyer_id
`

func scanOrder(row pgx.Row) (Order, error) {
	var o Order
	err := row.Scan(&o.ID, &o.CheckoutID, &o.BuyerID, &o.BuyerName,
		&o.SellerID, &o.StoreName,
		&o.Status, &o.TotalCents, &o.CreatedAt, &o.UpdatedAt)
	o.Items = []Item{}
	return o, err
}

// ListByBuyer returns one page of the buyer's own orders (newest first) with items.
func (r *Repo) ListByBuyer(ctx context.Context, buyerID int64, limit, offset int) ([]Order, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM orders WHERE buyer_id = $1`, buyerID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		orderSelect+` WHERE o.buyer_id = $1 ORDER BY o.created_at DESC, o.id DESC LIMIT $2 OFFSET $3`,
		buyerID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	orders := make([]Order, 0)
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	rows.Close()

	if err := r.attachItems(ctx, orders); err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

// GetForBuyer returns an order only if it belongs to buyerID; otherwise ErrNotFound.
// The ownership check lives in the WHERE clause, so other buyers' orders simply "don't exist" for you.
func (r *Repo) GetForBuyer(ctx context.Context, id, buyerID int64) (Order, error) {
	o, err := scanOrder(r.pool.QueryRow(ctx, orderSelect+` WHERE o.id = $1 AND o.buyer_id = $2`, id, buyerID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, ErrNotFound
	}
	if err != nil {
		return Order{}, fmt.Errorf("get order: %w", err)
	}
	orders := []Order{o}
	if err := r.attachItems(ctx, orders); err != nil {
		return Order{}, err
	}
	return orders[0], nil
}

// GetForSeller returns an order only if it belongs to sellerID; otherwise ErrNotFound.
func (r *Repo) GetForSeller(ctx context.Context, id, sellerID int64) (Order, error) {
	o, err := scanOrder(r.pool.QueryRow(ctx, orderSelect+` WHERE o.id = $1 AND o.seller_id = $2`, id, sellerID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, ErrNotFound
	}
	if err != nil {
		return Order{}, fmt.Errorf("get seller order: %w", err)
	}
	orders := []Order{o}
	if err := r.attachItems(ctx, orders); err != nil {
		return Order{}, err
	}
	return orders[0], nil
}

// ListBySeller returns one page of orders placed against sellerID, optionally filtered by status.
func (r *Repo) ListBySeller(ctx context.Context, sellerID int64, statusFilter string, limit, offset int) ([]Order, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM orders
		WHERE seller_id = $1 AND ($2 = '' OR status = $2)
	`, sellerID, statusFilter).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count seller orders: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		orderSelect+` WHERE o.seller_id = $1 AND ($2 = '' OR o.status = $2)
		ORDER BY o.created_at DESC, o.id DESC LIMIT $3 OFFSET $4`,
		sellerID, statusFilter, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list seller orders: %w", err)
	}
	defer rows.Close()

	orders := make([]Order, 0)
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan seller order: %w", err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	rows.Close()

	if err := r.attachItems(ctx, orders); err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

// UpdateStatusBySeller advances a seller's order along Pending -> Shipped -> Delivered.
// Uses compare-and-swap (WHERE status = currentStatus) so a concurrent buyer cancellation
// cannot be silently overwritten by a seller marking the order as Shipped.
func (r *Repo) UpdateStatusBySeller(ctx context.Context, orderID, sellerID int64, newStatus string) (Order, error) {
	current, err := r.GetForSeller(ctx, orderID, sellerID)
	if err != nil {
		return Order{}, err
	}
	if !CanSellerTransition(current.Status, newStatus) {
		return Order{}, fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidTransition, current.Status, newStatus)
	}

	tag, err := r.pool.Exec(ctx, `
		UPDATE orders
		SET status = $3, updated_at = now()
		WHERE id = $1 AND seller_id = $2 AND status = $4
	`, orderID, sellerID, newStatus, current.Status)
	if err != nil {
		return Order{}, fmt.Errorf("update order status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		// Another request changed the status between our SELECT and UPDATE.
		return Order{}, fmt.Errorf("%w: order status changed concurrently", ErrInvalidTransition)
	}
	return r.GetForSeller(ctx, orderID, sellerID)
}

// CancelByBuyer cancels a buyer's own Pending order and atomically restores product stock.
func (r *Repo) CancelByBuyer(ctx context.Context, orderID, buyerID int64) (Order, error) {
	err := db.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		var status string
		err := tx.QueryRow(ctx, `
			SELECT status FROM orders
			WHERE id = $1 AND buyer_id = $2
			FOR UPDATE
		`, orderID, buyerID).Scan(&status)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("lock order for cancel: %w", err)
		}

		if !CanBuyerCancel(status) {
			return fmt.Errorf("%w: only Pending orders can be cancelled (current status: %s)", ErrInvalidTransition, status)
		}

		if _, err := tx.Exec(ctx, `
			UPDATE orders SET status = $2, updated_at = now() WHERE id = $1
		`, orderID, StatusCancelled); err != nil {
			return fmt.Errorf("mark order cancelled: %w", err)
		}

		// Restore stock for each item whose product still exists (product_id IS NOT NULL).
		if _, err := tx.Exec(ctx, `
			UPDATE products p
			SET stock = p.stock + oi.quantity, updated_at = now()
			FROM order_items oi
			WHERE oi.order_id = $1 AND oi.product_id = p.id
		`, orderID); err != nil {
			return fmt.Errorf("restore stock on cancel: %w", err)
		}
		return nil
	})
	if err != nil {
		return Order{}, err
	}
	return r.GetForBuyer(ctx, orderID, buyerID)
}

// attachItems loads items for all given orders with ONE query (avoids N+1 queries).
func (r *Repo) attachItems(ctx context.Context, orders []Order) error {
	if len(orders) == 0 {
		return nil
	}
	ids := make([]int64, len(orders))
	index := make(map[int64]int, len(orders))
	for i, o := range orders {
		ids[i] = o.ID
		index[o.ID] = i
	}

	rows, err := r.pool.Query(ctx, `
		SELECT order_id, id, product_id, product_name, unit_price_cents, quantity
		FROM order_items WHERE order_id = ANY($1) ORDER BY id`, ids)
	if err != nil {
		return fmt.Errorf("load order items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var orderID int64
		var it Item
		if err := rows.Scan(&orderID, &it.ID, &it.ProductID, &it.ProductName, &it.UnitPriceCents, &it.Quantity); err != nil {
			return fmt.Errorf("scan order item: %w", err)
		}
		it.LineTotalCents = it.UnitPriceCents * int64(it.Quantity)
		i := index[orderID]
		orders[i].Items = append(orders[i].Items, it)
	}
	return rows.Err()
}
