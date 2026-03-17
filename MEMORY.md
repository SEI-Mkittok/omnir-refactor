# MEMORY.md - Forge's Long-Term Memory

## Project: Omnir CRM
- Go backend + React/TypeScript frontend replacing vtiger CE
- Stack: Go 1.22, Chi, pgx/v5, goose migrations, PostgreSQL 16, React 18, Vite, shadcn/ui, TanStack Query/Router, Zustand
- Repo structure: `omnir-go-backend/`, `omnir-frontend/`, docker-compose stack
- Docs: `docs/frontend-architecture.md`, `docs/vtiger-modernization-scope.md`, `omnir-go-backend/ARCHITECTURE.md`

## Agent Setup
- **Paperclip** is the agent orchestration layer (paperclipai/paperclip) — manages the AI company
- 6 agents: **odin** (me/Forge), **ceo**, **freya**, **heimdall**, **skadi**, **tyr**
- Each agent has workspace under `~/.openclaw/workspace/agents/<name>/`
- CEO agent has full SOUL.md + para-memory-files skill
- Paperclip API key stored at: `~/.openclaw/workspace/paperclip-claimed-api-key.json`
- Paperclip context: `~/.paperclip/context.json` (companyId: 3adbd3b9-1581-461b-a070-8ae4576d56cf)

## Infrastructure
- **OpenClaw gateway**: loopback, port 18789, running as systemd service
- **Paperclip server**: port 3100, running via `nohup node dist/index.js` from `~/.npm/_npx/43414d9b790239bb/node_modules/@paperclipai/server/`
  - Config: `~/.paperclip/instances/default/config.json`
  - Mode: `authenticated` (switched from local_trusted to allow Tailscale access)
  - Logs: `/tmp/paperclip.log`
  - Systemd service: `paperclip.service` (user), enabled + running
- **Tailscale**: enabled, this machine = `omnir-claw`, IP = `100.109.245.95`
  - Tailnet also has: bonsai (windows), laptop-ia97p41g (windows), omnir-dev, omnir-dev-2, omnir-dev-qa (linux)
- Paperclip accessible remotely at: `http://100.109.245.95:3100`

## Cron Jobs
- `paperclip-monitor` (id: 550c4373-4393-4c2c-9db9-fb37e4ec2937): every 30min, checks Paperclip health on port 3100, restarts if down, notifies Matthias on Telegram

## Known Issues / Notes
- Paperclip server is NOT set up as a systemd service — it will die on reboot. TODO: make it persistent.
- OpenClaw `groupPolicy` warning: allowlist mode but no group IDs configured — group messages dropped silently
- Gateway contention: multiple agents firing simultaneously caused timeout for Matthias's Telegram messages (2026-03-17)

## About Matthias
- Co-founder, Omnir CRM
- Timezone: CDT (America/Chicago)
- Prefers direct communication, technically sophisticated
- Contact: Telegram
