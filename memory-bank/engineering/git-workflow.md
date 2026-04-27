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

**Создание** (выполняется на Bootstrap — сразу после обсуждения, до создания feature-файлов):

```bash
# из корня основного репозитория
git worktree add ../lfcru_forum-FT-XXX -b feat/FT-XXX-slug
cd ../lfcru_forum-FT-XXX
gh pr create --repo Pegorino82/lfcru_forum --draft --title "[WIP][FT-XXX] Краткое описание" --body "..."
```

Draft PR создаётся сразу — до первого коммита с кодом.

**Вся разработка** ведётся внутри worktree-папки. Все commits/push/CI привязаны к feature PR.

**Cleanup** (выполняется после merge PR):

```bash
git worktree remove ../lfcru_forum-FT-XXX
git branch -d feat/FT-XXX-slug
```

## Remote Safety

> **⛔ СТРОГИЙ ЗАПРЕТ на работу с upstream.**

Репозиторий имеет два remote:

| Remote | Назначение | Разрешено |
|---|---|---|
| `origin` | `Pegorino82/lfcru_forum` — рабочий форк | push, PR, CI |
| `upstream` | исходный репозиторий | только `git fetch` для синхронизации |

**Правила:**

1. `git push` — только в `origin`. Никогда в `upstream`.
2. `gh pr create` — всегда с явным флагом `--repo Pegorino82/lfcru_forum`. Без этого флага `gh` может использовать дефолтный репозиторий из контекста (который может быть чужим).
3. Перед `gh pr create` обязательно выполнить `gh repo view` и убедиться, что контекст — `Pegorino82/lfcru_forum`.
4. `gh pr`, `gh run`, `gh issue` без `--repo` — допустимы только после проверки контекста через `gh repo view`.
