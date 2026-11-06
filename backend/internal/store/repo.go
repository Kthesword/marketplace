// Package store manages seller stores (one store per seller).
package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound    = errors.New("store not found")
	ErrStoreExists = errors.New("seller already has a store")
)

// Store represents a row in the stores table.
type Store struct {
	ID           int64     `json:"id"`
	SellerID     int64     `json:"seller_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	LogoURL      string    `json:"logo_url"`
	ContactEmail string    `json:"contact_email"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SellerProfile is returned by GET /api/sellers/{id}.
type SellerProfile struct {
	SellerID   int64     `json:"seller_id"`
	SellerName string    `json:"seller_name"`
	JoinedAt   time.Time `json:"joined_at"`
	Store      Store     `json:"store"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

// Create inserts a store for sellerID. Returns ErrStoreExists if the seller already has one.
func (r *Repo) Create(ctx context.Context, sellerID int64, name, description, logoURL, contactEmail string) (Store, error) {
	const q = `
		INSERT INTO stores (seller_id, name, description, logo_url, contact_email)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, seller_id, name, description, logo_url, contact_email, created_at, updated_at
	`
	var s Store
	err := r.pool.QueryRow(ctx, q, sellerID, name, description, logoURL, contactEmail).
		Scan(&s.ID, &s.SellerID, &s.Name, &s.Description, &s.LogoURL, &s.ContactEmail, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Store{}, ErrStoreExists
		}
		return Store{}, fmt.Errorf("create store: %w", err)
	}
	return s, nil
}

// GetBySellerID returns the store belonging to sellerID, or ErrNotFound.
func (r *Repo) GetBySellerID(ctx context.Context, sellerID int64) (Store, error) {
	const q = `
		SELECT id, seller_id, name, description, logo_url, contact_email, created_at, updated_at
		FROM stores
		WHERE seller_id = $1
	`
	var s Store
	err := r.pool.QueryRow(ctx, q, sellerID).
		Scan(&s.ID, &s.SellerID, &s.Name, &s.Description, &s.LogoURL, &s.ContactEmail, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Store{}, ErrNotFound
	}
	if err != nil {
		return Store{}, fmt.Errorf("get store by seller: %w", err)
	}
	return s, nil
}

// Update modifies the store belonging to sellerID. Returns ErrNotFound if not set up yet.
func (r *Repo) Update(ctx context.Context, sellerID int64, name, description, logoURL, contactEmail string) (Store, error) {
	const q = `
		UPDATE stores
		SET name = $2, description = $3, logo_url = $4, contact_email = $5, updated_at = now()
		WHERE seller_id = $1
		RETURNING id, seller_id, name, description, logo_url, contact_email, created_at, updated_at
	`
	var s Store
	err := r.pool.QueryRow(ctx, q, sellerID, name, description, logoURL, contactEmail).
		Scan(&s.ID, &s.SellerID, &s.Name, &s.Description, &s.LogoURL, &s.ContactEmail, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Store{}, ErrNotFound
	}
	if err != nil {
		return Store{}, fmt.Errorf("update store: %w", err)
	}
	return s, nil
}

// GetSellerProfile joins users and stores for GET /api/sellers/{id}.
func (r *Repo) GetSellerProfile(ctx context.Context, sellerID int64) (SellerProfile, error) {
	const q = `
		SELECT u.id, u.name, u.created_at,
		       s.id, s.seller_id, s.name, s.description, s.logo_url, s.contact_email, s.created_at, s.updated_at
		FROM users u
		JOIN stores s ON s.seller_id = u.id
		WHERE u.id = $1 AND u.role = 'seller'
	`
	var p SellerProfile
	err := r.pool.QueryRow(ctx, q, sellerID).Scan(
		&p.SellerID, &p.SellerName, &p.JoinedAt,
		&p.Store.ID, &p.Store.SellerID, &p.Store.Name, &p.Store.Description,
		&p.Store.LogoURL, &p.Store.ContactEmail, &p.Store.CreatedAt, &p.Store.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return SellerProfile{}, ErrNotFound
	}
	if err != nil {
		return SellerProfile{}, fmt.Errorf("get seller profile: %w", err)
	}
	return p, nil
}
