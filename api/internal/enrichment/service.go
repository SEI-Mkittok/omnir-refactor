// Package enrichment provides domain-based contact data lookup with caching.
// External API keys (e.g. CLEARBIT_API_KEY) are optional; the feature degrades
// gracefully to DNS-based fallback when no key is present.
package enrichment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

const cacheTTL = 30 * 24 * time.Hour

// Service handles domain enrichment lookups with caching.
type Service struct {
	repo        repository.EnrichmentCacheRepository
	clearbitKey string
	httpClient  *http.Client
}

// New creates a new enrichment Service. clearbitKey may be empty.
func New(repo repository.EnrichmentCacheRepository, clearbitKey string) *Service {
	return &Service{
		repo:        repo,
		clearbitKey: clearbitKey,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

// LookupDomain returns enrichment data for the given domain, using cache when fresh.
func (s *Service) LookupDomain(ctx context.Context, d string) (*domain.EnrichmentCache, error) {
	d = normalizeDomain(d)
	if d == "" {
		return nil, fmt.Errorf("empty domain")
	}

	// Check cache first.
	cached, err := s.repo.GetByDomain(ctx, d)
	if err != nil {
		return nil, err
	}
	if cached != nil && time.Since(cached.FetchedAt) < cacheTTL {
		return cached, nil
	}

	// Cache miss or expired — fetch fresh data.
	data, err := s.fetch(d)
	if err != nil {
		// If cache exists but is stale, return stale rather than erroring.
		if cached != nil {
			return cached, nil
		}
		return nil, err
	}

	entry := &domain.EnrichmentCache{
		Domain: d,
		Data:   *data,
	}
	return s.repo.Upsert(ctx, entry)
}

// EnrichContact extracts the email domain from a contact and runs LookupDomain.
// It is safe to call with a nil or empty email — returns nil, nil in that case.
func (s *Service) EnrichContact(ctx context.Context, c *domain.Contact) (*domain.EnrichmentCache, error) {
	if c.Email == nil || *c.Email == "" {
		return nil, nil
	}
	d := domainFromEmail(*c.Email)
	if d == "" {
		return nil, nil
	}
	return s.LookupDomain(ctx, d)
}

// fetch queries Clearbit (if configured) or falls back to DNS.
func (s *Service) fetch(d string) (*domain.EnrichmentData, error) {
	if s.clearbitKey != "" {
		data, err := s.fetchClearbit(d)
		if err == nil {
			return data, nil
		}
		// Fall through to DNS on any Clearbit error.
	}
	return s.fetchDNS(d)
}

// fetchClearbit calls the Clearbit Company API.
func (s *Service) fetchClearbit(d string) (*domain.EnrichmentData, error) {
	url := fmt.Sprintf("https://company.clearbit.com/v2/companies/find?domain=%s", d)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.clearbitKey, "")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("clearbit returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Name    string `json:"name"`
		Logo    string `json:"logo"`
		Domain  string `json:"domain"`
		Metrics struct {
			Employees string `json:"employeesRange"`
		} `json:"metrics"`
		Category struct {
			Industry string `json:"industry"`
		} `json:"category"`
		LinkedIn struct {
			Handle string `json:"handle"`
		} `json:"linkedin"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	data := &domain.EnrichmentData{
		CompanyName: result.Name,
		Industry:    result.Category.Industry,
		Size:        result.Metrics.Employees,
		LogoURL:     result.Logo,
	}
	if result.LinkedIn.Handle != "" {
		data.LinkedInURL = "https://www.linkedin.com/company/" + result.LinkedIn.Handle
	}
	return data, nil
}

// fetchDNS performs a minimal DNS check and derives a company name from the domain.
func (s *Service) fetchDNS(d string) (*domain.EnrichmentData, error) {
	// Verify the domain is resolvable; ignore error — degrade gracefully.
	_, _ = net.LookupHost(d)

	// Derive company name: strip TLD and capitalise.
	parts := strings.Split(d, ".")
	name := d
	if len(parts) >= 2 {
		name = parts[len(parts)-2]
	}
	if len(name) > 0 {
		name = strings.ToUpper(name[:1]) + name[1:]
	}

	return &domain.EnrichmentData{
		CompanyName: name,
	}, nil
}

// domainFromEmail extracts the domain portion of an email address.
func domainFromEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return ""
	}
	return normalizeDomain(parts[1])
}

// normalizeDomain lowercases and trims the domain string.
func normalizeDomain(d string) string {
	return strings.ToLower(strings.TrimSpace(d))
}
