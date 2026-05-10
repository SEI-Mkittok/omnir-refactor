package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

// leadsStore is the set of operations the handler needs from the leads persistence layer.
type leadsStore interface {
	repository.LeadRepository
	ConvertToContact(ctx context.Context, leadID, contactID uuid.UUID) (*domain.Lead, error)
	DefaultPipelineID(ctx context.Context, orgID uuid.UUID) (uuid.UUID, error)
}

type LeadHandler struct {
	leads       leadsStore
	contacts    repository.ContactRepository
	accounts    repository.AccountRepository
	deals       repository.DealRepository
	mappingRepo repository.LeadConversionMappingRepository
	cfDefs      repository.CustomFieldDefinitionRepository
}

func NewLeadHandler(
	leads leadsStore,
	contacts repository.ContactRepository,
	accounts repository.AccountRepository,
	deals repository.DealRepository,
	mappingRepo repository.LeadConversionMappingRepository,
	cfDefs repository.CustomFieldDefinitionRepository,
) *LeadHandler {
	return &LeadHandler{
		leads:       leads,
		contacts:    contacts,
		accounts:    accounts,
		deals:       deals,
		mappingRepo: mappingRepo,
		cfDefs:      cfDefs,
	}
}

func (h *LeadHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/sources", h.ListSources)
	r.Get("/{id}", h.GetByID)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(domain.UserRoleAdmin, domain.UserRoleAgent))
		r.Post("/", h.Create)
		r.Patch("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
		r.Post("/{id}/convert", h.Convert)
	})
	return r
}

func (h *LeadHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	// Accept both canonical names and frontend aliases.
	searchQ := q.Get("q")
	if searchQ == "" {
		searchQ = q.Get("search")
	}
	filter := domain.LeadFilter{
		Q:     searchQ,
		Sort:  q.Get("sort_by"),
		Order: q.Get("sort_dir"),
	}
	// Also accept legacy sort/order params.
	if filter.Sort == "" {
		filter.Sort = q.Get("sort")
	}
	if filter.Order == "" {
		filter.Order = q.Get("order")
	}

	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.Page = n
		}
	}
	// Accept ?limit (canonical) or ?per_page (frontend alias).
	limitParam := q.Get("limit")
	if limitParam == "" {
		limitParam = q.Get("per_page")
	}
	if limitParam != "" {
		if n, err := strconv.Atoi(limitParam); err == nil && n <= 200 {
			filter.Limit = n
		}
	}
	if v := q.Get("owner_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.OwnerID = &id
		}
	}
	if v := q.Get("status"); v != "" {
		s := domain.LeadStatus(v)
		filter.Status = &s
	}
	if v := q.Get("source"); v != "" {
		filter.Source = &v
	}
	if v := q.Get("score_min"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.ScoreMin = &n
		}
	}
	if v := q.Get("score_max"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.ScoreMax = &n
		}
	}

	if filter.Limit == 0 {
		filter.Limit = 50
	}
	if filter.Page == 0 {
		filter.Page = 1
	}

	leads, total, err := h.leads.List(r.Context(), filter)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, paginated(leads, total, filter.Page, filter.Limit))
}

func (h *LeadHandler) ListSources(w http.ResponseWriter, r *http.Request) {
	sources, err := h.leads.ListSources(r.Context())
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	if sources == nil {
		sources = []string{}
	}
	writeJSON(w, http.StatusOK, sources)
}

func (h *LeadHandler) Create(w http.ResponseWriter, r *http.Request) {
	var l domain.Lead
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}

	created, err := h.leads.Create(r.Context(), &l)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *LeadHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	l, err := h.leads.GetByID(r.Context(), id)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, l)
}

func (h *LeadHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	var patch domain.LeadPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}
	// Auto-derive score from status when status changes and score is not explicitly set.
	if patch.Status != nil && patch.Score == nil {
		derived := domain.ScoreForStatus(*patch.Status)
		patch.Score = &derived
	}
	// Snap manually-provided scores to the nearest 10% step.
	if patch.Score != nil && patch.Status == nil {
		snapped := domain.SnapScoreToStep(*patch.Score)
		patch.Score = &snapped
	}
	l, err := h.leads.Update(r.Context(), id, patch)
	if err != nil {
		handleDomainErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, l)
}

