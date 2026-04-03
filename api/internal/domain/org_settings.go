package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DocType identifies document types that have auto-incrementing numbers.
type DocType string

const (
	DocTypeQuote     DocType = "quote"
	DocTypeTicket    DocType = "ticket"
	DocTypeKBArticle DocType = "kb_article"
	DocTypeInvoice   DocType = "invoice"
)

// OrgSettings holds per-org configuration, including document numbering start values.
type OrgSettings struct {
	ID                   uuid.UUID `json:"id"`
	OrgID                uuid.UUID `json:"org_id"`
	QuoteNumberStart     int64     `json:"quote_number_start"`
	TicketNumberStart    int64     `json:"ticket_number_start"`
	KBArticleNumberStart int64     `json:"kb_article_number_start"`
	InvoiceNumberStart   int64     `json:"invoice_number_start"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// OrgSettingsPatch holds the updatable fields for org settings.
type OrgSettingsPatch struct {
	QuoteNumberStart     *int64 `json:"quote_number_start"`
	TicketNumberStart    *int64 `json:"ticket_number_start"`
	KBArticleNumberStart *int64 `json:"kb_article_number_start"`
	InvoiceNumberStart   *int64 `json:"invoice_number_start"`
}

// FormattedDocNumber formats a number with the given prefix, e.g. QUO-00001.
func FormattedDocNumber(prefix string, number int64) string {
	return fmt.Sprintf("%s-%05d", prefix, number)
}
