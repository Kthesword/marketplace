// Package order handles checkout and order queries.
package order

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
	SELECT o.id, o.checkout_id::text, o.buyer_id, o.seller_id, COALESCE(s.name, ''),
	       o.status, o.total_cents, o.created_at, o.updated_at
	FROM orders o
	LEFT JOIN stores s ON s.seller_id = o.seller_id
`

func scanOrder(row pgx.Row) (Order, error) {
	var o Order
	err := row.Scan(&o.ID, &o.CheckoutID, &o.BuyerID, &o.SellerID, &o.StoreName,
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
