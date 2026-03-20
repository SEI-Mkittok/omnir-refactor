// Package seqtoken creates and verifies HMAC-SHA256 signed tokens used for
// email sequence tracking (open pixel, click redirect, unsubscribe).
//
// Token format (URL-safe, no padding):
//
//	base64url(json_payload) + "." + base64url(hmac_sha256_of_first_part)
package seqtoken

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Claims holds the data encoded in a sequence tracking token.
type Claims struct {
	OrgID        uuid.UUID `json:"o"`
	SequenceID   uuid.UUID `json:"s"`
	EnrollmentID uuid.UUID `json:"e"`
	StepID       uuid.UUID `json:"t"`
	ExpiresAt    int64     `json:"x"` // unix seconds
}

// TokenTTL is the validity window for tracking tokens.
const TokenTTL = 90 * 24 * time.Hour

var (
	ErrInvalidToken = errors.New("seqtoken: invalid token")
	ErrExpiredToken = errors.New("seqtoken: token expired")
)

func enc(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func dec(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

// Issue signs and encodes a Claims struct into a URL-safe token string.
func Issue(secret string, c Claims) (string, error) {
	if c.ExpiresAt == 0 {
		c.ExpiresAt = time.Now().Add(TokenTTL).Unix()
	}
	payload, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	encoded := enc(payload)
	mac := sign(secret, encoded)
	return encoded + "." + enc(mac), nil
}

// Verify parses and validates a token. Returns ErrExpiredToken when past expiry.
func Verify(secret, token string) (*Claims, error) {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return nil, ErrInvalidToken
	}
	payload, err := dec(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}
	sig, err := dec(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}
	expected := sign(secret, parts[0])
	if !hmac.Equal(sig, expected) {
		return nil, ErrInvalidToken
	}
	var c Claims
	if err := json.Unmarshal(payload, &c); err != nil {
		return nil, ErrInvalidToken
	}
	if time.Now().Unix() > c.ExpiresAt {
		return nil, ErrExpiredToken
	}
	return &c, nil
}

func sign(secret, data string) []byte {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return h.Sum(nil)
}
