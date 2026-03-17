package auth_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"omnir/internal/auth"
)

func TestJWT_RoundTrip(t *testing.T) {
	svc := auth.NewJWTService("test-secret-key")
	userID := uuid.New()

	claims := auth.Claims{
		UserID:    userID,
		Role:      "admin",
		CompanyID: uuid.New(),
	}

	token, err := svc.Issue(claims, 15*time.Minute)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	parsed, err := svc.Verify(token)
	require.NoError(t, err)
	assert.Equal(t, userID, parsed.UserID)
	assert.Equal(t, "admin", parsed.Role)
}

func TestJWT_ExpiredToken(t *testing.T) {
	svc := auth.NewJWTService("test-secret-key")

	claims := auth.Claims{
		UserID: uuid.New(),
		Role:   "user",
	}

	// Issue a token that expires in the past
	token, err := svc.Issue(claims, -1*time.Second)
	require.NoError(t, err)

	_, err = svc.Verify(token)
	assert.ErrorIs(t, err, auth.ErrTokenExpired)
}

func TestJWT_InvalidSignature(t *testing.T) {
	svc1 := auth.NewJWTService("secret-one")
	svc2 := auth.NewJWTService("secret-two")

	claims := auth.Claims{UserID: uuid.New(), Role: "user"}
	token, err := svc1.Issue(claims, 15*time.Minute)
	require.NoError(t, err)

	_, err = svc2.Verify(token)
	assert.ErrorIs(t, err, auth.ErrInvalidToken)
}

func TestJWT_TamperedToken(t *testing.T) {
	svc := auth.NewJWTService("test-secret-key")
	claims := auth.Claims{UserID: uuid.New(), Role: "user"}

	token, err := svc.Issue(claims, 15*time.Minute)
	require.NoError(t, err)

	// Corrupt the token
	tampered := token[:len(token)-4] + "XXXX"
	_, err = svc.Verify(tampered)
	assert.Error(t, err)
}
