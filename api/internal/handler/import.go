package handler

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

// ImportResult is the response body from a CSV import operation.
type ImportResult struct {
	Processed int           `json:"processed"`
	Created   int           `json:"created"`
	Updated   int           `json:"updated"`
	Failed    int           `json:"failed"`
	Errors    []ImportError `json:"errors,omitempty"`
}

// ImportError describes a single row-level failure.
type ImportError struct {
	Row   int    `json:"row"`
	Error string `json:"error"`
}

// ImportHandler handles bulk CSV import for contacts, accounts, and leads.
type ImportHandler struct {
	contacts repository.ContactRepository
	accounts repository.AccountRepository
	leads    repository.LeadRepository
}

func NewImportHandler(
	contacts repository.ContactRepository,
	accounts repository.AccountRepository,
	leads repository.LeadRepository,
) *ImportHandler {
	return &ImportHandler{contacts: contacts, accounts: accounts, leads: leads}
}

func (h *ImportHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequireRole(domain.UserRoleAdmin, domain.UserRoleAgent))
	r.Post("/contacts", h.ImportContacts)
	r.Post("/accounts", h.ImportAccounts)
	r.Post("/leads", h.ImportLeads)
	return r
}

// parseImportRequest parses a multipart form upload and returns a CSV reader
// plus an optional field mapping (csv header → domain field name).
// Max upload size is 100 MB.
func parseImportRequest(r *http.Request) (*csv.Reader, map[string]string, error) {
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		return nil, nil, err
	}

	f, _, err := r.FormFile("file")
	if err != nil {
		return nil, nil, err
	}

	var mapping map[string]string
	if raw := r.FormValue("mapping"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &mapping); err != nil {
			return nil, nil, err
		}
	}

	csvReader := csv.NewReader(f)
	csvReader.TrimLeadingSpace = true
	csvReader.LazyQuotes = true
	return csvReader, mapping, nil
}

// normalizeHeader converts a CSV column header to a canonical domain field name.
func normalizeHeader(h string) string {
	h = strings.ToLower(strings.TrimSpace(h))
	h = strings.ReplaceAll(h, " ", "_")
	h = strings.ReplaceAll(h, "-", "_")
	aliases := map[string]string{
		"email_address": "email",
		"e_mail":        "email",
		"mobile":        "phone",
		"mobile_phone":  "phone",
		"telephone":     "phone",
		"firstname":     "first_name",
		"givenname":     "first_name",
		"lastname":      "last_name",
		"surname":       "last_name",
		"familyname":    "last_name",
		"source":        "lead_source",
		"account":       "company",
		"organization":  "company",
		"company_name":  "company",
	}
	if canonical, ok := aliases[h]; ok {
		return canonical
	}
	return h
}

// buildIndex maps domain field names to CSV column indices.
// An explicit mapping takes priority over auto-detection.
func buildIndex(headers []string, mapping map[string]string) map[string]int {
	idx := make(map[string]int)
	for i, h := range headers {
		var field string
		if mapping != nil {
			if f, ok := mapping[h]; ok {
				field = f
			}
		}
		if field == "" {
			field = normalizeHeader(h)
		}
		idx[field] = i
	}
	return idx
}

func getField(row []string, idx map[string]int, field string) string {
	i, ok := idx[field]
	if !ok || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func importAccessContext(w http.ResponseWriter, r *http.Request) (*domain.AccessContext, bool) {
	access, ok := domain.AccessContextFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "access context required")
		return nil, false
	}
	return access, true
}

func deniedImportField(access *domain.AccessContext, module domain.ACLModule, fields ...string) string {
	if access != nil && domain.IsAdminRole(access.PlatformRole) {
		return ""
	}
	seen := map[string]bool{}
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" || seen[field] {
			continue
		}
		seen[field] = true
		if !access.CanWriteField(module, field) {
			return field
		}
	}
	return ""
}

func importFieldDeniedMessage(field string) string {
	return `field "` + field + `" is not writable`
}

// callerUserID returns the authenticated user's ID from the request context.
func callerUserID(r *http.Request) uuid.UUID {
	if claims, ok := middleware.ClaimsFromContext(r); ok {
		return claims.UserID
	}
	return uuid.Nil
}

