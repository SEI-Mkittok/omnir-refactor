package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

type ReportsHandler struct {
	repo repository.ReportsRepository
}

func NewReportsHandler(repo repository.ReportsRepository) *ReportsHandler {
	return &ReportsHandler{repo: repo}
}

func (h *ReportsHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.Summary)
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
