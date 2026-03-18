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

// TicketHandler serves the /tickets resource and its sub-resources.
type TicketHandler struct {
	tickets     repository.TicketRepository
	comments    repository.TicketCommentRepository
	attachments repository.TicketAttachmentRepository
}

func NewTicketHandler(
	tickets repository.TicketRepository,
	comments repository.TicketCommentRepository,
	attachments repository.TicketAttachmentRepository,
) *TicketHandler {
	return &TicketHandler{tickets: tickets, comments: comments, attachments: attachments}
}

func (h *TicketHandler) Router() chi.Router {
	r := chi.NewRouter()

	// All ticket routes require at least agent role.
	r.Use(middleware.RequireRole(domain.UserRoleAdmin, domain.UserRoleAgent))

	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.GetByID)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	// Nested sub-resources
	r.Get("/{id}/comments", h.ListComments)
	r.Post("/{id}/comments", h.CreateComment)
	r.Delete("/{id}/comments/{commentID}", h.DeleteComment)

	r.Get("/{id}/attachments", h.ListAttachments)
	r.Post("/{id}/attachments", h.CreateAttachment)
	r.Delete("/{id}/attachments/{attachmentID}", h.DeleteAttachment)

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

// CreateAttachment accepts a JSON body with pre-upload metadata.
// Actual file bytes are stored externally (object storage); this endpoint
// records the attachment metadata and returns the record.
func (h *TicketHandler) CreateAttachment(w http.ResponseWriter, r *http.Request) {
	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid ticket id")
		return
	}

	var req struct {
		Filename    string `json:"filename"`
		ContentType string `json:"content_type"`
		SizeBytes   *int64 `json:"size_bytes"`
		StorageURL  string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if req.Filename == "" || req.StorageURL == "" {
		writeError(w, http.StatusUnprocessableEntity, "filename and url are required")
		return
	}

	var uploadedBy *uuid.UUID
	if claims, ok := middleware.ClaimsFromContext(r); ok {
		id := claims.UserID
		uploadedBy = &id
	}

	a := &domain.TicketAttachment{
		TicketID:    ticketID,
		UploadedBy:  uploadedBy,
		Filename:    req.Filename,
		ContentType: req.ContentType,
		SizeBytes:   req.SizeBytes,
		StorageURL:  req.StorageURL,
	}
	if a.ContentType == "" {
		a.ContentType = "application/octet-stream"
	}

	created, err := h.attachments.Create(r.Context(), a)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
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
