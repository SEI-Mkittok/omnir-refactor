package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EnrichmentData is the JSONB payload stored per domain in the enrichment cache.
type EnrichmentData struct {
	CompanyName string `json:"company_name,omitempty"`
	Industry    string `json:"industry,omitempty"`
	Size        string `json:"size,omitempty"`
	LogoURL     string `json:"logo_url,omitempty"`
	LinkedInURL string `json:"linkedin_url,omitempty"`
}

// EnrichmentCache is one cached domain lookup result, scoped to an org.
type EnrichmentCache struct {
	ID        uuid.UUID      `json:"id"`
	OrgID     uuid.UUID      `json:"org_id"`
	Domain    string         `json:"domain"`
	Data      EnrichmentData `json:"data"`
	FetchedAt time.Time      `json:"fetched_at"`
	CreatedAt time.Time      `json:"created_at"`
}

// RawData returns Data marshalled as json.RawMessage for DB storage.
func (e *EnrichmentCache) RawData() (json.RawMessage, error) {
	return json.Marshal(e.Data)
}
