package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
	"github.com/omnir/crm-api/internal/storage"
)

// TicketHandler serves the /tickets resource and its sub-resources.
type TicketHandler struct {
	tickets     repository.TicketRepository
	comments    repository.TicketCommentRepository
	attachments repository.TicketAttachmentRepository
	storage     storage.Backend
}

func NewTicketHandler(
	tickets repository.TicketRepository,
	comments repository.TicketCommentRepository,
	attachments repository.TicketAttachmentRepository,
	store storage.Backend,
) *TicketHandler {
	return &TicketHandler{tickets: tickets, comments: comments, attachments: attachments, storage: store}
}

func (h *TicketHandler) Router() chi.Router {
	r := chi.NewRouter()

	// agentOnly restricts mutation operations to admin and agent roles.
	// Read operations and comment creation are also open to client role,
	// with per-handler logic filtering what clients can see/do.
	agentOnly := middleware.RequireRole(domain.UserRoleAdmin, domain.UserRoleAgent)

	r.Get("/", h.List)
	r.With(agentOnly).Post("/", h.Create)
	r.Get("/{id}", h.GetByID)
	r.With(agentOnly).Patch("/{id}", h.Update)
	r.With(agentOnly).Delete("/{id}", h.Delete)

	// Nested sub-resources
	r.Get("/{id}/comments", h.ListComments)   // clients see public comments only (filtered in handler)
	r.Post("/{id}/comments", h.CreateComment) // clients can comment but not mark internal (enforced in handler)
	r.With(agentOnly).Delete("/{id}/comments/{commentID}", h.DeleteComment)

	r.Get("/{id}/attachments", h.ListAttachments)
	r.With(agentOnly).Post("/{id}/attachments", h.CreateAttachment)
	r.Get("/{id}/attachments/{attachmentID}", h.GetAttachment)
	r.With(agentOnly).Delete("/{id}/attachments/{attachmentID}", h.DeleteAttachment)

	return r
}

// ---- Ticket CRUD ----

func (h *TicketHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := domain.TicketFilter{
		Q:     q.Get("search"),
		Sort:  q.Get("sort_by"),
		Order: q.Get("sort_dir"),
	}

	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.Page = n
		}
	}
	if v := q.Get("per_page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n <= 200 {
			filter.Limit = n
		}
	}
	if v := q.Get("status"); v != "" {
		s := domain.TicketStatus(v)
		filter.Status = &s
	}
	if v := q.Get("priority"); v != "" {
		p := domain.TicketPriority(v)
		filter.Priority = &p
	}
	if v := q.Get("assignee_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.AssigneeID = &id
		}
	}
	if v := q.Get("contact_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.ContactID = &id
		}
	}

	if filter.Limit == 0 {
		filter.Limit = 50
	}
	if filter.Page == 0 {
		filter.Page = 1
	}

	tickets, total, err := h.tickets.List(r.Context(), filter)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, paginated(tickets, total, filter.Page, filter.Limit))
}

