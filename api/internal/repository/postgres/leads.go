package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

type LeadRepo struct {
	db *pgxpool.Pool
}

func NewLeadRepo(db *pgxpool.Pool) *LeadRepo {
	return &LeadRepo{db: db}
}

const leadCols = `
	id, org_id, first_name, last_name, email, phone,
	company, lead_source, lead_score, status, owner_id,
	converted_contact_id, custom_fields, vtiger_legacy_id,
	created_at, updated_at, deleted_at
`

func scanLead(row pgx.Row) (*domain.Lead, error) {
	var l domain.Lead
	var customFields []byte
	err := row.Scan(
		&l.ID, &l.OrgID, &l.FirstName, &l.LastName, &l.Email, &l.Phone,
		&l.Company, &l.LeadSource, &l.Score, &l.Status, &l.OwnerID,
		&l.ConvertedContactID, &customFields, &l.VtigerLegacyID,
		&l.CreatedAt, &l.UpdatedAt, &l.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if customFields != nil {
		l.CustomFields = &customFields
	}
	return &l, nil
}

func (r *LeadRepo) Create(ctx context.Context, l *domain.Lead) (*domain.Lead, error) {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		l.OrgID = orgID
	}
	now := time.Now().UTC()
	l.CreatedAt = now
	l.UpdatedAt = now

	if l.Status == "" {
		l.Status = domain.LeadStatusNew
	}

	var customFields *[]byte
	if l.CustomFields != nil {
		customFields = l.CustomFields
	}

	row := r.db.QueryRow(ctx, `
		INSERT INTO leads
			(id, org_id, first_name, last_name, email, phone,
			 company, lead_source, lead_score, status, owner_id,
			 converted_contact_id, custom_fields, vtiger_legacy_id,
			 created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		RETURNING `+leadCols,
		l.ID, l.OrgID, l.FirstName, l.LastName, l.Email, l.Phone,
		l.Company, l.LeadSource, l.Score, l.Status, l.OwnerID,
		l.ConvertedContactID, customFields, l.VtigerLegacyID,
		l.CreatedAt, l.UpdatedAt,
	)
	return scanLead(row)
}

func (r *LeadRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Lead, error) {
	q := `SELECT ` + leadCols + ` FROM leads WHERE id=$1 AND deleted_at IS NULL`
	args := []any{id}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}

	return scanLead(r.db.QueryRow(ctx, q, args...))
}

func (r *LeadRepo) Update(ctx context.Context, id uuid.UUID, patch domain.LeadPatch) (*domain.Lead, error) {
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	i := 1

	addArg := func(col string, val any) {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}

	if patch.FirstName != nil {
		addArg("first_name", *patch.FirstName)
	}
	if patch.LastName != nil {
		addArg("last_name", *patch.LastName)
	}
	if patch.Email != nil {
		addArg("email", *patch.Email)
	}
	if patch.Phone != nil {
		addArg("phone", *patch.Phone)
	}
	if patch.Company != nil {
		addArg("company", *patch.Company)
	}
	if patch.LeadSource != nil {
		addArg("lead_source", *patch.LeadSource)
	}
	if patch.Score != nil {
		// Clamp score to [0, 100]
		score := *patch.Score
		if score < 0 {
			score = 0
		}
		if score > 100 {
			score = 100
		}
		addArg("lead_score", score)
	}
	if patch.Status != nil {
		addArg("status", *patch.Status)
	}
	if patch.OwnerID != nil {
		addArg("owner_id", *patch.OwnerID)
	}

	whereClause := fmt.Sprintf(`id=$%d AND deleted_at IS NULL`, i)
	args = append(args, id)
	i++

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		whereClause += fmt.Sprintf(` AND org_id=$%d`, i)
		args = append(args, orgID)
	}

	query := fmt.Sprintf(
		`UPDATE leads SET %s WHERE %s RETURNING %s`,
		strings.Join(sets, ", "), whereClause, leadCols,
	)
	return scanLead(r.db.QueryRow(ctx, query, args...))
}