// ImportContacts handles POST /api/v1/import/contacts.
// Upserts contacts by email (org-scoped); rows without email are always inserted.
func (h *ImportHandler) ImportContacts(w http.ResponseWriter, r *http.Request) {
	csvReader, mapping, err := parseImportRequest(r)
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "could not parse upload: "+err.Error())
		return
	}

	ownerID := callerUserID(r)

	headers, err := csvReader.Read()
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "could not read CSV headers")
		return
	}
	idx := buildIndex(headers, mapping)
	access, ok := importAccessContext(w, r)
	if !ok {
		return
	}
	if claims, hasClaims := middleware.ClaimsFromContext(r); hasClaims && access.PlatformRole == "" {
		access.PlatformRole = claims.Role
	}

	result := ImportResult{}
	rowNum := 1 // header is row 0; data starts at row 1

	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		rowNum++
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{Row: rowNum, Error: "CSV parse error: " + err.Error()})
			continue
		}
		result.Processed++

		firstName := getField(row, idx, "first_name")
		lastName := getField(row, idx, "last_name")
		if firstName == "" || lastName == "" {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{Row: rowNum, Error: "first_name and last_name are required"})
			continue
		}

		email := getField(row, idx, "email")
		phone := getField(row, idx, "phone")
		leadSource := getField(row, idx, "lead_source")
		fields := []string{"first_name", "last_name"}
		if email != "" {
			fields = append(fields, "email")
		}
		if phone != "" {
			fields = append(fields, "phone")
		}
		if leadSource != "" {
			fields = append(fields, "lead_source")
		}

		// Validate email format when provided.
		if email != "" && !strings.Contains(email, "@") {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{Row: rowNum, Error: "invalid email format"})
			continue
		}

		// Upsert by email when present.
		if email != "" {
			existing, lookupErr := h.contacts.GetByEmail(r.Context(), email)
			if lookupErr == nil && existing != nil {
				updateFields := []string{"first_name", "last_name"}
				if phone != "" {
					updateFields = append(updateFields, "phone")
				}
				if leadSource != "" {
					updateFields = append(updateFields, "lead_source")
				}
				if denied := deniedImportField(access, domain.ACLModuleContacts, updateFields...); denied != "" {
					result.Failed++
					result.Errors = append(result.Errors, ImportError{Row: rowNum, Error: importFieldDeniedMessage(denied)})
					continue
				}
				patch := domain.ContactPatch{
					FirstName:  &firstName,
					LastName:   &lastName,
					Phone:      strPtr(phone),
					LeadSource: strPtr(leadSource),
				}
				if _, updateErr := h.contacts.Update(r.Context(), existing.ID, patch); updateErr != nil {
					result.Failed++
					result.Errors = append(result.Errors, ImportError{Row: rowNum, Error: updateErr.Error()})
					continue
				}
				result.Updated++
				continue
			}
		}

		if stageStr := getField(row, idx, "stage"); stageStr != "" {
			s := domain.ContactStage(stageStr)
			if s.IsValid() {
				fields = append(fields, "stage")
			}
		}
		if denied := deniedImportField(access, domain.ACLModuleContacts, fields...); denied != "" {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{Row: rowNum, Error: importFieldDeniedMessage(denied)})
			continue
		}

		c := &domain.Contact{
			FirstName:  firstName,
			LastName:   lastName,
			Email:      strPtr(email),
			Phone:      strPtr(phone),
			OwnerID:    ownerID,
			Stage:      domain.ContactStageLead,
			LeadSource: strPtr(leadSource),
		}
		if stageStr := getField(row, idx, "stage"); stageStr != "" {
			s := domain.ContactStage(stageStr)
			if s.IsValid() {
				c.Stage = s
			}
		}

		if _, createErr := h.contacts.Create(r.Context(), c); createErr != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{Row: rowNum, Error: createErr.Error()})
			continue
		}
		result.Created++
	}

	status := http.StatusOK
	if result.Failed > 0 {
		status = http.StatusUnprocessableEntity
	}
	writeJSON(w, status, result)
}

