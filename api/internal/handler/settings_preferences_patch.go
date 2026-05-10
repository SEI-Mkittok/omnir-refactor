package handler

import (
	"encoding/json"

	"github.com/omnir/crm-api/internal/domain"
)

type repositoryPatch struct {
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

func decodePatchMapIntoPreferences(raw map[string]json.RawMessage, out *repositoryPatch) error {
	b, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

func (p repositoryPatch) toDomainPatch() domain.UserPreferencesPatch {
	return domain.UserPreferencesPatch{
		DefaultCurrency:               p.DefaultCurrency,
		NumberFormat:                  p.NumberFormat,
		DefaultRecordView:             p.DefaultRecordView,
		LandingPage:                   p.LandingPage,
		AddressLine1:                  p.AddressLine1,
		AddressLine2:                  p.AddressLine2,
		City:                          p.City,
		State:                         p.State,
		PostalCode:                    p.PostalCode,
		Country:                       p.Country,
		PhotoURL:                      p.PhotoURL,
		Tags:                          p.Tags,
		ServicePreferences:            p.ServicePreferences,
		CalendarStartDay:              p.CalendarStartDay,
		CalendarDateFormat:            p.CalendarDateFormat,
		CalendarTimeZone:              p.CalendarTimeZone,
		CalendarDefaultActivityStatus: p.CalendarDefaultActivityStatus,
		CalendarDefaultDurationMins:   p.CalendarDefaultDurationMins,
		CalendarReminderIntervalMins:  p.CalendarReminderIntervalMins,
		CalendarDefaultView:           p.CalendarDefaultView,
		CalendarDayStartHour:          p.CalendarDayStartHour,
		CalendarHourFormat:            p.CalendarHourFormat,
		CalendarDefaultActivityType:   p.CalendarDefaultActivityType,
		CalendarShowCompletedEvents:   p.CalendarShowCompletedEvents,
	}
}

