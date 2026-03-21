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

// PortalHandler serves the /portal/* client-facing endpoints.
// All routes require role=client (enforced via RequireRole middleware).
type PortalHandler struct {
	tickets  repository.TicketRepository
	comments repository.TicketCommentRepository
}

func NewPortalHandler(
	tickets repository.TicketRepository,
	comments repository.TicketCommentRepository,
) *PortalHandler {
	return &PortalHandler{tickets: tickets, comments: comments}
}

func (h *PortalHandler) Router() chi.Router {
	r := chi.NewRouter()

	// All portal routes require client role.
	clientOnly := middleware.RequireRole(domain.UserRoleClient)
	r.Use(clientOnly)

	r.Get("/tickets", h.ListTickets)
	r.Post("/tickets", h.CreateTicket)
	r.Get("/tickets/{id}", h.GetTicket)
	r.Post("/tickets/{id}/comments", h.CreateComment)

	return r
}

// CreateTicket handles POST /portal/tickets.
// Creates a ticket owned by the authenticated client user.
func (h *PortalHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Subject     string  `json:"subject"`
		Description *string `json:"description,omitempty"`
		Priority    string  `json:"priority,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if req.Subject == "" {
		writeError(w, http.StatusUnprocessableEntity, "subject is required")
		return
	}

	t := &domain.Ticket{
		Subject:           req.Subject,
		Description:       req.Description,
		SubmittedByUserID: &claims.UserID,
	}
	if req.Priority != "" {
		p := domain.TicketPriority(req.Priority)
		if !p.IsValid() {
			writeError(w, http.StatusUnprocessableEntity, "invalid priority: must be one of low, medium, high, critical")
			return
		}
		t.Priority = p
	}

	created, err := h.tickets.Create(r.Context(), t)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// ListTickets handles GET /portal/tickets.
// Returns only tickets submitted by the authenticated client user.
func (h *PortalHandler) ListTickets(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	q := r.URL.Query()
	filter := domain.TicketFilter{
		SubmittedByUserID: &claims.UserID,
		Q:                 q.Get("search"),
		Sort:              q.Get("sort_by"),
		Order:             q.Get("sort_dir"),
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

// portalTicketDetail is the response shape for GET /portal/tickets/{id}.
// It includes the ticket and all public (non-internal) comments.
type portalTicketDetail struct {
	*domain.Ticket
	Comments []*domain.TicketComment `json:"comments"`
}

// GetTicket handles GET /portal/tickets/{id}.
// Returns the ticket with its public comments if submitted by the authenticated client user.
func (h *PortalHandler) GetTicket(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

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

	// Enforce ownership: client can only view their own tickets.
	if t.SubmittedByUserID == nil || *t.SubmittedByUserID != claims.UserID {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	// Fetch public (non-internal) comments for this ticket.
	isInternal := false
	comments, err := h.comments.List(r.Context(), domain.TicketCommentFilter{
		TicketID:   id,
		IsInternal: &isInternal,
	})
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if comments == nil {
		comments = []*domain.TicketComment{}
	}

	writeJSON(w, http.StatusOK, portalTicketDetail{Ticket: t, Comments: comments})
}

// CreateComment handles POST /portal/tickets/{id}/comments.
// Adds a public comment to a ticket owned by the authenticated client user.
func (h *PortalHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid ticket id")
		return
	}

	// Verify ticket ownership before allowing the comment.
	t, err := h.tickets.GetByID(r.Context(), ticketID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if t.SubmittedByUserID == nil || *t.SubmittedByUserID != claims.UserID {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	var req struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	if req.Body == "" {
		writeError(w, http.StatusUnprocessableEntity, "body is required")
		return
	}

	c := &domain.TicketComment{
		TicketID:   ticketID,
		AuthorID:   &claims.UserID,
		Body:       req.Body,
		IsInternal: false, // clients cannot post internal notes
	}
	created, err := h.comments.Create(r.Context(), c)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}
