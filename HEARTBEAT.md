# HEARTBEAT.md — Völundr (CTO) Checklist

Run every heartbeat. In order.

## 1. Board Review (ALL statuses)

Check for issues needing action:
- `in_review` → my queue, review and merge/close immediately
- `blocked` → unblock or cancel if stale
- `backlog` → activate (set to `todo`) or cancel if old roadmap
- `todo` unassigned → assign to right agent

## 2. PR Review

- Check open PRs: `gh pr list` or GitHub API
- For each open PR: check CI status, diff, approve + merge or request changes
- No PR should sit open >1 heartbeat without action

## 3. Branch Cleanup

After every merge, delete the branch:
```bash
git push origin --delete <branch>
```
Also periodically: delete all branches fully merged into develop:
```bash
git fetch origin --prune
# delete any branch with 0 commits ahead of develop
```

## 4. Odin Watch

- Check OMN-57 comments for new proposals from Odin
- Review proposed issues — spec them out and assign to engineers or cancel
- Ensure Odin is NOT assigning directly to engineers

## 5. CI Health

- Check latest CI run on develop — must be green
- If red, fix before assigning new work to agents

## 6. Stale Issue Check

- Any issue with `issueNumber < 107` that isn't OMN-57/58/84 → cancel it
- Any Phase 9/10/11 issue → cancel immediately
