package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// allowedExtensions is the set of file extensions accepted for upload.
// Any extension not in this set is rejected server-side regardless of content.
var allowedExtensions = map[string]bool{
	".pdf":  true,
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".webp": true,
	".svg":  true,
	".txt":  true,
	".csv":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
	".ppt":  true,
	".pptx": true,
	".zip":  true,
	".mp4":  true,
	".mp3":  true,
}

// allowedMIMEPrefixes is the set of MIME type prefixes accepted for upload.
// The sniffed MIME type (from the first 512 bytes) must match one of these.
var allowedMIMEPrefixes = []string{
	"application/pdf",
	"image/",
	"text/",
	"application/msword",
	"application/vnd.openxmlformats",
	"application/vnd.ms-",
	"application/zip",
	"application/x-zip",
	"video/mp4",
	"audio/mpeg",
	"audio/mp4",
	"application/octet-stream", // kept for binary attachments but gated by extension allowlist
}

// EntityAttachmentHandler serves attachment sub-resources for contacts, accounts, and deals.
type EntityAttachmentHandler struct {
	repo        repository.EntityAttachmentRepository
	uploadsDir  string
	baseURL     string
	entityType  domain.EntityType
	parentParam string
}

func NewEntityAttachmentHandler(
	repo repository.EntityAttachmentRepository,
	uploadsDir string,
	baseURL string,
	entityType domain.EntityType,
	parentParam string,
) *EntityAttachmentHandler {
	if uploadsDir == "" {
		uploadsDir = defaultUploadsDir
	}
	return &EntityAttachmentHandler{
		repo:        repo,
		uploadsDir:  uploadsDir,
		baseURL:     baseURL,
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
		a.URL = h.attachmentDownloadURL(a.ID)
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

	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedExtensions[ext] {
		writeError(w, http.StatusUnprocessableEntity, fmt.Sprintf("file extension %q is not allowed", ext))
		return
	}

	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	// Use the sniffed MIME type from actual file bytes — never trust the client-supplied header.
	contentType := http.DetectContentType(buf[:n])

	mimeOK := false
	for _, prefix := range allowedMIMEPrefixes {
		if strings.HasPrefix(contentType, prefix) {
			mimeOK = true
			break
		}
	}
	if !mimeOK {
		writeError(w, http.StatusUnprocessableEntity, fmt.Sprintf("file content type %q is not allowed", contentType))
		return
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

	created.URL = h.attachmentDownloadURL(created.ID)
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

// attachmentDownloadURL builds a download URL from a configured base URL, never from
// the request Host header (which can be spoofed by an attacker).
func (h *EntityAttachmentHandler) attachmentDownloadURL(id uuid.UUID) string {
	base := strings.TrimRight(h.baseURL, "/")
	return fmt.Sprintf("%s/api/v1/attachments/%s", base, id)
}

// AttachmentDownloadHandler serves raw attachment files.
type AttachmentDownloadHandler struct {
	repo         repository.EntityAttachmentRepository
	recordAccess repository.RecordAccessRepository
}

func NewAttachmentDownloadHandler(repo repository.EntityAttachmentRepository, recordAccess repository.RecordAccessRepository) *AttachmentDownloadHandler {
	return &AttachmentDownloadHandler{repo: repo, recordAccess: recordAccess}
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
	if !h.canDownloadAttachment(w, r, a) {
		return
	}

	f, err := os.Open(a.StoragePath)
	if err != nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "file not found on server")
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", a.ContentType)
	// Use "attachment" (forces download) with RFC 5987 percent-encoded filename to prevent
	// header injection and stored XSS via MIME-type confusion (OMN-613).
	encodedName := url.PathEscape(a.Filename)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename*=UTF-8''%s`, encodedName))
	http.ServeContent(w, r, a.Filename, a.CreatedAt, f)
}

func (h *AttachmentDownloadHandler) canDownloadAttachment(w http.ResponseWriter, r *http.Request, a *domain.EntityAttachment) bool {
	module, ok := attachmentEntityModule(a.EntityType)
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return false
	}
	access, ok := domain.AccessContextFromContext(r.Context())
	if !ok || !access.HasPermission(module, domain.ACLActionRead) {
		writeError(w, http.StatusNotFound, "not found")
		return false
	}
	if h.recordAccess == nil {
		writeError(w, http.StatusInternalServerError, "access check failed")
		return false
	}
	canAccess, err := h.recordAccess.CanAccessRecord(r.Context(), module, a.EntityID, domain.SharingAccessRead)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "access check failed")
		return false
	}
	if !canAccess {
		writeError(w, http.StatusNotFound, "not found")
		return false
	}
	return true
}

func attachmentEntityModule(entityType domain.EntityType) (domain.ACLModule, bool) {
	switch entityType {
	case domain.EntityTypeContact:
		return domain.ACLModuleContacts, true
	case domain.EntityTypeAccount:
		return domain.ACLModuleAccounts, true
	case domain.EntityTypeDeal:
		return domain.ACLModuleDeals, true
	default:
		return "", false
	}
}
