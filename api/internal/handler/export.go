package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

const exportPageSize = 500
const exportMaxRecords = 50_000

// ExportHandler handles CSV export endpoints.
type ExportHandler struct {
	contacts repository.ContactRepository
	accounts repository.AccountRepository
	deals    repository.DealRepository
	reports  repository.ReportsRepository
	auditor  Auditor
}

func NewExportHandler(
	contacts repository.ContactRepository,
	accounts repository.AccountRepository,
	deals repository.DealRepository,
	reports repository.ReportsRepository,
) *ExportHandler {
	return &ExportHandler{
		contacts: contacts,
		accounts: accounts,
		deals:    deals,
		reports:  reports,
	}
}

func (h *ExportHandler) WithAuditLog(r repository.AuditLogRepository) *ExportHandler {
	h.auditor = newAuditor(r)
	return h
}

func (h *ExportHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/contacts", h.Contacts)
	r.Get("/accounts", h.Accounts)
	r.Get("/deals", h.Deals)
	r.Get("/reports", h.Reports)
	return r
}

// Contacts streams all contacts matching the filter as CSV.
func (h *ExportHandler) Contacts(w http.ResponseWriter, r *http.Request) {
	h.auditor.logExport(r, domain.AuditEntityContact)
	q := r.URL.Query()
	filter := domain.ContactFilter{
		Q:     q.Get("q"),
		Sort:  q.Get("sort"),
		Order: q.Get("order"),
		Limit: exportPageSize,
		Page:  1,
	}
	if v := q.Get("owner_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid owner_id: must be a UUID")
			return
		}
		filter.OwnerID = &id
	}
	if v := q.Get("stage"); v != "" {
		s := domain.ContactStage(v)
		filter.Stage = &s
	}

	setCsvHeaders(w, "contacts.csv")
	cw := csv.NewWriter(w)

	if err := cw.Write([]string{
		"id", "first_name", "last_name", "email", "phone",
		"account_id", "owner_id", "stage", "lead_source",
		"tags", "created_at", "updated_at",
	}); err != nil {
		return
	}

	total := 0
	for {
		rows, _, err := h.contacts.List(r.Context(), filter)
		if err != nil {
			_ = cw.Write([]string{"#error", err.Error(), "", "", "", "", "", "", "", "", "", ""})
			break
		}
		for _, c := range rows {
			if err := cw.Write([]string{
				c.ID.String(),
				c.FirstName,
				c.LastName,
				derefStr(c.Email),
				derefStr(c.Phone),
				optUUID(c.AccountID),
				c.OwnerID.String(),
				string(c.Stage),
				derefStr(c.LeadSource),
				strings.Join(c.Tags, ";"),
				c.CreatedAt.Format("2006-01-02T15:04:05Z"),
				c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			}); err != nil {
				cw.Flush()
				return
			}
		}
		total += len(rows)
		if len(rows) < exportPageSize || total >= exportMaxRecords {
			break
		}
		filter.Page++
	}
	cw.Flush()
}

// Accounts streams all accounts matching the filter as CSV.
func (h *ExportHandler) Accounts(w http.ResponseWriter, r *http.Request) {
	h.auditor.logExport(r, domain.AuditEntityAccount)
	q := r.URL.Query()
	filter := domain.AccountFilter{
		Q:     q.Get("q"),
		Sort:  q.Get("sort"),
		Order: q.Get("order"),
		Limit: exportPageSize,
		Page:  1,
	}
	if v := q.Get("owner_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid owner_id: must be a UUID")
			return
		}
		filter.OwnerID = &id
	}
	if v := q.Get("industry"); v != "" {
		filter.Industry = &v
	}

	setCsvHeaders(w, "accounts.csv")
	cw := csv.NewWriter(w)

	if err := cw.Write([]string{
		"id", "name", "domain", "industry", "size",
		"owner_id", "tags", "created_at", "updated_at",
	}); err != nil {
		return
	}

	total := 0
	for {
		rows, _, err := h.accounts.List(r.Context(), filter)
		if err != nil {
			_ = cw.Write([]string{"#error", err.Error(), "", "", "", "", "", "", ""})
			break
		}
		for _, a := range rows {
			size := ""
			if a.Size != nil {
				size = string(*a.Size)
			}
			if err := cw.Write([]string{
				a.ID.String(),
				a.Name,
				derefStr(a.Domain),
				derefStr(a.Industry),
				size,
				a.OwnerID.String(),
				strings.Join(a.Tags, ";"),
				a.CreatedAt.Format("2006-01-02T15:04:05Z"),
				a.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			}); err != nil {
				cw.Flush()
				return
			}
		}
		total += len(rows)
		if len(rows) < exportPageSize || total >= exportMaxRecords {
			break
		}
		filter.Page++
	}
	cw.Flush()
}

