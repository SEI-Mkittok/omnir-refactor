package domain

// SearchResultItem is a single result returned by the global search endpoint.
type SearchResultItem struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Title   string `json:"title"`
	Excerpt string `json:"excerpt"`
	URL     string `json:"url"`
}

// GlobalSearchResponse is the response for GET /api/v1/search.
type GlobalSearchResponse struct {
	Results []SearchResultItem `json:"results"`
	Total   int                `json:"total"`
}
