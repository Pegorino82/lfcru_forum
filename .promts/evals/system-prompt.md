# Agent System Context — LFC.ru Forum

## Stack
Go + Echo, html/template, HTMX + Alpine.js, PostgreSQL (pgx), goose migrations.
Architecture: Handler → Service → Repo → PostgreSQL. DI in main.go.

## Feature Package Structure
All feature artifacts live in `memory-bank/features/FT-XXX/`:
- `README.md` — routing layer index
- `feature.md` — canonical owner intent (draft → active)
- `implementation-plan.md` — derived execution doc; MUST NOT exist until `feature.md` is design-ready

## Feature Flow Lifecycle
Draft Feature → Design Ready → Plan Ready → Execution → Done

## Transition Gate: Plan Ready → Execution

**⛔ HARD STOP. ALL steps below MUST happen BEFORE the first code commit. CANNOT be skipped or deferred.**

Execute IN THIS EXACT ORDER:
1. Move Trello card TODO → IN PROGRESS via Trello API — **THIS IS FIRST, BEFORE ANYTHING ELSE**
2. Create feature branch: `feat/FT-XXX-slug` (from main repo root)
3. Create git worktree: `git worktree add ../lfcru_forum-FT-XXX -b feat/FT-XXX-slug`
4. Switch into worktree: all subsequent work happens in `../lfcru_forum-FT-XXX`
5. Create draft PR: `gh pr create --repo Pegorino82/lfcru_forum --draft --title "[WIP][FT-XXX] ..." --body "..."`
6. ALL file creation and commits happen EXCLUSIVELY inside `../lfcru_forum-FT-XXX`

**Direct work in the main directory after worktree creation is PROHIBITED.**

## Trello Sync Rules

| Column      | Trigger                                              | Confirmation |
|-------------|------------------------------------------------------|--------------|
| TODO        | Card not started                                     | —            |
| IN PROGRESS | Before worktree + branch + PR creation (gate step 1) | Not required |
| DONE        | After PR merged                                      | Required     |

## Trello API

```
# Move card to IN PROGRESS
PUT https://api.trello.com/1/cards/{shortLink}?key={TRELLO_API_KEY}&token={TRELLO_TOKEN}&idList=69e908732098656229043150
```

Stable Trello Board IDs:
- Board ID: `69e90873209865622904312c`
- TODO list: `69e90873209865622904314f`
- IN PROGRESS list: `69e908732098656229043150`
- DONE list: `69e908732098656229043151`

Use these IDs directly — no dynamic lookup needed.

## Git Workflow

```bash
# Worktree creation (run from main repo root)
git worktree add ../lfcru_forum-FT-XXX -b feat/FT-XXX-slug
cd ../lfcru_forum-FT-XXX
gh pr create --repo Pegorino82/lfcru_forum --draft \
  --title "[WIP][FT-XXX] Short description" \
  --body "Closes #issue"
```

Branch naming: `feat/FT-XXX-slug` (feature), `fix/FT-XXX-slug` (bugfix).
Remote safety: always use `--repo Pegorino82/lfcru_forum` with `gh pr create`.

## Feature.md Minimum Requirements
- `status: active` for design-ready
- `delivery_status: in_progress` for execution
- Sections: What (≥1 REQ-*, ≥1 NS-*), How, Verify (≥1 SC-*, ≥1 CHK-*, ≥1 EVID-*)

## implementation-plan.md Minimum Requirements
- ≥1 PRE-*, ≥1 STEP-*, ≥1 CHK-*, ≥1 EVID-*
- Discovery context: relevant paths, local patterns, unresolved questions (OQ-*)

## Autonomy Rules
- Without confirmation: read files, create/update feature package artifacts
- Show plan before: architectural decisions, DB schema changes, code deletion
- Stop and ask: contradictory requirements, out-of-scope requests, thin card descriptions
