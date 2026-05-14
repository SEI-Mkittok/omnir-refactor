# Environment Variable Reference

All environment variables used by the Omnir CRM stack.

## Core

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `POSTGRES_USER` | yes | — | PostgreSQL username |
| `POSTGRES_PASSWORD` | yes | — | PostgreSQL password |
| `POSTGRES_DB` | yes | — | PostgreSQL database name |
| `DATABASE_URL` | yes | — | Full Postgres connection string |
| `JWT_SECRET` | yes | — | Secret for signing JWTs |
| `CORS_ORIGINS` | yes | — | Allowed CORS origins (comma-separated) |
| `SEQUENCE_TOKEN_SECRET` | yes | — | Secret for sequence tracking tokens |

## Object Storage (S3 / MinIO)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `S3_ENDPOINT` | no | `http://minio:9000` | S3-compatible endpoint |
| `S3_BUCKET` | no | `omnir-attachments` | Bucket name for file attachments |
| `S3_ACCESS_KEY` | no | `omnir` | S3 access key |
| `S3_SECRET_KEY` | no | `minioadmin-staging` | S3 secret key |

## Product Help Wiki

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PRODUCT_HELP_WIKI_ENABLED` | no | `false` | Enables periodic background sync for global Omnir product help. Manual sync remains available from Settings > Product Help. |
| `PRODUCT_HELP_WIKI_RAW_BASE_URL` | no | `https://raw.githubusercontent.com/wiki/SEI-Mkittok/omnir-refactor` | Raw GitHub Wiki base URL used to fetch the manifest and Markdown pages. |
| `PRODUCT_HELP_WIKI_BASE_URL` | no | `https://github.com/SEI-Mkittok/omnir-refactor/wiki` | Human GitHub Wiki base URL used for article view/edit links. |
| `PRODUCT_HELP_WIKI_MANIFEST_PATH` | no | `omnir-product-help-manifest.json` | Manifest filename in the GitHub Wiki. |
| `PRODUCT_HELP_WIKI_SYNC_INTERVAL` | no | `1h` | Background sync interval when product-help wiki sync is enabled. |

## Container Images

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `REGISTRY` | no | `ghcr.io` | Container registry |
| `API_IMAGE` | yes | — | API image path (org/repo/image) |
| `FRONTEND_IMAGE` | yes | — | Frontend image path |
| `IMAGE_TAG` | no | `develop` | Image tag to deploy |

## Phase 12: SSO & Push Notifications

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `VAPID_PUBLIC_KEY` | no* | — | VAPID public key for Web Push (OMN-438). Required once push notifications are enabled. |
| `VAPID_PRIVATE_KEY` | no* | — | VAPID private key for Web Push (OMN-438). Keep secret. |
| `SSO_ENCRYPTION_KEY` | no* | — | 32-byte hex key for encrypting OIDC client secrets at rest (OMN-441). Required once SSO is enabled. |

*These become required when their respective features are activated.
