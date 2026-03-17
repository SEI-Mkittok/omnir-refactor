package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrTokenExpired = errors.New("token expired")
	ErrInvalidToken = errors.New("invalid token")
)

// Claims holds the data embedded in a JWT.
type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	Role      string    `json:"role"`
	CompanyID uuid.UUID `json:"company_id,omitempty"`
}

type jwtClaims struct {
	jwt.RegisteredClaims
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
	CompanyID string `json:"company_id,omitempty"`
}

// JWTService issues and verifies JWT tokens.
type JWTService struct {
	secret []byte
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{secret: []byte(secret)}
}

// Issue creates a signed JWT with the given claims, valid for ttl duration.
func (s *JWTService) Issue(c Claims, ttl time.Duration) (string, error) {
	now := time.Now()
	tc := jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		UserID:    c.UserID.String(),
		Role:      c.Role,
		CompanyID: c.CompanyID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tc)
	return token.SignedString(s.secret)
}

// Verify parses and validates a JWT, returning the embedded Claims.
func (s *JWTService) Verify(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	tc, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(tc.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}
	companyID, _ := uuid.Parse(tc.CompanyID)

	return &Claims{
		UserID:    userID,
		Role:      tc.Role,
		CompanyID: companyID,
	}, nil
}
