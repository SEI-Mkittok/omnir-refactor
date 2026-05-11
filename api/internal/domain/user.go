package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	UserRoleSuperAdmin UserRole = "super_admin" // cross-org operator; can list orgs and switch context
	UserRoleAdmin      UserRole = "admin"
	UserRoleAgent      UserRole = "agent"
	UserRoleClient     UserRole = "client"
)

// IsAdminRole reports whether a role has workspace-admin privileges inside
// the active tenant. It intentionally does not make admin equivalent to
// super_admin for cross-org operations.
func IsAdminRole(role string) bool {
	return role == string(UserRoleAdmin) || role == string(UserRoleSuperAdmin)
}

// IsValidUserRole reports whether the provided role is one of the accepted
// persisted user role values.
func IsValidUserRole(role UserRole) bool {
	switch role {
	case UserRoleSuperAdmin, UserRoleAdmin, UserRoleAgent, UserRoleClient:
		return true
	default:
		return false
	}
}

type User struct {
	ID          uuid.UUID                        `json:"id"`
	OrgID       uuid.UUID                        `json:"org_id"`
	Email       string                           `json:"email"`
	Name        string                           `json:"name"`
	Role        UserRole                         `json:"role"`
	RoleID      *uuid.UUID                       `json:"role_id,omitempty"`
	ProfileID   *uuid.UUID                       `json:"profile_id,omitempty"`
	RoleName    *string                          `json:"role_name,omitempty"`
	ProfileName *string                          `json:"profile_name,omitempty"`
	Permissions map[ACLModule]map[ACLAction]bool `json:"permissions,omitempty"`
	AvatarURL   *string                          `json:"avatar_url,omitempty"`
	CreatedAt   time.Time                        `json:"created_at"`
	UpdatedAt   time.Time                        `json:"updated_at"`
	DeletedAt   *time.Time                       `json:"deleted_at,omitempty"`
}

// UserPatch holds optional fields for partial user updates.
type UserPatch struct {
	Name      *string    `json:"name,omitempty"`
	Email     *string    `json:"email,omitempty"`
	Role      *UserRole  `json:"role,omitempty"`
	RoleID    *uuid.UUID `json:"role_id,omitempty"`
	ProfileID *uuid.UUID `json:"profile_id,omitempty"`
}

// UserFilter holds query parameters for listing users.
type UserFilter struct {
	OrgID uuid.UUID
	Q     string
	Role  *UserRole
	Page  int
	Limit int
	Sort  string
	Order string
}
