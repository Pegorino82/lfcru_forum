---
title: Bug Fix Execution Loop
doc_kind: governance
doc_function: canonical
purpose: "Цикл выполнения баг-фикса: от облегчённого FT-пакета до closure. Аналог feature-execution-loop.md для workflow «Баг-фикс»."
derived_from:
  - workflows.md
  - ../engineering/git-workflow.md
  - ../engineering/testing-policy.md
canonical_for:
  - bugfix_execution_loop_stages
  - bugfix_execution_loop_state
  - bugfix_run_state_structure
must_not_define:
  - feature_flow_gate_predicates
  - session_protocol
status: active
audience: humans_and_agents
---

# Bug Fix Execution Loop

Цикл выполнения баг-фикса: облегчённый пакет без `feature.md` и `implementation-plan.md`.

## State-pack

После каждого значимого этапа обновляются три артефакта:

| Артефакт | Роль |
|---|---|
| `run-state/FT-XXX/active-context.md` | текущая стадия, blocked/pending, ключевые решения |
| `run-state/FT-XXX/stage-log.md` | журнал пройденных этапов с outcome |
| `HANDOFF.md` (корень проекта) | сессионный entry point для следующего агента |

## FT-пакет баг-фикса

Создаётся в начале работы (внутри worktree):

- `memory-bank/features/FT-XXX/README.md` — структурированный артефакт:
  - `## Bug` — что сломано (одно предложение)
  - `## Repro` — минимальные шаги воспроизведения
  - `## Root Cause` — file:line + объяснение (заполнить после анализа)
  - `## Fix` — что изменено, хеш коммита (заполнить после fix)
  - `## Regression` — CHK-1 + EVID-1 (заполнить после верификации)
- `run-state/FT-XXX/active-context.md` — инициализировать по шаблону (тип: bugfix)
- `run-state/FT-XXX/stage-log.md` — инициализировать: 5 стадий pending

## Стадии

| Стадия | Что делать | Обновление run-state |
|--------|-----------|----------------------|
| `analysis` | Воспроизвести баг, найти root cause (file:line), заполнить `## Root Cause` в README | stage-log: analysis → done; active-context → Current: fix |
| `fix` | Сделать коммит, заполнить `## Fix` в README; получить хеш (`git log --oneline -1`), вписать в README и `HANDOFF.md`, сделать `docs:`-коммит | stage-log: fix → done; active-context → Current: tests |
| `tests` | Добавить regression test или задокументировать CHK-1, заполнить `## Regression` | stage-log: tests → done; active-context → Current: verification |
| `verification` | Вручную проверить что баг не воспроизводится, заполнить EVID-1 | stage-log: verification → done; active-context → Current: closure |
| `closure` | `gh pr ready`, обновить `## PR` в README, убедиться CI green | stage-log: closure → done; active-context → Status: awaiting-human |

> **Session Protocol** (`workflows.md` § «В конце сеанса») выполняется дополнительно к стадии `fix` — оба протокола действуют, не заменяют друг друга.

## Resume Protocol

При возобновлении после прерывания:

1. Прочитать `HANDOFF.md` → найти FT_ID и текущий stage
2. Прочитать `run-state/FT-XXX/active-context.md` → восстановить контекст (поле `Type: bugfix`)
3. Прочитать `run-state/FT-XXX/stage-log.md` → первая строка со статусом `pending`
4. Продолжить с этой стадии

## После merge PR

- Обновить `memory-bank/features/README.md` → статус `done`
- Удалить worktree: `git worktree remove ../lfcru_forum-FT-XXX && git branch -d fix/FT-XXX-slug`

## Формат state-артефактов

### active-context.md (bugfix)

```markdown
# Active Context: FT-XXX

**Updated:** YYYY-MM-DD
**Type:** bugfix
**Status:** in_progress

## Completed
<!-- заполняется по мере прохождения стадий -->

## Current
**<stage>** — in_progress
- [ ] <шаг>

## Blocked / Pending
—

## Key Decisions
—
```

### stage-log.md (bugfix)

```markdown
# Stage Log: FT-XXX

| Stage        | Status  | Outcome | Date       | Ref |
|-------------|---------|---------|------------|-----|
| analysis    | pending | —       | —          | —   |
| fix         | pending | —       | —          | —   |
| tests       | pending | —       | —          | —   |
| verification| pending | —       | —          | —   |
| closure     | pending | —       | —          | —   |
```
