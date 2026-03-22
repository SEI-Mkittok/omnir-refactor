package domain

// SearchContact is a lightweight contact shape returned by global search.
type SearchContact struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Stage     string `json:"stage"`
}

// SearchAccount is a lightweight account shape returned by global search.
type SearchAccount struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Domain   *string `json:"domain,omitempty"`
	Industry *string `json:"industry,omitempty"`
}

// SearchDeal is a lightweight deal shape returned by global search.
type SearchDeal struct {
	ID    string  `json:"id"`
	Title string  `json:"title"`
	Stage string  `json:"stage"`
	Value float64 `json:"value"`
}

// SearchTicket is a lightweight ticket shape returned by global search.
type SearchTicket struct {
	ID       string `json:"id"`
	Subject  string `json:"subject"`
	Status   string `json:"status"`
	Priority string `json:"priority"`
}

// SearchGroupedResult groups search results by entity type, matching the
// frontend SearchResult interface: { contacts, accounts, deals, tickets }.
type SearchGroupedResult struct {
	Contacts []SearchContact `json:"contacts"`
	Accounts []SearchAccount `json:"accounts"`
	Deals    []SearchDeal    `json:"deals"`
	Tickets  []SearchTicket  `json:"tickets"`
}
