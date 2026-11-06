package main

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kthesword/marketplace/backend/internal/auth"
	"github.com/kthesword/marketplace/backend/internal/cart"
	"github.com/kthesword/marketplace/backend/internal/config"
	"github.com/kthesword/marketplace/backend/internal/httpx"
	"github.com/kthesword/marketplace/backend/internal/middleware"
	"github.com/kthesword/marketplace/backend/internal/product"
	"github.com/kthesword/marketplace/backend/internal/store"
	"github.com/kthesword/marketplace/backend/internal/user"
)

// newRouter wires repositories, services, and handlers onto a Go 1.22+ ServeMux.
func newRouter(cfg config.Config, pool *pgxpool.Pool) http.Handler {
	mux := http.NewServeMux()

	userRepo := user.NewRepo(pool)
	storeRepo := store.NewRepo(pool)
	productRepo := product.NewRepo(pool)
	cartRepo := cart.NewRepo(pool)

	tokenMgr := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTTTL)
	authHandler := auth.NewHandler(userRepo, tokenMgr)
	storeHandler := store.NewHandler(storeRepo)
	productHandler := product.NewHandler(productRepo, storeRepo)
	cartHandler := cart.NewHandler(cartRepo)

	// Health
	mux.HandleFunc("GET /api/health", healthHandler(pool))

	// Auth
	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("GET /api/auth/me", middleware.Authenticate(tokenMgr, authHandler.Me(func(r *http.Request) (int64, bool) {
		c, ok := middleware.ClaimsFromContext(r.Context())
		if !ok {
			return 0, false
		}
		return c.UserID, true
	})))

	// Seller store (seller role required)
	mux.HandleFunc("POST /api/seller/store", middleware.RequireRole(tokenMgr, user.RoleSeller, storeHandler.Create))
	mux.HandleFunc("GET /api/seller/store", middleware.RequireRole(tokenMgr, user.RoleSeller, storeHandler.GetMine))
	mux.HandleFunc("PUT /api/seller/store", middleware.RequireRole(tokenMgr, user.RoleSeller, storeHandler.Update))

	// Seller products (seller role required)
	mux.HandleFunc("GET /api/seller/products", middleware.RequireRole(tokenMgr, user.RoleSeller, productHandler.ListMine))
	mux.HandleFunc("POST /api/seller/products", middleware.RequireRole(tokenMgr, user.RoleSeller, productHandler.Create))
	mux.HandleFunc("PUT /api/seller/products/{id}", middleware.RequireRole(tokenMgr, user.RoleSeller, productHandler.Update))
	mux.HandleFunc("DELETE /api/seller/products/{id}", middleware.RequireRole(tokenMgr, user.RoleSeller, productHandler.Delete))

	// Public catalog & seller profile
	mux.HandleFunc("GET /api/sellers/{id}", storeHandler.GetPublic)
	mux.HandleFunc("GET /api/sellers/{id}/products", productHandler.ListBySellerPublic)
	mux.HandleFunc("GET /api/products", productHandler.ListPublic)
	mux.HandleFunc("GET /api/products/{id}", productHandler.GetPublic)

	// Buyer cart (buyer role required)
	mux.HandleFunc("GET /api/cart/items", middleware.RequireRole(tokenMgr, user.RoleBuyer, cartHandler.Get))
	mux.HandleFunc("POST /api/cart/items", middleware.RequireRole(tokenMgr, user.RoleBuyer, cartHandler.Add))
	mux.HandleFunc("PATCH /api/cart/items/{productId}", middleware.RequireRole(tokenMgr, user.RoleBuyer, cartHandler.UpdateQuantity))
	mux.HandleFunc("DELETE /api/cart/items/{productId}", middleware.RequireRole(tokenMgr, user.RoleBuyer, cartHandler.Remove))

	// Catch-all: any unmatched path gets our JSON error format, not Go's plain text.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteError(w, httpx.NotFound("route not found"))
	})
	return mux
}

// healthHandler is a closure: it "captures" pool so the handler can use it,
// without a global variable.
func healthHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			httpx.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "degraded", "db": "down"})
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "db": "ok"})
	}
}
