package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

type OperationsFinanceHandler struct {
	repo repository.OperationsFinanceRepository
}

func NewOperationsFinanceHandler(repo repository.OperationsFinanceRepository) *OperationsFinanceHandler {
	return &OperationsFinanceHandler{repo: repo}
}

func (h *OperationsFinanceHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/service-contracts", h.ListServiceContracts)
	r.Post("/service-contracts", h.CreateServiceContract)
	r.Get("/assets", h.ListAssets)
	r.Post("/assets", h.CreateAsset)
	r.Get("/projects", h.ListProjects)
	r.Post("/projects", h.CreateProject)
	r.Get("/projects/{projectId}/tasks", h.ListProjectTasks)
	r.Post("/project-tasks", h.CreateProjectTask)
	r.Get("/time-entries", h.ListTimeEntries)
	r.Post("/time-entries", h.CreateTimeEntry)
	r.Get("/expenses", h.ListExpenses)
	r.Post("/expenses", h.CreateExpense)
	r.Post("/invoices", h.CreateInvoice)
	r.Get("/invoices/{id}", h.GetInvoice)
	r.Post("/invoices/{id}/line-items", h.CreateInvoiceLineItem)
	r.Post("/payments", h.CreatePayment)
	r.Post("/invoices/assemble", h.BuildInvoice)
	return r
}

func currentUserID(r *http.Request) uuid.UUID {
	if claims, ok := middleware.ClaimsFromContext(r); ok {
		return claims.UserID
	}
	return uuid.Nil
}

func (h *OperationsFinanceHandler) CreateServiceContract(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}
	userID := currentUserID(r)
	var in domain.ServiceContract
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	in.OrgID = orgID
	in.CreatedBy = userID
	if in.OwnerID == uuid.Nil {
		in.OwnerID = userID
	}
	out, err := h.repo.CreateServiceContract(r.Context(), &in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, out)
}
func (h *OperationsFinanceHandler) ListServiceContracts(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}
	out, err := h.repo.ListServiceContracts(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": out})
}
func (h *OperationsFinanceHandler) CreateAsset(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}
	userID := currentUserID(r)
	var in domain.Asset
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	in.OrgID = orgID
	in.CreatedBy = userID
	if in.OwnerID == uuid.Nil {
		in.OwnerID = userID
	}
	out, err := h.repo.CreateAsset(r.Context(), &in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, out)
}
func (h *OperationsFinanceHandler) ListAssets(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}
	out, err := h.repo.ListAssets(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": out})
}
func (h *OperationsFinanceHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}
	userID := currentUserID(r)
	var in domain.Project
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	in.OrgID = orgID
	in.CreatedBy = userID
	if in.OwnerID == uuid.Nil {
		in.OwnerID = userID
	}
	out, err := h.repo.CreateProject(r.Context(), &in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, out)
}
func (h *OperationsFinanceHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}
	out, err := h.repo.ListProjects(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": out})
}
func (h *OperationsFinanceHandler) CreateProjectTask(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}
	userID := currentUserID(r)
	var in domain.ProjectTask
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	in.OrgID = orgID
	in.CreatedBy = userID
	if in.OwnerID == uuid.Nil {
		in.OwnerID = userID
	}
	out, err := h.repo.CreateProjectTask(r.Context(), &in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, out)
}
func (h *OperationsFinanceHandler) ListProjectTasks(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}
	pid, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	out, err := h.repo.ListProjectTasks(r.Context(), orgID, pid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": out})
}
func (h *OperationsFinanceHandler) CreateTimeEntry(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}
	userID := currentUserID(r)
	var in domain.TimeEntry
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	in.OrgID = orgID
	in.CreatedBy = userID
	if in.OwnerID == uuid.Nil {
		in.OwnerID = userID
	}
	out, err := h.repo.CreateTimeEntry(r.Context(), &in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, out)
}
func (h *OperationsFinanceHandler) ListTimeEntries(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}
	aid, err := uuid.Parse(r.URL.Query().Get("account_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "account_id query param required")
		return
	}
	out, err := h.repo.ListTimeEntries(r.Context(), orgID, aid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": out})
}
func (h *OperationsFinanceHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}
	userID := currentUserID(r)
	var in domain.Expense
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	in.OrgID = orgID
	in.CreatedBy = userID
	if in.OwnerID == uuid.Nil {
		in.OwnerID = userID
	}
	out, err := h.repo.CreateExpense(r.Context(), &in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, out)
}
func (h *OperationsFinanceHandler) ListExpenses(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}
	aid, err := uuid.Parse(r.URL.Query().Get("account_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "account_id query param required")
		return
	}
	out, err := h.repo.ListExpenses(r.Context(), orgID, aid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": out})
}

func (h *OperationsFinanceHandler) BuildInvoice(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}
	userID := currentUserID(r)
	var req struct {
		AccountID  uuid.UUID  `json:"account_id"`
		ContactID  *uuid.UUID `json:"contact_id"`
		ContractID *uuid.UUID `json:"contract_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	asm, err := h.repo.BuildInvoiceAssembly(r.Context(), orgID, req.AccountID, req.ContactID, req.ContractID, userID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, asm)
}

func (h *OperationsFinanceHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}
	userID := currentUserID(r)
	var in domain.OpsInvoice
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	in.OrgID = orgID
	in.CreatedBy = userID
	if in.OwnerID == uuid.Nil {
		in.OwnerID = userID
	}
	out, err := h.repo.CreateInvoice(r.Context(), &in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, out)
}
func (h *OperationsFinanceHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}
	out, err := h.repo.GetInvoiceAssembly(r.Context(), orgID, id)
	if err != nil {
		if err == domain.ErrNotFound {
			writeError(w, http.StatusNotFound, "invoice not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}
func (h *OperationsFinanceHandler) CreateInvoiceLineItem(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}
	userID := currentUserID(r)
	invoiceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}
	var in domain.InvoiceLineItem
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	in.OrgID = orgID
	in.InvoiceID = invoiceID
	in.CreatedBy = userID
	if in.OwnerID == uuid.Nil {
		in.OwnerID = userID
	}
	out, err := h.repo.CreateInvoiceLineItem(r.Context(), &in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, out)
}
func (h *OperationsFinanceHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	orgID, ok := domain.OrgIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "org context required")
		return
	}
	userID := currentUserID(r)
	var in domain.Payment
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	in.OrgID = orgID
	in.CreatedBy = userID
	if in.OwnerID == uuid.Nil {
		in.OwnerID = userID
	}
	out, err := h.repo.CreatePayment(r.Context(), &in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, out)
}
