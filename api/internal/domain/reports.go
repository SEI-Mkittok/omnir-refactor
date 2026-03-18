package domain

import (
	"time"

	"github.com/google/uuid"
)

// DealStageMetric holds aggregated deal data for one pipeline stage.
type DealStageMetric struct {
	Stage           DealStage `json:"stage"`
	Count           int       `json:"count"`
	TotalValueCents int64     `json:"total_value_cents"`
}

// ContactMonthlyMetric holds the number of contacts created in a given month.
type ContactMonthlyMetric struct {
	Month string `json:"month"` // format: "2006-01"
	Count int    `json:"count"`
}

// ActivityTypeMetric holds the count of activities for a given type.
type ActivityTypeMetric struct {
	Type  ActivityType `json:"type"`
	Count int          `json:"count"`
}

// ReportsSummary is the full analytics response.
type ReportsSummary struct {
	DealsByStage     []DealStageMetric      `json:"deals_by_stage"`
	ContactsMonthly  []ContactMonthlyMetric `json:"contacts_monthly"`
	ActivitiesByType []ActivityTypeMetric   `json:"activities_by_type"`
}

// ReportFilter holds common date-range and org-scoping options for report queries.
// OrgID is only respected for admin callers; regular users always see their own org.
type ReportFilter struct {
	From  *time.Time
	To    *time.Time
	OrgID *uuid.UUID // admin-only override
}

// TicketStatusCount holds a per-status ticket count.
type TicketStatusCount struct {
	Status TicketStatus `json:"status"`
	Count  int          `json:"count"`
}

// TicketDailyMetric holds the number of tickets created on a given day.
type TicketDailyMetric struct {
	Date  string `json:"date"`  // YYYY-MM-DD
	Count int    `json:"count"`
}

// TicketReport is the response payload for GET /reports/tickets.
type TicketReport struct {
	TotalOpen          int                 `json:"total_open"`
	TotalClosed        int                 `json:"total_closed"`
	AvgResolutionHours float64             `json:"avg_resolution_hours"`
	ByStatus           []TicketStatusCount `json:"by_status"`
	BreachRate         float64             `json:"breach_rate"` // % of open tickets older than 48 h
	OverTime           []TicketDailyMetric `json:"over_time"`
}

// ContactOverTimeMetric holds the count of contacts created in one monthly bucket.
type ContactOverTimeMetric struct {
	Month string `json:"month"` // "YYYY-MM"
	Count int    `json:"count"`
}

// ContactReport is the response payload for GET /reports/contacts.
type ContactReport struct {
	NewCount   int                     `json:"new_count"`   // contacts created within the date range
	TotalCount int                     `json:"total_count"` // all-time total for the org
	OverTime   []ContactOverTimeMetric `json:"over_time"`
}

// DealStageCount holds per-stage deal count and aggregate value.
type DealStageCount struct {
	Stage           DealStage `json:"stage"`
	Count           int       `json:"count"`
	TotalValueCents int64     `json:"total_value_cents"`
}

// DealReport is the response payload for GET /reports/deals.
type DealReport struct {
	PipelineValueCents int64            `json:"pipeline_value_cents"` // sum of active (non-closed) deals
	WonCount           int              `json:"won_count"`
	LostCount          int              `json:"lost_count"`
	ByStage            []DealStageCount `json:"by_stage"`
}

// LeadFunnelMetric holds the count of leads at one stage of the funnel.
type LeadFunnelMetric struct {
	Stage string `json:"stage"` // e.g. "new", "contacted", "qualified", "converted"
	Label string `json:"label"` // title-cased display label
	Count int    `json:"count"`
}

// LeadReport is the response payload for GET /reports/leads.
type LeadReport struct {
	NewCount       int                `json:"new_count"`       // leads created within the date range
	ConvertedCount int                `json:"converted_count"` // converted leads within the date range
	ConversionRate float64            `json:"conversion_rate"` // converted / total (0.0-1.0)
	Funnel         []LeadFunnelMetric `json:"funnel"`
}
