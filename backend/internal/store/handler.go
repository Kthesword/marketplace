package store

import (
	"errors"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"strings"

	"github.com/kthesword/marketplace/backend/internal/httpx"
	"github.com/kthesword/marketplace/backend/internal/middleware"
)

type Handler struct {
	repo *Repo
}

func NewHandler(repo *Repo) *Handler {
	return &Handler{repo: repo}
}

type storeInput struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	LogoURL      string `json:"logo_url"`
	ContactEmail string `json:"contact_email"`
}

func validateInput(in *storeInput) map[string]string {
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	in.LogoURL = strings.TrimSpace(in.LogoURL)
	in.ContactEmail = strings.ToLower(strings.TrimSpace(in.ContactEmail))

	fields := make(map[string]string)
	if in.Name == "" {
		fields["name"] = "store name is required"
	} else if len(in.Name) > 100 {
		fields["name"] = "store name must not exceed 100 characters"
	}
	if len(in.Description) > 1000 {
		fields["description"] = "description must not exceed 1000 characters"
	}
	if in.LogoURL != "" {
		u, err := url.ParseRequestURI(in.LogoURL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
			fields["logo_url"] = "logo_url must be a valid http or https URL"
		}
	}
	if _, err := mail.ParseAddress(in.ContactEmail); err != nil || in.ContactEmail == "" {
		fields["contact_email"] = "contact_email must be a valid email address"
	}
	return fields
}

// Create handles POST /api/seller/store.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())

	var in storeInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if fields := validateInput(&in); len(fields) > 0 {
		httpx.WriteError(w, httpx.Validation(fields))
		return
	}

	s, err := h.repo.Create(r.Context(), claims.UserID, in.Name, in.Description, in.LogoURL, in.ContactEmail)
	if err != nil {
		if errors.Is(err, ErrStoreExists) {
			httpx.WriteError(w, httpx.Conflict("store_exists", "you have already created a store; use PUT to update it"))
			return
		}
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, map[string]any{"store": s})
}

// GetMine handles GET /api/seller/store.
func (h *Handler) GetMine(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())

	s, err := h.repo.GetBySellerID(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, httpx.NotFound("store not set up yet"))
			return
		}
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"store": s})
}

// Update handles PUT /api/seller/store.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.ClaimsFromContext(r.Context())

	var in storeInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if fields := validateInput(&in); len(fields) > 0 {
		httpx.WriteError(w, httpx.Validation(fields))
		return
	}

	s, err := h.repo.Update(r.Context(), claims.UserID, in.Name, in.Description, in.LogoURL, in.ContactEmail)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, httpx.NotFound("store not set up yet; use POST first"))
			return
		}
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"store": s})
}

// GetPublic handles GET /api/sellers/{id}.
func (h *Handler) GetPublic(w http.ResponseWriter, r *http.Request) {
	sellerID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || sellerID <= 0 {
		httpx.WriteError(w, httpx.BadRequest("invalid seller id"))
		return
	}

	profile, err := h.repo.GetSellerProfile(r.Context(), sellerID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, httpx.NotFound("seller store not found"))
			return
		}
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, profile)
}
