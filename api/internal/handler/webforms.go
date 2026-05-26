package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

type WebformHandler struct {
	repo     repository.WebformRepository
	leads    repository.LeadRepository
	contacts repository.ContactRepository
	tickets  repository.TicketRepository
}

func NewWebformHandler(
	repo repository.WebformRepository,
	leads repository.LeadRepository,
	contacts repository.ContactRepository,
	tickets repository.TicketRepository,
) *WebformHandler {
	return &WebformHandler{repo: repo, leads: leads, contacts: contacts, tickets: tickets}
}

func (h *WebformHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.Get)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/preview", h.Preview)
	return r
}

func (h *WebformHandler) PublicRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/{publicId}", h.PublicGet)
	r.Post("/{publicId}/submissions", h.PublicSubmit)
	return r
}

func (h *WebformHandler) List(w http.ResponseWriter, r *http.Request) {
	if !hasModuleAdminAccess(r, domain.ACLModuleSettings) {
		writeError(w, http.StatusForbidden, "admin access required")
		return
	}
	filter := domain.WebformFilter{Page: parseIntDefault(r.URL.Query().Get("page"), 1), Limit: parseIntDefault(r.URL.Query().Get("limit"), 50), Q: r.URL.Query().Get("q")}
	if v := r.URL.Query().Get("status"); v != "" {
		status := domain.WebformStatus(v)
		filter.Status = &status
	}
	if v := r.URL.Query().Get("target_module"); v != "" {
		module := domain.WebformTargetModule(v)
		filter.TargetModule = &module
	}
	if v := r.URL.Query().Get("campaign_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.CampaignID = &id
		}
	}
	items, total, err := h.repo.ListWebforms(r.Context(), filter)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, paginated(items, total, filter.Page, filter.Limit))
}

