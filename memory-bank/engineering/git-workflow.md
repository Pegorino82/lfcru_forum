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

## Worktrees

Не используются в проекте.
