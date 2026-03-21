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
- Verify Tyr/Freya are only working on their assigned issue
- Cancel any issues with issueNumber > 457 that aren't standing tasks (Phase 2 only: 454-457)
- If new rogue issues exist: cancel them and re-pause the offending agent

## 6. CI Health
- Check latest CI run on develop — must be green
- If red, fix before anything else

## 7. Phase Progress Gate
- Phase 2 tasks: OMN-455 (Tyr RLS), OMN-456 (Tyr org signup), OMN-457 (Freya tenant switcher)
- Sequence: 455 → merge → assign 456 → merge → activate Freya → assign 457
- When Phase 2 done: run `bash scripts/plan-next.sh` for Phase 3

## Paused agents (do NOT unpause without explicit reason)
- Odin: paused — use scripts/plan-next.sh for planning
- Skadi: paused — activate per-task after each merge
- Freya: paused — activate only after OMN-455+456 merged
