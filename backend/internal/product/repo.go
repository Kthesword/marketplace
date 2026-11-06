// Package product manages product CRUD for sellers and catalog browsing for buyers.
package product

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("product not found")

// Product represents a row in the products table, enriched with StoreName when joined.
type Product struct {
	ID          int64     `json:"id"`
	SellerID    int64     `json:"seller_id"`
	StoreName   string    `json:"store_name"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	PriceCents  int64     `json:"price_cents"`
	Stock       int       `json:"stock"`
	ImageURL    string    `json:"image_url"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Input struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int64  `json:"price_cents"`
	Stock       int    `json:"stock"`
	ImageURL    string `json:"image_url"`
	Active      *bool  `json:"active"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

// Create inserts a new product for sellerID.
func (r *Repo) Create(ctx context.Context, sellerID int64, storeName string, in Input, active bool) (Product, error) {
	const q = `
		INSERT INTO products (seller_id, name, description, price_cents, stock, image_url, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, seller_id, name, description, price_cents, stock, image_url, active, created_at, updated_at
	`
	var p Product
	err := r.pool.QueryRow(ctx, q, sellerID, in.Name, in.Description, in.PriceCents, in.Stock, in.ImageURL, active).
		Scan(&p.ID, &p.SellerID, &p.Name, &p.Description, &p.PriceCents, &p.Stock, &p.ImageURL, &p.Active, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return Product{}, fmt.Errorf("create product: %w", err)
	}
	p.StoreName = storeName
	return p, nil
}

// Update modifies a product ONLY if it is owned by sellerID (`WHERE id = $1 AND seller_id = $2`).
// If the product does not exist OR belongs to another seller, pgx.ErrNoRows -> ErrNotFound.
func (r *Repo) Update(ctx context.Context, id, sellerID int64, storeName string, in Input, active bool) (Product, error) {
	const q = `
		UPDATE products
		SET name = $3, description = $4, price_cents = $5, stock = $6, image_url = $7, active = $8, updated_at = now()
		WHERE id = $1 AND seller_id = $2
		RETURNING id, seller_id, name, description, price_cents, stock, image_url, active, created_at, updated_at
	`
	var p Product
	err := r.pool.QueryRow(ctx, q, id, sellerID, in.Name, in.Description, in.PriceCents, in.Stock, in.ImageURL, active).
		Scan(&p.ID, &p.SellerID, &p.Name, &p.Description, &p.PriceCents, &p.Stock, &p.ImageURL, &p.Active, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Product{}, ErrNotFound
	}
	if err != nil {
		return Product{}, fmt.Errorf("update product: %w", err)
	}
	p.StoreName = storeName
	return p, nil
}

// Delete removes a product ONLY if owned by sellerID (`WHERE id = $1 AND seller_id = $2`).
func (r *Repo) Delete(ctx context.Context, id, sellerID int64) error {
	const q = `DELETE FROM products WHERE id = $1 AND seller_id = $2`
	tag, err := r.pool.Exec(ctx, q, id, sellerID)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListBySeller returns a seller's products. When includeInactive is true (for /api/seller/products),
// both active and inactive products are returned; otherwise only active ones.
func (r *Repo) ListBySeller(ctx context.Context, sellerID int64, includeInactive bool, limit, offset int) ([]Product, int, error) {
	const countQ = `
		SELECT COUNT(*)
		FROM products p
		LEFT JOIN stores s ON s.seller_id = p.seller_id
		WHERE p.seller_id = $1 AND ($2::bool OR p.active = true)
	`
	var total int
	if err := r.pool.QueryRow(ctx, countQ, sellerID, includeInactive).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count seller products: %w", err)
	}

	const listQ = `
		SELECT p.id, p.seller_id, COALESCE(s.name, ''), p.name, p.description,
		       p.price_cents, p.stock, p.image_url, p.active, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN stores s ON s.seller_id = p.seller_id
		WHERE p.seller_id = $1 AND ($2::bool OR p.active = true)
		ORDER BY p.created_at DESC, p.id DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.pool.Query(ctx, listQ, sellerID, includeInactive, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list seller products: %w", err)
	}
	defer rows.Close()

	items := make([]Product, 0)
	for rows.Next() {
		var p Product
		if err := rows.Scan(
			&p.ID, &p.SellerID, &p.StoreName, &p.Name, &p.Description,
			&p.PriceCents, &p.Stock, &p.ImageURL, &p.Active, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan product: %w", err)
		}
		items = append(items, p)
	}
	return items, total, rows.Err()
}

// ListPublic returns active products from sellers with a store, with optional case-insensitive search.
func (r *Repo) ListPublic(ctx context.Context, search string, limit, offset int) ([]Product, int, error) {
	search = strings.TrimSpace(search)
	pattern := "%" + search + "%"

	const countQ = `
		SELECT COUNT(*)
		FROM products p
		JOIN stores s ON s.seller_id = p.seller_id
		WHERE p.active = true
		  AND ($1 = '' OR p.name ILIKE $2 OR p.description ILIKE $2)
	`
	var total int
	if err := r.pool.QueryRow(ctx, countQ, search, pattern).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count public products: %w", err)
	}

	const listQ = `
		SELECT p.id, p.seller_id, s.name, p.name, p.description,
		       p.price_cents, p.stock, p.image_url, p.active, p.created_at, p.updated_at
		FROM products p
		JOIN stores s ON s.seller_id = p.seller_id
		WHERE p.active = true
		  AND ($1 = '' OR p.name ILIKE $2 OR p.description ILIKE $2)
		ORDER BY p.created_at DESC, p.id DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.pool.Query(ctx, listQ, search, pattern, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list public products: %w", err)
	}
	defer rows.Close()

	items := make([]Product, 0)
	for rows.Next() {
		var p Product
		if err := rows.Scan(
			&p.ID, &p.SellerID, &p.StoreName, &p.Name, &p.Description,
			&p.PriceCents, &p.Stock, &p.ImageURL, &p.Active, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan public product: %w", err)
		}
		items = append(items, p)
	}
	return items, total, rows.Err()
}

// GetPublicByID fetches a single active product with its seller's store name.
func (r *Repo) GetPublicByID(ctx context.Context, id int64) (Product, error) {
	const q = `
		SELECT p.id, p.seller_id, s.name, p.name, p.description,
		       p.price_cents, p.stock, p.image_url, p.active, p.created_at, p.updated_at
		FROM products p
		JOIN stores s ON s.seller_id = p.seller_id
		WHERE p.id = $1 AND p.active = true
	`
	var p Product
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&p.ID, &p.SellerID, &p.StoreName, &p.Name, &p.Description,
		&p.PriceCents, &p.Stock, &p.ImageURL, &p.Active, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Product{}, ErrNotFound
	}
	if err != nil {
		return Product{}, fmt.Errorf("get product by id: %w", err)
	}
	return p, nil
}
