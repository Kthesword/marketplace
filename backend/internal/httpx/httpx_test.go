package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	payload := map[string]string{"message": "hello world"}

	WriteJSON(rec, http.StatusCreated, payload)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusCreated)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q; want application/json", contentType)
	}

	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if got["message"] != "hello world" {
		t.Errorf("got message %q; want %q", got["message"], "hello world")
	}
}

func TestWriteError(t *testing.T) {
	t.Run("AppError serializes correctly", func(t *testing.T) {
		rec := httptest.NewRecorder()
		appErr := BadRequest("bad payload data")

		WriteError(rec, appErr)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d; want %d", rec.Code, http.StatusBadRequest)
		}

		var body struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if body.Error.Code != "bad_request" {
			t.Errorf("code = %q; want bad_request", body.Error.Code)
		}
		if body.Error.Message != "bad payload data" {
			t.Errorf("message = %q; want bad payload data", body.Error.Message)
		}
	})

	t.Run("generic error defaults to 500 internal_error", func(t *testing.T) {
		rec := httptest.NewRecorder()
		genericErr := errors.New("database connection refused")

		WriteError(rec, genericErr)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d; want %d", rec.Code, http.StatusInternalServerError)
		}

		var body struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if body.Error.Code != "internal_error" {
			t.Errorf("code = %q; want internal_error", body.Error.Code)
		}
		if body.Error.Message != "something went wrong" {
			t.Errorf("message = %q; want something went wrong", body.Error.Message)
		}
	})
}

func TestDecodeJSON(t *testing.T) {
	type sample struct {
		Name  string `json:"name"`
		Price int    `json:"price"`
	}

	t.Run("valid payload decodes successfully", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Mug","price":1200}`))

		var s sample
		if err := DecodeJSON(rec, req, &s); err != nil {
			t.Fatalf("DecodeJSON failed unexpectedly: %v", err)
		}
		if s.Name != "Mug" || s.Price != 1200 {
			t.Errorf("decoded %+v; want name=Mug, price=1200", s)
		}
	})

	t.Run("unknown fields are rejected", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Mug","price":1200,"extra":"hacked"}`))

		var s sample
		err := DecodeJSON(rec, req, &s)
		if err == nil {
			t.Fatal("expected error on unknown field, got nil")
		}
		var appErr *AppError
		if !errors.As(err, &appErr) || appErr.Status != http.StatusBadRequest {
			t.Errorf("expected 400 BadRequest AppError, got %v", err)
		}
	})
}

func TestParsePagination(t *testing.T) {
	t.Run("defaults when params omitted", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/products", nil)
		page, pageSize, offset, err := ParsePagination(req, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if page != 1 || pageSize != 10 || offset != 0 {
			t.Errorf("got (%d, %d, %d); want (1, 10, 0)", page, pageSize, offset)
		}
	})

	t.Run("valid page and page_size calculate offset correctly", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/products?page=3&page_size=25", nil)
		page, pageSize, offset, err := ParsePagination(req, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if page != 3 || pageSize != 25 || offset != 50 {
			t.Errorf("got (%d, %d, %d); want (3, 25, 50)", page, pageSize, offset)
		}
	})

	t.Run("invalid page format returns 400 BadRequest", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/products?page=abc", nil)
		_, _, _, err := ParsePagination(req, 10)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

