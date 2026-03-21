package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/config"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/testutil/mocks"
)

// postSignup fires a POST /signup request against an OrgHandler router.
func postSignup(t *testing.T, h *handler.OrgHandler, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)
	return rec
}

// ---------------------------------------------------------------------------
// Gate: single-tenant mode must block signup
// ---------------------------------------------------------------------------

func TestOrgSignup_SingleMode_Returns404(t *testing.T) {
	orgRepo := new(mocks.MockOrgRepository)
	userRepo := new(mocks.MockUserRepository)
	h := handler.NewOrgHandler(orgRepo, userRepo, auth.NewJWTService("s"), config.OrgModeSingle)

	rec := postSignup(t, h, map[string]string{
		"orgName": "Acme", "adminName": "Alice",
		"email": "alice@acme.com", "password": "password123",
	})

	assert.Equal(t, http.StatusNotFound, rec.Code)
	// No repo calls expected.
	orgRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

// ---------------------------------------------------------------------------
// Validation: required fields / format
// ---------------------------------------------------------------------------

func TestOrgSignup_Validation(t *testing.T) {
	tests := []struct {
		name string
		body map[string]string
	}{
		{
			"missing orgName",
			map[string]string{"adminName": "Alice", "email": "a@a.com", "password": "password1"},
		},
		{
			"missing adminName",
			map[string]string{"orgName": "Acme", "email": "a@a.com", "password": "password1"},
		},
		{
			"missing email",
			map[string]string{"orgName": "Acme", "adminName": "Alice", "password": "password1"},
		},
		{
			"invalid email (no @)",
			map[string]string{"orgName": "Acme", "adminName": "Alice", "email": "notanemail", "password": "password1"},
		},
		{
			"password too short",
			map[string]string{"orgName": "Acme", "adminName": "Alice", "email": "a@a.com", "password": "short"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orgRepo := new(mocks.MockOrgRepository)
			userRepo := new(mocks.MockUserRepository)
			h := handler.NewOrgHandler(orgRepo, userRepo, auth.NewJWTService("s"), config.OrgModeSaaS)

			rec := postSignup(t, h, tc.body)
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		})
	}
}

func TestOrgSignup_InvalidJSON(t *testing.T) {
	orgRepo := new(mocks.MockOrgRepository)
	userRepo := new(mocks.MockUserRepository)
	h := handler.NewOrgHandler(orgRepo, userRepo, auth.NewJWTService("s"), config.OrgModeSaaS)

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ---------------------------------------------------------------------------
// Happy path: org + admin user created, JWTs issued as httpOnly cookies
// ---------------------------------------------------------------------------

func TestOrgSignup_SaaS_Success(t *testing.T) {
	orgRepo := new(mocks.MockOrgRepository)
	userRepo := new(mocks.MockUserRepository)
	h := handler.NewOrgHandler(orgRepo, userRepo, auth.NewJWTService("test-secret"), config.OrgModeSaaS)

	orgID := uuid.New()
	userID := uuid.New()

	orgRepo.On("SlugExists", mock.Anything, "acme-corp").Return(false, nil)
	orgRepo.On("Create", mock.Anything, mock.MatchedBy(func(o *domain.Organization) bool {
		return o.Name == "Acme Corp" && o.Plan == domain.OrgPlanStarter
	})).Return(&domain.Organization{
		ID: orgID, Name: "Acme Corp", Slug: "acme-corp", Plan: domain.OrgPlanStarter,
	}, nil)
	userRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == "alice@acme.com" && u.Role == domain.UserRoleAdmin
	}), mock.AnythingOfType("string")).Return(&domain.User{
		ID: userID, OrgID: orgID, Email: "alice@acme.com", Name: "Alice", Role: domain.UserRoleAdmin,
	}, nil)

	rec := postSignup(t, h, map[string]string{
		"orgName":   "Acme Corp",
		"adminName": "Alice",
		"email":     "alice@acme.com",
		"password":  "password123",
	})

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.NotNil(t, resp["org"], "response must include org")
	assert.NotNil(t, resp["user"], "response must include user")

	var hasAccess, hasRefresh bool
	for _, c := range rec.Result().Cookies() {
		switch c.Name {
		case "access_token":
			hasAccess = true
		case "refresh_token":
			hasRefresh = true
		}
	}
	assert.True(t, hasAccess, "access_token cookie must be set")
	assert.True(t, hasRefresh, "refresh_token cookie must be set")
}

// ---------------------------------------------------------------------------
// Slug collision: handler appends -2 suffix when base slug is taken
// ---------------------------------------------------------------------------

