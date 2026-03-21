# Workflow: Git + Paperclip

## GitHub Push Policy

**All pushes must go through Völundr (CTO) for review before reaching origin/develop.**

1. **Create a feature branch** from `develop`: `git checkout -b feature/OMN-XX-description`
2. **Commit your work** locally
3. **Push to your branch** (not develop): `git push origin feature/OMN-XX-description`
4. **Message Völundr in Paperclip** with: branch name, commit count, what you changed
5. **Völundr reviews** → either approves or requests changes
6. **Create a Pull Request** on GitHub (Völundr will review + merge)
7. **Never push directly to develop** — all changes go through PRs

## GitHub PR Requirements

- Title: "feat/fix: description" matching the issue
- Linked issue: "Fixes OMN-XX" in the PR body
- Branch: `feature/OMN-XX-...` or `fix/OMN-XX-...`
- Approval: Völundr (CTO) must approve before merge
- CI must pass before merge

## Paperclip Issue Updates

After Völundr approves and merges your PR:
- Update the issue status to `done`
- Comment: "Merged in PR #NNN"
- Link any QA subtasks that need verification

## Emergency Hotfixes

If something is broken on staging:
1. Notify Völundr immediately in Paperclip
2. Create a `hotfix/OMN-XX` branch from `main`
3. Fix + test
4. Submit for Völundr review (expedited)
5. Merge to both `develop` and `main` after approval

---

**Bottom line: All code goes through Völundr before it reaches production.**

## Pre-Push Validation (MANDATORY)

Before pushing ANY branch to GitHub, you MUST run:
```bash
cd /home/omnirdev/.openclaw/workspace
bash scripts/pre-push-check.sh
```

If it fails, FIX the issue before pushing. Do NOT push code that does not compile or pass tests.

Common failures to watch for:
- Interface not fully implemented (missing methods)
- Import cycle or unused imports
- go.mod version mismatch with CI Go version
- Test assertions using wrong field names after schema changes



## Sequential Merge Rule (MANDATORY)

**Never open a PR or request a merge if a previous PR is still open or failed CI.**

Workflow:
1. Check if any PR is currently open — ask Völundr or run: gh pr list
2. If a PR is open → wait for it to be reviewed, CI to pass, and merge to complete
3. Only then push your branch and open the next PR
4. One PR in flight at a time — no parallel merges

If you are unsure whether a PR is open, ask Völundr before pushing.


## QA Scope — What Skadi Reviews

QA runs on code AFTER it merges to develop, not before.

**Do NOT:**
- Create QA tasks for PRs that have not yet merged
- Test features on feature branches
- Block PR merges with QA findings (QA is post-merge)

**DO:**
- After a merge to develop is deployed to staging, test end-to-end
- Create bug issues for regressions found on staging
- Assign bugs back to the original author (Tyr or Freya)

**When you find a bug post-merge:**
1. Create a bug issue with clear repro steps
2. Assign to Tyr (backend) or Freya (frontend)
3. Mark priority based on severity (critical = blocks users, low = cosmetic)
4. Do NOT create a new QA subtask for the same bug — the bug issue IS the task

**What good QA output looks like:**
- Specific repro steps
- Expected vs actual behavior
- Screenshot/log if available
- Severity assessment


## Open PR Immediately After Pushing (MANDATORY)

After pushing a feature branch, you MUST open a PR within the same task:

```
gh pr create --base develop --head <your-branch> --title "<OMN-XXX> title" --body "Closes OMN-XXX"
```

Do NOT wait to be told. Push branch → open PR → notify Völundr in Paperclip.
A branch with no PR is invisible to the review pipeline.


## Workflow Role: Phase 5 — Post-Merge QA

See `docs/workflow.md` for the full workflow.

Only QA features that are:
- Merged to develop ✅
- Deployed to staging (http://100.73.134.90) ✅

QA output: comment on the issue with:
- ✅ / ❌ per acceptance criterion
- For failures: create a bug issue assigned to the original author
- Set issue to `done` if all criteria pass

Do NOT create QA tasks for unmerged features.
