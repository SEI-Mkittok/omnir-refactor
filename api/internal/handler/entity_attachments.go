package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

const (
	maxUploadSize     = 25 << 20 // 25 MB
	defaultUploadsDir = "uploads"
)

// EntityAttachmentHandler serves attachment sub-resources for contacts, accounts, and deals.
type EntityAttachmentHandler struct {
	repo        repository.EntityAttachmentRepository
	uploadsDir  string
	entityType  domain.EntityType
	parentParam string
}

func NewEntityAttachmentHandler(
	repo repository.EntityAttachmentRepository,
	uploadsDir string,
	entityType domain.EntityType,
	parentParam string,
) *EntityAttachmentHandler {
	if uploadsDir == "" {
		uploadsDir = defaultUploadsDir
	}
	return &EntityAttachmentHandler{
		repo:        repo,
		uploadsDir:  uploadsDir,
		entityType:  entityType,
		parentParam: parentParam,
	}
}

func (h *EntityAttachmentHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Upload)
	r.Delete("/{attachmentID}", h.Delete)
	return r
}

func (h *EntityAttachmentHandler) List(w http.ResponseWriter, r *http.Request) {
	entityID, err := uuid.Parse(chi.URLParam(r, h.parentParam))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid entity id")
		return
	}

	attachments, err := h.repo.List(r.Context(), h.entityType, entityID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}

	if attachments == nil {
		attachments = []*domain.EntityAttachment{}
	}
	for _, a := range attachments {
		a.URL = attachmentDownloadURL(r, a.ID)
	}

	writeJSON(w, http.StatusOK, attachments)
}

func (h *EntityAttachmentHandler) Upload(w http.ResponseWriter, r *http.Request) {
	entityID, err := uuid.Parse(chi.URLParam(r, h.parentParam))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid entity id")
		return
	}

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "file too large or invalid multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "missing file field")
		return
	}
	defer file.Close()

	if header.Size > maxUploadSize {
		writeError(w, http.StatusRequestEntityTooLarge, fmt.Sprintf("file exceeds %d MB limit", maxUploadSize>>20))
		return
	}

	filename := filepath.Base(header.Filename)
	if filename == "." || filename == "/" {
		filename = "upload"
	}

	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	contentType := http.DetectContentType(buf[:n])
	if ct := header.Header.Get("Content-Type"); ct != "" && ct != "application/octet-stream" {
		contentType = ct
	}

	attachmentID := uuid.New()
	dir := filepath.Join(h.uploadsDir, attachmentID.String())
	if err := os.MkdirAll(dir, 0755); err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", "could not create upload directory")
		return
	}

	storagePath := filepath.Join(dir, filename)
	out, err := os.Create(storagePath)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", "could not save file")
		return
	}
	defer out.Close()

	if _, err := out.Write(buf[:n]); err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", "could not write file")
		return
	}
	written, err := io.Copy(out, file)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", "could not write file")
		return
	}
	totalSize := int64(n) + written

	var uploadedBy *uuid.UUID
	if claims, ok := middleware.ClaimsFromContext(r); ok {
		id := claims.UserID
		uploadedBy = &id
	}

	a := &domain.EntityAttachment{
		ID:          attachmentID,
		EntityType:  h.entityType,
		EntityID:    entityID,
		UploadedBy:  uploadedBy,
		Filename:    filename,
		ContentType: contentType,
		SizeBytes:   &totalSize,
		StoragePath: storagePath,
	}

	created, err := h.repo.Create(r.Context(), a)
	if err != nil {
		_ = os.RemoveAll(dir)
		handleDomainErr(w, err)
		return
	}

	created.URL = attachmentDownloadURL(r, created.ID)
	writeJSON(w, http.StatusCreated, created)
}

func (h *EntityAttachmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	entityID, err := uuid.Parse(chi.URLParam(r, h.parentParam))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid entity id")
		return
	}
	attachmentID, err := uuid.Parse(chi.URLParam(r, "attachmentID"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid attachment id")
		return
	}

	existing, err := h.repo.GetByID(r.Context(), attachmentID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}

	if err := h.repo.Delete(r.Context(), attachmentID, h.entityType, entityID); err != nil {
		handleDomainErr(w, err)
		return
	}

	if existing.StoragePath != "" {
		_ = os.RemoveAll(filepath.Dir(existing.StoragePath))
	}

	w.WriteHeader(http.StatusNoContent)
}

func attachmentDownloadURL(r *http.Request, id uuid.UUID) string {
	scheme := "http"
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/api/v1/attachments/%s", scheme, r.Host, id)
}

// AttachmentDownloadHandler serves raw attachment files.
type AttachmentDownloadHandler struct {
	repo repository.EntityAttachmentRepository
}

func NewAttachmentDownloadHandler(repo repository.EntityAttachmentRepository) *AttachmentDownloadHandler {
	return &AttachmentDownloadHandler{repo: repo}
}

func (h *AttachmentDownloadHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/{attachmentID}", h.Download)
	return r
}

func (h *AttachmentDownloadHandler) Download(w http.ResponseWriter, r *http.Request) {
	attachmentID, err := uuid.Parse(chi.URLParam(r, "attachmentID"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid attachment id")
		return
	}

	a, err := h.repo.GetByID(r.Context(), attachmentID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}

	f, err := os.Open(a.StoragePath)
	if err != nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "file not found on server")
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", a.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, a.Filename))
	http.ServeContent(w, r, a.Filename, a.CreatedAt, f)
}
