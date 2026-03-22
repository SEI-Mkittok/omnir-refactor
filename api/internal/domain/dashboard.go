package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// WidgetPosition defines the grid layout position for a dashboard widget.
type WidgetPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// Widget is a single panel on a custom dashboard.
type Widget struct {
	Type        string         `json:"type"`         // e.g. "deals_by_stage", "contacts_monthly"
	QueryParams map[string]any `json:"query_params"` // filter parameters forwarded to the query
	Position    WidgetPosition `json:"position"`
}

// CustomDashboard is a user-configured analytics dashboard.
type CustomDashboard struct {
	ID        uuid.UUID `json:"id"`
	OrgID     uuid.UUID `json:"org_id"`
	Name      string    `json:"name"`
	Widgets   []Widget  `json:"widgets"`
	CreatedBy uuid.UUID `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CustomDashboardPatch is the optional-field update payload for a dashboard.
type CustomDashboardPatch struct {
	Name    *string  `json:"name"`
	Widgets []Widget `json:"widgets"` // nil means "don't update"
}

// ScheduledReport defines a recurring email delivery for a dashboard.
type ScheduledReport struct {
	ID          uuid.UUID  `json:"id"`
	OrgID       uuid.UUID  `json:"org_id"`
	DashboardID uuid.UUID  `json:"dashboard_id"`
	Schedule    string     `json:"schedule"`   // standard cron expression (5-field)
	Recipients  []string   `json:"recipients"` // list of email addresses
	LastSentAt  *time.Time `json:"last_sent_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ScheduledReportPatch is the optional-field update payload for a scheduled report.
type ScheduledReportPatch struct {
	DashboardID *uuid.UUID `json:"dashboard_id"`
	Schedule    *string    `json:"schedule"`
	Recipients  []string   `json:"recipients"` // nil means "don't update"
}

// DashboardRunResult is the response for POST /dashboards/:id/run.
type DashboardRunResult struct {
	DashboardID uuid.UUID    `json:"dashboard_id"`
	Widgets     []WidgetData `json:"widgets"`
}

// WidgetData pairs a widget definition with its query results.
type WidgetData struct {
	Widget Widget `json:"widget"`
	Data   any    `json:"data"`
}

// WidgetsToJSON serialises a widget slice for JSONB storage.
func WidgetsToJSON(ws []Widget) ([]byte, error) {
	return json.Marshal(ws)
}

// WidgetsFromJSON deserialises widgets from a JSONB column.
func WidgetsFromJSON(b []byte) ([]Widget, error) {
	var ws []Widget
	if err := json.Unmarshal(b, &ws); err != nil {
		return nil, err
	}
	return ws, nil
}

// RecipientsToJSON serialises a recipient list for JSONB storage.
func RecipientsToJSON(rs []string) ([]byte, error) {
	return json.Marshal(rs)
}

// RecipientsFromJSON deserialises a recipient list from a JSONB column.
func RecipientsFromJSON(b []byte) ([]string, error) {
	var rs []string
	if err := json.Unmarshal(b, &rs); err != nil {
		return nil, err
	}
	return rs, nil
}
