package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/middleware"
)

// cfdMock is an inline testify mock for CustomFieldDefinitionRepository.
type cfdMock struct {
	mock.Mock
}

func (m *cfdMock) Create(ctx context.Context, def *domain.CustomFieldDefinition) (*domain.CustomFieldDefinition, error) {
	args := m.Called(ctx, def)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CustomFieldDefinition), args.Error(1)
}

func (m *cfdMock) GetByID(ctx context.Context, id uuid.UUID) (*domain.CustomFieldDefinition, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CustomFieldDefinition), args.Error(1)
}

func (m *cfdMock) Update(ctx context.Context, id uuid.UUID, patch domain.CustomFieldDefinitionPatch) (*domain.CustomFieldDefinition, error) {
	args := m.Called(ctx, id, patch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CustomFieldDefinition), args.Error(1)
}

func (m *cfdMock) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *cfdMock) List(ctx context.Context, filter domain.CustomFieldDefinitionFilter) ([]*domain.CustomFieldDefinition, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.CustomFieldDefinition), args.Error(1)
}

func cfdAdminReq(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	r := httptest.NewRequest(method, path, &buf)
	r.Header.Set("Content-Type", "application/json")
	orgID := uuid.New()
	r = r.WithContext(middleware.WithClaims(r.Context(), &auth.Claims{
		UserID: uuid.New(), OrgID: orgID, Role: "admin",
	}))
	return r
}

func cfdRouter(h *handler.CustomFieldHandler) chi.Router {
	r := chi.NewRouter()
	r.Mount("/custom-fields", h.Router())
	return r
}

func TestCustomFieldHandler_Create_Valid(t *testing.T) {
	repo := &cfdMock{}
	h := handler.NewCustomFieldHandler(repo)

	def := &domain.CustomFieldDefinition{
		ID:         uuid.New(),
		EntityType: domain.CustomFieldEntityTicket,
		Name:       "account_type",
		Label:      "Account Type",
		FieldType:  domain.CustomFieldTypeSelect,
		Options:    []string{"enterprise", "smb", "startup"},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	repo.On("List", mock.Anything, mock.Anything).Return([]*domain.CustomFieldDefinition{}, nil)
	repo.On("Create", mock.Anything, mock.Anything).Return(def, nil)

	req := cfdAdminReq(t, http.MethodPost, "/custom-fields", map[string]any{
		"entity_type": "ticket",
		"name":        "account_type",
		"label":       "Account Type",
		"field_type":  "select",
		"options":     []string{"enterprise", "smb", "startup"},
	})
	w := httptest.NewRecorder()
	cfdRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var got domain.CustomFieldDefinition
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "account_type", got.Name)
}

func TestCustomFieldHandler_Create_MissingSelectOptions(t *testing.T) {
	h := handler.NewCustomFieldHandler(&cfdMock{})

	req := cfdAdminReq(t, http.MethodPost, "/custom-fields", map[string]any{
		"entity_type": "ticket",
		"name":        "tier",
		"label":       "Tier",
		"field_type":  "select",
	})
	w := httptest.NewRecorder()
	cfdRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestCustomFieldHandler_Create_InvalidEntityType(t *testing.T) {
	h := handler.NewCustomFieldHandler(&cfdMock{})

	req := cfdAdminReq(t, http.MethodPost, "/custom-fields", map[string]any{
		"entity_type": "invoice",
		"name":        "tier",
		"label":       "Tier",
		"field_type":  "text",
	})
	w := httptest.NewRecorder()
	cfdRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestCustomFieldHandler_Create_InvalidName(t *testing.T) {
	cases := []string{"TitleCase", "123start", "has-hyphen", "has space", ""}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			h := handler.NewCustomFieldHandler(&cfdMock{})
			req := cfdAdminReq(t, http.MethodPost, "/custom-fields", map[string]any{
				"entity_type": "contact",
				"name":        name,
				"label":       "Label",
				"field_type":  "text",
			})
			w := httptest.NewRecorder()
			cfdRouter(h).ServeHTTP(w, req)
			assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "name=%q should be rejected", name)
		})
	}
}

func TestCustomFieldHandler_Create_MaxFieldsExceeded(t *testing.T) {
	repo := &cfdMock{}
	h := handler.NewCustomFieldHandler(repo)

	existing := make([]*domain.CustomFieldDefinition, 50)
	for i := range existing {
		existing[i] = &domain.CustomFieldDefinition{ID: uuid.New()}
	}
	repo.On("List", mock.Anything, mock.Anything).Return(existing, nil)

	req := cfdAdminReq(t, http.MethodPost, "/custom-fields", map[string]any{
		"entity_type": "contact",
		"name":        "overflow_field",
		"label":       "Overflow",
		"field_type":  "text",
	})
	w := httptest.NewRecorder()
	cfdRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestCustomFieldHandler_List(t *testing.T) {
	repo := &cfdMock{}
	h := handler.NewCustomFieldHandler(repo)

	defs := []*domain.CustomFieldDefinition{
		{ID: uuid.New(), Name: "tier", EntityType: domain.CustomFieldEntityContact},
	}
	repo.On("List", mock.Anything, mock.Anything).Return(defs, nil)

	req := cfdAdminReq(t, http.MethodGet, "/custom-fields", nil)
	w := httptest.NewRecorder()
	cfdRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var got []*domain.CustomFieldDefinition
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Len(t, got, 1)
}

func TestCustomFieldHandler_GetByID(t *testing.T) {
	repo := &cfdMock{}
	h := handler.NewCustomFieldHandler(repo)
	id := uuid.New()

	def := &domain.CustomFieldDefinition{ID: id, Name: "region"}
	repo.On("GetByID", mock.Anything, id).Return(def, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id.String())
	req := cfdAdminReq(t, http.MethodGet, fmt.Sprintf("/custom-fields/%s", id), nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Get("/custom-fields/{id}", h.GetByID)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var got domain.CustomFieldDefinition
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, id, got.ID)
}

func TestCustomFieldHandler_Delete(t *testing.T) {
	repo := &cfdMock{}
	h := handler.NewCustomFieldHandler(repo)
	id := uuid.New()

	repo.On("Delete", mock.Anything, id).Return(nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id.String())
	req := cfdAdminReq(t, http.MethodDelete, fmt.Sprintf("/custom-fields/%s", id), nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Delete("/custom-fields/{id}", h.Delete)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}
