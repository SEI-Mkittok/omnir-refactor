package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

// SequenceHandler serves /sequences.
type SequenceHandler struct {
	repo repository.SequenceRepository
}

func NewSequenceHandler(repo repository.SequenceRepository) *SequenceHandler {
	return &SequenceHandler{repo: repo}
}

func (h *SequenceHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.Get)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Get("/{id}/steps", h.ListSteps)
	r.Get("/{id}/enrollments", h.ListEnrollments)
	r.Post("/{id}/enroll", h.Enroll)
	r.Get("/{id}/analytics", h.Analytics)
	r.Patch("/{id}/enrollments/{enrollmentId}", h.UpdateEnrollment)
	return r
}

// List handles GET /sequences
func (h *SequenceHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := domain.SequenceFilter{
		Page:  1,
		Limit: 50,
	}
	q := r.URL.Query()
	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			filter.Page = n
		}
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			filter.Limit = n
		}
	}
	if v := q.Get("status"); v != "" {
		st := domain.SequenceStatus(v)
		filter.Status = &st
	}

	seqs, total, err := h.repo.ListSequences(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list sequences")
		return
	}
	if seqs == nil {
		seqs = []*domain.EmailSequence{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": seqs, "total": total})
}

// Create handles POST /sequences
func (h *SequenceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateSequenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	s := &domain.EmailSequence{
		Name:        req.Name,
		Description: req.Description,
	}
	created, err := h.repo.CreateSequence(r.Context(), s, req.Steps)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create sequence")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// Get handles GET /sequences/{id}
func (h *SequenceHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	s, err := h.repo.GetSequence(r.Context(), id)
	if err != nil {
		if err == domain.ErrNotFound {
			writeError(w, http.StatusNotFound, "sequence not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get sequence")
		return
	}
	writeJSON(w, http.StatusOK, s)
}

// Update handles PATCH /sequences/{id}
func (h *SequenceHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req domain.UpdateSequenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	s, err := h.repo.UpdateSequence(r.Context(), id, req)
	if err != nil {
		if err == domain.ErrNotFound {
			writeError(w, http.StatusNotFound, "sequence not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update sequence")
		return
	}
	writeJSON(w, http.StatusOK, s)
}

// Delete handles DELETE /sequences/{id}
func (h *SequenceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.repo.DeleteSequence(r.Context(), id); err != nil {
		if err == domain.ErrNotFound {
			writeError(w, http.StatusNotFound, "sequence not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete sequence")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Enroll handles POST /sequences/{id}/enroll
func (h *SequenceHandler) Enroll(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req domain.EnrollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	enrolled, err := h.repo.Enroll(r.Context(), id, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to enroll contacts")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"enrolled": enrolled})
}

// ListEnrollments handles GET /sequences/{id}/enrollments
func (h *SequenceHandler) ListEnrollments(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	enrollments, err := h.repo.ListEnrollments(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list enrollments")
		return
	}
	if enrollments == nil {
		enrollments = []*domain.SequenceEnrollment{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": enrollments})
}

// UpdateEnrollment handles PATCH /sequences/{id}/enrollments/{enrollmentId}
func (h *SequenceHandler) UpdateEnrollment(w http.ResponseWriter, r *http.Request) {
	enrollmentID, err := uuid.Parse(chi.URLParam(r, "enrollmentId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid enrollmentId")
		return
	}
	var body struct {
		Status domain.EnrollmentStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.repo.UpdateEnrollmentStatus(r.Context(), enrollmentID, body.Status); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update enrollment")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListSteps handles GET /sequences/{id}/steps
func (h *SequenceHandler) ListSteps(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	steps, err := h.repo.ListSteps(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list steps")
		return
	}
	if steps == nil {
		steps = []domain.SequenceStep{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": steps})
}

// Analytics handles GET /sequences/{id}/analytics
func (h *SequenceHandler) Analytics(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	analytics, err := h.repo.GetAnalytics(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load analytics")
		return
	}
	writeJSON(w, http.StatusOK, analytics)
}