func (r *LeadRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := `UPDATE leads SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`
	args := []any{id}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}

	result, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *LeadRepo) List(ctx context.Context, f domain.LeadFilter) ([]*domain.Lead, int, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	where := []string{"deleted_at IS NULL"}
	args := []any{}
	i := 1

	addWhere := func(expr string, val any) {
		where = append(where, fmt.Sprintf("%s = $%d", expr, i))
		args = append(args, val)
		i++
	}

	orgID, hasCtxOrg := domain.OrgIDFromContext(ctx)
	if !hasCtxOrg {
		orgID = f.OrgID
	}
	if orgID != uuid.Nil {
		addWhere("org_id", orgID)
	}

	if f.Status != nil {
		addWhere("status", *f.Status)
	}
	if f.OwnerID != nil {
		addWhere("owner_id", *f.OwnerID)
	}
	if f.Source != nil {
		addWhere("lead_source", *f.Source)
	}
	if f.ScoreMin != nil {
		where = append(where, fmt.Sprintf("lead_score >= $%d", i))
		args = append(args, *f.ScoreMin)
		i++
	}
	if f.ScoreMax != nil {
		where = append(where, fmt.Sprintf("lead_score <= $%d", i))
		args = append(args, *f.ScoreMax)
		i++
	}
	if f.Q != "" {
		// Use the GIN FTS index (idx_leads_fts) rather than ILIKE with a
		// leading wildcard, which cannot use btree indexes and causes seq scans.
		where = append(where, fmt.Sprintf(
			`to_tsvector('english', first_name || ' ' || last_name || ' ' || coalesce(email, '') || ' ' || coalesce(company, '')) @@ plainto_tsquery('english', $%d)`, i,
		))
		args = append(args, f.Q)
		i++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM leads WHERE `+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sortCol := "created_at"
	allowedSorts := map[string]bool{
		"created_at": true, "updated_at": true,
		"first_name": true, "last_name": true, "email": true,
		"lead_score": true, "status": true,
	}
	if allowedSorts[f.Sort] {
		sortCol = f.Sort
	}
	order := "DESC"
	if strings.ToUpper(f.Order) == "ASC" {
		order = "ASC"
	}

	rows, err := r.db.Query(ctx,
		fmt.Sprintf(
			`SELECT %s FROM leads WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
			leadCols, whereClause, sortCol, order, i, i+1,
		),
		append(args, f.Limit, offset)...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var leads []*domain.Lead
	for rows.Next() {
		l, err := scanLead(rows)
		if err != nil {
			return nil, 0, err
		}
		leads = append(leads, l)
	}
	return leads, total, rows.Err()
}

// ConvertToContact marks the lead as converted and sets the converted_contact_id.
// The caller is responsible for creating the contact beforehand.
func (r *LeadRepo) ConvertToContact(ctx context.Context, leadID, contactID uuid.UUID) (*domain.Lead, error) {
	q := `UPDATE leads SET status=$1, converted_contact_id=$2, updated_at=NOW()
	      WHERE id=$3 AND deleted_at IS NULL`
	args := []any{domain.LeadStatusConverted, contactID, leadID}
	i := 4

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += fmt.Sprintf(` AND org_id=$%d`, i)
		args = append(args, orgID)
	}

	q += ` RETURNING ` + leadCols
	return scanLead(r.db.QueryRow(ctx, q, args...))
}

// ListSources returns distinct non-null lead_source values for the org.
func (r *LeadRepo) ListSources(ctx context.Context) ([]string, error) {
	q := `SELECT DISTINCT lead_source FROM leads WHERE lead_source IS NOT NULL AND deleted_at IS NULL`
	args := []any{}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$1`
		args = append(args, orgID)
	}
	q += ` ORDER BY lead_source`

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		sources = append(sources, s)
	}
	return sources, rows.Err()
}

func (r *LeadRepo) DefaultPipelineID(ctx context.Context, orgID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx, `
		SELECT id
		FROM pipelines
		WHERE org_id = $1
		ORDER BY created_at ASC
		LIMIT 1
	`, orgID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, domain.ErrNotFound
	}
	return id, err
}