func TestOrgSignup_SaaS_SlugCollision(t *testing.T) {
	orgRepo := new(mocks.MockOrgRepository)
	userRepo := new(mocks.MockUserRepository)
	h := handler.NewOrgHandler(orgRepo, userRepo, auth.NewJWTService("s"), config.OrgModeSaaS)

	orgID := uuid.New()
	userID := uuid.New()

	// "acme" taken; "acme-2" free.
	orgRepo.On("SlugExists", mock.Anything, "acme").Return(true, nil)
	orgRepo.On("SlugExists", mock.Anything, "acme-2").Return(false, nil)
	orgRepo.On("Create", mock.Anything, mock.MatchedBy(func(o *domain.Organization) bool {
		return o.Slug == "acme-2"
	})).Return(&domain.Organization{
		ID: orgID, Name: "Acme", Slug: "acme-2", Plan: domain.OrgPlanStarter,
	}, nil)
	userRepo.On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("string")).Return(&domain.User{
		ID: userID, OrgID: orgID, Email: "bob@acme.com", Name: "Bob", Role: domain.UserRoleAdmin,
	}, nil)

	rec := postSignup(t, h, map[string]string{
		"orgName":   "Acme",
		"adminName": "Bob",
		"email":     "bob@acme.com",
		"password":  "password123",
	})

	require.Equal(t, http.StatusCreated, rec.Code)
	orgRepo.AssertCalled(t, "Create", mock.Anything, mock.MatchedBy(func(o *domain.Organization) bool {
		return o.Slug == "acme-2"
	}))
}

// ---------------------------------------------------------------------------
// Duplicate email → 409 Conflict
// ---------------------------------------------------------------------------

func TestOrgSignup_SaaS_DuplicateEmail(t *testing.T) {
	orgRepo := new(mocks.MockOrgRepository)
	userRepo := new(mocks.MockUserRepository)
	h := handler.NewOrgHandler(orgRepo, userRepo, auth.NewJWTService("s"), config.OrgModeSaaS)

	orgID := uuid.New()

	orgRepo.On("SlugExists", mock.Anything, "acme").Return(false, nil)
	orgRepo.On("Create", mock.Anything, mock.Anything).Return(&domain.Organization{
		ID: orgID, Name: "Acme", Slug: "acme", Plan: domain.OrgPlanStarter,
	}, nil)
	userRepo.On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("string")).
		Return(nil, &pgconn.PgError{Code: "23505"})

	rec := postSignup(t, h, map[string]string{
		"orgName":   "Acme",
		"adminName": "Alice",
		"email":     "alice@acme.com",
		"password":  "password123",
	})

	assert.Equal(t, http.StatusConflict, rec.Code)
}

// ---------------------------------------------------------------------------
// ORG_MODE=multitenant is a valid alias for saas
// ---------------------------------------------------------------------------

func TestOrgSignup_MultitenantMode_Accepted(t *testing.T) {
	orgRepo := new(mocks.MockOrgRepository)
	userRepo := new(mocks.MockUserRepository)
	h := handler.NewOrgHandler(orgRepo, userRepo, auth.NewJWTService("s"), config.OrgModeMultitenant)

	orgID := uuid.New()
	userID := uuid.New()

	orgRepo.On("SlugExists", mock.Anything, "test-org").Return(false, nil)
	orgRepo.On("Create", mock.Anything, mock.Anything).Return(&domain.Organization{
		ID: orgID, Name: "Test Org", Slug: "test-org", Plan: domain.OrgPlanStarter,
	}, nil)
	userRepo.On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("string")).Return(&domain.User{
		ID: userID, OrgID: orgID, Email: "admin@test.com", Name: "Admin", Role: domain.UserRoleAdmin,
	}, nil)

	rec := postSignup(t, h, map[string]string{
		"orgName":   "Test Org",
		"adminName": "Admin",
		"email":     "admin@test.com",
		"password":  "password123",
	})

	assert.Equal(t, http.StatusCreated, rec.Code)
}

// ---------------------------------------------------------------------------
// Slug generation: special characters are normalized to hyphens
// ---------------------------------------------------------------------------

func TestOrgSignup_SaaS_SlugNormalization(t *testing.T) {
	orgRepo := new(mocks.MockOrgRepository)
	userRepo := new(mocks.MockUserRepository)
	h := handler.NewOrgHandler(orgRepo, userRepo, auth.NewJWTService("s"), config.OrgModeSaaS)

	orgID := uuid.New()
	userID := uuid.New()

	// "My Amazing Corp!" → "my-amazing-corp"
	orgRepo.On("SlugExists", mock.Anything, "my-amazing-corp").Return(false, nil)
	orgRepo.On("Create", mock.Anything, mock.MatchedBy(func(o *domain.Organization) bool {
		return o.Slug == "my-amazing-corp"
	})).Return(&domain.Organization{
		ID: orgID, Name: "My Amazing Corp!", Slug: "my-amazing-corp", Plan: domain.OrgPlanStarter,
	}, nil)
	userRepo.On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("string")).Return(&domain.User{
		ID: userID, OrgID: orgID, Email: "ceo@amazing.com", Name: "CEO", Role: domain.UserRoleAdmin,
	}, nil)

	rec := postSignup(t, h, map[string]string{
		"orgName":   "My Amazing Corp!",
		"adminName": "CEO",
		"email":     "ceo@amazing.com",
		"password":  "password123",
	})

	require.Equal(t, http.StatusCreated, rec.Code)
	orgRepo.AssertCalled(t, "Create", mock.Anything, mock.MatchedBy(func(o *domain.Organization) bool {
		return o.Slug == "my-amazing-corp"
	}))
}
