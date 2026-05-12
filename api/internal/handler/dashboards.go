package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

// DashboardHandler serves /dashboards and /reports/schedules.
type DashboardHandler struct {
	repo        repository.DashboardRepository
	reportsRepo repository.ReportsRepository
}

func NewDashboardHandler(repo repository.DashboardRepository, reportsRepo repository.ReportsRepository) *DashboardHandler {
	return &DashboardHandler{repo: repo, reportsRepo: reportsRepo}
}

// Router mounts dashboard CRUD + run.
func (h *DashboardHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.Get)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/run", h.Run)
	return r
}

// ScheduleRouter mounts scheduled report CRUD under /reports/schedules.
func (h *DashboardHandler) ScheduleRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(h.requireScheduleAdmin)
	r.Get("/", h.ListSchedules)
	r.Post("/", h.CreateSchedule)
	r.Get("/{id}", h.GetSchedule)
	r.Patch("/{id}", h.UpdateSchedule)
	r.Delete("/{id}", h.DeleteSchedule)
	return r
}

// ── Dashboard CRUD ────────────────────────────────────────────────────────────

func (h *DashboardHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.repo.ListDashboards(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list dashboards")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": list})
}

func (h *DashboardHandler) Create(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}
	var req struct {
		Name    string          `json:"name"`
		Widgets []domain.Widget `json:"widgets"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Widgets == nil {
		req.Widgets = []domain.Widget{}
	}

	d := &domain.CustomDashboard{
		ID:      uuid.New(),
		OrgID:   orgID,
		Name:    req.Name,
		Widgets: req.Widgets,
	}
	if claims, ok := middleware.ClaimsFromContext(r); ok {
		d.CreatedBy = claims.UserID
	}

	created, err := h.repo.CreateDashboard(r.Context(), d)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create dashboard")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *DashboardHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid dashboard id")
		return
	}
	d, err := h.repo.GetDashboardByID(r.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "dashboard not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get dashboard")
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (h *DashboardHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid dashboard id")
		return
	}
	var patch domain.CustomDashboardPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	updated, err := h.repo.UpdateDashboard(r.Context(), id, patch)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "dashboard not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update dashboard")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *DashboardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid dashboard id")
		return
	}
	if err := h.repo.DeleteDashboard(r.Context(), id); errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "dashboard not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete dashboard")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Run executes dashboard queries and returns widget data.
// POST /api/v1/dashboards/:id/run
func (h *DashboardHandler) Run(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid dashboard id")
		return
	}
	d, err := h.repo.GetDashboardByID(r.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "dashboard not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get dashboard")
		return
	}

	result := domain.DashboardRunResult{
		DashboardID: d.ID,
		Widgets:     make([]domain.WidgetData, 0, len(d.Widgets)),
	}
	for _, widget := range d.Widgets {
		data, runErr := h.runWidget(r, widget)
		if runErr != nil {
			data = map[string]string{"error": runErr.Error()}
		}
		result.Widgets = append(result.Widgets, domain.WidgetData{Widget: widget, Data: data})
	}
	writeJSON(w, http.StatusOK, result)
}

// runWidget dispatches a widget type to the appropriate reports query.
func (h *DashboardHandler) runWidget(r *http.Request, widget domain.Widget) (any, error) {
	ctx := r.Context()
	switch widget.Type {
	case "deals_by_stage":
		return h.reportsRepo.DealsByStage(ctx)
	case "contacts_monthly":
		return h.reportsRepo.ContactsMonthly(ctx)
	case "activities_by_type":
		return h.reportsRepo.ActivitiesByType(ctx)
	case "ticket_metrics":
		return h.reportsRepo.TicketMetrics(ctx, domain.ReportFilter{})
	case "deal_metrics":
		return h.reportsRepo.DealMetrics(ctx, domain.ReportFilter{})
	case "contact_metrics":
		return h.reportsRepo.ContactMetrics(ctx, domain.ReportFilter{})
	case "lead_metrics":
		return h.reportsRepo.LeadMetrics(ctx, domain.ReportFilter{})
	case "revenue_projection":
		return h.reportsRepo.RevenueProjection(ctx, 3)
	case "activity_summary":
		return h.reportsRepo.ActivitySummary(ctx, domain.ReportFilter{})
	case "pipeline_funnel":
		return h.reportsRepo.PipelineFunnel(ctx, nil, domain.ReportFilter{})
	case "conversion_rates":
		return h.reportsRepo.ConversionRates(ctx, domain.ReportFilter{})
	default:
		return map[string]string{"note": "unsupported widget type"}, nil
	}
}

// ── Scheduled report CRUD ─────────────────────────────────────────────────────

func (h *DashboardHandler) requireScheduleAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !hasModuleAdminAccess(r, domain.ACLModuleReports) {
			writeError(w, http.StatusForbidden, "report admin access required")
			return
		}
		access, ok := domain.AccessContextFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusForbidden, "access context required")
			return
		}
		if access.SuperAdminBypass {
			next.ServeHTTP(w, r)
			return
		}
		if !access.CanAccessAllRecords(domain.ACLModuleContacts, domain.SharingAccessRead) ||
			!access.CanAccessAllRecords(domain.ACLModuleAccounts, domain.SharingAccessRead) ||
			!access.CanAccessAllRecords(domain.ACLModuleDeals, domain.SharingAccessRead) {
			writeError(w, http.StatusForbidden, "all-record access required to schedule reports")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *DashboardHandler) ListSchedules(w http.ResponseWriter, r *http.Request) {
	list, err := h.repo.ListSchedules(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list schedules")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": list})
}

func (h *DashboardHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}
	var req struct {
		DashboardID uuid.UUID `json:"dashboard_id"`
		Schedule    string    `json:"schedule"`
		Recipients  []string  `json:"recipients"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.DashboardID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "dashboard_id is required")
		return
	}
	if req.Schedule == "" {
		writeError(w, http.StatusBadRequest, "schedule is required")
		return
	}
	if len(req.Recipients) == 0 {
		writeError(w, http.StatusBadRequest, "at least one recipient is required")
		return
	}

	s := &domain.ScheduledReport{
		ID:          uuid.New(),
		OrgID:       orgID,
		DashboardID: req.DashboardID,
		Schedule:    req.Schedule,
		Recipients:  req.Recipients,
	}
	created, err := h.repo.CreateSchedule(r.Context(), s)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create schedule")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *DashboardHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid schedule id")
		return
	}
	s, err := h.repo.GetScheduleByID(r.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "schedule not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get schedule")
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func (h *DashboardHandler) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid schedule id")
		return
	}
	var patch domain.ScheduledReportPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	updated, err := h.repo.UpdateSchedule(r.Context(), id, patch)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "schedule not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update schedule")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *DashboardHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid schedule id")
		return
	}
	if err := h.repo.DeleteSchedule(r.Context(), id); errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "schedule not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete schedule")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
