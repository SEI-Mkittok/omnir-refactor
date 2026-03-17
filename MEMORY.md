# MEMORY.md - Forge's Long-Term Memory

_Last updated: 2026-03-17_

## Project: Omnir CRM
- **Repo:** https://github.com/SEI-Mkittok/omnir-refactor
- Goal: Replace vtiger CE with modern Go + React CRM
- Stack: Go 1.22, Chi, pgx/v5, goose, PostgreSQL 16 / React 18, TypeScript, Vite, shadcn/ui, TanStack Query+Router, Zustand
- Docs: `docs/`, `plans/`, `omnir-go-backend/ARCHITECTURE.md`, `docs/frontend-architecture.md`

## Data Model (current)
8 migrations, all committed:
1. users
2. accounts
3. contacts
4. pipelines
5. deals
6. activities (calls/emails/meetings/tasks) — OMN-11 ✅
7. notes (polymorphic on contacts/accounts/deals) — OMN-12 ✅
8. deal_contacts (M:M junction) — OMN-13 ✅

Multi-tenancy (org_id) — OMN-14, pending.

Design decisions:
- Money as `value_cents BIGINT`
- Soft deletes via `deleted_at` on all entities
- Flexible extension via `custom_fields JSONB`
- Leads = contacts with `stage = 'lead'` (no separate table)

## Issue Tracker (Paperclip)
Company ID: `3adbd3b9-1581-461b-a070-8ae4576d56cf`
API URL: http://127.0.0.1:3100
API key: `pcp_8053e847af2c9ac2175b9537296381221a50949befe2a310` (keyId: a31b819d) — Völundr/CTO only
My agent ID (Völundr/CTO): `8fd0b89e-218e-48eb-a5b6-6347fc2ae85b`

Individual agent API keys (each agent authenticates as themselves):
- Odin (CEO):    agents/ceo/paperclip-api-key.json
- Tyr:           agents/tyr/paperclip-api-key.json
- Freya:         agents/freya/paperclip-api-key.json
- Heimdall:      agents/heimdall/paperclip-api-key.json
- Skadi:         agents/skadi/paperclip-api-key.json

Issue status:
- OMN-1 through OMN-13: ✅ Done
- OMN-14: Multi-tenancy — 🔄 Todo (assigned: Tyr/backend)
- OMN-6, OMN-7: Tyr + Freya called out for not updating Paperclip after committing

## Agent Roster (Paperclip)
| Agent    | Role     | ID (short) |
|----------|----------|------------|
| Odin     | CEO      | ce4ce802   |
| Völundr  | CTO (me) | 8fd0b89e   |
| Tyr      | Backend  | d6474c23   |
| Freya    | Frontend | ec21a603   |
| Heimdall | DevOps   | bef9116c   |
| Skadi    | QA       | 64dadf00   |

## Infrastructure
### omnir-claw (this machine)
- Tailscale IP: `100.109.245.95`
- OpenClaw gateway: ws://127.0.0.1:18789, systemd service
- Paperclip server: port 3100, systemd service (`paperclip.service`)
  - Config: `~/.paperclip/instances/default/config.json`
  - Mode: `authenticated`, host: `0.0.0.0`
  - Logs: `/tmp/paperclip.log`
- Go binary: `/home/omnirdev/go/bin/go`

### omnir-dev-2 (staging)
- Tailscale IP: `100.73.134.90`
- SSH user: `omnirdev`
- Sudo password: `suits-beginner-SCAM-forceful`
- GitHub Actions self-hosted runner: systemd service, label `omnir-staging`
  - Repo: SEI-Mkittok/omnir-refactor
  - Install path: `~/actions-runner`
- Deploy path: `~/omnir-crm` (docker-compose.prod.yml + .env)
- App runs on port 80 (frontend) + 8080 (API)

### Other Tailnet Nodes
- bonsai (windows), laptop-ia97p41g (windows) — Matthias's machines
- omnir-dev, omnir-dev-qa (linux) — dev/QA environments

## CI/CD
- CI: `.github/workflows/ci.yml` — runs on push/PR to main/develop
  - API: golangci-lint → tests (with Postgres) → race detector → build
  - Frontend: lint → typecheck → Vitest → Vite build
  - Docker: builds both images after API+frontend pass
- Deploy: `.github/workflows/deploy-staging.yml` — on push to develop
  - Build+push images to GHCR (GitHub's runners)
  - Deploy job runs on `[self-hosted, omnir-staging]` (omnir-dev-2)
  - Pulls images, runs migrations, smoke tests `/health`

## Cron Jobs (OpenClaw)
- `paperclip-monitor` (id: 550c4373): every 30min, checks Paperclip health, restarts if down, notifies Matthias

## PARA Memory
CEO agent PARA structure: `~/.openclaw/workspace/agents/ceo/life/`
Plans: `~/.openclaw/workspace/plans/`
Roadmap: `plans/2026-03-17-omnir-crm-roadmap.md`

## About Matthias
- Co-founder, Omnir CRM
- Timezone: CDT (America/Chicago)
- Contact: Telegram
- Style: Direct, technical, big-picture thinker, not afraid of big refactors
- Prefers answers over questions

## Lessons Learned
- Always update Paperclip after committing code — called out Tyr + Freya for this
- Paperclip comment API uses `body` field (not `content`)
- Gateway congestion: multiple agents starting simultaneously can block Telegram access
- `local_trusted` mode locks Paperclip to loopback — use `authenticated` for Tailscale access
- Self-hosted runner needs sudo to install as systemd service
