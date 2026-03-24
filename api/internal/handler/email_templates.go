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

type EmailTemplateHandler struct {
	repo repository.EmailTemplateRepository
}

func NewEmailTemplateHandler(repo repository.EmailTemplateRepository) *EmailTemplateHandler {
	return &EmailTemplateHandler{repo: repo}
}

func (h *EmailTemplateHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(domain.UserRoleAdmin, domain.UserRoleAgent))
		r.Post("/", h.Create)
		r.Patch("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})

	return r
}

func (h *EmailTemplateHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing org context")
		return
	}

	q := r.URL.Query()
	filter := domain.EmailTemplateFilter{OrgID: orgID, Q: q.Get("q")}

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

	templates, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": templates, "total": total})
}

func (h *EmailTemplateHandler) Create(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing org context")
		return
	}

	var pt domain.EmailTemplate
	if err := json.NewDecoder(r.Body).Decode(&pt); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON")
		return
	}
	pt.OrgID = orgID

	if err := pt.Validate(); err != nil {
		writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", err.Error())
		return
	}

	created, err := h.repo.Create(r.Context(), &pt)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (h *EmailTemplateHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid template id")
		return
	}

	tmpl, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if err == domain.ErrNotFound {
			writeProblem(w, http.StatusNotFound, "Not Found", "template not found")
			return
		}
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}

	// Make sure it belongs to the current org
	orgID, _ := domain.OrgIDFromContext(r.Context())
	if tmpl.OrgID != orgID {
		writeProblem(w, http.StatusForbidden, "Forbidden", "not your template")
		return
	}

	writeJSON(w, http.StatusOK, tmpl)
}

func (h *EmailTemplateHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid template id")
		return
	}

	orgID, _ := domain.OrgIDFromContext(r.Context())
	existing, err := h.repo.GetByID(r.Context(), id)
	if err != nil || existing.OrgID != orgID {
		writeProblem(w, http.StatusNotFound, "Not Found", "template not found")
		return
	}

	var patch domain.EmailTemplatePatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}

	updated, err := h.repo.Update(r.Context(), id, patch)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *EmailTemplateHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid template id")
		return
	}

	orgID, _ := domain.OrgIDFromContext(r.Context())
	existing, err := h.repo.GetByID(r.Context(), id)
	if err != nil || existing.OrgID != orgID {
		writeProblem(w, http.StatusNotFound, "Not Found", "template not found")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
