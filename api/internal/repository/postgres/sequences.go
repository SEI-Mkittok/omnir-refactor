package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

// SequenceRepo implements repository.SequenceRepository.
type SequenceRepo struct {
	db *pgxpool.Pool
}

func NewSequenceRepo(db *pgxpool.Pool) *SequenceRepo {
	return &SequenceRepo{db: db}
}

// ---- helpers ----

func scanStep(row pgx.Row) (*domain.SequenceStep, error) {
	var s domain.SequenceStep
	err := row.Scan(
		&s.ID, &s.SequenceID, &s.OrgID, &s.Position, &s.Kind,
		&s.Subject, &s.Body, &s.WaitDurationHours,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *SequenceRepo) listSteps(ctx context.Context, sequenceID uuid.UUID) ([]domain.SequenceStep, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, sequence_id, org_id, position, kind,
		       COALESCE(subject,''), COALESCE(body,''), wait_duration_hours,
		       created_at, updated_at
		FROM sequence_steps
		WHERE sequence_id = $1
		ORDER BY position ASC
	`, sequenceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []domain.SequenceStep
	for rows.Next() {
		s, err := scanStep(rows)
		if err != nil {
			return nil, err
		}
		steps = append(steps, *s)
	}
	return steps, rows.Err()
}

func (r *SequenceRepo) replaceSteps(ctx context.Context, tx pgx.Tx, sequenceID, orgID uuid.UUID, stepReqs []domain.CreateStepRequest) error {
	if _, err := tx.Exec(ctx, `DELETE FROM sequence_steps WHERE sequence_id = $1`, sequenceID); err != nil {
		return err
	}
	now := time.Now().UTC()
	for _, sr := range stepReqs {
		id := uuid.New()
		if _, err := tx.Exec(ctx, `
			INSERT INTO sequence_steps
				(id, sequence_id, org_id, position, kind, subject, body, wait_duration_hours, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		`, id, sequenceID, orgID, sr.Position, sr.Kind, sr.Subject, sr.Body, sr.WaitDurationHours, now, now); err != nil {
			return err
		}
	}
	return nil
}

// ---- CreateSequence ----

func (r *SequenceRepo) CreateSequence(ctx context.Context, s *domain.EmailSequence, steps []domain.CreateStepRequest) (*domain.EmailSequence, error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		s.OrgID = orgID
	}
	now := time.Now().UTC()
	s.CreatedAt = now
	s.UpdatedAt = now
	if s.Status == "" {
		s.Status = domain.SequenceStatusDraft
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `
		INSERT INTO email_sequences (id, org_id, name, description, status, created_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, s.ID, s.OrgID, s.Name, s.Description, s.Status, s.CreatedBy, s.CreatedAt, s.UpdatedAt); err != nil {
		return nil, err
	}

	if len(steps) > 0 {
		if err := r.replaceSteps(ctx, tx, s.ID, s.OrgID, steps); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetSequence(ctx, s.ID)
}

// ---- GetSequence ----

func (r *SequenceRepo) GetSequence(ctx context.Context, id uuid.UUID) (*domain.EmailSequence, error) {
	q := `
		SELECT id, org_id, name, COALESCE(description,''), status, created_by, created_at, updated_at
		FROM email_sequences
		WHERE id = $1 AND deleted_at IS NULL
	`
	args := []any{id}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id = $2`
		args = append(args, orgID)
	}

	var s domain.EmailSequence
	err := r.db.QueryRow(ctx, q, args...).Scan(
		&s.ID, &s.OrgID, &s.Name, &s.Description, &s.Status, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	steps, err := r.listSteps(ctx, s.ID)
	if err != nil {
		return nil, err
	}
	s.Steps = steps

	return &s, nil
}

// ---- ListSequences ----

func (r *SequenceRepo) ListSequences(ctx context.Context, filter domain.SequenceFilter) ([]*domain.EmailSequence, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	offset := 0
	if filter.Page > 1 {
		offset = (filter.Page - 1) * filter.Limit
	}

	where := `WHERE es.deleted_at IS NULL`
	args := []any{}
	idx := 1

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		where += ` AND es.org_id = $` + itoa(idx)
		args = append(args, orgID)
		idx++
	}
	if filter.Status != nil {
		where += ` AND es.status = $` + itoa(idx)
		args = append(args, *filter.Status)
		idx++
	}

	// count
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM email_sequences es `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, filter.Limit, offset)
	rows, err := r.db.Query(ctx, `
		SELECT
			es.id, es.org_id, es.name, COALESCE(es.description,''), es.status, es.created_by, es.created_at, es.updated_at,
			COUNT(DISTINCT se.id) FILTER (WHERE se.status IN ('active','completed')) AS enrolled_count,
			COALESCE(
				COUNT(ev.id) FILTER (WHERE ev.kind = 'opened')::float /
				NULLIF(COUNT(ev.id) FILTER (WHERE ev.kind = 'sent'), 0),
				0
			) AS open_rate
		FROM email_sequences es
		LEFT JOIN sequence_enrollments se ON se.sequence_id = es.id
		LEFT JOIN sequence_events ev ON ev.sequence_id = es.id
		`+where+`
		GROUP BY es.id
		ORDER BY es.created_at DESC
		LIMIT $`+itoa(idx)+` OFFSET $`+itoa(idx+1),
		args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var seqs []*domain.EmailSequence
	for rows.Next() {
		var s domain.EmailSequence
		if err := rows.Scan(
			&s.ID, &s.OrgID, &s.Name, &s.Description, &s.Status, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
			&s.EnrolledCount, &s.OpenRate,
		); err != nil {
			return nil, 0, err
		}
		seqs = append(seqs, &s)
	}
	return seqs, total, rows.Err()
}

// ---- UpdateSequence ----

func (r *SequenceRepo) UpdateSequence(ctx context.Context, id uuid.UUID, req domain.UpdateSequenceRequest) (*domain.EmailSequence, error) {
	existing, err := r.GetSequence(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}
	existing.UpdatedAt = time.Now().UTC()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `
		UPDATE email_sequences
		SET name=$1, description=$2, status=$3, updated_at=$4
		WHERE id=$5
	`, existing.Name, existing.Description, existing.Status, existing.UpdatedAt, id); err != nil {
		return nil, err
	}

	if req.Steps != nil {
		if err := r.replaceSteps(ctx, tx, id, existing.OrgID, req.Steps); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetSequence(ctx, id)
}

// ---- DeleteSequence (soft) ----

func (r *SequenceRepo) DeleteSequence(ctx context.Context, id uuid.UUID) error {
	q := `UPDATE email_sequences SET deleted_at=$1, updated_at=$1 WHERE id=$2 AND deleted_at IS NULL`
	args := []any{time.Now().UTC(), id}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id = $3`
		args = append(args, orgID)
	}

	tag, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ---- Enroll ----

func (r *SequenceRepo) Enroll(ctx context.Context, sequenceID uuid.UUID, req domain.EnrollRequest) (int, error) {
	orgID, _ := domain.OrgIDFromContext(ctx)
	now := time.Now().UTC()
	enrolled := 0

	for _, contactID := range req.ContactIDs {
		id := uuid.New()
		tag, err := r.db.Exec(ctx, `
			INSERT INTO sequence_enrollments (id, sequence_id, contact_id, org_id, status, current_step, enrolled_at, next_step_at)
			VALUES ($1,$2,$3,$4,'active',0,$5,$5)
			ON CONFLICT (sequence_id, contact_id) DO NOTHING
		`, id, sequenceID, contactID, orgID, now)
		if err != nil {
			return enrolled, err
		}
		if tag.RowsAffected() > 0 {
			enrolled++
		}
	}
	return enrolled, nil
}

// ---- ListEnrollments ----

func (r *SequenceRepo) ListEnrollments(ctx context.Context, sequenceID uuid.UUID) ([]*domain.SequenceEnrollment, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			se.id, se.sequence_id, se.contact_id, se.org_id, se.status,
			se.current_step, se.enrolled_at, se.completed_at,
			COALESCE(c.first_name||' '||c.last_name, '') AS contact_name,
			COALESCE(c.email, '') AS contact_email
		FROM sequence_enrollments se
		LEFT JOIN contacts c ON c.id = se.contact_id
		WHERE se.sequence_id = $1
		ORDER BY se.enrolled_at DESC
	`, sequenceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enrollments []*domain.SequenceEnrollment
	for rows.Next() {
		var e domain.SequenceEnrollment
		if err := rows.Scan(
			&e.ID, &e.SequenceID, &e.ContactID, &e.OrgID, &e.Status,
			&e.CurrentStep, &e.EnrolledAt, &e.CompletedAt,
			&e.ContactName, &e.ContactEmail,
		); err != nil {
			return nil, err
		}
		enrollments = append(enrollments, &e)
	}
	return enrollments, rows.Err()
}

// ---- UpdateEnrollmentStatus ----

func (r *SequenceRepo) UpdateEnrollmentStatus(ctx context.Context, enrollmentID uuid.UUID, status domain.EnrollmentStatus) error {
	var completedAt *time.Time
	if status == domain.EnrollmentStatusCompleted {
		now := time.Now().UTC()
		completedAt = &now
	}
	_, err := r.db.Exec(ctx, `
		UPDATE sequence_enrollments
		SET status=$1, completed_at=$2
		WHERE id=$3
	`, status, completedAt, enrollmentID)
	return err
}

// ---- GetAnalytics ----

func (r *SequenceRepo) GetAnalytics(ctx context.Context, sequenceID uuid.UUID) (*domain.SequenceAnalytics, error) {
	// Sequence-level aggregates
	var a domain.SequenceAnalytics
	a.SequenceID = sequenceID

	row := r.db.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE kind = 'sent')        AS sent,
			COUNT(*) FILTER (WHERE kind = 'opened')      AS opened,
			COUNT(*) FILTER (WHERE kind = 'clicked')     AS clicked,
			COUNT(*) FILTER (WHERE kind = 'completed')   AS completed,
			COUNT(*) FILTER (WHERE kind = 'bounced')     AS bounced,
			COUNT(*) FILTER (WHERE kind = 'unsubscribed') AS unsubscribed
		FROM sequence_events
		WHERE sequence_id = $1
	`, sequenceID)
	if err := row.Scan(&a.Sent, &a.Opened, &a.Clicked, &a.Completed, &a.Bounced, &a.Unsubscribed); err != nil {
		return nil, err
	}
	if a.Sent > 0 {
		a.OpenRate = float64(a.Opened) / float64(a.Sent)
		a.ClickRate = float64(a.Clicked) / float64(a.Sent)
	}

	// Per-step aggregates
	stepRows, err := r.db.Query(ctx, `
		SELECT
			ss.id, ss.position,
			COUNT(ev.id) FILTER (WHERE ev.kind = 'sent')   AS sent,
			COUNT(ev.id) FILTER (WHERE ev.kind = 'opened') AS opened,
			COUNT(ev.id) FILTER (WHERE ev.kind = 'clicked') AS clicked
		FROM sequence_steps ss
		LEFT JOIN sequence_events ev ON ev.step_id = ss.id
		WHERE ss.sequence_id = $1
		GROUP BY ss.id, ss.position
		ORDER BY ss.position ASC
	`, sequenceID)
	if err != nil {
		return nil, err
	}
	defer stepRows.Close()

	for stepRows.Next() {
		var s domain.StepAnalytics
		if err := stepRows.Scan(&s.StepID, &s.Position, &s.Sent, &s.Opened, &s.Clicked); err != nil {
			return nil, err
		}
		a.Steps = append(a.Steps, s)
	}
	return &a, stepRows.Err()
}

// ---- PendingEnrollments ----

// PendingEnrollments returns all active enrollments whose next_step_at is due.
// This is used by the background worker; no org scoping is applied.
// It joins contacts to include the contact email for sending.
func (r *SequenceRepo) PendingEnrollments(ctx context.Context) ([]*domain.SequenceEnrollment, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			se.id, se.sequence_id, se.contact_id, se.org_id, se.status,
			se.current_step, se.enrolled_at, se.completed_at,
			COALESCE(c.first_name||' '||c.last_name, '') AS contact_name,
			COALESCE(c.email, '') AS contact_email
		FROM sequence_enrollments se
		LEFT JOIN contacts c ON c.id = se.contact_id
		WHERE se.status = 'active'
		  AND se.next_step_at IS NOT NULL
		  AND se.next_step_at <= NOW()
		ORDER BY se.next_step_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enrollments []*domain.SequenceEnrollment
	for rows.Next() {
		var e domain.SequenceEnrollment
		if err := rows.Scan(
			&e.ID, &e.SequenceID, &e.ContactID, &e.OrgID, &e.Status,
			&e.CurrentStep, &e.EnrolledAt, &e.CompletedAt,
			&e.ContactName, &e.ContactEmail,
		); err != nil {
			return nil, err
		}
		enrollments = append(enrollments, &e)
	}
	return enrollments, rows.Err()
}

// ---- AdvanceEnrollment ----

// AdvanceEnrollment moves an enrollment to the next step and schedules it.
func (r *SequenceRepo) AdvanceEnrollment(ctx context.Context, id uuid.UUID, nextStep int, nextStepAt *time.Time) error {
	_, err := r.db.Exec(ctx, `
		UPDATE sequence_enrollments
		SET current_step=$1, next_step_at=$2
		WHERE id=$3
	`, nextStep, nextStepAt, id)
	return err
}

// ---- CompleteEnrollment ----

// CompleteEnrollment marks an enrollment as completed.
func (r *SequenceRepo) CompleteEnrollment(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(ctx, `
		UPDATE sequence_enrollments
		SET status='completed', completed_at=$1, next_step_at=NULL
		WHERE id=$2
	`, now, id)
	return err
}

// ---- RecordEvent ----

// RecordEvent inserts a new sequence event record.
func (r *SequenceRepo) RecordEvent(ctx context.Context, e *domain.SequenceEvent) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	e.OccurredAt = time.Now().UTC()
	_, err := r.db.Exec(ctx, `
		INSERT INTO sequence_events
			(id, sequence_id, step_id, enrollment_id, contact_id, org_id, kind, occurred_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, e.ID, e.SequenceID, e.StepID, e.EnrollmentID, e.ContactID, e.OrgID, e.Kind, e.OccurredAt)
	return err
}
