package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/enrichment"
	"github.com/omnir/crm-api/internal/repository"
)

// EnrichmentHandler serves the domain enrichment endpoints.
type EnrichmentHandler struct {
	svc         *enrichment.Service
	contactRepo repository.ContactRepository
}

// NewEnrichmentHandler creates an EnrichmentHandler.
func NewEnrichmentHandler(svc *enrichment.Service, contactRepo repository.ContactRepository) *EnrichmentHandler {
	return &EnrichmentHandler{svc: svc, contactRepo: contactRepo}
}

// Router wires the enrichment routes.
func (h *EnrichmentHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/domain", h.LookupDomain)
	return r
}

// ContactEnrichRouter returns the sub-router for POST /contacts/{id}/enrich.
func (h *EnrichmentHandler) ContactEnrichRouter() chi.Router {
	r := chi.NewRouter()
	r.Post("/enrich", h.EnrichContact)
	return r
}

// LookupDomain handles GET /api/v1/enrich/domain?domain=<domain>
func (h *EnrichmentHandler) LookupDomain(w http.ResponseWriter, r *http.Request) {
	d := r.URL.Query().Get("domain")
	if d == "" {
		writeError(w, http.StatusBadRequest, "domain query parameter is required")
		return
	}

	result, err := h.svc.LookupDomain(r.Context(), d)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "enrichment lookup failed")
		return
	}
	if result == nil {
		writeError(w, http.StatusNotFound, "no enrichment data found for domain")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// EnrichContact handles POST /api/v1/contacts/{id}/enrich
func (h *EnrichmentHandler) EnrichContact(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid contact id")
		return
	}

	c, err := h.contactRepo.GetByID(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}

	result, err := h.svc.EnrichContact(r.Context(), c)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "enrichment failed")
		return
	}
	if result == nil {
		writeError(w, http.StatusUnprocessableEntity, "contact has no email address to enrich from")
		return
	}
	writeJSON(w, http.StatusOK, result)
}
