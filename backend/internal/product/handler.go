package product

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/kthesword/marketplace/backend/internal/httpx"
	"github.com/kthesword/marketplace/backend/internal/middleware"
	"github.com/kthesword/marketplace/backend/internal/store"
)

type Handler struct {
	products *Repo
	stores   *store.Repo
}

func NewHandler(products *Repo, stores *store.Repo) *Handler {
	return &Handler{products: products, stores: stores}
}

func validateInput(in *Input) (bool, map[string]string) {
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	in.ImageURL = strings.TrimSpace(in.ImageURL)

	active := true
	if in.Active != nil {
		active = *in.Active
	}

	fields := make(map[string]string)
	if in.Name == "" {
		fields["name"] = "product name is required"
	} else if len(in.Name) > 140 {
		fields["name"] = "product name must not exceed 140 characters"
	}
	if len(in.Description) > 2000 {
		fields["description"] = "description must not exceed 2000 characters"
	}
	if in.PriceCents <= 0 {
		fields["price_cents"] = "price_cents must be greater than 0"
	}
	if in.Stock < 0 {
		fields["stock"] = "stock cannot be negative"
	}
	if in.ImageURL != "" {
		u, err := url.ParseRequestURI(in.ImageURL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
			fields["image_url"] = "image_url must be a valid http or https URL"
		}
	}
	return active, fields
}

// Create handles POST /api/seller/products.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())

	// Ensure the seller has set up a store before creating products.
	st, err := h.stores.GetBySellerID(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			httpx.WriteError(w, httpx.Conflict("store_required", "please set up your store before adding products"))
			return
		}
		httpx.WriteError(w, err)
		return
	}

	var in Input
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.WriteError(w, err)
		return
	}
	active, fields := validateInput(&in)
	if len(fields) > 0 {
		httpx.WriteError(w, httpx.Validation(fields))
		return
	}

	p, err := h.products.Create(r.Context(), claims.UserID, st.Name, in, active)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, map[string]any{"product": p})
}

// Update handles PUT /api/seller/products/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		httpx.WriteError(w, httpx.BadRequest("invalid product id"))
		return
	}

	st, err := h.stores.GetBySellerID(r.Context(), claims.UserID)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		httpx.WriteError(w, err)
		return
	}

	var in Input
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.WriteError(w, err)
		return
	}
	active, fields := validateInput(&in)
	if len(fields) > 0 {
		httpx.WriteError(w, httpx.Validation(fields))
		return
	}

	p, err := h.products.Update(r.Context(), id, claims.UserID, st.Name, in, active)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, httpx.NotFound("product not found"))
			return
		}
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"product": p})
}

// Delete handles DELETE /api/seller/products/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		httpx.WriteError(w, httpx.BadRequest("invalid product id"))
		return
	}

	if err := h.products.Delete(r.Context(), id, claims.UserID); err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, httpx.NotFound("product not found"))
			return
		}
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListMine handles GET /api/seller/products (includes inactive products).
func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())

	page, pageSize, offset, err := httpx.ParsePagination(r, 20)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	items, total, err := h.products.ListBySeller(r.Context(), claims.UserID, true, pageSize, offset)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.Page[Product]{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	})
}

// ListPublic handles GET /api/products?q=&page=&page_size=.
func (h *Handler) ListPublic(w http.ResponseWriter, r *http.Request) {
	page, pageSize, offset, err := httpx.ParsePagination(r, 12)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	q := r.URL.Query().Get("q")

	items, total, err := h.products.ListPublic(r.Context(), q, pageSize, offset)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.Page[Product]{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	})
}

// GetPublic handles GET /api/products/{id}.
func (h *Handler) GetPublic(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		httpx.WriteError(w, httpx.BadRequest("invalid product id"))
		return
	}

	p, err := h.products.GetPublicByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, httpx.NotFound("product not found"))
			return
		}
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"product": p})
}

// ListBySellerPublic handles GET /api/sellers/{id}/products.
func (h *Handler) ListBySellerPublic(w http.ResponseWriter, r *http.Request) {
	sellerID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || sellerID <= 0 {
		httpx.WriteError(w, httpx.BadRequest("invalid seller id"))
		return
	}

	page, pageSize, offset, err := httpx.ParsePagination(r, 12)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	items, total, err := h.products.ListBySeller(r.Context(), sellerID, false, pageSize, offset)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.Page[Product]{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	})
}
