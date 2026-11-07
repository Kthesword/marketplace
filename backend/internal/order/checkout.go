package order

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kthesword/marketplace/backend/internal/db"
)

// CheckoutError is a business-rule failure (maps to HTTP 409), as opposed to a system error.
type CheckoutError struct {
	Code    string
	Message string
}

func (e *CheckoutError) Error() string { return e.Code + ": " + e.Message }

// CheckoutResult is what a successful checkout returns.
type CheckoutResult struct {
	CheckoutID string  `json:"checkout_id"`
	Orders     []Order `json:"orders"`
}

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service { return &Service{pool: pool} }

type cartLine struct {
	productID int64
	sellerID  int64
	name      string
	price     int64
	stock     int
	active    bool
	qty       int
}

// Checkout turns the buyer's cart into one order per seller, atomically:
//  1. lock the cart's product rows (FOR UPDATE, ordered by id => no deadlocks)
//  2. validate stock and decrement it, line by line
//  3. create one order per seller, with name/price SNAPSHOTS in order_items
//  4. clear the cart
//
// Any error rolls back EVERYTHING (stock, orders, cart stay untouched).
// Payment is mocked: it always "succeeds".
func (s *Service) Checkout(ctx context.Context, buyerID int64) (CheckoutResult, error) {
	var result CheckoutResult

	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		// 1. Read cart + lock product rows until this transaction ends.
		rows, err := tx.Query(ctx, `
			SELECT p.id, p.seller_id, p.name, p.price_cents, p.stock, p.active, c.quantity
			FROM cart_items c
			JOIN products p ON p.id = c.product_id
			WHERE c.user_id = $1
			ORDER BY p.id
			FOR UPDATE OF p`, buyerID)
		if err != nil {
			return fmt.Errorf("lock cart products: %w", err)
		}
		var lines []cartLine
		for rows.Next() {
			var l cartLine
			if err := rows.Scan(&l.productID, &l.sellerID, &l.name, &l.price, &l.stock, &l.active, &l.qty); err != nil {
				rows.Close()
				return fmt.Errorf("scan cart line: %w", err)
			}
			lines = append(lines, l)
		}
		rows.Close() // must close before running another query on the same tx
		if err := rows.Err(); err != nil {
			return err
		}
		if len(lines) == 0 {
			return &CheckoutError{"cart_empty", "your cart is empty"}
		}

		// 2. Validate and decrement stock. A failure on line N rolls back lines 1..N-1 too.
		for _, l := range lines {
			if !l.active {
				return &CheckoutError{"product_unavailable", fmt.Sprintf("%q is no longer available", l.name)}
			}
			if l.stock < l.qty {
				return &CheckoutError{"insufficient_stock",
					fmt.Sprintf("only %d of %q left in stock (you have %d in your cart)", l.stock, l.name, l.qty)}
			}
			// The "AND stock >= $2" guard is a second safety net next to the row lock.
			tag, err := tx.Exec(ctx,
				`UPDATE products SET stock = stock - $2, updated_at = now() WHERE id = $1 AND stock >= $2`,
				l.productID, l.qty)
			if err != nil {
				return fmt.Errorf("decrement stock: %w", err)
			}
			if tag.RowsAffected() == 0 {
				return &CheckoutError{"insufficient_stock", fmt.Sprintf("%q just sold out", l.name)}
			}
		}

		// 3. Group lines by seller (keeping first-seen order) and create one order each.
		var sellerOrder []int64
		bySeller := make(map[int64][]cartLine)
		for _, l := range lines {
			if _, seen := bySeller[l.sellerID]; !seen {
				sellerOrder = append(sellerOrder, l.sellerID)
			}
			bySeller[l.sellerID] = append(bySeller[l.sellerID], l)
		}

		if err := tx.QueryRow(ctx, `SELECT gen_random_uuid()::text`).Scan(&result.CheckoutID); err != nil {
			return fmt.Errorf("generate checkout id: %w", err)
		}

		for _, sellerID := range sellerOrder {
			group := bySeller[sellerID]

			var total int64
			for _, l := range group {
				total += l.price * int64(l.qty)
			}

			var o Order
			o.CheckoutID = result.CheckoutID
			o.BuyerID = buyerID
			o.SellerID = sellerID
			o.Status = StatusPending
			o.TotalCents = total
			err := tx.QueryRow(ctx, `
				INSERT INTO orders (checkout_id, buyer_id, seller_id, status, total_cents)
				VALUES ($1, $2, $3, $4, $5)
				RETURNING id, created_at, updated_at`,
				result.CheckoutID, buyerID, sellerID, StatusPending, total).
				Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt)
			if err != nil {
				return fmt.Errorf("insert order: %w", err)
			}
			// Store name is display-only; a missing store just yields "".
			_ = tx.QueryRow(ctx, `SELECT name FROM stores WHERE seller_id = $1`, sellerID).Scan(&o.StoreName)

			o.Items = make([]Item, 0, len(group))
			for _, l := range group {
				pid := l.productID
				it := Item{ProductID: &pid, ProductName: l.name, UnitPriceCents: l.price, Quantity: l.qty}
				if err := tx.QueryRow(ctx, `
					INSERT INTO order_items (order_id, product_id, product_name, unit_price_cents, quantity)
					VALUES ($1, $2, $3, $4, $5) RETURNING id`,
					o.ID, l.productID, l.name, l.price, l.qty).Scan(&it.ID); err != nil {
					return fmt.Errorf("insert order item: %w", err)
				}
				it.LineTotalCents = l.price * int64(l.qty)
				o.Items = append(o.Items, it)
			}
			result.Orders = append(result.Orders, o)
		}

		// 4. Clear the cart.
		if _, err := tx.Exec(ctx, `DELETE FROM cart_items WHERE user_id = $1`, buyerID); err != nil {
			return fmt.Errorf("clear cart: %w", err)
		}
		return nil
	})
	if err != nil {
		return CheckoutResult{}, err
	}
	return result, nil
}
