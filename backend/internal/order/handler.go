package order

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/kthesword/marketplace/backend/internal/httpx"
	"github.com/kthesword/marketplace/backend/internal/middleware"
)

type Handler struct {
	repo    *Repo
	service *Service
}

func NewHandler(repo *Repo, service *Service) *Handler {
	return &Handler{repo: repo, service: service}
}

// Checkout handles POST /api/orders/checkout (buyer only). The body is ignored:
// the cart lives in the database, so the client cannot tamper with prices or items.
func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())

	res, err := h.service.Checkout(r.Context(), claims.UserID)
	if err != nil {
		var ce *CheckoutError
		if errors.As(err, &ce) {
			httpx.WriteError(w, httpx.Conflict(ce.Code, ce.Message))
			return
		}
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, res)
}

// List handles GET /api/orders.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())

	page, pageSize, offset, err := httpx.ParsePagination(r, 10)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	orders, total, err := h.repo.ListByBuyer(r.Context(), claims.UserID, pageSize, offset)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.Page[Order]{Items: orders, Page: page, PageSize: pageSize, Total: total})
}

// Get handles GET /api/orders/{id}.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		httpx.WriteError(w, httpx.BadRequest("invalid order id"))
		return
	}
	o, err := h.repo.GetForBuyer(r.Context(), id, claims.UserID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, httpx.NotFound("order not found"))
			return
		}
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"order": o})
}

// Cancel handles POST /api/orders/{id}/cancel (buyer only).
// Cancels a Pending order owned by the buyer and restores product stock atomically.
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		httpx.WriteError(w, httpx.BadRequest("invalid order id"))
		return
	}

	o, err := h.repo.CancelByBuyer(r.Context(), id, claims.UserID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, httpx.NotFound("order not found"))
			return
		}
		if errors.Is(err, ErrInvalidTransition) {
			httpx.WriteError(w, httpx.Conflict("invalid_transition", err.Error()))
			return
		}
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"order": o})
}

// ListSeller handles GET /api/seller/orders (seller only).
// Supports optional ?status=Pending|Shipped|Delivered|Cancelled plus pagination.
func (h *Handler) ListSeller(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())

	statusFilter := strings.TrimSpace(r.URL.Query().Get("status"))
	if statusFilter != "" && !IsValidStatus(statusFilter) {
		httpx.WriteError(w, httpx.BadRequest("invalid status filter: must be Pending, Shipped, Delivered, or Cancelled"))
		return
	}

	page, pageSize, offset, err := httpx.ParsePagination(r, 10)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	orders, total, err := h.repo.ListBySeller(r.Context(), claims.UserID, statusFilter, pageSize, offset)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.Page[Order]{Items: orders, Page: page, PageSize: pageSize, Total: total})
}

type updateStatusReq struct {
	Status string `json:"status"`
}

// UpdateStatusSeller handles PATCH /api/seller/orders/{id}/status (seller only).
func (h *Handler) UpdateStatusSeller(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		httpx.WriteError(w, httpx.BadRequest("invalid order id"))
		return
	}

	var req updateStatusReq
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	req.Status = strings.TrimSpace(req.Status)
	if !IsValidStatus(req.Status) {
		httpx.WriteError(w, httpx.Validation(map[string]string{
			"status": "must be Pending, Shipped, Delivered, or Cancelled",
		}))
		return
	}

	o, err := h.repo.UpdateStatusBySeller(r.Context(), id, claims.UserID, req.Status)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, httpx.NotFound("order not found"))
			return
		}
		if errors.Is(err, ErrInvalidTransition) {
			httpx.WriteError(w, httpx.Conflict("invalid_transition", err.Error()))
			return
		}
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"order": o})
}
