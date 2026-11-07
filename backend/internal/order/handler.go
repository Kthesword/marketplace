package order

import (
	"errors"
	"net/http"
	"strconv"

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
