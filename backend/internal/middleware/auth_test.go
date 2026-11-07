package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kthesword/marketplace/backend/internal/auth"
)

func setupTestTokenManager(t *testing.T) *auth.TokenManager {
	t.Helper()
	return auth.NewTokenManager("test-secret-key-32-bytes-long!!", time.Hour)
}

func TestAuthenticate(t *testing.T) {
	tm := setupTestTokenManager(t)

	// Create a dummy next handler that echoes the authenticated user's ID
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := ClaimsFromContext(r.Context())
		if !ok {
			t.Fatal("expected claims in context, got none")
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"user_id": claims.UserID, "role": claims.Role})
	})

	handler := Authenticate(tm, nextHandler)

	validToken, err := tm.Issue(42, "buyer@shop.com", "buyer")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "missing Authorization header",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
			wantCode:   "unauthorized",
		},
		{
			name:       "malformed Authorization header without Bearer prefix",
			authHeader: "Basic dXNlcjpwYXNz",
			wantStatus: http.StatusUnauthorized,
			wantCode:   "unauthorized",
		},
		{
			name:       "invalid or forged token",
			authHeader: "Bearer invalid.jwt.token",
			wantStatus: http.StatusUnauthorized,
			wantCode:   "unauthorized",
		},
		{
			name:       "valid token passes authentication",
			authHeader: "Bearer " + validToken,
			wantStatus: http.StatusOK,
			wantCode:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d; want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var body struct {
					Error struct {
						Code string `json:"code"`
					} `json:"error"`
				}
				if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode error body: %v", err)
				}
				if body.Error.Code != tt.wantCode {
					t.Errorf("error.code = %q; want %q", body.Error.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestRequireRole(t *testing.T) {
	tm := setupTestTokenManager(t)

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	sellerHandler := RequireRole(tm, "seller", okHandler)

	sellerToken, err := tm.Issue(1, "seller@shop.com", "seller")
	if err != nil {
		t.Fatalf("issue seller token: %v", err)
	}

	buyerToken, err := tm.Issue(2, "buyer@shop.com", "buyer")
	if err != nil {
		t.Fatalf("issue buyer token: %v", err)
	}

	tests := []struct {
		name       string
		token      string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "seller accessing seller route succeeds",
			token:      sellerToken,
			wantStatus: http.StatusOK,
			wantCode:   "",
		},
		{
			name:       "buyer accessing seller route gets 403 Forbidden",
			token:      buyerToken,
			wantStatus: http.StatusForbidden,
			wantCode:   "forbidden",
		},
		{
			name:       "unauthenticated request gets 401 Unauthorized",
			token:      "",
			wantStatus: http.StatusUnauthorized,
			wantCode:   "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/seller/secret", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			rec := httptest.NewRecorder()

			sellerHandler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d; want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var body struct {
					Error struct {
						Code string `json:"code"`
					} `json:"error"`
				}
				if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode error body: %v", err)
				}
				if body.Error.Code != tt.wantCode {
					t.Errorf("error.code = %q; want %q", body.Error.Code, tt.wantCode)
				}
			}
		})
	}
}

