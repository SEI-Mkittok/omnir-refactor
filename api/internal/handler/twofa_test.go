package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
)

type fakeTOTPRepo struct {
	enabled bool
	userID  uuid.UUID
}

func (f *fakeTOTPRepo) SetSecret(context.Context, uuid.UUID, string) error { return nil }
func (f *fakeTOTPRepo) GetSecret(context.Context, uuid.UUID) (string, error) {
	return "", nil
}
func (f *fakeTOTPRepo) Activate(context.Context, uuid.UUID, []string) error { return nil }
func (f *fakeTOTPRepo) Disable(context.Context, uuid.UUID) error            { return nil }
func (f *fakeTOTPRepo) IsEnabled(_ context.Context, userID uuid.UUID) (bool, error) {
	f.userID = userID
	return f.enabled, nil
}
func (f *fakeTOTPRepo) FindUnusedBackupCode(context.Context, uuid.UUID) ([]*domain.TOTPBackupCode, error) {
	return nil, nil
}
func (f *fakeTOTPRepo) MarkBackupCodeUsed(context.Context, uuid.UUID) error { return nil }

func TestTwoFAHandler_StatusReturnsEnabledState(t *testing.T) {
	userID := uuid.New()
	totp := &fakeTOTPRepo{enabled: true}
	h := handler.NewTwoFAHandler(totp, nil, nil, "").Router()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = withClaims(req, &auth.Claims{UserID: userID, OrgID: uuid.New(), Role: string(domain.UserRoleAdmin)})
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, userID, totp.userID)

	var body map[string]bool
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	require.True(t, body["enabled"])
}
