package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

// EntityAttachmentHandler handles file attachment endpoints for a given entity type.
type EntityAttachmentHandler struct {
	repo       repository.EntityAttachmentRepository
	uploadsDir string
	entityType domain.EntityType
	idParam    string
}

func NewEntityAttachmentHandler(
	repo repository.EntityAttachmentRepository,
	uploadsDir string,
	entityType domain.EntityType,
	idParam string,
) *EntityAttachmentHandler {
	return &EntityAttachmentHandler{
		repo:       repo,
		uploadsDir: uploadsDir,
		entityType: entityType,
		idParam:    idParam,
	}
}

func (h *EntityAttachmentHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Upload)
	r.Delete("/{attachmentId}", h.Delete)
	return r
}

func (h *EntityAttachmentHandler) List(w http.ResponseWriter, r *http.Request) {
	entityID, err := uuid.Parse(chi.URLParam(r, h.idParam))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	items, err := h.repo.List(r.Context(), h.entityType, entityID)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *EntityAttachmentHandler) Upload(w http.ResponseWriter, r *http.Request) {
	writeProblem(w, http.StatusNotImplemented, "Not Implemented", "file upload not yet implemented")
}

func (h *EntityAttachmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	writeProblem(w, http.StatusNotImplemented, "Not Implemented", "attachment delete not yet implemented")
}

// AttachmentDownloadHandler handles secure attachment downloads.
type AttachmentDownloadHandler struct {
	repo repository.EntityAttachmentRepository
}

func NewAttachmentDownloadHandler(repo repository.EntityAttachmentRepository) *AttachmentDownloadHandler {
	return &AttachmentDownloadHandler{repo: repo}
}

func (h *AttachmentDownloadHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/{attachmentId}", h.Download)
	return r
}

func (h *AttachmentDownloadHandler) Download(w http.ResponseWriter, r *http.Request) {
	writeProblem(w, http.StatusNotImplemented, "Not Implemented", "attachment download not yet implemented")
}
