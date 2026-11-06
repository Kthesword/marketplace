package auth

import (
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"github.com/kthesword/marketplace/backend/internal/httpx"
	"github.com/kthesword/marketplace/backend/internal/user"
)

// dummyHash is a precomputed bcrypt hash used when a login email does not exist,
// so that non-existent emails take the same ~70ms as existing ones (no timing leak).
var dummyHash, _ = HashPassword("dummy-password-for-timing-safety")

// Handler serves /api/auth/* endpoints.
type Handler struct {
	users  *user.Repo
	tokens *TokenManager
}

func NewHandler(users *user.Repo, tokens *TokenManager) *Handler {
	return &Handler{users: users, tokens: tokens}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

type authResponse struct {
	Token string    `json:"token"`
	User  user.User `json:"user"`
}

// Register handles POST /api/auth/register.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Name = strings.TrimSpace(req.Name)

	// Validate fields and collect all problems into a 422 map.
	fields := make(map[string]string)
	if _, err := mail.ParseAddress(req.Email); err != nil || req.Email == "" {
		fields["email"] = "must be a valid email address"
	}
	if len(req.Password) < 8 {
		fields["password"] = "must be at least 8 characters"
	} else if len(req.Password) > 72 {
		fields["password"] = "must not exceed 72 bytes"
	}
	if req.Name == "" {
		fields["name"] = "name is required"
	}
	if req.Role != user.RoleSeller && req.Role != user.RoleBuyer {
		fields["role"] = "must be 'seller' or 'buyer'"
	}
	if len(fields) > 0 {
		httpx.WriteError(w, httpx.Validation(fields))
		return
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	u, err := h.users.Create(r.Context(), req.Email, hash, req.Name, req.Role)
	if err != nil {
		if errors.Is(err, user.ErrEmailTaken) {
			httpx.WriteError(w, httpx.Conflict("email_taken", "an account with this email already exists"))
			return
		}
		httpx.WriteError(w, err)
		return
	}

	tok, err := h.tokens.Issue(u.ID, u.Email, u.Role)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, authResponse{Token: tok, User: u})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login handles POST /api/auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" || req.Password == "" {
		httpx.WriteError(w, httpx.Validation(map[string]string{
			"credentials": "email and password are required",
		}))
		return
	}

	u, err := h.users.GetByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			// Run dummy bcrypt so timing matches a real account check.
			_ = CheckPassword(dummyHash, req.Password)
			httpx.WriteError(w, httpx.Unauthorized("invalid email or password"))
			return
		}
		httpx.WriteError(w, err)
		return
	}

	if !CheckPassword(u.PasswordHash, req.Password) {
		httpx.WriteError(w, httpx.Unauthorized("invalid email or password"))
		return
	}

	tok, err := h.tokens.Issue(u.ID, u.Email, u.Role)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, authResponse{Token: tok, User: u})
}

// Me returns a handler for GET /api/auth/me that fetches the current user by ID.
func (h *Handler) Me(extractUserID func(*http.Request) (int64, bool)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := extractUserID(r)
		if !ok {
			httpx.WriteError(w, httpx.Unauthorized("unauthenticated"))
			return
		}
		u, err := h.users.GetByID(r.Context(), uid)
		if err != nil {
			if errors.Is(err, user.ErrNotFound) {
				httpx.WriteError(w, httpx.Unauthorized("user no longer exists"))
				return
			}
			httpx.WriteError(w, err)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"user": u})
	}
}