// Deals streams all deals matching the filter as CSV.
func (h *ExportHandler) Deals(w http.ResponseWriter, r *http.Request) {
	h.auditor.logExport(r, domain.AuditEntityDeal)
	q := r.URL.Query()
	filter := domain.DealFilter{
		Q:     q.Get("q"),
		Sort:  q.Get("sort"),
		Order: q.Get("order"),
		Limit: exportPageSize,
		Page:  1,
	}
	if v := q.Get("owner_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid owner_id: must be a UUID")
			return
		}
		filter.OwnerID = &id
	}
	if v := q.Get("stage"); v != "" {
		s := domain.DealStage(v)
		filter.Stage = &s
	}
	if v := q.Get("account_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid account_id: must be a UUID")
			return
		}
		filter.AccountID = &id
	}
	if v := q.Get("pipeline_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid pipeline_id: must be a UUID")
			return
		}
		filter.PipelineID = &id
	}

	setCsvHeaders(w, "deals.csv")
	cw := csv.NewWriter(w)

	if err := cw.Write([]string{
		"id", "title", "stage", "value_cents", "currency",
		"probability", "expected_close_date",
		"account_id", "owner_id", "pipeline_id",
		"created_at", "updated_at",
	}); err != nil {
		return
	}

	total := 0
	for {
		rows, _, err := h.deals.List(r.Context(), filter)
		if err != nil {
			_ = cw.Write([]string{"#error", err.Error(), "", "", "", "", "", "", "", "", "", ""})
			break
		}
		for _, d := range rows {
			closeDate := ""
			if d.ExpectedCloseDate != nil {
				closeDate = d.ExpectedCloseDate.Format("2006-01-02")
			}
			if err := cw.Write([]string{
				d.ID.String(),
				d.Title,
				string(d.Stage),
				strconv.FormatInt(d.ValueCents, 10),
				d.Currency,
				strconv.Itoa(d.Probability),
				closeDate,
				optUUID(d.AccountID),
				d.OwnerID.String(),
				d.PipelineID.String(),
				d.CreatedAt.Format("2006-01-02T15:04:05Z"),
				d.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			}); err != nil {
				cw.Flush()
				return
			}
		}
		total += len(rows)
		if len(rows) < exportPageSize || total >= exportMaxRecords {
			break
		}
		filter.Page++
	}
	cw.Flush()
}

// Reports exports the current report summary as CSV (deals by stage + contacts monthly + activities by type).
func (h *ExportHandler) Reports(w http.ResponseWriter, r *http.Request) {
	h.auditor.logExport(r, domain.AuditEntityView)
	ctx := r.Context()

	setCsvHeaders(w, "reports.csv")
	cw := csv.NewWriter(w)

	// Section: deals by stage
	if err := cw.Write([]string{"section", "stage", "count", "total_value_cents"}); err != nil {
		return
	}
	if rows, err := h.reports.DealsByStage(ctx); err == nil {
		for _, m := range rows {
			if err := cw.Write([]string{
				"deals_by_stage",
				string(m.Stage),
				strconv.Itoa(m.Count),
				fmt.Sprintf("%d", m.TotalValueCents),
			}); err != nil {
				cw.Flush()
				return
			}
		}
	}

	// Section: contacts monthly (last 12 months)
	if err := cw.Write([]string{"section", "month", "count", ""}); err != nil {
		cw.Flush()
		return
	}
	if rows, err := h.reports.ContactsMonthly(ctx); err == nil {
		for _, m := range rows {
			if err := cw.Write([]string{
				"contacts_monthly",
				m.Month,
				strconv.Itoa(m.Count),
				"",
			}); err != nil {
				cw.Flush()
				return
			}
		}
	}

	// Section: activities by type
	if err := cw.Write([]string{"section", "type", "count", ""}); err != nil {
		cw.Flush()
		return
	}
	if rows, err := h.reports.ActivitiesByType(ctx); err == nil {
		for _, m := range rows {
			if err := cw.Write([]string{
				"activities_by_type",
				string(m.Type),
				strconv.Itoa(m.Count),
				"",
			}); err != nil {
				cw.Flush()
				return
			}
		}
	}

	cw.Flush()
}

func setCsvHeaders(w http.ResponseWriter, filename string) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Transfer-Encoding", "chunked")
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func optUUID(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}
