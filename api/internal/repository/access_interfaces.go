package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
)

type AccessRepository interface {
	RecordAccessRepository

	ResolveAccess(ctx context.Context, userID, orgID uuid.UUID, platformRole string) (*domain.AccessContext, error)

	ListRoles(ctx context.Context, orgID uuid.UUID) ([]*domain.ACLRole, error)
	CreateRole(ctx context.Context, role *domain.ACLRole) (*domain.ACLRole, error)
	UpdateRole(ctx context.Context, id uuid.UUID, patch domain.ACLRolePatch) (*domain.ACLRole, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error
	MoveRole(ctx context.Context, id uuid.UUID, parentID *uuid.UUID) (*domain.ACLRole, error)

	ListProfiles(ctx context.Context, orgID uuid.UUID) ([]*domain.ACLProfile, error)
	CreateProfile(ctx context.Context, profile *domain.ACLProfile) (*domain.ACLProfile, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, patch domain.ACLProfilePatch) (*domain.ACLProfile, error)
	DeleteProfile(ctx context.Context, id uuid.UUID) error
	ListProfilePermissions(ctx context.Context, profileID uuid.UUID) ([]domain.ACLProfilePermission, []domain.ACLProfileFieldPermission, error)
	ReplaceProfilePermissions(ctx context.Context, profileID uuid.UUID, permissions []domain.ACLProfilePermission, fieldPermissions []domain.ACLProfileFieldPermission) error

	ListGroups(ctx context.Context, orgID uuid.UUID) ([]*domain.ACLGroup, error)
	CreateGroup(ctx context.Context, group *domain.ACLGroup) (*domain.ACLGroup, error)
	UpdateGroup(ctx context.Context, id uuid.UUID, patch domain.ACLGroupPatch) (*domain.ACLGroup, error)
	DeleteGroup(ctx context.Context, id uuid.UUID) error
	ReplaceGroupMembers(ctx context.Context, groupID uuid.UUID, userIDs []uuid.UUID) error

	GetSharingRules(ctx context.Context, orgID uuid.UUID) (*domain.ACLSharingRules, error)
	ReplaceSharingRules(ctx context.Context, orgID uuid.UUID, rules *domain.ACLSharingRules) (*domain.ACLSharingRules, error)
}

type RecordAccessRepository interface {
	CanAccessRecord(ctx context.Context, module domain.ACLModule, id uuid.UUID, access domain.SharingAccessLevel) (bool, error)
}
