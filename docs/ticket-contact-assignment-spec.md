# Feature Specification: Assign Contact to Support Ticket

## Overview
Add the ability to assign a **single primary contact** to a support ticket. This allows agents to explicitly link each ticket to the correct customer/person record and keeps support workflows and reporting customer-centric.

## Purpose
- Ensure every ticket is tied to the right contact.
- Reduce communication errors caused by ambiguous or missing requester identity.
- Improve support analytics by enabling reliable contact-level reporting.

## Expected Benefits
- Faster ticket handling (less manual lookup).
- Better historical context for agents (all ticket/contact interactions connected).
- Lower risk of duplicate outreach or missed follow-up.
- Clearer operational reporting (tickets per contact, response/resolution trends).

## Scope
### In Scope
- Assign, change, and (if permitted) remove a contact on a ticket.
- Search/select contact from existing records.
- Display assigned contact in ticket detail and list views.
- Record contact assignment changes in ticket activity history.

### Out of Scope (Initial Release)
- Multiple contacts per ticket.
- Complex account-contact relationship modeling beyond existing system rules.
- Automated contact inference from inbound channels (future enhancement).

## User Stories
1. As a support agent, I can assign a contact to a ticket so the issue is linked to the correct person.
2. As a support agent, I can reassign a ticket’s contact if the initial match is wrong.
3. As a support manager, I can audit who changed ticket contact assignment and when.
4. As a reporting user, I can filter ticket lists by contact.

## Functional Requirements

### FR-1: Ticket Contact Field
- Add a `contact_id` reference field to the ticket model.
- The ticket stores at most one primary contact.
- If business policy requires it, contact can become mandatory before status transitions (e.g., `resolved`, `closed`).

### FR-2: Contact Assignment Actions
Authorized users can:
- Assign contact when none exists.
- Replace existing contact.
- Remove contact assignment (if org policy allows null contact).

### FR-3: Contact Picker
The contact selector must support:
- Typeahead search by name, email, and phone.
- Fast selection via keyboard and mouse.
- Empty-state message when no results found.
- Optional "Create Contact" action from picker context (if user has create permission).

### FR-4: Validation and Rules
- Reject assignment to deleted/disabled contacts.
- If tickets are account-scoped, validate contact-account compatibility (block or warn per business rule).
- Return user-friendly error messages for all validation failures.

### FR-5: Activity and Audit Logging
On every assign/change/remove action:
- Add ticket timeline event with previous/new value.
- Capture actor and timestamp.
- Persist change in system audit logs (if audit module exists).

### FR-6: Permissions
- `ticket.contact.read`: can view assigned contact.
- `ticket.contact.write`: can assign/change/remove contact.
- `contact.create` (optional): can create a new contact from ticket workflow.

### FR-7: List and Filter Support
- Ticket table displays assigned contact column.
- Ticket filters include contact selection.

## User Experience / Workflow

### Ticket Detail View
- Add `Contact` field in ticket header or right-side metadata panel.
- If assigned: show name, email, and profile link.
- If unassigned: show `No contact assigned` placeholder.

### Assignment Flow
1. Agent opens ticket.
2. Agent clicks `Contact` field.
3. Searchable picker opens.
4. Agent selects contact (or creates one if enabled).
5. System saves and updates UI.
6. Activity entry appears in ticket timeline.
7. Success toast appears: `Contact assigned` or `Contact updated`.

### Reassignment / Removal
- Reassignment uses same picker.
- Removal requires clear action and optional confirmation if policy requires.

## Data/API Design Guidance

### Data Model
- Add nullable FK: `tickets.contact_id -> contacts.id`.
- Add DB index on `tickets.contact_id` for filter/report performance.

### API Expectations
- `GET /tickets/:id` returns contact summary object.
- `PATCH /tickets/:id` accepts `contactId` updates.
- Response returns updated ticket and contact payload.

### Events (Optional but Recommended)
- `ticket.contact_assigned`
- `ticket.contact_changed`
- `ticket.contact_removed`

## Edge Cases
- Contact merged into another: maintain pointer to surviving contact record.
- Contact archived/deleted after assignment: preserve historical timeline context.
- Concurrent edits: prevent silent overwrite (optimistic locking or conflict handling).

## Acceptance Criteria
1. Authorized user can assign a contact from ticket detail UI.
2. Assigned contact persists and displays after refresh.
3. Contact reassignment/removal works according to policy.
4. All changes create timeline entries with actor + timestamp.
5. Unauthorized users cannot modify contact assignment.
6. Ticket list can display and filter by contact.
7. Validation messages are clear and actionable.

## Non-Functional Notes
- Contact search in picker should return results within acceptable UI latency for standard dataset size.
- Assignment actions must be fully covered by automated tests (API + UI where applicable).
- Audit events should be retained according to current support/compliance policies.
