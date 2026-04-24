---
title: Git Workflow
doc_kind: engineering
doc_function: canonical
purpose: Git-конвенции проекта LFC.ru forum — коммиты, ветки, PR.
derived_from:
  - ../dna/governance.md
status: active
audience: humans_and_agents
---

# Git Workflow

**Main branch:** `main`

## Commits

- Conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`
- Concise, present-tense subject (`fix: replies displayed after parent post`)
- Issue refs не обязательны, но желательны для фич и фиксов

## Pull Requests

- Перед PR: unit-тесты зелёные, integration-тесты затронутых пакетов зелёные
- PR title: короткий (до 70 символов), детали — в body

## Branches

Каждая задача ведётся на изолированной ветке:

| Тип задачи | Шаблон ветки |
|---|---|
| Фича | `feat/FT-XXX-slug` |
| Баг-фикс | `fix/FT-XXX-slug` |

`slug` — 2–4 слова через дефис, описывающие суть (`add-pagination`, `fix-reply-order`).

## Worktrees

Каждая фича или баг-фикс ведётся в отдельном git worktree — изолированной копии репозитория на feature-ветке.

**Naming convention:** worktree создаётся как sibling-директория рядом с основным репозиторием:

```
../lfcru_forum-FT-XXX/
```

**Создание** (выполняется при переходе Plan Ready → Execution):

```bash
# из корня основного репозитория
git worktree add ../lfcru_forum-FT-XXX -b feat/FT-XXX-slug
gh pr create --draft --title "[WIP][FT-XXX] Краткое описание" --body "..."
```

Draft PR создаётся сразу — до первого коммита с кодом.

**Вся разработка** ведётся внутри worktree-папки. Все commits/push/CI привязаны к feature PR.

**Cleanup** (выполняется после merge PR):

```bash
git worktree remove ../lfcru_forum-FT-XXX
git branch -d feat/FT-XXX-slug
```
