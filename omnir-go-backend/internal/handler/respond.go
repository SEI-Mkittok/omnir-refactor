package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/omnir/crm-api/internal/domain"
)

// Problem is an RFC 7807 error response.
type Problem struct {
	Status int    `json:"status"`
	Title  string `json:"title"`
	Detail string `json:"detail,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeProblem(w http.ResponseWriter, status int, title string, detail ...string) {
	p := Problem{Status: status, Title: title}
	if len(detail) > 0 {
		p.Detail = detail[0]
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(p)
}

func domainErrStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, domain.ErrValidation):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

func handleDomainErr(w http.ResponseWriter, err error) {
	status := domainErrStatus(err)
	title := http.StatusText(status)
	writeProblem(w, status, title, err.Error())
}

// PaginatedResponse wraps a list result with pagination metadata.
type PaginatedResponse[T any] struct {
	Data       []T `json:"data"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"total_pages"`
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
		Data:       data,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}
}
