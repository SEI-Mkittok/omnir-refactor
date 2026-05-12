package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

type AccountHandler struct {
	repo   repository.AccountRepository
	cfDefs repository.CustomFieldDefinitionRepository
}

func NewAccountHandler(repo repository.AccountRepository) *AccountHandler {
	return &AccountHandler{repo: repo}
}

func (h *AccountHandler) WithCustomFields(r repository.CustomFieldDefinitionRepository) *AccountHandler {
	h.cfDefs = r
	return h
}

func (h *AccountHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Get("/{id}/hierarchy/descendants", h.ListDescendants)
	r.Get("/{id}/hierarchy/ancestors", h.ListAncestors)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(domain.UserRoleAdmin, domain.UserRoleAgent))
		r.Post("/", h.Create)
		r.Patch("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
		r.Post("/{id}/relationships", h.CreateRelationship)
		r.Delete("/relationships/{relationshipID}", h.DeleteRelationship)
	})
	return r
}

func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := domain.AccountFilter{
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
	if v := q.Get("industry"); v != "" {
		filter.Industry = &v
	}
	if v := q.Get("size"); v != "" {
		s := domain.AccountSize(v)
		filter.Size = &s
	}

	if filter.Limit == 0 {
		filter.Limit = 50
	}
	if filter.Page == 0 {
		filter.Page = 1
	}

	accounts, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, paginated(accounts, total, filter.Page, filter.Limit))
}

func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	var a domain.Account
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if a.OwnerID == uuid.Nil {
		if claims, ok := middleware.ClaimsFromContext(r); ok {
			a.OwnerID = claims.UserID
		}
	}

	if err := a.Validate(); err != nil {
		handleDomainErr(w, err)
		return
	}

	created, err := h.repo.Create(r.Context(), &a)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *AccountHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	a, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if h.cfDefs != nil {
		et := domain.CustomFieldEntityAccount
		defs, err := h.cfDefs.List(r.Context(), domain.CustomFieldDefinitionFilter{EntityType: &et})
		if err == nil && len(defs) > 0 {
			a.CustomFields = domain.ExpandCustomFields(a.CustomFields, defs)
		}
	}
	linkedFilter, err := parseLinkedEntityFilter(r)
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid linked entity filters")
		return
	}
	linked, total, err := h.repo.ListLinkedEntities(r.Context(), id, linkedFilter)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	resp := struct {
		*domain.Account
		LinkedEntities     []domain.LinkedEntity `json:"linked_entities"`
		Associations       []domain.LinkedEntity `json:"associations"`
		LinkedEntitiesMeta linkedEntitiesMeta    `json:"linked_entities_meta"`
	}{
		Account:        a,
		LinkedEntities: linked,
		Associations:   linked, // compatibility alias during rollout
		LinkedEntitiesMeta: linkedEntitiesMeta{
			Total: total,
			Page:  linkedFilter.Page,
			Limit: linkedFilter.Limit,
		},
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *AccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	var patch domain.AccountPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if h.cfDefs != nil && len(patch.CustomFields) > 0 {
		et := domain.CustomFieldEntityAccount
		defs, err := h.cfDefs.List(r.Context(), domain.CustomFieldDefinitionFilter{EntityType: &et})
		if err != nil {
			handleDomainErr(w, err)
			return
		}
		if err := domain.ValidateCustomFields(patch.CustomFields, defs); err != nil {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
	}
	a, err := h.repo.Update(r.Context(), id, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AccountHandler) CreateRelationship(w http.ResponseWriter, r *http.Request) {
	parentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	var rel domain.AccountRelationship
	if err := json.NewDecoder(r.Body).Decode(&rel); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	rel.ParentAccountID = parentID
	if rel.EffectiveFrom.IsZero() {
		rel.EffectiveFrom = time.Now().UTC()
	}
	if err := rel.Validate(); err != nil {
		handleDomainErr(w, err)
		return
	}
	if _, ok := domain.AccessContextFromContext(r.Context()); ok {
		canAccessChild, err := h.repo.CanAccess(r.Context(), rel.ChildAccountID, domain.SharingAccessWrite)
		if err != nil {
			handleDomainErr(w, err)
			return
		}
		if !canAccessChild {
			handleDomainErr(w, domain.ErrNotFound)
			return
		}
	}
	created, err := h.repo.CreateRelationship(r.Context(), &rel)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *AccountHandler) DeleteRelationship(w http.ResponseWriter, r *http.Request) {
	relationshipID, err := uuid.Parse(chi.URLParam(r, "relationshipID"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid relationshipID")
		return
	}
	var deletedBy *uuid.UUID
	if claims, ok := middleware.ClaimsFromContext(r); ok && claims.UserID != uuid.Nil {
		deletedBy = &claims.UserID
	}
	if err := h.repo.DeleteRelationship(r.Context(), relationshipID, deletedBy); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AccountHandler) ListDescendants(w http.ResponseWriter, r *http.Request) {
	accountID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	ids, err := h.repo.ListDescendants(r.Context(), accountID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"account_id": accountID, "descendant_ids": ids})
}

func (h *AccountHandler) ListAncestors(w http.ResponseWriter, r *http.Request) {
	accountID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	ids, err := h.repo.ListAncestors(r.Context(), accountID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"account_id": accountID, "ancestor_ids": ids})
}
