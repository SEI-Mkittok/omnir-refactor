package domain

import (
	"time"

	"github.com/google/uuid"
)

type CalendarDefaultView string

const (
	CalendarDefaultViewMonth CalendarDefaultView = "month"
	CalendarDefaultViewWeek  CalendarDefaultView = "week"
)

type CalendarHourFormat string

const (
	CalendarHourFormat12 CalendarHourFormat = "12h"
	CalendarHourFormat24 CalendarHourFormat = "24h"
)

type UserPreferences struct {
	ID                            uuid.UUID      `json:"id"`
	UserID                        uuid.UUID      `json:"user_id"`
	OrgID                         uuid.UUID      `json:"org_id"`
	DefaultCurrency               *string        `json:"default_currency,omitempty"`
	NumberFormat                  *string        `json:"number_format,omitempty"`
	DefaultRecordView             *string        `json:"default_record_view,omitempty"`
	LandingPage                   *string        `json:"landing_page,omitempty"`
	AddressLine1                  *string        `json:"address_line1,omitempty"`
	AddressLine2                  *string        `json:"address_line2,omitempty"`
	City                          *string        `json:"city,omitempty"`
	State                         *string        `json:"state,omitempty"`
	PostalCode                    *string        `json:"postal_code,omitempty"`
	Country                       *string        `json:"country,omitempty"`
	PhotoURL                      *string        `json:"photo_url,omitempty"`
	Tags                          []string       `json:"tags"`
	ServicePreferences            map[string]any `json:"service_preferences"`
	CalendarStartDay              *string        `json:"calendar_start_day,omitempty"`
	CalendarDateFormat            *string        `json:"calendar_date_format,omitempty"`
	CalendarTimeZone              *string        `json:"calendar_time_zone,omitempty"`
	CalendarDefaultActivityStatus *string        `json:"calendar_default_activity_status,omitempty"`
	CalendarDefaultDurationMins   *int           `json:"calendar_default_duration_minutes,omitempty"`
	CalendarReminderIntervalMins  *int           `json:"calendar_reminder_interval_minutes,omitempty"`
	CalendarDefaultView           *string        `json:"calendar_default_view,omitempty"`
	CalendarDayStartHour          *int           `json:"calendar_day_start_hour,omitempty"`
	CalendarHourFormat            *string        `json:"calendar_hour_format,omitempty"`
	CalendarDefaultActivityType   *string        `json:"calendar_default_activity_type,omitempty"`
	CalendarShowCompletedEvents   bool           `json:"calendar_show_completed_events"`
	CreatedAt                     time.Time      `json:"created_at"`
	UpdatedAt                     time.Time      `json:"updated_at"`
}

type UserPreferencesPatch struct {
	DefaultCurrency               *string         `json:"default_currency,omitempty"`
	NumberFormat                  *string         `json:"number_format,omitempty"`
	DefaultRecordView             *string         `json:"default_record_view,omitempty"`
	LandingPage                   *string         `json:"landing_page,omitempty"`
	AddressLine1                  *string         `json:"address_line1,omitempty"`
	AddressLine2                  *string         `json:"address_line2,omitempty"`
	City                          *string         `json:"city,omitempty"`
	State                         *string         `json:"state,omitempty"`
	PostalCode                    *string         `json:"postal_code,omitempty"`
	Country                       *string         `json:"country,omitempty"`
	PhotoURL                      *string         `json:"photo_url,omitempty"`
	Tags                          *[]string       `json:"tags,omitempty"`
	ServicePreferences            *map[string]any `json:"service_preferences,omitempty"`
	CalendarStartDay              *string         `json:"calendar_start_day,omitempty"`
	CalendarDateFormat            *string         `json:"calendar_date_format,omitempty"`
	CalendarTimeZone              *string         `json:"calendar_time_zone,omitempty"`
	CalendarDefaultActivityStatus *string         `json:"calendar_default_activity_status,omitempty"`
	CalendarDefaultDurationMins   *int            `json:"calendar_default_duration_minutes,omitempty"`
	CalendarReminderIntervalMins  *int            `json:"calendar_reminder_interval_minutes,omitempty"`
	CalendarDefaultView           *string         `json:"calendar_default_view,omitempty"`
	CalendarDayStartHour          *int            `json:"calendar_day_start_hour,omitempty"`
	CalendarHourFormat            *string         `json:"calendar_hour_format,omitempty"`
	CalendarDefaultActivityType   *string         `json:"calendar_default_activity_type,omitempty"`
	CalendarShowCompletedEvents   *bool           `json:"calendar_show_completed_events,omitempty"`
}
