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
}

func NewLeadHandler(leads leadsStore, contacts repository.ContactRepository) *LeadHandler {
	return &LeadHandler{leads: leads, contacts: contacts}
}

func (h *LeadHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
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
	filter := domain.LeadFilter{
		Q:     q.Get("q"),
		Sort:  q.Get("sort"),
		Order: q.Get("order"),
	}

	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.Page = n
		}
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n <= 200 {
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
	l, err := h.leads.Update(r.Context(), id, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
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
		FirstName:  lead.FirstName,
		LastName:   lead.LastName,
		Email:      lead.Email,
		Phone:      lead.Phone,
		OwnerID:    ownerID,
		Stage:      domain.ContactStageLead,
		LeadSource: lead.LeadSource,
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

	writeJSON(w, http.StatusOK, map[string]any{
		"lead":    updatedLead,
		"contact": createdContact,
	})
}
