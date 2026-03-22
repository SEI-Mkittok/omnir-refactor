package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/omnir/crm-api/internal/domain"
)

// ErrorDetail describes a single field-level validation failure.
type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ErrorResponse is the standard API error shape: {error, code, details[]}.
type ErrorResponse struct {
	Error   string        `json:"error"`
	Code    string        `json:"code"`
	Details []ErrorDetail `json:"details,omitempty"`
}

// errorCodes maps HTTP status to a machine-readable code.
var errorCodes = map[int]string{
	http.StatusBadRequest:          "bad_request",
	http.StatusUnauthorized:        "unauthorized",
	http.StatusForbidden:           "forbidden",
	http.StatusNotFound:            "not_found",
	http.StatusConflict:            "conflict",
	http.StatusUnprocessableEntity: "validation_error",
	http.StatusInternalServerError: "internal_error",
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeProblem is a legacy helper kept for handler compatibility.
// New code should use writeError directly.
func writeProblem(w http.ResponseWriter, status int, _ string, detail ...string) {
	msg := http.StatusText(status)
	if len(detail) > 0 {
		msg = detail[0]
	}
	writeError(w, status, msg)
}

func writeError(w http.ResponseWriter, status int, message string, details ...ErrorDetail) {
	code, ok := errorCodes[status]
	if !ok {
		code = "error"
	}
	resp := ErrorResponse{Error: message, Code: code}
	if len(details) > 0 {
		resp.Details = details
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

func handleDomainErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrConflict):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrValidation):
		// Attempt to parse field-level detail from the error message.
		writeError(w, http.StatusUnprocessableEntity, err.Error(), parseValidationDetails(err)...)
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

// parseValidationDetails extracts field-level details from a validation error.
// Format expected: "validation error: <field> <message>".
func parseValidationDetails(err error) []ErrorDetail {
	msg := err.Error()
	prefix := "validation error: "
	if !strings.HasPrefix(msg, prefix) {
		return nil
	}
	rest := strings.TrimPrefix(msg, prefix)
	// Split on first space to get field name.
	parts := strings.SplitN(rest, " ", 2)
	if len(parts) == 2 {
		return []ErrorDetail{{Field: parts[0], Message: parts[1]}}
	}
	return []ErrorDetail{{Field: "_", Message: rest}}
}

// PaginatedMeta holds pagination metadata nested under "meta" in the response.
type PaginatedMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// PaginatedResponse wraps a list result with pagination metadata.
// Shape: { "data": [...], "meta": { "page", "per_page", "total", "total_pages" } }
type PaginatedResponse[T any] struct {
	Data []T          `json:"data"`
	Meta PaginatedMeta `json:"meta"`
}

func paginated[T any](data []T, total, page, limit int) PaginatedResponse[T] {
	if data == nil {
		data = []T{}
	}
	totalPages := 1
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	}
	return PaginatedResponse[T]{
		Data: data,
		Meta: PaginatedMeta{
			Page:       page,
			PerPage:    limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}
