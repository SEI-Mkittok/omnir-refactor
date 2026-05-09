package domain

import "github.com/google/uuid"

type SearchEntityType string

const (
	SearchEntityContacts SearchEntityType = "contacts"
	SearchEntityAccounts SearchEntityType = "accounts"
	SearchEntityDeals    SearchEntityType = "deals"
	SearchEntityTickets  SearchEntityType = "tickets"
)

func (e SearchEntityType) IsValid() bool {
	switch e {
	case SearchEntityContacts, SearchEntityAccounts, SearchEntityDeals, SearchEntityTickets:
		return true
	default:
		return false
	}
}

type SearchFilter struct {
	Query            string
	Limit            int
	EntityType       SearchEntityType
	AccountID        *uuid.UUID
	ContactID        *uuid.UUID
	RelationshipType string
}

type SearchRelationship struct {
	EntityType string `json:"entity_type"`
	ID         string `json:"id"`
	Name       string `json:"name"`
}

// SearchContact is a lightweight contact shape returned by global search.
type SearchContact struct {
	ID               string              `json:"id"`
	FirstName        string              `json:"first_name"`
	LastName         string              `json:"last_name"`
	Email            string              `json:"email"`
	Stage            string              `json:"stage"`
	RelatedAccount   *SearchRelationship `json:"related_account,omitempty"`
	RelationshipType string              `json:"relationship_type,omitempty"`
}

// SearchAccount is a lightweight account shape returned by global search.
type SearchAccount struct {
	ID               string              `json:"id"`
	Name             string              `json:"name"`
	Domain           *string             `json:"domain,omitempty"`
	Industry         *string             `json:"industry,omitempty"`
	RelatedAccount   *SearchRelationship `json:"related_account,omitempty"`
	RelatedContact   *SearchRelationship `json:"related_contact,omitempty"`
	RelationshipType string              `json:"relationship_type,omitempty"`
}

// SearchDeal is a lightweight deal shape returned by global search.
type SearchDeal struct {
	ID               string              `json:"id"`
	Title            string              `json:"title"`
	Stage            string              `json:"stage"`
	Value            float64             `json:"value"`
	RelatedAccount   *SearchRelationship `json:"related_account,omitempty"`
	RelatedContact   *SearchRelationship `json:"related_contact,omitempty"`
	RelationshipType string              `json:"relationship_type,omitempty"`
}

// SearchTicket is a lightweight ticket shape returned by global search.
type SearchTicket struct {
	ID               string              `json:"id"`
	Subject          string              `json:"subject"`
	Status           string              `json:"status"`
	Priority         string              `json:"priority"`
	RelatedAccount   *SearchRelationship `json:"related_account,omitempty"`
	RelatedContact   *SearchRelationship `json:"related_contact,omitempty"`
	RelationshipType string              `json:"relationship_type,omitempty"`
}

// SearchGroupedResult groups search results by entity type, matching the
// frontend SearchResult interface: { contacts, accounts, deals, tickets }.
type SearchGroupedResult struct {
	Contacts []SearchContact `json:"contacts"`
	Accounts []SearchAccount `json:"accounts"`
	Deals    []SearchDeal    `json:"deals"`
	Tickets  []SearchTicket  `json:"tickets"`
}
