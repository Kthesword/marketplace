package order

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5433/marketplace?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("skipping Postgres integration test: failed to parse db url: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("skipping Postgres integration test: database not reachable on %s (%v)", dbURL, err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}

func TestCheckoutAndCancelIntegration(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()

	service := NewService(pool)
	repo := NewRepo(pool)

	// Generate unique email timestamps to avoid colliding with previous seed data
	suffix := time.Now().UnixNano()
	seller1Email := fmt.Sprintf("seller1_%d@test.com", suffix)
	seller2Email := fmt.Sprintf("seller2_%d@test.com", suffix)
	buyerEmail := fmt.Sprintf("buyer_%d@test.com", suffix)

	var seller1ID, seller2ID, buyerID int64
	err := pool.QueryRow(ctx, `INSERT INTO users (email, password_hash, name, role) VALUES ($1, 'hash', 'Seller One', 'seller') RETURNING id`, seller1Email).Scan(&seller1ID)
	if err != nil {
		t.Fatalf("create seller1: %v", err)
	}
	err = pool.QueryRow(ctx, `INSERT INTO users (email, password_hash, name, role) VALUES ($1, 'hash', 'Seller Two', 'seller') RETURNING id`, seller2Email).Scan(&seller2ID)
	if err != nil {
		t.Fatalf("create seller2: %v", err)
	}
	err = pool.QueryRow(ctx, `INSERT INTO users (email, password_hash, name, role) VALUES ($1, 'hash', 'Test Buyer', 'buyer') RETURNING id`, buyerEmail).Scan(&buyerID)
	if err != nil {
		t.Fatalf("create buyer: %v", err)
	}

	// Create stores
	_, err = pool.Exec(ctx, `INSERT INTO stores (seller_id, name, description, contact_email) VALUES ($1, 'Store One', 'Desc', $2)`, seller1ID, seller1Email)
	if err != nil {
		t.Fatalf("create store1: %v", err)
	}
	_, err = pool.Exec(ctx, `INSERT INTO stores (seller_id, name, description, contact_email) VALUES ($1, 'Store Two', 'Desc', $2)`, seller2ID, seller2Email)
	if err != nil {
		t.Fatalf("create store2: %v", err)
	}

	// Create products: Product 1 has stock=10, Product 2 has stock=5
	var prod1ID, prod2ID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO products (seller_id, name, description, price_cents, stock, active)
		VALUES ($1, 'Handmade Vase', 'A ceramic vase', 2500, 10, true)
		RETURNING id`, seller1ID).Scan(&prod1ID)
	if err != nil {
		t.Fatalf("create product 1: %v", err)
	}

	err = pool.QueryRow(ctx, `
		INSERT INTO products (seller_id, name, description, price_cents, stock, active)
		VALUES ($1, 'Wooden Bowl', 'Hand-carved bowl', 4000, 5, true)
		RETURNING id`, seller2ID).Scan(&prod2ID)
	if err != nil {
		t.Fatalf("create product 2: %v", err)
	}

	// Step 1: Add items to buyer's cart (2 of Product 1, 1 of Product 2)
	_, err = pool.Exec(ctx, `INSERT INTO cart_items (user_id, product_id, quantity) VALUES ($1, $2, 2), ($1, $3, 1)`, buyerID, prod1ID, prod2ID)
	if err != nil {
		t.Fatalf("add to cart: %v", err)
	}

	// Step 2: Perform atomic Checkout
	res, err := service.Checkout(ctx, buyerID)
	if err != nil {
		t.Fatalf("checkout failed unexpectedly: %v", err)
	}

	if res.CheckoutID == "" {
		t.Error("expected non-empty checkout_id")
	}

	// Multi-seller split: 2 sellers -> 2 orders
	if len(res.Orders) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(res.Orders))
	}

	for _, o := range res.Orders {
		if o.CheckoutID != res.CheckoutID {
			t.Errorf("order checkout_id = %s; want %s", o.CheckoutID, res.CheckoutID)
		}
		if o.Status != StatusPending {
			t.Errorf("order status = %s; want Pending", o.Status)
		}
	}

	// Verify stock was decremented in database
	var stock1, stock2 int
	_ = pool.QueryRow(ctx, `SELECT stock FROM products WHERE id = $1`, prod1ID).Scan(&stock1)
	_ = pool.QueryRow(ctx, `SELECT stock FROM products WHERE id = $2`, prod2ID).Scan(&stock2)
	if stock1 != 8 {
		t.Errorf("product 1 stock = %d; want 8 (10 - 2)", stock1)
	}
	if stock2 != 4 {
		t.Errorf("product 2 stock = %d; want 4 (5 - 1)", stock2)
	}

	// Verify cart is empty
	var cartCount int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM cart_items WHERE user_id = $1`, buyerID).Scan(&cartCount)
	if cartCount != 0 {
		t.Errorf("cart items remaining = %d; want 0", cartCount)
	}

	// Step 3: Test Rollback on Insufficient Stock
	// Add 10 of Product 2 to cart (only 4 remain in stock) and 1 of Product 1
	_, err = pool.Exec(ctx, `INSERT INTO cart_items (user_id, product_id, quantity) VALUES ($1, $2, 1), ($1, $3, 10)`, buyerID, prod1ID, prod2ID)
	if err != nil {
		t.Fatalf("add items for rollback test: %v", err)
	}

	_, err = service.Checkout(ctx, buyerID)
	if err == nil {
		t.Fatal("expected checkout to fail with insufficient stock, got nil")
	}
	var ce *CheckoutError
	if !errors.As(err, &ce) || ce.Code != "insufficient_stock" {
		t.Errorf("expected insufficient_stock CheckoutError, got %v", err)
	}

	// PROOF OF ROLLBACK: product 1 stock must still be 8, not 7!
	_ = pool.QueryRow(ctx, `SELECT stock FROM products WHERE id = $1`, prod1ID).Scan(&stock1)
	if stock1 != 8 {
		t.Errorf("ROLLBACK FAILED: product 1 stock decremented to %d; want 8", stock1)
	}

	// Clear cart from failed test
	_, _ = pool.Exec(ctx, `DELETE FROM cart_items WHERE user_id = $1`, buyerID)

	// Step 4: Test Buyer Cancel with Stock Restoration
	// Find order 1 (which purchased 2 of Product 1)
	order1 := res.Orders[0]
	cancelledOrder, err := repo.CancelByBuyer(ctx, order1.ID, buyerID)
	if err != nil {
		t.Fatalf("buyer cancel failed: %v", err)
	}
	if cancelledOrder.Status != StatusCancelled {
		t.Errorf("cancelled order status = %s; want Cancelled", cancelledOrder.Status)
	}

	// Verify stock of Product 1 was restored from 8 -> 10!
	_ = pool.QueryRow(ctx, `SELECT stock FROM products WHERE id = $1`, prod1ID).Scan(&stock1)
	if stock1 != 10 {
		t.Errorf("STOCK RESTORATION FAILED: product 1 stock = %d; want 10", stock1)
	}

	// Verify double cancellation is rejected with ErrInvalidTransition
	_, err = repo.CancelByBuyer(ctx, order1.ID, buyerID)
	if err == nil {
		t.Fatal("expected double cancellation to fail, got nil")
	}
	if !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("expected ErrInvalidTransition on double cancel, got %v", err)
	}
}