func (h *WebformHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !hasModuleAdminAccess(r, domain.ACLModuleSettings) {
		writeError(w, http.StatusForbidden, "admin access required")
		return
	}
	var form domain.Webform
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if claims, ok := middleware.ClaimsFromContext(r); ok {
		form.CreatedBy = &claims.UserID
	}
	created, err := h.repo.CreateWebform(r.Context(), &form)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *WebformHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !hasModuleAdminAccess(r, domain.ACLModuleSettings) {
		writeError(w, http.StatusForbidden, "admin access required")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	form, err := h.repo.GetWebform(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, form)
}

func (h *WebformHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !hasModuleAdminAccess(r, domain.ACLModuleSettings) {
		writeError(w, http.StatusForbidden, "admin access required")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var patch domain.WebformPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	updated, err := h.repo.UpdateWebform(r.Context(), id, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *WebformHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !hasModuleAdminAccess(r, domain.ACLModuleSettings) {
		writeError(w, http.StatusForbidden, "admin access required")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.repo.DeleteWebform(r.Context(), id); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WebformHandler) Preview(w http.ResponseWriter, r *http.Request) {
	if !hasModuleAdminAccess(r, domain.ACLModuleSettings) {
		writeError(w, http.StatusForbidden, "admin access required")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	form, err := h.repo.GetWebform(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	var req domain.WebformPreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	writeJSON(w, http.StatusOK, previewWebform(form, req.Payload))
}

func (h *WebformHandler) PublicGet(w http.ResponseWriter, r *http.Request) {
	form, err := h.repo.GetWebformByPublicID(r.Context(), chi.URLParam(r, "publicId"))
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if form.Status != domain.WebformStatusActive {
		writeError(w, http.StatusNotFound, "webform not found")
		return
	}
	writeJSON(w, http.StatusOK, form)
}

func (h *WebformHandler) PublicSubmit(w http.ResponseWriter, r *http.Request) {
	form, err := h.repo.GetWebformByPublicID(r.Context(), chi.URLParam(r, "publicId"))
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if form.Status != domain.WebformStatusActive {
		writeError(w, http.StatusNotFound, "webform not found")
		return
	}

	var req domain.PublicWebformSubmissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Payload == nil {
		req.Payload = map[string]interface{}{}
	}
	if trap := strings.TrimSpace(fmt.Sprintf("%v", req.Payload[form.SpamTrapField])); trap != "" && trap != "<nil>" {
		writeJSON(w, http.StatusOK, domain.WebformSubmissionResult{
			Success:        true,
			ReturnURL:      form.ReturnURL,
			SuccessMessage: form.SuccessMessage,
		})
		return
	}
	preview := previewWebform(form, req.Payload)
	if len(preview.MissingFields) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, preview)
		return
	}

	publicCtx := domain.WithOrgID(r.Context(), form.OrgID)
	recordType, recordID, err := h.createRecord(publicCtx, form, preview.MappedFields, req.Payload)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	payload, _ := json.Marshal(req.Payload)
	submission := &domain.WebformSubmission{
		OrgID:             form.OrgID,
		WebformID:         form.ID,
		CampaignID:        form.CampaignID,
		TargetModule:      form.TargetModule,
		Payload:           payload,
		CreatedRecordType: recordType,
		CreatedRecordID:   recordID,
		IPAddress:         requestIP(r),
		UserAgent:         stringPtr(strings.TrimSpace(r.UserAgent())),
	}
	if submission.UserAgent != nil && *submission.UserAgent == "" {
		submission.UserAgent = nil
	}
	created, err := h.repo.CreateWebformSubmission(publicCtx, submission)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, domain.WebformSubmissionResult{
		Success:           true,
		ReturnURL:         form.ReturnURL,
		SuccessMessage:    form.SuccessMessage,
		SubmissionID:      &created.ID,
		CreatedRecordType: recordType,
		CreatedRecordID:   recordID,
	})
}

func previewWebform(form *domain.Webform, payload map[string]interface{}) domain.WebformPreviewResponse {
	mapped := map[string]interface{}{}
	missing := []string{}
	for _, field := range form.Fields {
		val, ok := payload[field.Key]
		if field.Required && (!ok || strings.TrimSpace(fmt.Sprintf("%v", val)) == "" || fmt.Sprintf("%v", val) == "<nil>") {
			missing = append(missing, field.Key)
			continue
		}
		if !ok {
			continue
		}
		target := field.TargetField
		if target == "" {
			target = field.Key
		}
		mapped[target] = val
	}
	return domain.WebformPreviewResponse{
		TargetModule:  form.TargetModule,
		MappedFields:  mapped,
		MissingFields: missing,
	}
}

func (h *WebformHandler) createRecord(ctx context.Context, form *domain.Webform, mapped map[string]interface{}, payload map[string]interface{}) (*string, *uuid.UUID, error) {
	switch form.TargetModule {
	case domain.WebformTargetLead:
		owner := parseUUIDPtr(mappedString(mapped, "owner_id"))
		lead := &domain.Lead{
			FirstName:  defaultString(mappedString(mapped, "first_name"), firstNameFallback(mapped)),
			LastName:   defaultString(mappedString(mapped, "last_name"), lastNameFallback(mapped)),
			Email:      stringPtrNonEmpty(mappedString(mapped, "email")),
			Phone:      stringPtrNonEmpty(mappedString(mapped, "phone")),
			Company:    stringPtrNonEmpty(mappedString(mapped, "company")),
			LeadSource: stringPtr("webform"),
			Status:     domain.LeadStatusNew,
			OwnerID:    owner,
		}
		created, err := h.leads.Create(ctx, lead)
		if err != nil {
			return nil, nil, err
		}
		t := "lead"
		return &t, &created.ID, nil
	case domain.WebformTargetContact:
		ownerID := uuid.Nil
		if owner := parseUUIDPtr(mappedString(mapped, "owner_id")); owner != nil {
			ownerID = *owner
		} else if form.CreatedBy != nil {
			ownerID = *form.CreatedBy
		}
		contact := &domain.Contact{
			FirstName:  defaultString(mappedString(mapped, "first_name"), firstNameFallback(mapped)),
			LastName:   defaultString(mappedString(mapped, "last_name"), lastNameFallback(mapped)),
			Email:      stringPtrNonEmpty(mappedString(mapped, "email")),
			Phone:      stringPtrNonEmpty(mappedString(mapped, "phone")),
			OwnerID:    ownerID,
			LeadSource: stringPtr("webform"),
			Stage:      domain.ContactStageLead,
		}
		if err := contact.Validate(); err != nil {
			return nil, nil, err
		}
		created, err := h.contacts.Create(ctx, contact)
		if err != nil {
			return nil, nil, err
		}
		t := "contact"
		return &t, &created.ID, nil
	case domain.WebformTargetTicket:
		description := mappedString(mapped, "description")
		if description == "" {
			description = compactPayload(payload)
		}
		source := "webform"
		ticket := &domain.Ticket{
			Subject:     defaultString(mappedString(mapped, "subject"), "Webform submission: "+form.Name),
			Description: stringPtrNonEmpty(description),
			Priority:    domain.TicketPriorityMedium,
			Status:      domain.TicketStatusOpen,
			Source:      &source,
		}
		if contactID := parseUUIDPtr(mappedString(mapped, "contact_id")); contactID != nil {
			ticket.ContactID = contactID
		}
		if accountID := parseUUIDPtr(mappedString(mapped, "account_id")); accountID != nil {
			ticket.AccountID = accountID
		}
		created, err := h.tickets.Create(ctx, ticket)
		if err != nil {
			return nil, nil, err
		}
		t := "ticket"
		return &t, &created.ID, nil
	default:
		return nil, nil, fmt.Errorf("%w: unsupported target_module", domain.ErrValidation)
	}
}

func mappedString(mapped map[string]interface{}, key string) string {
	v, ok := mapped[key]
	if !ok || v == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprintf("%v", v))
}

func firstNameFallback(mapped map[string]interface{}) string {
	name := defaultString(mappedString(mapped, "name"), mappedString(mapped, "full_name"))
	parts := strings.Fields(name)
	if len(parts) > 0 {
		return parts[0]
	}
	email := mappedString(mapped, "email")
	if at := strings.Index(email, "@"); at > 0 {
		return email[:at]
	}
	return "Webform"
}

func lastNameFallback(mapped map[string]interface{}) string {
	name := defaultString(mappedString(mapped, "name"), mappedString(mapped, "full_name"))
	parts := strings.Fields(name)
	if len(parts) > 1 {
		return strings.Join(parts[1:], " ")
	}
	return "Submission"
}

func defaultString(v string, fallback string) string {
	if strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return fallback
}

func stringPtr(v string) *string {
	return &v
}

func stringPtrNonEmpty(v string) *string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return &v
}

func parseUUIDPtr(v string) *uuid.UUID {
	if v == "" {
		return nil
	}
	id, err := uuid.Parse(v)
	if err != nil {
		return nil
	}
	return &id
}

func requestIP(r *http.Request) *string {
	host := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if host != "" {
		if idx := strings.Index(host, ","); idx >= 0 {
			host = host[:idx]
		}
		return stringPtr(strings.TrimSpace(host))
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || host == "" {
		host = r.RemoteAddr
	}
	if strings.TrimSpace(host) == "" {
		return nil
	}
	return stringPtr(host)
}

func compactPayload(payload map[string]interface{}) string {
	b, err := json.Marshal(payload)
	if err != nil {
		return "Webform submission"
	}
	return string(b)
}

func parseIntDefault(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	var out int
	if _, err := fmt.Sscanf(raw, "%d", &out); err != nil || out <= 0 {
		return fallback
	}
	return out
}
