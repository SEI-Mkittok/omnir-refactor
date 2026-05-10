package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

type UserPreferenceRepo struct {
	db *pgxpool.Pool
}

func NewUserPreferenceRepo(db *pgxpool.Pool) *UserPreferenceRepo {
	return &UserPreferenceRepo{db: db}
}

const userPrefCols = `
	id, user_id, org_id,
	default_currency, number_format, default_record_view, landing_page,
	address_line1, address_line2, city, state, postal_code, country, photo_url,
	tags, service_preferences,
	calendar_start_day, calendar_date_format, calendar_time_zone, calendar_default_activity_status,
	calendar_default_duration_minutes, calendar_reminder_interval_minutes, calendar_default_view,
	calendar_day_start_hour, calendar_hour_format, calendar_default_activity_type, calendar_show_completed_events,
	created_at, updated_at
`

func scanUserPreferences(row pgx.Row) (*domain.UserPreferences, error) {
	p := &domain.UserPreferences{
		Tags:               []string{},
		ServicePreferences: map[string]any{},
	}
	var tagsRaw []byte
	var servicePrefsRaw []byte
	err := row.Scan(
		&p.ID, &p.UserID, &p.OrgID,
		&p.DefaultCurrency, &p.NumberFormat, &p.DefaultRecordView, &p.LandingPage,
		&p.AddressLine1, &p.AddressLine2, &p.City, &p.State, &p.PostalCode, &p.Country, &p.PhotoURL,
		&tagsRaw, &servicePrefsRaw,
		&p.CalendarStartDay, &p.CalendarDateFormat, &p.CalendarTimeZone, &p.CalendarDefaultActivityStatus,
		&p.CalendarDefaultDurationMins, &p.CalendarReminderIntervalMins, &p.CalendarDefaultView,
		&p.CalendarDayStartHour, &p.CalendarHourFormat, &p.CalendarDefaultActivityType, &p.CalendarShowCompletedEvents,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if len(tagsRaw) > 0 {
		_ = json.Unmarshal(tagsRaw, &p.Tags)
	}
	if len(servicePrefsRaw) > 0 {
		_ = json.Unmarshal(servicePrefsRaw, &p.ServicePreferences)
	}
	if p.Tags == nil {
		p.Tags = []string{}
	}
	if p.ServicePreferences == nil {
		p.ServicePreferences = map[string]any{}
	}
	return p, nil
}

func defaultUserPreferences(userID, orgID uuid.UUID) *domain.UserPreferences {
	return &domain.UserPreferences{
		ID:                          uuid.Nil,
		UserID:                      userID,
		OrgID:                       orgID,
		Tags:                        []string{},
		ServicePreferences:          map[string]any{},
		CalendarShowCompletedEvents: true,
	}
}

func (r *UserPreferenceRepo) GetByUser(ctx context.Context, userID, orgID uuid.UUID) (*domain.UserPreferences, error) {
	row := r.db.QueryRow(ctx, `SELECT `+userPrefCols+` FROM user_preferences WHERE user_id=$1 AND org_id=$2`, userID, orgID)
	p, err := scanUserPreferences(row)
	if errors.Is(err, domain.ErrNotFound) {
		return defaultUserPreferences(userID, orgID), nil
	}
	return p, err
}

func normalizeStrPtr(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	return &v
}

func (r *UserPreferenceRepo) Upsert(ctx context.Context, userID, orgID uuid.UUID, patch domain.UserPreferencesPatch) (*domain.UserPreferences, error) {
	current, err := r.GetByUser(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}

	if patch.DefaultCurrency != nil {
		current.DefaultCurrency = normalizeStrPtr(patch.DefaultCurrency)
	}
	if patch.NumberFormat != nil {
		current.NumberFormat = normalizeStrPtr(patch.NumberFormat)
	}
	if patch.DefaultRecordView != nil {
		current.DefaultRecordView = normalizeStrPtr(patch.DefaultRecordView)
	}
	if patch.LandingPage != nil {
		current.LandingPage = normalizeStrPtr(patch.LandingPage)
	}
	if patch.AddressLine1 != nil {
		current.AddressLine1 = normalizeStrPtr(patch.AddressLine1)
	}
	if patch.AddressLine2 != nil {
		current.AddressLine2 = normalizeStrPtr(patch.AddressLine2)
	}
	if patch.City != nil {
		current.City = normalizeStrPtr(patch.City)
	}
	if patch.State != nil {
		current.State = normalizeStrPtr(patch.State)
	}
	if patch.PostalCode != nil {
		current.PostalCode = normalizeStrPtr(patch.PostalCode)
	}
	if patch.Country != nil {
		current.Country = normalizeStrPtr(patch.Country)
	}
	if patch.PhotoURL != nil {
		current.PhotoURL = normalizeStrPtr(patch.PhotoURL)
	}
	if patch.Tags != nil {
		current.Tags = *patch.Tags
	}
	if patch.ServicePreferences != nil {
		current.ServicePreferences = *patch.ServicePreferences
	}
	if patch.CalendarStartDay != nil {
		current.CalendarStartDay = normalizeStrPtr(patch.CalendarStartDay)
	}
	if patch.CalendarDateFormat != nil {
		current.CalendarDateFormat = normalizeStrPtr(patch.CalendarDateFormat)
	}
	if patch.CalendarTimeZone != nil {
		current.CalendarTimeZone = normalizeStrPtr(patch.CalendarTimeZone)
	}
	if patch.CalendarDefaultActivityStatus != nil {
		current.CalendarDefaultActivityStatus = normalizeStrPtr(patch.CalendarDefaultActivityStatus)
	}
	if patch.CalendarDefaultDurationMins != nil {
		current.CalendarDefaultDurationMins = patch.CalendarDefaultDurationMins
	}
	if patch.CalendarReminderIntervalMins != nil {
		current.CalendarReminderIntervalMins = patch.CalendarReminderIntervalMins
	}
	if patch.CalendarDefaultView != nil {
		current.CalendarDefaultView = normalizeStrPtr(patch.CalendarDefaultView)
	}
	if patch.CalendarDayStartHour != nil {
		current.CalendarDayStartHour = patch.CalendarDayStartHour
	}
	if patch.CalendarHourFormat != nil {
		current.CalendarHourFormat = normalizeStrPtr(patch.CalendarHourFormat)
	}
	if patch.CalendarDefaultActivityType != nil {
		current.CalendarDefaultActivityType = normalizeStrPtr(patch.CalendarDefaultActivityType)
	}
	if patch.CalendarShowCompletedEvents != nil {
		current.CalendarShowCompletedEvents = *patch.CalendarShowCompletedEvents
	}

	tagsRaw, err := json.Marshal(current.Tags)
	if err != nil {
		return nil, err
	}
	serviceRaw, err := json.Marshal(current.ServicePreferences)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	row := r.db.QueryRow(ctx, `
		INSERT INTO user_preferences (
			user_id, org_id, default_currency, number_format, default_record_view, landing_page,
			address_line1, address_line2, city, state, postal_code, country, photo_url,
			tags, service_preferences,
			calendar_start_day, calendar_date_format, calendar_time_zone, calendar_default_activity_status,
			calendar_default_duration_minutes, calendar_reminder_interval_minutes, calendar_default_view,
			calendar_day_start_hour, calendar_hour_format, calendar_default_activity_type, calendar_show_completed_events,
			updated_at
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,
			$7,$8,$9,$10,$11,$12,$13,
			$14::jsonb,$15::jsonb,
			$16,$17,$18,$19,
			$20,$21,$22,
			$23,$24,$25,$26,
			$27
		)
		ON CONFLICT (user_id, org_id) DO UPDATE SET
			default_currency = EXCLUDED.default_currency,
			number_format = EXCLUDED.number_format,
			default_record_view = EXCLUDED.default_record_view,
			landing_page = EXCLUDED.landing_page,
			address_line1 = EXCLUDED.address_line1,
			address_line2 = EXCLUDED.address_line2,
			city = EXCLUDED.city,
			state = EXCLUDED.state,
			postal_code = EXCLUDED.postal_code,
			country = EXCLUDED.country,
			photo_url = EXCLUDED.photo_url,
			tags = EXCLUDED.tags,
			service_preferences = EXCLUDED.service_preferences,
			calendar_start_day = EXCLUDED.calendar_start_day,
			calendar_date_format = EXCLUDED.calendar_date_format,
			calendar_time_zone = EXCLUDED.calendar_time_zone,
			calendar_default_activity_status = EXCLUDED.calendar_default_activity_status,
			calendar_default_duration_minutes = EXCLUDED.calendar_default_duration_minutes,
			calendar_reminder_interval_minutes = EXCLUDED.calendar_reminder_interval_minutes,
			calendar_default_view = EXCLUDED.calendar_default_view,
			calendar_day_start_hour = EXCLUDED.calendar_day_start_hour,
			calendar_hour_format = EXCLUDED.calendar_hour_format,
			calendar_default_activity_type = EXCLUDED.calendar_default_activity_type,
			calendar_show_completed_events = EXCLUDED.calendar_show_completed_events,
			updated_at = EXCLUDED.updated_at
		RETURNING `+userPrefCols,
		userID, orgID, current.DefaultCurrency, current.NumberFormat, current.DefaultRecordView, current.LandingPage,
		current.AddressLine1, current.AddressLine2, current.City, current.State, current.PostalCode, current.Country, current.PhotoURL,
		string(tagsRaw), string(serviceRaw),
		current.CalendarStartDay, current.CalendarDateFormat, current.CalendarTimeZone, current.CalendarDefaultActivityStatus,
		current.CalendarDefaultDurationMins, current.CalendarReminderIntervalMins, current.CalendarDefaultView,
		current.CalendarDayStartHour, current.CalendarHourFormat, current.CalendarDefaultActivityType, current.CalendarShowCompletedEvents,
		now,
	)
	return scanUserPreferences(row)
}

