package domain

import "time"

// LinkedEntity describes an association from a primary CRM record
// (e.g. account/contact detail endpoint) to another entity.
type LinkedEntity struct {
	Type      string    `json:"type"`
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}

// LinkedEntityFilter controls pagination and filters when listing associations.
type LinkedEntityFilter struct {
	Type  *string
	Role  *string
	Since *time.Time
	Page  int
	Limit int
}
