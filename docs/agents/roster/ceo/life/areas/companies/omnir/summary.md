# Omnir — Company Summary

_Last updated: 2026-03-17_

## What
Building a modern, self-hostable CRM. Replacing vtiger Community Edition with a clean Go + React stack.

## Infrastructure
- **Server:** Omnir-Claw (Linux), Tailscale IP: 100.109.245.95
- **OpenClaw gateway:** ws://127.0.0.1:18789, systemd service
- **Paperclip:** http://127.0.0.1:3100 (also http://100.109.245.95:3100 via Tailscale), systemd service
- **Tailnet nodes:** omnir-claw, bonsai, laptop-ia97p41g, omnir-dev, omnir-dev-2, omnir-dev-qa

## Agent Roster
| Name      | Role      | Adapter         |
|-----------|-----------|-----------------|
| CEO       | CEO       | openclaw_gateway |
| Völundr   | CTO       | openclaw_gateway |
| Backend   | Engineer  | openclaw_gateway |
| Frontend  | Engineer  | openclaw_gateway |
| DevOps    | DevOps    | openclaw_gateway |
| QA        | QA        | openclaw_gateway |
