-- +goose Up

-- Full-text search GIN indexes for global search across CRM entities.
-- Uses expression indexes on to_tsvector so queries can leverage them
-- with the matching to_tsvector(...) @@ plainto_tsquery(...) predicate.

-- Tickets: subject + description
CREATE INDEX idx_tickets_fts ON tickets
    USING GIN (to_tsvector('english', subject || ' ' || coalesce(description, '')))
    WHERE deleted_at IS NULL;

-- Ticket comments: body (searched as part of ticket results)
CREATE INDEX idx_ticket_comments_fts ON ticket_comments
    USING GIN (to_tsvector('english', body))
    WHERE deleted_at IS NULL;

-- Contacts: first_name + last_name + email + phone
CREATE INDEX idx_contacts_fts ON contacts
    USING GIN (to_tsvector('english',
        first_name || ' ' || last_name || ' ' ||
        coalesce(email, '') || ' ' || coalesce(phone, '')))
    WHERE deleted_at IS NULL;

-- Accounts: name + domain + industry
CREATE INDEX idx_accounts_fts ON accounts
    USING GIN (to_tsvector('english',
        name || ' ' || coalesce(domain, '') || ' ' || coalesce(industry, '')))
    WHERE deleted_at IS NULL;

-- Deals: title
CREATE INDEX idx_deals_fts ON deals
    USING GIN (to_tsvector('english', title))
    WHERE deleted_at IS NULL;

-- +goose Down

DROP INDEX IF EXISTS idx_deals_fts;
DROP INDEX IF EXISTS idx_accounts_fts;
DROP INDEX IF EXISTS idx_contacts_fts;
DROP INDEX IF EXISTS idx_ticket_comments_fts;
DROP INDEX IF EXISTS idx_tickets_fts;
