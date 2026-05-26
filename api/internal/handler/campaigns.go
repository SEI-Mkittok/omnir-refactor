package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

type CampaignHandler struct {
	campaigns repository.CampaignRepository
	webforms  repository.WebformRepository
}

func NewCampaignHandler(campaigns repository.CampaignRepository, webforms repository.WebformRepository) *CampaignHandler {
	return &CampaignHandler{campaigns: campaigns, webforms: webforms}
}

func (h *CampaignHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.Get)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Get("/{id}/members", h.ListMembers)
	r.Post("/{id}/members", h.AddMembers)
	return r
}

func (h *CampaignHandler) SubmissionsRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListSubmissions)
	return r
}

func (h *CampaignHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := domain.CampaignFilter{Page: parseIntDefault(r.URL.Query().Get("page"), 1), Limit: parseIntDefault(r.URL.Query().Get("limit"), 50), Q: r.URL.Query().Get("q")}
	if v := r.URL.Query().Get("status"); v != "" {
		status := domain.CampaignStatus(v)
		filter.Status = &status
	}
	campaigns, total, err := h.campaigns.ListCampaigns(r.Context(), filter)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, paginated(campaigns, total, filter.Page, filter.Limit))
}

func (h *CampaignHandler) Create(w http.ResponseWriter, r *http.Request) {
	var campaign domain.Campaign
	if err := json.NewDecoder(r.Body).Decode(&campaign); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if claims, ok := middleware.ClaimsFromContext(r); ok {
		campaign.CreatedBy = &claims.UserID
	}
	created, err := h.campaigns.CreateCampaign(r.Context(), &campaign)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *CampaignHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	campaign, err := h.campaigns.GetCampaign(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, campaign)
}

func (h *CampaignHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var patch domain.CampaignPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	updated, err := h.campaigns.UpdateCampaign(r.Context(), id, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *CampaignHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.campaigns.DeleteCampaign(r.Context(), id); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CampaignHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	members, err := h.campaigns.ListCampaignMembers(r.Context(), id, parseIntDefault(r.URL.Query().Get("limit"), 100))
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": members})
}

func (h *CampaignHandler) AddMembers(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req domain.CampaignMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if len(req.Members) == 0 {
		writeError(w, http.StatusUnprocessableEntity, "members are required")
		return
	}
	members, err := h.campaigns.AddCampaignMembers(r.Context(), id, req.Members)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": members})
}

func (h *CampaignHandler) ListSubmissions(w http.ResponseWriter, r *http.Request) {
	filter := domain.WebformSubmissionFilter{Page: parseIntDefault(r.URL.Query().Get("page"), 1), Limit: parseIntDefault(r.URL.Query().Get("limit"), 50)}
	if v := r.URL.Query().Get("campaign_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.CampaignID = &id
		}
	}
	if v := r.URL.Query().Get("webform_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.WebformID = &id
		}
	}
	submissions, total, err := h.webforms.ListWebformSubmissions(r.Context(), filter)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, paginated(submissions, total, filter.Page, filter.Limit))
}
