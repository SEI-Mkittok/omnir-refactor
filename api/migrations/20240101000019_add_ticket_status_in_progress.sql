-- Allow 'in_progress' as a valid ticket status value.
-- The original CHECK constraint only included: open, pending, resolved, closed.

ALTER TABLE tickets
  DROP CONSTRAINT IF EXISTS tickets_status_check;

ALTER TABLE tickets
  ADD CONSTRAINT tickets_status_check
    CHECK (status IN ('open', 'in_progress', 'pending', 'resolved', 'closed'));
