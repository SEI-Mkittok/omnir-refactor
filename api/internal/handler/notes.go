package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

// NoteHandler serves notes nested under a parent entity
// (e.g. /contacts/:id/notes, /accounts/:id/notes, /deals/:id/notes).
type NoteHandler struct {
	repo        repository.NoteRepository
	entityType  domain.NoteEntityType
	parentParam string // chi URL param name for the parent entity ID
}

// NewNoteHandler constructs a NoteHandler.
func NewNoteHandler(repo repository.NoteRepository, entityType domain.NoteEntityType, parentParam string) *NoteHandler {
	return &NoteHandler{repo: repo, entityType: entityType, parentParam: parentParam}
}

// Router returns the sub-router to be mounted under e.g. /contacts/{contactID}/notes.
func (h *NoteHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(domain.UserRoleAdmin, domain.UserRoleAgent))
		r.Post("/", h.Create)
		r.Delete("/{noteID}", h.Delete)
	})
	return r
}

// List returns paginated notes for the parent entity.
func (h *NoteHandler) List(w http.ResponseWriter, r *http.Request) {
	parentID, err := uuid.Parse(chi.URLParam(r, h.parentParam))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid entity id")
		return
	}

	q := r.URL.Query()
	filter := domain.NoteFilter{
		EntityType: h.entityType,
		EntityID:   parentID,
	}
	if v := q.Get("page"); v != "" {
		if n, err2 := strconv.Atoi(v); err2 == nil {
			filter.Page = n
		}
	}
	// Accept ?per_page (frontend) or ?limit (canonical).
	if limitVal := q.Get("per_page"); limitVal == "" {
		if v := q.Get("limit"); v != "" {
			if n, err2 := strconv.Atoi(v); err2 == nil && n <= 200 {
				filter.Limit = n
			}
		}
	} else if n, err2 := strconv.Atoi(limitVal); err2 == nil && n <= 200 {
		filter.Limit = n
	}
	if filter.Limit == 0 {
		filter.Limit = 50
	}
	if filter.Page == 0 {
		filter.Page = 1
	}

	notes, total, err := h.repo.ListByEntity(r.Context(), filter)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, paginated(notes, total, filter.Page, filter.Limit))
}

// Create appends a note to the parent entity.
func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	parentID, err := uuid.Parse(chi.URLParam(r, h.parentParam))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid entity id")
		return
	}

	var n domain.Note
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}

	// Override entity fields from the URL — callers must not set these themselves.
	n.EntityType = h.entityType
	n.EntityID = parentID

	if err := n.Validate(); err != nil {
		writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", err.Error())
		return
	}

	created, err := h.repo.Create(r.Context(), &n)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// Delete soft-deletes a note.
func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	noteID, err := uuid.Parse(chi.URLParam(r, "noteID"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid note id")
		return
	}

	if err := h.repo.Delete(r.Context(), noteID); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
