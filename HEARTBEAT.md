# HEARTBEAT.md — Völundr (CTO) Checklist

Run every heartbeat. In order.

## 1. Self-assign standing task
PATCH OMN-58 to `in_progress` assigned to Völundr if not already.

## 2. Board Review (ALL statuses)
Check for issues needing action:
- `in_review` → my queue, review PR and merge/close immediately
- `blocked` → unblock or cancel if stale
- `backlog` → activate (set `todo`) or cancel if old roadmap
- `todo` unassigned → assign to right agent

## 3. PR Review
- Check open PRs: GitHub API
- For each open PR: check CI status, diff, approve + merge or request changes
- Delete branch after merge
- No PR should sit open >1 heartbeat without action

## 4. Branch Cleanup
After every merge, delete the branch.
Periodically delete branches fully merged into develop.

## 5. Agent Sanity Check
- Verify Tyr/Freya are only working on their assigned issues
- If rogue self-created issues appear (agent created without CTO authorization): cancel them and comment why
- Standing tasks (always active): OMN-58 (Völundr), OMN-84 (Heimdall), OMN-530 (Völundr validation)

## 6. CI Health
- Check latest CI run on develop — must be green
- If red, fix before anything else

## 7. Phase Progress Gate
- Phase 2 complete. Project is in active development — follow run `bash scripts/plan-next.sh` for next planning cycle.
- Current active agents: Tyr (backend), Freya (frontend), Heimdall (DevOps)
- Skadi: activate per-task after each merge for QA

## Paused agents (do NOT unpause without explicit reason)
- Odin: paused — use scripts/plan-next.sh for planning
