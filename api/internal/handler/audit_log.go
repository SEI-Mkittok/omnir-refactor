package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

type AuditLogHandler struct {
	repo repository.AuditLogRepository
}

func NewAuditLogHandler(repo repository.AuditLogRepository) *AuditLogHandler {
	return &AuditLogHandler{repo: repo}
}

func (h *AuditLogHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequireRole(domain.UserRoleAdmin))
	r.Get("/", h.List)
	r.Get("/export", h.Export)
	return r
}

func (h *AuditLogHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	filter := domain.AuditLogFilter{
		OrgID: claims.OrgID,
		Page:  1,
		Limit: 50,
	}

	if v := r.URL.Query().Get("entityType"); v != "" {
		et := domain.AuditEntityType(v)
		filter.EntityType = &et
	}
	if v := r.URL.Query().Get("entityId"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.EntityID = &id
		}
	}
	if v := r.URL.Query().Get("userId"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.UserID = &id
		}
	}
	if v := r.URL.Query().Get("action"); v != "" {
		a := domain.AuditAction(v)
		filter.Action = &a
	}
	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.From = &t
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.To = &t
		}
	}
	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			filter.Page = n
		}
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			filter.Limit = n
		}
	}

	entries, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		handleDomainErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, paginated(entries, total, filter.Page, filter.Limit))
}

func (h *AuditLogHandler) Export(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Export up to 10 000 rows without pagination.
	filter := domain.AuditLogFilter{
		OrgID: claims.OrgID,
		Page:  1,
		Limit: 10000,
	}

	entries, _, err := h.repo.List(r.Context(), filter)
	if err != nil {
		handleDomainErr(w, err)
		return
	}

	filename := fmt.Sprintf("audit-log-%s.csv", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"id", "action", "entity_type", "entity_id", "entity_name", "user_id", "ip_address", "created_at"})
	for _, e := range entries {
		row := []string{
			e.ID.String(),
			string(e.Action),
			string(e.EntityType),
			uuidStr(e.EntityID),
			strStr(e.EntityName),
			uuidStr(e.UserID),
			strStr(e.IPAddress),
			e.CreatedAt.Format(time.RFC3339),
		}
		_ = cw.Write(row)
	}
	cw.Flush()
}

func uuidStr(u *uuid.UUID) string {
	if u == nil {
		return ""
	}
	return u.String()
}

func strStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
