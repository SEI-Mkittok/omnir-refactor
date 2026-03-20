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

type ProductHandler struct {
	repo repository.ProductRepository
}

func NewProductHandler(repo repository.ProductRepository) *ProductHandler {
	return &ProductHandler{repo: repo}
}

func (h *ProductHandler) Router() chi.Router {
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

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing org context")
		return
	}

	q := r.URL.Query()
	filter := domain.ProductFilter{OrgID: orgID, Q: q.Get("q")}
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
	if v := q.Get("active"); v != "" {
		b := v == "true"
		filter.IsActive = &b
	}

	products, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": products, "total": total})
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid product id")
		return
	}
	p, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if err == domain.ErrNotFound {
			writeProblem(w, http.StatusNotFound, "Not Found", "product not found")
			return
		}
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing org context")
		return
	}

	var p domain.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	p.OrgID = orgID
	if err := p.Validate(); err != nil {
		writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", err.Error())
		return
	}

	created, err := h.repo.Create(r.Context(), &p)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid product id")
		return
	}
	var patch domain.ProductPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	updated, err := h.repo.Update(r.Context(), id, patch)
	if err != nil {
		if err == domain.ErrNotFound {
			writeProblem(w, http.StatusNotFound, "Not Found", "product not found")
			return
		}
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid product id")
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		if err == domain.ErrNotFound {
			writeProblem(w, http.StatusNotFound, "Not Found", "product not found")
			return
		}
		writeProblem(w, http.StatusInternalServerError, "Internal Error", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
