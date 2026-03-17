# MEMORY.md - Forge's Long-Term Memory

_Last updated: 2026-03-17_

## Project: Omnir CRM
- **Repo:** https://github.com/SEI-Mkittok/omnir-refactor
- **Branch:** `develop` (active), `main` (legacy vtiger)
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
My API key (Völundr/CTO): `pcp_8053e847af2c9ac2175b9537296381221a50949befe2a310`
My agent ID: `8fd0b89e-218e-48eb-a5b6-6347fc2ae85b`
Goal ID: `5c93e45b-32b8-44a3-a09a-0e11814abd3d`
Project ID: `224df1f1-39f9-4f1c-b08c-e963fb819349`

Individual agent API keys stored at `agents/<name>/paperclip-api-key.json`
Paperclip comment API uses `body` field (not `content`)

## Agent Roster (Paperclip)
| Agent    | Role     | ID (short) | Adapter        | Model           |
|----------|----------|------------|----------------|-----------------|
| Odin     | CEO      | ce4ce802   | claude_local   | sonnet-4-6      |
| Völundr  | CTO (me) | 8fd0b89e   | openclaw_gw    | sonnet-4-6      |
| Tyr      | Backend  | d6474c23   | claude_local   | sonnet-4-6      |
| Freya    | Frontend | ec21a603   | claude_local   | sonnet-4-6      |
| Heimdall | DevOps   | bef9116c   | claude_local   | sonnet-4-5      |
| Skadi    | QA       | 64dadf00   | claude_local   | sonnet-4-5      |

## Agent Workflows
- **Tyr/Freya:** When marking task done → create unassigned `QA: <task>` subtask
- **Skadi:** Each heartbeat scans for unassigned QA tasks + self-assigns
- **Heimdall:** Each heartbeat checks GitHub Actions CI status, creates issues for failures
- **All agents use:** `scripts/claude-wrapper.sh` (bakes in ANTHROPIC_API_KEY + --dangerously-skip-permissions)

## Infrastructure
### omnir-claw (this machine)
- Tailscale IP: `100.109.245.95`
- OpenClaw gateway: ws://127.0.0.1:18789, systemd service
- Paperclip server: port 3100, systemd service (`paperclip.service`)
  - Config: `~/.paperclip/instances/default/config.json`
  - Mode: `authenticated`, host: `0.0.0.0`
  - Logs: `/tmp/paperclip.log`
  - Heartbeat scheduler: 300000ms (5min)
- Go binary: `/home/omnirdev/go/bin/go`
- Claude Code: `/home/omnirdev/.npm-global/bin/claude` v2.1.77

### omnir-dev-2 (staging)
- Tailscale IP: `100.73.134.90`
- SSH user: `omnirdev`, sudo pw: `suits-beginner-SCAM-forceful`
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
- **CI is currently failing:** goimports formatting, missing ESLint config, lock file sync — OMN-15

## Cron Jobs (OpenClaw)
- `paperclip-monitor` (id: 550c4373): every 30min, checks Paperclip health, restarts if down

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
- OpenClaw gateway is single-agent by design — don't run 6 agents through it
- adapter-claude-local only passes plain string env values — use wrapper script to bake in secrets
- Claude Code needs `--dangerously-skip-permissions` for autonomous Paperclip heartbeats
- GitHub push protection blocks tokens in code — use gitignored files for secrets
- All Paperclip issues need projectId set or workspace resolution fails
- Paperclip checkout requires the agent's own API key (can't checkout as another agent)