func (h *LeadHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}
	if err := h.leads.Delete(r.Context(), id); err != nil {
		handleDomainErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Convert creates a Contact from the lead's data, marks the lead as converted,
// and returns the updated lead with converted_contact_id set.
func (h *LeadHandler) Convert(w http.ResponseWriter, r *http.Request) {
	leadID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid id")
		return
	}

	lead, err := h.leads.GetByID(r.Context(), leadID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}

	if lead.Status == domain.LeadStatusConverted {
		writeProblem(w, http.StatusConflict, "Conflict", "lead is already converted")
		return
	}

	// Derive owner_id: prefer authenticated user, fall back to lead's owner.
	ownerID := uuid.Nil
	if claims, ok := middleware.ClaimsFromContext(r); ok {
		ownerID = claims.UserID
	} else if lead.OwnerID != nil {
		ownerID = *lead.OwnerID
	}
	if ownerID == uuid.Nil {
		writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", "lead owner is required")
		return
	}

	orgID, hasOrg := domain.OrgIDFromContext(r.Context())
	if !hasOrg {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "missing org context")
		return
	}

	mappings := []*domain.LeadConversionMapping{}
	if h.mappingRepo != nil {
		rows, err := h.mappingRepo.List(r.Context(), orgID)
		if err != nil {
			handleDomainErr(w, err)
			return
		}
		mappings = rows
	}

	leadCustom := map[string]any{}
	if lead.CustomFields != nil && len(*lead.CustomFields) > 0 {
		_ = json.Unmarshal(*lead.CustomFields, &leadCustom)
	}

	resolveLeadValue := func(field string) any {
		if strings.HasPrefix(field, "custom:") {
			return leadCustom[strings.TrimPrefix(field, "custom:")]
		}
		switch field {
		case "first_name":
			return lead.FirstName
		case "last_name":
			return lead.LastName
		case "email":
			if lead.Email != nil {
				return *lead.Email
			}
			return nil
		case "phone":
			if lead.Phone != nil {
				return *lead.Phone
			}
			return nil
		case "company":
			if lead.Company != nil {
				return *lead.Company
			}
			return nil
		case "lead_source":
			if lead.LeadSource != nil {
				return *lead.LeadSource
			}
			return nil
		case "lead_score":
			return lead.Score
		case "status":
			return string(lead.Status)
		default:
			if val, ok := leadCustom[field]; ok {
				return val
			}
			return nil
		}
	}

	contactCustom := map[string]any{}
	accountCustom := map[string]any{}
	dealCustom := map[string]any{}
	var mappedAccountName *string
	var mappedDealTitle *string
	var mappedDealValueCents *int64
	var mappedDealCurrency *string

	for _, row := range mappings {
		if row == nil || !row.IsActive {
			continue
		}
		value := resolveLeadValue(row.LeadField)
		if value == nil {
			continue
		}

		target := strings.TrimSpace(row.TargetField)
		switch row.TargetEntity {
		case domain.LeadConversionTargetContact:
			switch target {
			case "first_name":
				if v, ok := value.(string); ok && strings.TrimSpace(v) != "" {
					lead.FirstName = strings.TrimSpace(v)
				}
			case "last_name":
				if v, ok := value.(string); ok && strings.TrimSpace(v) != "" {
					lead.LastName = strings.TrimSpace(v)
				}
			case "email":
				if v, ok := value.(string); ok && strings.TrimSpace(v) != "" {
					val := strings.TrimSpace(v)
					lead.Email = &val
				}
			case "phone":
				if v, ok := value.(string); ok {
					val := strings.TrimSpace(v)
					lead.Phone = &val
				}
			case "lead_source":
				if v, ok := value.(string); ok {
					val := strings.TrimSpace(v)
					lead.LeadSource = &val
				}
			default:
				customKey := strings.TrimSpace(strings.TrimPrefix(target, "custom:"))
				if customKey != "" {
					contactCustom[customKey] = value
				}
			}
		case domain.LeadConversionTargetAccount:
			switch target {
			case "name":
				if v, ok := value.(string); ok && strings.TrimSpace(v) != "" {
					val := strings.TrimSpace(v)
					mappedAccountName = &val
				}
			case "domain":
				if v, ok := value.(string); ok {
					trimmed := strings.TrimSpace(v)
					if trimmed != "" {
						accountCustom["domain"] = trimmed
					}
				}
			default:
				customKey := strings.TrimSpace(strings.TrimPrefix(target, "custom:"))
				if customKey != "" {
					accountCustom[customKey] = value
				}
			}
		case domain.LeadConversionTargetDeal:
			switch target {
			case "title":
				if v, ok := value.(string); ok && strings.TrimSpace(v) != "" {
					val := strings.TrimSpace(v)
					mappedDealTitle = &val
				}
			case "value_cents":
				switch typed := value.(type) {
				case float64:
					val := int64(typed)
					mappedDealValueCents = &val
				case int:
					val := int64(typed)
					mappedDealValueCents = &val
				case int64:
					val := typed
					mappedDealValueCents = &val
				}
			case "currency":
				if v, ok := value.(string); ok && strings.TrimSpace(v) != "" {
					val := strings.ToUpper(strings.TrimSpace(v))
					mappedDealCurrency = &val
				}
			default:
				customKey := strings.TrimSpace(strings.TrimPrefix(target, "custom:"))
				if customKey != "" {
					dealCustom[customKey] = value
				}
			}
		}
	}

	contact := &domain.Contact{
		FirstName:           lead.FirstName,
		LastName:            lead.LastName,
		Email:               lead.Email,
		Phone:               lead.Phone,
		OwnerID:             ownerID,
		Stage:               domain.ContactStageLead,
		LeadSource:          lead.LeadSource,
		ConvertedFromLeadID: &leadID,
	}
	if len(contactCustom) > 0 {
		if raw, err := json.Marshal(contactCustom); err == nil {
			contact.CustomFields = raw
		}
	}

	createdContact, err := h.contacts.Create(r.Context(), contact)
	if err != nil {
		handleDomainErr(w, err)
		return
	}

	accountName := "Converted Lead"
	if lead.Company != nil && strings.TrimSpace(*lead.Company) != "" {
		accountName = strings.TrimSpace(*lead.Company)
	}
	if mappedAccountName != nil && strings.TrimSpace(*mappedAccountName) != "" {
		accountName = strings.TrimSpace(*mappedAccountName)
	}
	account := &domain.Account{
		Name:    accountName,
		OwnerID: ownerID,
	}
	if domainValue, ok := accountCustom["domain"].(string); ok && domainValue != "" {
		account.Domain = &domainValue
		delete(accountCustom, "domain")
	}
	if len(accountCustom) > 0 {
		if raw, err := json.Marshal(accountCustom); err == nil {
			account.CustomFields = raw
		}
	}
	createdAccount, err := h.accounts.Create(r.Context(), account)
	if err != nil {
		handleDomainErr(w, err)
		return
	}

	pipelineID, err := h.leads.DefaultPipelineID(r.Context(), orgID)
	if err != nil {
		writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", "no pipeline available for lead conversion")
		return
	}

	dealTitle := "Converted Lead"
	if mappedDealTitle != nil && strings.TrimSpace(*mappedDealTitle) != "" {
		dealTitle = strings.TrimSpace(*mappedDealTitle)
	} else if lead.Company != nil && strings.TrimSpace(*lead.Company) != "" {
		dealTitle = strings.TrimSpace(*lead.Company) + " Opportunity"
	} else if lead.Email != nil {
		dealTitle = "Opportunity: " + strings.TrimSpace(*lead.Email)
	}
	deal := &domain.Deal{
		Title:      dealTitle,
		Stage:      domain.DealStageLead,
		Probability: 0,
		OwnerID:    ownerID,
		PipelineID: pipelineID,
		AccountID:  &createdAccount.ID,
		ContactID:  &createdContact.ID,
	}
	if mappedDealCurrency != nil {
		deal.Currency = *mappedDealCurrency
	}
	if mappedDealValueCents != nil {
		deal.ValueCents = *mappedDealValueCents
	}
	if len(dealCustom) > 0 {
		if raw, err := json.Marshal(dealCustom); err == nil {
			deal.CustomFields = raw
		}
	}
	createdDeal, err := h.deals.Create(r.Context(), deal)
	if err != nil {
		handleDomainErr(w, err)
		return
	}

	updatedLead, err := h.leads.ConvertToContact(r.Context(), leadID, createdContact.ID)
	if err != nil {
		handleDomainErr(w, err)
		return
	}

	// Return shape includes all conversion outputs in one deterministic flow.
	writeJSON(w, http.StatusOK, map[string]any{
		"contact": createdContact,
		"account": createdAccount,
		"deal":    createdDeal,
		"lead":    updatedLead,
	})
}
