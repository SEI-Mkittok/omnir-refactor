# Integrations and Webhooks

Integrations connect Omnir to external systems.

## Common Integration Areas

- Email inbox.
- Calendar sync.
- Teams or Slack-style notifications.
- Webhooks.
- API keys.
- Enrichment providers.

## Email and Calendar

Admins configure provider credentials. Users authorize their own accounts where required.

If OAuth fails, check redirect URLs, provider credentials, and tenant permissions.

## Webhooks

Webhooks send event notifications to external endpoints.

Before enabling a webhook:

- Confirm the receiving URL.
- Use HTTPS.
- Confirm the event scope.
- Confirm who owns the receiving system.
- Test with a safe event.

## API Keys

API keys are credentials. Treat them like passwords.

Use keys only for approved systems. Rotate them when ownership changes or if exposure is suspected.

## Best Practice

Document each integration's purpose, owner, and failure path. Integrations without owners become operational debt.
