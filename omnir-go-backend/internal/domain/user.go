package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	UserRoleAdmin  UserRole = "admin"
	UserRoleUser   UserRole = "user"
	UserRoleViewer UserRole = "viewer"
)

type User struct {
	ID        uuid.UUID  `json:"id"`
	OrgID     uuid.UUID  `json:"org_id"`
	Email     string     `json:"email"`
	Name      string     `json:"name"`
	Role      UserRole   `json:"role"`
	AvatarURL *string    `json:"avatar_url,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
