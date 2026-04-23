package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/omnir/crm-api/internal/domain"
)

type linkedEntitiesMeta struct {
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

func parseLinkedEntityFilter(r *http.Request) (domain.LinkedEntityFilter, error) {
	q := r.URL.Query()
	filter := domain.LinkedEntityFilter{Page: 1, Limit: 25}

	if v := q.Get("linked_page"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return filter, domain.ErrValidation
		}
		filter.Page = n
	}
	if v := q.Get("linked_limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 200 {
			return filter, domain.ErrValidation
		}
		filter.Limit = n
	}
	if v := q.Get("type"); v != "" {
		filter.Type = &v
	}
	if v := q.Get("role"); v != "" {
		filter.Role = &v
	}
	if v := q.Get("since"); v != "" {
		formats := []string{time.RFC3339, "2006-01-02"}
		var parsed time.Time
		var err error
		for _, f := range formats {
			parsed, err = time.Parse(f, v)
			if err == nil {
				break
			}
		}
		if err != nil {
			return filter, domain.ErrValidation
		}
		parsed = parsed.UTC()
		filter.Since = &parsed
	}
	return filter, nil
}
