# PraestOS + n8n Integration Guide

This guide explains how to connect PraestOS to n8n using the **HTTP Request** node. No custom nodes are required — PraestOS exposes a standard REST API and an outbound webhook system that n8n can consume directly.

## Prerequisites

- A running n8n instance (self-hosted or n8n Cloud)
- A PraestOS API key — generate one at **Settings → API Keys**
- Your PraestOS base URL (e.g. `https://app.praestos.io` or your self-hosted domain)

---

## Authentication

All PraestOS API requests require a `Bearer` token header:

```
Authorization: Bearer <your-api-key>
```

### Storing credentials in n8n

Use **n8n Credentials → Header Auth** and set:

| Field  | Value                       |
|--------|-----------------------------|
| Name   | `Authorization`             |
| Value  | `Bearer <your-api-key>`     |

Reference this credential in every HTTP Request node that calls PraestOS.

---

## Base URL

All REST endpoints are under:

```
https://<your-praestos-domain>/api/v1
```

Example: `https://app.praestos.io/api/v1/tickets`

---

## Example Workflow 1 — Create a Ticket on Incoming Webhook

**Use case:** A form submission or external alert triggers an n8n webhook, which automatically creates a PraestOS support ticket.

### Nodes

```
Webhook (n8n) → HTTP Request (PraestOS Create Ticket)
```

### Step 1 — Add a Webhook node

- **HTTP Method:** `POST`
- **Path:** `/ticket-intake` (or any path you choose)
- **Response Mode:** `Immediately`

### Step 2 — Add an HTTP Request node

| Field             | Value                                                   |
|-------------------|---------------------------------------------------------|
| **Method**        | `POST`                                                  |
| **URL**           | `https://<your-domain>/api/v1/tickets`                  |
| **Authentication**| Header Auth (your PraestOS credential)                  |
| **Body Type**     | `JSON`                                                  |

**JSON Body:**

```json
{
  "subject": "={{ $json.subject }}",
  "description": "={{ $json.description }}",
  "priority": "medium",
  "status": "open"
}
```

Replace `$json.subject` and `$json.description` with the actual field names from your incoming webhook payload.

**Successful response (201):**

```json
{
  "id": "a1b2c3d4-...",
  "org_id": "...",
  "subject": "Cannot log in",
  "status": "open",
  "priority": "medium",
  "tags": [],
  "created_at": "2026-03-01T10:00:00Z",
  "updated_at": "2026-03-01T10:00:00Z"
}
```

---

## Example Workflow 2 — Notify n8n When a Deal Stage Changes

**Use case:** When a deal moves to `closed_won` in PraestOS, send a Slack message or create a task in your project management tool.

### Step 1 — Register a PraestOS outbound webhook

Call the PraestOS webhooks API once to register your n8n endpoint:

```http
POST /api/v1/webhooks
Authorization: Bearer <your-api-key>
Content-Type: application/json

{
  "url": "https://your-n8n-instance.com/webhook/praestos-deals",
  "events": ["deal.stage_changed"]
}
```

PraestOS will `POST` a payload to your n8n webhook URL whenever a deal's stage changes.

### Step 2 — Add a Webhook node in n8n

- **HTTP Method:** `POST`
- **Path:** `/praestos-deals`
- **Authentication:** None (PraestOS signs payloads with a shared secret — see [Webhook Security](#webhook-security) below)

### Step 3 — Add an IF node to filter for `closed_won`

**Condition:** `{{ $json.data.stage }}` equals `closed_won`

### Step 4 — Act on the event

Connect the `true` branch to a Slack node, HubSpot node, or any action.

**Incoming payload shape:**

```json
{
  "event": "deal.stage_changed",
  "org_id": "...",
  "data": {
    "id": "a1b2c3d4-...",
    "title": "Acme Corp — Enterprise Plan",
    "value_cents": 1200000,
    "currency": "USD",
    "stage": "closed_won",
    "probability": 100,
    "owner_id": "...",
    "pipeline_id": "...",
    "updated_at": "2026-03-01T12:00:00Z"
  }
}
```

Note: `value_cents` is an integer representing the deal value in the smallest currency unit (e.g. cents for USD). Divide by 100 for display.

---

## Supported Outbound Webhook Events

Register any combination of these events when calling `POST /api/v1/webhooks`:

| Event                  | Description                                  |
|------------------------|----------------------------------------------|
| `ticket.created`       | A new ticket was created                     |
| `ticket.updated`       | An existing ticket was updated               |
| `deal.created`         | A new deal was created                       |
| `deal.updated`         | A deal was updated                           |
| `deal.stage_changed`   | A deal's pipeline stage changed              |
| `deal.deleted`         | A deal was deleted                           |
| `contact.created`      | A new contact was created                    |
| `contact.updated`      | A contact was updated                        |
| `activity.created`     | A new activity was logged                    |

---

## Webhook Security

PraestOS signs every outbound webhook with a shared secret using HMAC-SHA256. The `X-PraestOS-Signature` header contains the hex-encoded signature of the raw request body.

To verify in n8n, add a **Function** node before processing:

```javascript
const crypto = require('crypto');

const secret = 'your-webhook-secret'; // stored in n8n credentials
const signature = $input.first().headers['x-praestos-signature'];
const body = JSON.stringify($input.first().body);

const expected = crypto
  .createHmac('sha256', secret)
  .update(body)
  .digest('hex');

if (signature !== expected) {
  throw new Error('Invalid webhook signature');
}

return $input.all();
```

---

## Key REST Endpoints

| Resource    | Endpoint                              | Methods                   |
|-------------|---------------------------------------|---------------------------|
| Tickets     | `/api/v1/tickets`                     | GET, POST                 |
| Ticket      | `/api/v1/tickets/{id}`                | GET, PATCH, DELETE        |
| Deals       | `/api/v1/deals`                       | GET, POST                 |
| Deal        | `/api/v1/deals/{id}`                  | GET, PATCH, DELETE        |
| Contacts    | `/api/v1/contacts`                    | GET, POST                 |
| Contact     | `/api/v1/contacts/{id}`               | GET, PATCH, DELETE        |
| Webhooks    | `/api/v1/webhooks`                    | GET, POST                 |
| Webhook     | `/api/v1/webhooks/{id}`               | GET, PATCH, DELETE        |
| API Keys    | `/api/v1/api-keys`                    | GET, POST, DELETE         |

---

## Troubleshooting

**401 Unauthorized** — Check that the `Authorization: Bearer <key>` header is present and the key has not been revoked (Settings → API Keys).

**422 Unprocessable Entity** — A required field is missing or has an invalid value. Check the `error` field in the response body.

**Webhook not firing** — Confirm the webhook was registered with `GET /api/v1/webhooks` and that the correct events are listed. Use `POST /api/v1/webhooks/{id}/test` to send a test payload to your n8n URL.
