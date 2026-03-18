package handler

import (
	"context"
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

// EmailEnqueuer is satisfied by worker.EmailNotifier. Defined here to avoid
// an import cycle and to make the handler independently testable.
type EmailEnqueuer interface {
	Enqueue(job domain.EmailJob)
}

// TicketHandler serves the /tickets resource and its sub-resources.
type TicketHandler struct {
	tickets     repository.TicketRepository
	comments    repository.TicketCommentRepository
	attachments repository.TicketAttachmentRepository
	slaPolicies repository.SLAPolicyRepository
	users       repository.UserRepository
	notifPrefs  repository.NotificationPrefRepository
	emailQueue  EmailEnqueuer // nil → email notifications disabled
}

func NewTicketHandler(
	tickets repository.TicketRepository,
	comments repository.TicketCommentRepository,
	attachments repository.TicketAttachmentRepository,
	slaPolicies repository.SLAPolicyRepository,
) *TicketHandler {
	return &TicketHandler{tickets: tickets, comments: comments, attachments: attachments, slaPolicies: slaPolicies}
}

// ticketResponse wraps a Ticket with computed SLA status for API responses.
type ticketResponse struct {
	*domain.Ticket
	SLA *domain.SLAStatus `json:"sla,omitempty"`
}

// WithEmailNotifications attaches the dependencies needed for outbound email notifications.
func (h *TicketHandler) WithEmailNotifications(
	users repository.UserRepository,
	notifPrefs repository.NotificationPrefRepository,
	queue EmailEnqueuer,
) *TicketHandler {
	h.users = users
	h.notifPrefs = notifPrefs
	h.emailQueue = queue
	return h
}

func (h *TicketHandler) Router() chi.Router {
	r := chi.NewRouter()

	agentOnly := middleware.RequireRole(domain.UserRoleAdmin, domain.UserRoleAgent)

	r.Get("/", h.List)
	r.With(agentOnly).Post("/", h.Create)
	r.Get("/{id}", h.GetByID)
	r.With(agentOnly).Patch("/{id}", h.Update)
	r.With(agentOnly).Delete("/{id}", h.Delete)

	r.Get("/{id}/comments", h.ListComments)
	r.Post("/{id}/comments", h.CreateComment)
	r.With(agentOnly).Delete("/{id}/comments/{commentID}", h.DeleteComment)

	r.Get("/{id}/attachments", h.ListAttachments)
	r.With(agentOnly).Post("/{id}/attachments", h.CreateAttachment)
	r.With(agentOnly).Delete("/{id}/attachments/{attachmentID}", h.DeleteAttachment)

	return r
}

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
	if t.Status != "" && !t.Status.IsValid() {
		writeError(w, http.StatusUnprocessableEntity, "invalid status: must be one of open, in_progress, pending, resolved, closed")
		return
	}
	if t.Priority != "" && !t.Priority.IsValid() {
		writeError(w, http.StatusUnprocessableEntity, "invalid priority: must be one of low, medium, high, critical")
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

	resp := &ticketResponse{Ticket: t}
	if t.SLAPolicyID != nil && h.slaPolicies != nil {
		if policy, err := h.slaPolicies.GetByID(r.Context(), *t.SLAPolicyID); err == nil {
			now := time.Now().UTC()
			responseDue := t.CreatedAt.Add(time.Duration(float64(time.Hour) * policy.ResponseTimeHours))
			resolutionDue := t.CreatedAt.Add(time.Duration(float64(time.Hour) * policy.ResolutionTimeHours))

			responseBreached := now.After(responseDue) && t.FirstRespondedAt == nil
			resolutionBreached := now.After(resolutionDue)

			status := "on_track"
			if responseBreached || resolutionBreached {
				status = "breached"
			} else {
				responseTotal := responseDue.Sub(t.CreatedAt)
				resolutionTotal := resolutionDue.Sub(t.CreatedAt)
				responseRemaining := responseDue.Sub(now)
				resolutionRemaining := resolutionDue.Sub(now)
				if (t.FirstRespondedAt == nil && responseRemaining < responseTotal/5) ||
					resolutionRemaining < resolutionTotal/5 {
					status = "at_risk"
				}
			}

			resp.SLA = &domain.SLAStatus{
				PolicyID:           t.SLAPolicyID,
				PolicyName:         policy.Name,
				ResponseDueAt:      &responseDue,
				ResolutionDueAt:    &resolutionDue,
				ResponseBreached:   responseBreached,
				ResolutionBreached: resolutionBreached,
				FirstRespondedAt:   t.FirstRespondedAt,
				Status:             status,
			}
		}
	}

	writeJSON(w, http.StatusOK, resp)
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
	if patch.Status != nil && !(*patch.Status).IsValid() {
		writeError(w, http.StatusUnprocessableEntity, "invalid status: must be one of open, in_progress, pending, resolved, closed")
		return
	}
	if patch.Priority != nil && !(*patch.Priority).IsValid() {
		writeError(w, http.StatusUnprocessableEntity, "invalid priority: must be one of low, medium, high, critical")
		return
	}

	// Fetch old state before update so we can detect changes for notifications.
	var old *domain.Ticket
	if h.emailQueue != nil {
		old, _ = h.tickets.GetByID(r.Context(), id) // best-effort
	}

	t, err := h.tickets.Update(r.Context(), id, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}

	if h.emailQueue != nil && old != nil {
		go h.enqueueTicketNotifications(context.Background(), old, t)
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

// enqueueTicketNotifications detects state changes and enqueues email jobs.
func (h *TicketHandler) enqueueTicketNotifications(ctx context.Context, old, updated *domain.Ticket) {
	orgID := updated.OrgID

	// Assignee changed → notify new assignee.
	if h.assigneeChanged(old, updated) && updated.AssigneeID != nil {
		assignee, err := h.users.GetByID(ctx, *updated.AssigneeID)
		if err == nil {
			pref, _ := h.notifPrefs.GetByUser(ctx, assignee.ID, orgID)
			if pref == nil || pref.EmailOnAssigned {
				h.emailQueue.Enqueue(domain.EmailJob{
					Kind:          domain.EmailEventAssigned,
					ToEmail:       assignee.Email,
					ToName:        assignee.Name,
					TicketID:      updated.ID.String(),
					TicketSubject: updated.Subject,
				})
			}
		}
	}

	// Status changed to resolved or closed → notify reporter.
	if updated.SubmittedByUserID != nil && h.statusChangedTo(old, updated, domain.TicketStatusResolved, domain.TicketStatusClosed) {
		reporter, err := h.users.GetByID(ctx, *updated.SubmittedByUserID)
		if err == nil {
			pref, _ := h.notifPrefs.GetByUser(ctx, reporter.ID, orgID)
			var kind domain.EmailEventKind
			if updated.Status == domain.TicketStatusResolved {
				if pref == nil || pref.EmailOnResolved {
					kind = domain.EmailEventResolved
				}
			} else {
				if pref == nil || pref.EmailOnClosed {
					kind = domain.EmailEventClosed
				}
			}
			if kind != "" {
				h.emailQueue.Enqueue(domain.EmailJob{
					Kind:          kind,
					ToEmail:       reporter.Email,
					ToName:        reporter.Name,
					TicketID:      updated.ID.String(),
					TicketSubject: updated.Subject,
				})
			}
		}
	}
}

func (h *TicketHandler) assigneeChanged(old, updated *domain.Ticket) bool {
	if old.AssigneeID == nil && updated.AssigneeID == nil {
		return false
	}
	if old.AssigneeID == nil || updated.AssigneeID == nil {
		return true
	}
	return *old.AssigneeID != *updated.AssigneeID
}

func (h *TicketHandler) statusChangedTo(old, updated *domain.Ticket, statuses ...domain.TicketStatus) bool {
	if old.Status == updated.Status {
		return false
	}
	for _, s := range statuses {
		if updated.Status == s {
			return true
		}
	}
	return false
}

// ---- Comments ----

func (h *TicketHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid ticket id")
		return
	}
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
	var authorID *uuid.UUID
	if claims, ok := middleware.ClaimsFromContext(r); ok {
		id := claims.UserID
		authorID = &id
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
