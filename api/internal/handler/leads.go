package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

// leadsStore is the set of operations the handler needs from the leads persistence layer.
type leadsStore interface {
	repository.LeadRepository
	ConvertToContact(ctx context.Context, leadID, contactID uuid.UUID) (*domain.Lead, error)
}

type LeadHandler struct {
	leads    leadsStore
	contacts repository.ContactRepository
	auditor  Auditor
}

func NewLeadHandler(leads leadsStore, contacts repository.ContactRepository) *LeadHandler {
	return &LeadHandler{leads: leads, contacts: contacts}
}

func (h *LeadHandler) WithAuditLog(r repository.AuditLogRepository) *LeadHandler {
	h.auditor = newAuditor(r)
	return h
}

func (h *LeadHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/sources", h.ListSources)
	r.Get("/{id}", h.GetByID)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(domain.UserRoleAdmin, domain.UserRoleAgent))
		r.Post("/", h.Create)
		r.Patch("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
		r.Post("/{id}/convert", h.Convert)
	})
	return r
}

func (h *LeadHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	// Accept both canonical names and frontend aliases.
	searchQ := q.Get("q")
	if searchQ == "" {
		searchQ = q.Get("search")
	}
	filter := domain.LeadFilter{
		Q:     searchQ,
		Sort:  q.Get("sort_by"),
		Order: q.Get("sort_dir"),
	}
	// Also accept legacy sort/order params.
	if filter.Sort == "" {
		filter.Sort = q.Get("sort")
	}
	if filter.Order == "" {
		filter.Order = q.Get("order")
	}

	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.Page = n
		}
	}
	// Accept ?limit (canonical) or ?per_page (frontend alias).
	limitParam := q.Get("limit")
	if limitParam == "" {
		limitParam = q.Get("per_page")
	}
	if limitParam != "" {
		if n, err := strconv.Atoi(limitParam); err == nil && n <= 200 {
			filter.Limit = n
		}
	}
	if v := q.Get("owner_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.OwnerID = &id
		}
	}
	if v := q.Get("status"); v != "" {
		s := domain.LeadStatus(v)
		filter.Status = &s
	}
	if v := q.Get("source"); v != "" {
		filter.Source = &v
	}
	if v := q.Get("score_min"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.ScoreMin = &n
		}
	}
	if v := q.Get("score_max"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.ScoreMax = &n
		}
	}

	if filter.Limit == 0 {
		filter.Limit = 50
	}
	if filter.Page == 0 {
		filter.Page = 1
	}

	leads, total, err := h.leads.List(r.Context(), filter)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, paginated(leads, total, filter.Page, filter.Limit))
}

func (h *LeadHandler) ListSources(w http.ResponseWriter, r *http.Request) {
	sources, err := h.leads.ListSources(r.Context())
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if sources == nil {
		sources = []string{}
	}
	writeJSON(w, http.StatusOK, sources)
}

func (h *LeadHandler) Create(w http.ResponseWriter, r *http.Request) {
	var l domain.Lead
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}

	created, err := h.leads.Create(r.Context(), &l)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	leadName := created.FirstName + " " + created.LastName
	h.auditor.log(r, domain.AuditActionCreated, domain.AuditEntityLead, idPtr(created.ID), strPtr(leadName), nil)
	writeJSON(w, http.StatusCreated, created)
}

func (h *LeadHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	l, err := h.leads.GetByID(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, l)
}

func (h *LeadHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	var patch domain.LeadPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	old, err := h.leads.GetByID(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	l, err := h.leads.Update(r.Context(), id, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	leadName := l.FirstName + " " + l.LastName
	h.auditor.log(r, domain.AuditActionUpdated, domain.AuditEntityLead, idPtr(l.ID), strPtr(leadName), buildLeadChanges(old, patch))
	writeJSON(w, http.StatusOK, l)
}

func (h *LeadHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	if err := h.leads.Delete(r.Context(), id); err != nil {
		handleDomainErr(w, err)
		return
	}
	h.auditor.log(r, domain.AuditActionDeleted, domain.AuditEntityLead, idPtr(id), nil, nil)
	w.WriteHeader(http.StatusNoContent)
}

// Convert creates a Contact from the lead's data, marks the lead as converted,
// and returns the updated lead with converted_contact_id set.
func (h *LeadHandler) Convert(w http.ResponseWriter, r *http.Request) {
	leadID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}

	lead, err := h.leads.GetByID(r.Context(), leadID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}

	if lead.Status == domain.LeadStatusConverted {
		writeProblem(w, http.StatusConflict, "Conflict", "lead is already converted")
		return
	}

	// Derive owner_id: prefer authenticated user, fall back to lead's owner.
	ownerID := uuid.Nil
	if claims, ok := middleware.ClaimsFromContext(r); ok {
		ownerID = claims.UserID
	} else if lead.OwnerID != nil {
		ownerID = *lead.OwnerID
	}

	contact := &domain.Contact{
		FirstName:           lead.FirstName,
		LastName:            lead.LastName,
		Email:               lead.Email,
		Phone:               lead.Phone,
		OwnerID:             ownerID,
		Stage:               domain.ContactStageLead,
		LeadSource:          lead.LeadSource,
		ConvertedFromLeadID: &leadID,
	}

	createdContact, err := h.contacts.Create(r.Context(), contact)
	if err != nil {
		handleDomainErr(w, err)
		return
	}

	updatedLead, err := h.leads.ConvertToContact(r.Context(), leadID, createdContact.ID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}

	h.auditor.log(r, domain.AuditActionConverted, domain.AuditEntityLead, idPtr(updatedLead.ID), strPtr(updatedLead.FirstName+" "+updatedLead.LastName), nil)
	// Return shape: { contact, lead } — contact is the primary result,
	// lead is included for callers that need to update their local state.
	writeJSON(w, http.StatusOK, map[string]any{
		"contact": createdContact,
		"lead":    updatedLead,
	})
}
