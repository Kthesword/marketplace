// Package middleware provides HTTP middleware for authentication and role-based access control.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/kthesword/marketplace/backend/internal/auth"
	"github.com/kthesword/marketplace/backend/internal/httpx"
)

// claimsKey is an unexported type used as the context key for *auth.Claims.
// Using a private struct type guarantees no other package can collide with this key.
type claimsKey struct{}

// WithClaims stores verified JWT claims in ctx.
func WithClaims(ctx context.Context, c *auth.Claims) context.Context {
	return context.WithValue(ctx, claimsKey{}, c)
}

// ClaimsFromContext retrieves the verified *auth.Claims placed by Authenticate.
func ClaimsFromContext(ctx context.Context) (*auth.Claims, bool) {
	c, ok := ctx.Value(claimsKey{}).(*auth.Claims)
	return c, ok
}

// Authenticate verifies the `Authorization: Bearer <token>` header and attaches
// *auth.Claims to the request context. Missing or invalid tokens get 401.
func Authenticate(tm *auth.TokenManager, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		tokenStr, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || strings.TrimSpace(tokenStr) == "" {
			httpx.WriteError(w, httpx.Unauthorized("missing or malformed Authorization header"))
			return
		}

		claims, err := tm.Verify(strings.TrimSpace(tokenStr))
		if err != nil {
			httpx.WriteError(w, httpx.Unauthorized("invalid or expired token"))
			return
		}

		next(w, r.WithContext(WithClaims(r.Context(), claims)))
	}
}

// RequireRole authenticates the request AND verifies that claims.Role matches
// the required role ("seller" or "buyer"). Wrong role gets 403 Forbidden.
func RequireRole(tm *auth.TokenManager, role string, next http.HandlerFunc) http.HandlerFunc {
	return Authenticate(tm, func(w http.ResponseWriter, r *http.Request) {
		claims, ok := ClaimsFromContext(r.Context())
		if !ok || claims.Role != role {
			httpx.WriteError(w, httpx.Forbidden("insufficient permissions for role"))
			return
		}
		next(w, r)
	})
}
