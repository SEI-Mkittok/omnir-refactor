package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

// OnboardingRepo is the Postgres implementation of repository.OnboardingRepository.
type OnboardingRepo struct {
	db *pgxpool.Pool
}

func NewOnboardingRepo(db *pgxpool.Pool) *OnboardingRepo {
	return &OnboardingRepo{db: db}
}

func (r *OnboardingRepo) GetOrCreate(ctx context.Context, orgID uuid.UUID) (*domain.OrgOnboarding, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO org_onboarding (org_id, completed_steps, created_at, updated_at)
		VALUES ($1, '[]', NOW(), NOW())
		ON CONFLICT (org_id) DO UPDATE SET org_id = EXCLUDED.org_id
		RETURNING org_id, completed_steps, step_status, dismissed, completed_at, created_at, updated_at`,
		orgID,
	)
	return scanOnboarding(row)
}

func (r *OnboardingRepo) UpdateSteps(ctx context.Context, orgID uuid.UUID, steps []string, completedAt *time.Time, dismissed *bool) (*domain.OrgOnboarding, error) {
	if steps == nil {
		steps = []string{}
	}
	row := r.db.QueryRow(ctx, `
		UPDATE org_onboarding
		SET completed_steps = $2, completed_at = $3, dismissed = COALESCE($4, dismissed), updated_at = NOW()
		WHERE org_id = $1
		RETURNING org_id, completed_steps, step_status, dismissed, completed_at, created_at, updated_at`,
		orgID, steps, completedAt, dismissed,
	)
	return scanOnboarding(row)
}

func (r *OnboardingRepo) UpdateStepStatus(ctx context.Context, orgID uuid.UUID, stepID string, status domain.OnboardingStepStatus) (*domain.OrgOnboarding, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE org_onboarding
		SET step_status = jsonb_set(COALESCE(step_status, '{}'), $2, $3, true), updated_at = NOW()
		WHERE org_id = $1
		RETURNING org_id, completed_steps, step_status, dismissed, completed_at, created_at, updated_at`,
		orgID,
		[]string{stepID},
		`"`+string(status)+`"`,
	)
	return scanOnboarding(row)
}

func (r *OnboardingRepo) CreateInvite(ctx context.Context, invite *domain.OrgInvite) (*domain.OrgInvite, error) {
	if invite.ID == uuid.Nil {
		invite.ID = uuid.New()
	}
	invite.CreatedAt = time.Now().UTC()

	row := r.db.QueryRow(ctx, `
		INSERT INTO org_invites (id, org_id, email, role, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, org_id, email, role, token, expires_at, accepted_at, created_at`,
		invite.ID, invite.OrgID, invite.Email, invite.Role,
		invite.Token, invite.ExpiresAt, invite.CreatedAt,
	)
	return scanInvite(row)
}

func (r *OnboardingRepo) GetInviteByToken(ctx context.Context, token string) (*domain.OrgInvite, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, org_id, email, role, token, expires_at, accepted_at, created_at
		FROM org_invites
		WHERE token = $1`,
		token,
	)
	inv, err := scanInvite(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return inv, nil
}

func (r *OnboardingRepo) AcceptInvite(ctx context.Context, inviteID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE org_invites SET accepted_at = NOW() WHERE id = $1`,
		inviteID,
	)
	return err
}

func (r *OnboardingRepo) ListInvites(ctx context.Context, orgID uuid.UUID) ([]*domain.OrgInvite, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, org_id, email, role, token, expires_at, accepted_at, created_at
		FROM org_invites
		WHERE org_id = $1
		ORDER BY created_at DESC`,
		orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []*domain.OrgInvite
	for rows.Next() {
		inv, err := scanInvite(rows)
		if err != nil {
			return nil, err
		}
		invites = append(invites, inv)
	}
	return invites, rows.Err()
}

func scanOnboarding(row pgx.Row) (*domain.OrgOnboarding, error) {
	var o domain.OrgOnboarding
	var rawStatus []byte
	err := row.Scan(
		&o.OrgID, &o.CompletedSteps, &rawStatus, &o.Dismissed, &o.CompletedAt,
		&o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if len(rawStatus) > 0 {
		if err := json.Unmarshal(rawStatus, &o.StepStatus); err != nil {
			return nil, err
		}
	}
	if o.StepStatus == nil {
		o.StepStatus = map[string]domain.OnboardingStepStatus{}
	}
	return &o, nil
}

func scanInvite(row pgx.Row) (*domain.OrgInvite, error) {
	var i domain.OrgInvite
	err := row.Scan(
		&i.ID, &i.OrgID, &i.Email, &i.Role, &i.Token,
		&i.ExpiresAt, &i.AcceptedAt, &i.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &i, nil
}
