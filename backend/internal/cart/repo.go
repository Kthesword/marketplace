// Package cart manages a buyer's shopping cart items.
package cart

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrProductNotFound   = errors.New("product not found or inactive")
	ErrItemNotInCart     = errors.New("item not in cart")
	ErrInsufficientStock = errors.New("requested quantity exceeds available stock")
)

// Item is a single line in the buyer's cart, joined with live product & store info.
type Item struct {
	ProductID      int64  `json:"product_id"`
	SellerID       int64  `json:"seller_id"`
	StoreName      string `json:"store_name"`
	Name           string `json:"name"`
	ImageURL       string `json:"image_url"`
	PriceCents     int64  `json:"price_cents"`
	Stock          int    `json:"stock"`
	Quantity       int    `json:"quantity"`
	LineTotalCents int64  `json:"line_total_cents"`
}

// Cart is the full response returned by /api/cart/items endpoints.
type Cart struct {
	Items      []Item `json:"items"`
	TotalCents int64  `json:"total_cents"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

// GetCart fetches all active items in userID's cart and computes line totals and cart total.
func (r *Repo) GetCart(ctx context.Context, userID int64) (Cart, error) {
	const q = `
		SELECT c.product_id, p.seller_id, s.name, p.name, p.image_url,
		       p.price_cents, p.stock, c.quantity
		FROM cart_items c
		JOIN products p ON p.id = c.product_id
		JOIN stores s ON s.seller_id = p.seller_id
		WHERE c.user_id = $1 AND p.active = true
		ORDER BY c.added_at ASC, c.product_id ASC
	`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return Cart{}, fmt.Errorf("query cart: %w", err)
	}
	defer rows.Close()

	cart := Cart{Items: make([]Item, 0)}
	for rows.Next() {
		var it Item
		if err := rows.Scan(
			&it.ProductID, &it.SellerID, &it.StoreName, &it.Name, &it.ImageURL,
			&it.PriceCents, &it.Stock, &it.Quantity,
		); err != nil {
			return Cart{}, fmt.Errorf("scan cart item: %w", err)
		}
		it.LineTotalCents = it.PriceCents * int64(it.Quantity)
		cart.TotalCents += it.LineTotalCents
		cart.Items = append(cart.Items, it)
	}
	return cart, rows.Err()
}

// AddItem adds quantity of productID to userID's cart (or increments if already present),
// verifying the product is active and the resulting quantity does not exceed available stock.
func (r *Repo) AddItem(ctx context.Context, userID, productID int64, addQty int) error {
	var stock int
	var existingQty int
	const checkQ = `
		SELECT p.stock, COALESCE(c.quantity, 0)
		FROM products p
		JOIN stores s ON s.seller_id = p.seller_id
		LEFT JOIN cart_items c ON c.product_id = p.id AND c.user_id = $1
		WHERE p.id = $2 AND p.active = true
	`
	err := r.pool.QueryRow(ctx, checkQ, userID, productID).Scan(&stock, &existingQty)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrProductNotFound
	}
	if err != nil {
		return fmt.Errorf("check product stock: %w", err)
	}

	if existingQty+addQty > stock {
		return ErrInsufficientStock
	}

	const upsertQ = `
		INSERT INTO cart_items (user_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, product_id)
		DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity
	`
	_, err = r.pool.Exec(ctx, upsertQ, userID, productID, addQty)
	if err != nil {
		return fmt.Errorf("upsert cart item: %w", err)
	}
	return nil
}

// SetQuantity sets the exact quantity for an item already in userID's cart (idempotent).
func (r *Repo) SetQuantity(ctx context.Context, userID, productID int64, qty int) error {
	var stock int
	const stockQ = `
		SELECT p.stock
		FROM products p
		WHERE p.id = $1 AND p.active = true
	`
	err := r.pool.QueryRow(ctx, stockQ, productID).Scan(&stock)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrProductNotFound
	}
	if err != nil {
		return fmt.Errorf("check product stock: %w", err)
	}
	if qty > stock {
		return ErrInsufficientStock
	}

	const updateQ = `
		UPDATE cart_items
		SET quantity = $3
		WHERE user_id = $1 AND product_id = $2
	`
	tag, err := r.pool.Exec(ctx, updateQ, userID, productID, qty)
	if err != nil {
		return fmt.Errorf("update cart item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrItemNotInCart
	}
	return nil
}

// RemoveItem deletes productID from userID's cart.
func (r *Repo) RemoveItem(ctx context.Context, userID, productID int64) error {
	const q = `DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2`
	tag, err := r.pool.Exec(ctx, q, userID, productID)
	if err != nil {
		return fmt.Errorf("remove cart item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrItemNotInCart
	}
	return nil
}