// ImportAccounts handles POST /api/v1/import/accounts.
// Upserts accounts by name (org-scoped).
func (h *ImportHandler) ImportAccounts(w http.ResponseWriter, r *http.Request) {
	csvReader, mapping, err := parseImportRequest(r)
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "could not parse upload: "+err.Error())
		return
	}

	ownerID := callerUserID(r)

	headers, err := csvReader.Read()
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "could not read CSV headers")
		return
	}
	idx := buildIndex(headers, mapping)
	access, ok := importAccessContext(w, r)
	if !ok {
		return
	}
	if claims, hasClaims := middleware.ClaimsFromContext(r); hasClaims && access.PlatformRole == "" {
		access.PlatformRole = claims.Role
	}

	result := ImportResult{}
	rowNum := 1

	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		rowNum++
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{Row: rowNum, Error: "CSV parse error: " + err.Error()})
			continue
		}
		result.Processed++

		name := getField(row, idx, "name")
		if name == "" {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{Row: rowNum, Error: "name is required"})
			continue
		}

		domainValue := getField(row, idx, "domain")
		industry := getField(row, idx, "industry")
		sizeStr := getField(row, idx, "size")
		fields := []string{"name"}
		if domainValue != "" {
			fields = append(fields, "domain")
		}
		if industry != "" {
			fields = append(fields, "industry")
		}
		if sizeStr != "" {
			fields = append(fields, "size")
		}
		if denied := deniedImportField(access, domain.ACLModuleAccounts, fields...); denied != "" {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{Row: rowNum, Error: importFieldDeniedMessage(denied)})
			continue
		}

		// Upsert by name (exact match within org). Note: accounts with the same name but
		// different domains are treated as duplicates — intentional Phase 6 limitation.
		existing, lookupErr := h.accounts.GetByName(r.Context(), name)
		if lookupErr == nil && existing != nil {
			patch := domain.AccountPatch{
				Domain:   strPtr(domainValue),
				Industry: strPtr(industry),
			}
			if sizeStr != "" {
				s := domain.AccountSize(sizeStr)
				patch.Size = &s
			}
			if _, updateErr := h.accounts.Update(r.Context(), existing.ID, patch); updateErr != nil {
				result.Failed++
				result.Errors = append(result.Errors, ImportError{Row: rowNum, Error: updateErr.Error()})
				continue
			}
			result.Updated++
			continue
		}

		a := &domain.Account{
			Name:     name,
			Domain:   strPtr(domainValue),
			Industry: strPtr(industry),
			OwnerID:  ownerID,
		}
		if sizeStr != "" {
			s := domain.AccountSize(sizeStr)
			a.Size = &s
		}

		if _, createErr := h.accounts.Create(r.Context(), a); createErr != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{Row: rowNum, Error: createErr.Error()})
			continue
		}
		result.Created++
	}

	status := http.StatusOK
	if result.Failed > 0 {
		status = http.StatusUnprocessableEntity
	}
	writeJSON(w, status, result)
}

// ImportLeads handles POST /api/v1/import/leads.
// Creates a new lead record per row (no dedup).
func (h *ImportHandler) ImportLeads(w http.ResponseWriter, r *http.Request) {
	csvReader, mapping, err := parseImportRequest(r)
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "could not parse upload: "+err.Error())
		return
	}

	ownerID := callerUserID(r)

	headers, err := csvReader.Read()
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "could not read CSV headers")
		return
	}
	idx := buildIndex(headers, mapping)
	access, ok := importAccessContext(w, r)
	if !ok {
		return
	}
	if claims, hasClaims := middleware.ClaimsFromContext(r); hasClaims && access.PlatformRole == "" {
		access.PlatformRole = claims.Role
	}

	result := ImportResult{}
	rowNum := 1

	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		rowNum++
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{Row: rowNum, Error: "CSV parse error: " + err.Error()})
			continue
		}
		result.Processed++

		firstName := getField(row, idx, "first_name")
		lastName := getField(row, idx, "last_name")
		if firstName == "" || lastName == "" {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{Row: rowNum, Error: "first_name and last_name are required"})
			continue
		}

		email := getField(row, idx, "email")
		phone := getField(row, idx, "phone")
		company := getField(row, idx, "company")
		leadSource := getField(row, idx, "lead_source")
		fields := []string{"first_name", "last_name"}
		if email != "" {
			fields = append(fields, "email")
		}
		if phone != "" {
			fields = append(fields, "phone")
		}
		if company != "" {
			fields = append(fields, "company")
		}
		if leadSource != "" {
			fields = append(fields, "lead_source")
		}
		if denied := deniedImportField(access, domain.ACLModuleLeads, fields...); denied != "" {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{Row: rowNum, Error: importFieldDeniedMessage(denied)})
			continue
		}

		l := &domain.Lead{
			FirstName:  firstName,
			LastName:   lastName,
			Email:      strPtr(email),
			Phone:      strPtr(phone),
			Company:    strPtr(company),
			LeadSource: strPtr(leadSource),
			Status:     domain.LeadStatusNew,
			OwnerID:    &ownerID,
		}

		if _, createErr := h.leads.Create(r.Context(), l); createErr != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{Row: rowNum, Error: createErr.Error()})
			continue
		}
		result.Created++
	}

	status := http.StatusOK
	if result.Failed > 0 {
		status = http.StatusUnprocessableEntity
	}
	writeJSON(w, status, result)
}
