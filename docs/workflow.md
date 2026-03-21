# PraestOS Development Workflow

Adapted from BMAD Method for Paperclip + our agent roster.

---

## Core Principle

**Each phase produces an artifact that gates the next.** No work begins until the artifact exists and is approved. This prevents agents from building the wrong thing.

---

## Phases & Artifacts

### Phase 1 — Story (Odin → Völundr)
**Who:** Odin (CEO)  
**Input:** `plans/praestos-rewrite-plan.md`  
**Output:** A Paperclip issue assigned to Völundr with:
- Feature name and phase reference
- User story: "As a [user], I want [feature] so that [value]"
- Acceptance criteria (bullet list)
- Open questions (if any)

**Gate:** Völundr must comment "APPROVED" or request changes before any code starts.

---

### Phase 2 — Tech Spec (Völundr)
**Who:** Völundr (CTO)  
**Input:** Approved story from Phase 1  
**Output:** Updated issue with:
- API endpoints (method, path, request/response shape)
- DB schema changes (table, columns, migration number)
- Frontend routes/components affected
- Edge cases and security considerations

**Gate:** Tech spec must exist in the issue before Tyr/Freya start work.

---

### Phase 3 — Implementation (Tyr / Freya / Heimdall)
**Who:** Tyr (backend), Freya (frontend), Heimdall (infra)  
**Input:** Issue with approved tech spec  
**Output:**
- Feature branch: `feature/OMN-XXX-slug`
- Code compiles, tests pass locally (`bash scripts/pre-push-check.sh`)
- PR opened, Völundr notified in issue comment
- One PR at a time — no new PR until previous is merged

**Gate:** CI must pass. Völundr must review diff before merge.

---

### Phase 4 — Code Review (Völundr)
**Who:** Völundr (CTO)  
**Input:** Open PR with green CI  
**Output:**
- Approved → merge + delete branch + mark issue `done`
- Rejected → specific change requests as issue comments, back to Phase 3

**Gate:** No merge without Völundr approval.

---

### Phase 5 — QA (Skadi)
**Who:** Skadi (QA)  
**Input:** Feature merged to develop + deployed to staging  
**Output:** QA report as issue comment:
- ✅ Pass or ❌ Fail per acceptance criterion
- Bug issues created for failures (assigned to original author)

**Gate:** QA runs AFTER merge, not before.

---

## Agent Role Map

| BMAD Agent | Our Agent | Role |
|---|---|---|
| Analyst (Mary) | Odin | Story creation from spec |
| PM (John) | Odin | Acceptance criteria |
| Architect (Winston) | Völundr | Tech spec, ADRs |
| Scrum Master (Bob) | Völundr | Sprint gate, issue management |
| Developer (Amelia) | Tyr + Freya | Implementation |
| QA (Quinn) | Skadi | Post-merge validation |
| DevOps | Heimdall | CI/CD, infra |

---

## Issue Lifecycle

```
[backlog] → Odin creates story
[todo]    → Assigned to Völundr for tech spec
[in_progress] → Assigned to engineer after spec approved
[in_review]   → PR open, waiting on Völundr
[done]    → Merged, branch deleted, QA assigned
```

## Rules

1. **Odin** creates stories from spec only → assigns to Völundr
2. **Völundr** writes tech spec → assigns to engineer
3. **Engineer** implements → opens PR → notifies Völundr
4. **Völundr** reviews → merges or rejects
5. **Skadi** QAs on staging after merge
6. **One PR at a time** per engineer
7. **Delete branch** after every merge
8. **No Phase 9-11 work** — PraestOS phases 0-4 only
