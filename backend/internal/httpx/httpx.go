// Package httpx holds small HTTP helpers shared by every feature package:
// JSON responses and the project-wide error format.
package httpx

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
)

// AppError is an error that knows how to present itself over HTTP.
// Handlers return/construct these; WriteError turns them into the JSON body
//
//	{"error": {"code": "...", "message": "...", "fields": {...}}}
type AppError struct {
	Status  int               `json:"-"` // HTTP status; not part of the JSON body
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"` // per-field validation messages (422)
}

// Error makes AppError satisfy Go's built-in error interface.
func (e *AppError) Error() string { return e.Code + ": " + e.Message }

// Constructors for the status codes this API uses.
func BadRequest(msg string) *AppError {
	return &AppError{http.StatusBadRequest, "bad_request", msg, nil}
}
func Unauthorized(msg string) *AppError {
	return &AppError{http.StatusUnauthorized, "unauthorized", msg, nil}
}
func Forbidden(msg string) *AppError { return &AppError{http.StatusForbidden, "forbidden", msg, nil} }
func NotFound(msg string) *AppError  { return &AppError{http.StatusNotFound, "not_found", msg, nil} }

// Conflict takes a specific code (e.g. "email_taken") so clients can react to it.
func Conflict(code, msg string) *AppError { return &AppError{http.StatusConflict, code, msg, nil} }

// Validation builds a 422 with a message per invalid field.
func Validation(fields map[string]string) *AppError {
	return &AppError{http.StatusUnprocessableEntity, "validation_failed", "one or more fields are invalid", fields}
}

// WriteJSON writes v as a JSON response with the given status.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status) // headers must be set BEFORE WriteHeader; the body comes after
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("httpx: encode response: %v", err)
	}
}

// DecodeJSON reads at most 1MB from r.Body and decodes it into dst.
// Unknown fields or trailing garbage return a 400 BadRequest AppError.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return BadRequest("malformed JSON body")
	}
	return nil
}

// Page is the standard paginated response shape across all list endpoints.
type Page[T any] struct {
	Items    []T `json:"items"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
}

// ParsePagination extracts ?page= and ?page_size= from the query string with safe bounds.
func ParsePagination(r *http.Request, defaultSize int) (page, pageSize, offset int, err error) {
	q := r.URL.Query()
	page = 1
	pageSize = defaultSize

	if v := q.Get("page"); v != "" {
		page, err = strconv.Atoi(v)
		if err != nil || page < 1 {
			return 0, 0, 0, BadRequest("page must be a positive integer")
		}
	}
	if v := q.Get("page_size"); v != "" {
		pageSize, err = strconv.Atoi(v)
		if err != nil || pageSize < 1 || pageSize > 100 {
			return 0, 0, 0, BadRequest("page_size must be between 1 and 100")
		}
	}
	offset = (page - 1) * pageSize
	return page, pageSize, offset, nil
}

// WriteError renders any error. Known AppErrors keep their status and message;
// anything else becomes a generic 500 so internal details are never leaked.
func WriteError(w http.ResponseWriter, err error) {
	var appErr *AppError
	if !errors.As(err, &appErr) { // errors.As finds an *AppError anywhere in the wrap chain
		log.Printf("internal error: %v", err) // the real cause goes to the server log only
		appErr = &AppError{http.StatusInternalServerError, "internal_error", "something went wrong", nil}
	}
	WriteJSON(w, appErr.Status, map[string]any{"error": appErr})
}
