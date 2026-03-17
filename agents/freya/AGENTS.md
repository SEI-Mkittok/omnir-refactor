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