func (h *TicketHandler) Create(w http.ResponseWriter, r *http.Request) {
	var t domain.Ticket
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if t.Subject == "" {
		writeError(w, http.StatusUnprocessableEntity, "subject is required")
		return
	}

	created, err := h.tickets.Create(r.Context(), &t)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *TicketHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	t, err := h.tickets.GetByID(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *TicketHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	var patch domain.TicketPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	t, err := h.tickets.Update(r.Context(), id, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *TicketHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	if err := h.tickets.Delete(r.Context(), id); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- Comments ----

func (h *TicketHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid ticket id")
		return
	}

	// Clients (non-agents) can only see public comments.
	filter := domain.TicketCommentFilter{TicketID: ticketID}
	if claims, ok := middleware.ClaimsFromContext(r); ok {
		if claims.Role == string(domain.UserRoleClient) {
			f := false
			filter.IsInternal = &f
		}
	}

	comments, err := h.comments.List(r.Context(), filter)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, comments)
}

func (h *TicketHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid ticket id")
		return
	}

	var req struct {
		Body       string `json:"body"`
		IsInternal bool   `json:"is_internal"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if req.Body == "" {
		writeError(w, http.StatusUnprocessableEntity, "body is required")
		return
	}

	// Extract author from JWT claims.
	var authorID *uuid.UUID
	if claims, ok := middleware.ClaimsFromContext(r); ok {
		id := claims.UserID
		authorID = &id
		// Clients cannot post internal notes.
		if claims.Role == string(domain.UserRoleClient) {
			req.IsInternal = false
		}
	}

	c := &domain.TicketComment{
		TicketID:   ticketID,
		AuthorID:   authorID,
		Body:       req.Body,
		IsInternal: req.IsInternal,
	}
	created, err := h.comments.Create(r.Context(), c)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *TicketHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid ticket id")
		return
	}
	commentID, err := uuid.Parse(chi.URLParam(r, "commentID"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid comment id")
		return
	}
	if err := h.comments.Delete(r.Context(), commentID, ticketID); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- Attachments ----

func (h *TicketHandler) ListAttachments(w http.ResponseWriter, r *http.Request) {
	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid ticket id")
		return
	}
	attachments, err := h.attachments.List(r.Context(), ticketID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, attachments)
}

// CreateAttachment accepts a multipart/form-data upload.
// The file is stored via the configured storage backend (S3 or local).
// Field name: "file".
func (h *TicketHandler) CreateAttachment(w http.ResponseWriter, r *http.Request) {
	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid ticket id")
		return
	}

	// Limit memory to 32 MB; larger files are buffered to temp files automatically.
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "failed to parse multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "file field is required")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	var uploadedBy *uuid.UUID
	orgID := uuid.Nil
	if claims, ok := middleware.ClaimsFromContext(r); ok {
		id := claims.UserID
		uploadedBy = &id
		orgID = claims.OrgID
	}

	storageKey := fmt.Sprintf("orgs/%s/tickets/%s/%s/%s", orgID, ticketID, uuid.New(), header.Filename)
	if err := h.storage.Upload(r.Context(), storageKey, file, header.Size, contentType); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to upload file")
		return
	}

	// Pre-generate the attachment ID so the storage_url stored in the DB is complete.
	attachmentID := uuid.New()
	size := header.Size
	a := &domain.TicketAttachment{
		ID:             attachmentID,
		TicketID:       ticketID,
		UploadedBy:     uploadedBy,
		Filename:       header.Filename,
		ContentType:    contentType,
		SizeBytes:      &size,
		StorageKey:     storageKey,
		StorageBackend: h.storage.Type(),
		StorageURL:     fmt.Sprintf("/api/v1/tickets/%s/attachments/%s", ticketID, attachmentID),
	}

	created, err := h.attachments.Create(r.Context(), a)
	if err != nil {
		// Best-effort cleanup of uploaded object on DB failure.
		_ = h.storage.Delete(r.Context(), storageKey)
		handleDomainErr(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

// GetAttachment serves or redirects to the attachment file.
// S3 backend: issues a 302 redirect to a presigned URL.
// Local backend: streams the file directly.
func (h *TicketHandler) GetAttachment(w http.ResponseWriter, r *http.Request) {
	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid ticket id")
		return
	}
	attachmentID, err := uuid.Parse(chi.URLParam(r, "attachmentID"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid attachment id")
		return
	}

	a, err := h.attachments.GetByID(r.Context(), attachmentID, ticketID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}

	// Legacy record: no storage key — redirect to stored URL.
	if a.StorageKey == "" {
		if a.StorageURL != "" {
			http.Redirect(w, r, a.StorageURL, http.StatusFound)
			return
		}
		writeError(w, http.StatusNotFound, "attachment has no downloadable content")
		return
	}

	// Try presigned URL first (S3 backend returns non-empty string).
	presignedURL, err := h.storage.PresignURL(r.Context(), a.StorageKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate download URL")
		return
	}
	if presignedURL != "" {
		http.Redirect(w, r, presignedURL, http.StatusFound)
		return
	}

	// Local backend: stream the file.
	rc, err := h.storage.Open(r.Context(), a.StorageKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to open attachment")
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", a.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, a.Filename))
	if a.SizeBytes != nil {
		w.Header().Set("Content-Length", strconv.FormatInt(*a.SizeBytes, 10))
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, rc)
}

func (h *TicketHandler) DeleteAttachment(w http.ResponseWriter, r *http.Request) {
	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid ticket id")
		return
	}
	attachmentID, err := uuid.Parse(chi.URLParam(r, "attachmentID"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid attachment id")
		return
	}
	if err := h.attachments.Delete(r.Context(), attachmentID, ticketID); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
