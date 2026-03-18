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

// TicketReport is the response payload for GET /reports/tickets.
type TicketReport struct {
	TotalOpen          int                 `json:"total_open"`
	TotalClosed        int                 `json:"total_closed"`
	AvgResolutionHours float64             `json:"avg_resolution_hours"`
	ByStatus           []TicketStatusCount `json:"by_status"`
	BreachRate         float64             `json:"breach_rate"` // % of open tickets older than 48 h
}

// ContactPeriodMetric holds the count of contacts created in one time bucket.
type ContactPeriodMetric struct {
	Period string `json:"period"` // "YYYY-MM"
	Count  int    `json:"count"`
}

// ContactReport is the response payload for GET /reports/contacts.
type ContactReport struct {
	Total    int                   `json:"total"`
	ByPeriod []ContactPeriodMetric `json:"by_period"`
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

// LeadReport is the response payload for GET /reports/leads.
type LeadReport struct {
	TotalNew       int     `json:"total_new"`
	TotalConverted int     `json:"total_converted"`
	ConversionRate float64 `json:"conversion_rate"` // converted / total * 100
}
