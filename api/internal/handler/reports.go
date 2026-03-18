package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

// ReportsHandler serves the /reports sub-resources.
type ReportsHandler struct {
	repo repository.ReportsRepository
}

func NewReportsHandler(repo repository.ReportsRepository) *ReportsHandler {
	return &ReportsHandler{repo: repo}
}

func (h *ReportsHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.Summary)
	r.Get("/tickets", h.Tickets)
	r.Get("/contacts", h.Contacts)
	r.Get("/deals", h.Deals)
	r.Get("/leads", h.Leads)
	return r
}

// Summary returns all analytics metrics in a single response.
func (h *ReportsHandler) Summary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dealsByStage, err := h.repo.DealsByStage(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load deals by stage")
		return
	}

	contactsMonthly, err := h.repo.ContactsMonthly(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load contacts monthly")
		return
	}

	activitiesByType, err := h.repo.ActivitiesByType(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load activities by type")
		return
	}

	summary := domain.ReportsSummary{
		DealsByStage:     dealsByStage,
		ContactsMonthly:  contactsMonthly,
		ActivitiesByType: activitiesByType,
	}
	writeJSON(w, http.StatusOK, summary)
}

// parseReportFilter extracts from/to date params and an optional org_id (admin-only).
func parseReportFilter(r *http.Request) (domain.ReportFilter, error) {
	var f domain.ReportFilter
	q := r.URL.Query()

	if v := q.Get("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			t, err = time.Parse("2006-01-02", v)
			if err != nil {
				return f, fmt.Errorf("invalid query parameter \"from\": must be ISO 8601 date or datetime")
			}
		}
		f.From = &t
	}
	if v := q.Get("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			t, err = time.Parse("2006-01-02", v)
			if err != nil {
				return f, fmt.Errorf("invalid query parameter \"to\": must be ISO 8601 date or datetime")
			}
		}
		f.To = &t
	}

	// org_id is admin-only; silently ignored for non-admin callers.
	if v := q.Get("org_id"); v != "" {
		claims, ok := middleware.ClaimsFromContext(r)
		if ok && claims.Role == string(domain.UserRoleAdmin) {
			if id, err := uuid.Parse(v); err == nil {
				f.OrgID = &id
			}
		}
	}

	return f, nil
}

// Tickets handles GET /reports/tickets
func (h *ReportsHandler) Tickets(w http.ResponseWriter, r *http.Request) {
	f, err := parseReportFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	report, err := h.repo.TicketMetrics(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load ticket metrics")
		return
	}
	writeJSON(w, http.StatusOK, report)
}

// Contacts handles GET /reports/contacts
func (h *ReportsHandler) Contacts(w http.ResponseWriter, r *http.Request) {
	f, err := parseReportFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	report, err := h.repo.ContactMetrics(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load contact metrics")
		return
	}
	writeJSON(w, http.StatusOK, report)
}

// Deals handles GET /reports/deals
func (h *ReportsHandler) Deals(w http.ResponseWriter, r *http.Request) {
	f, err := parseReportFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	report, err := h.repo.DealMetrics(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load deal metrics")
		return
	}
	writeJSON(w, http.StatusOK, report)
}

// Leads handles GET /reports/leads
func (h *ReportsHandler) Leads(w http.ResponseWriter, r *http.Request) {
	f, err := parseReportFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	report, err := h.repo.LeadMetrics(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load lead metrics")
		return
	}
	writeJSON(w, http.StatusOK, report)
}
