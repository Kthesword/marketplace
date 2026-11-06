package cart

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/kthesword/marketplace/backend/internal/httpx"
	"github.com/kthesword/marketplace/backend/internal/middleware"
)

type Handler struct {
	repo *Repo
}

func NewHandler(repo *Repo) *Handler {
	return &Handler{repo: repo}
}

// Get handles GET /api/cart/items.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())

	c, err := h.repo.GetCart(r.Context(), claims.UserID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, c)
}

type addItemInput struct {
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

// Add handles POST /api/cart/items.
func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())

	var in addItemInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.WriteError(w, err)
		return
	}

	fields := make(map[string]string)
	if in.ProductID <= 0 {
		fields["product_id"] = "product_id is required"
	}
	if in.Quantity <= 0 {
		fields["quantity"] = "quantity must be at least 1"
	}
	if len(fields) > 0 {
		httpx.WriteError(w, httpx.Validation(fields))
		return
	}

	if err := h.repo.AddItem(r.Context(), claims.UserID, in.ProductID, in.Quantity); err != nil {
		switch {
		case errors.Is(err, ErrProductNotFound):
			httpx.WriteError(w, httpx.NotFound("product not found"))
		case errors.Is(err, ErrInsufficientStock):
			httpx.WriteError(w, httpx.Conflict("insufficient_stock", "requested quantity exceeds available stock"))
		default:
			httpx.WriteError(w, err)
		}
		return
	}

	c, err := h.repo.GetCart(r.Context(), claims.UserID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, c)
}

type updateQuantityInput struct {
	Quantity int `json:"quantity"`
}

// UpdateQuantity handles PATCH /api/cart/items/{productId}.
func (h *Handler) UpdateQuantity(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())

	productID, err := strconv.ParseInt(r.PathValue("productId"), 10, 64)
	if err != nil || productID <= 0 {
		httpx.WriteError(w, httpx.BadRequest("invalid product id"))
		return
	}

	var in updateQuantityInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if in.Quantity <= 0 {
		httpx.WriteError(w, httpx.Validation(map[string]string{
			"quantity": "quantity must be at least 1 (use DELETE to remove)",
		}))
		return
	}

	if err := h.repo.SetQuantity(r.Context(), claims.UserID, productID, in.Quantity); err != nil {
		switch {
		case errors.Is(err, ErrItemNotInCart), errors.Is(err, ErrProductNotFound):
			httpx.WriteError(w, httpx.NotFound("cart item not found"))
		case errors.Is(err, ErrInsufficientStock):
			httpx.WriteError(w, httpx.Conflict("insufficient_stock", "requested quantity exceeds available stock"))
		default:
			httpx.WriteError(w, err)
		}
		return
	}

	c, err := h.repo.GetCart(r.Context(), claims.UserID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, c)
}

// Remove handles DELETE /api/cart/items/{productId}.
func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())

	productID, err := strconv.ParseInt(r.PathValue("productId"), 10, 64)
	if err != nil || productID <= 0 {
		httpx.WriteError(w, httpx.BadRequest("invalid product id"))
		return
	}

	if err := h.repo.RemoveItem(r.Context(), claims.UserID, productID); err != nil {
		if errors.Is(err, ErrItemNotInCart) {
			httpx.WriteError(w, httpx.NotFound("cart item not found"))
			return
		}
		httpx.WriteError(w, err)
		return
	}

	c, err := h.repo.GetCart(r.Context(), claims.UserID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, c)
}

