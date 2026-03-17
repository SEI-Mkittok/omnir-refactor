package domain

// DealStageMetric holds aggregated deal data for one pipeline stage.
type DealStageMetric struct {
	Stage          DealStage `json:"stage"`
	Count          int       `json:"count"`
	TotalValueCents int64    `json:"total_value_cents"`
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
