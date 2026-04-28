---
title: Trello Integration
doc_kind: governance
doc_function: canonical
purpose: "Описывает интеграцию с Trello-доской: чтение карточек через API, создание feature package и синхронизацию статуса карточки со стадиями feature flow."
derived_from:
  - ../dna/governance.md
  - feature-flow.md
  - workflows.md
canonical_for:
  - trello_workflow_trigger
  - trello_card_to_feature_package_mapping
  - trello_column_sync_rules
  - trello_api_access
status: active
audience: humans_and_agents
---

# Trello Integration

## Workflow Trigger

Пользователь передаёт URL Trello-карточки в промпте. Агент извлекает идентификатор карточки из URL:

```
https://trello.com/c/{shortLink}/...
```

`shortLink` используется как ID карточки во всех последующих API-вызовах.

## API Access

Переменные окружения (задаются в `.env.local`, не коммитятся):

| Переменная | Описание |
|---|---|
| `TRELLO_API_KEY` | Trello REST API key |
| `TRELLO_TOKEN` | Trello user token |

### Чтение карточки

```
GET https://api.trello.com/1/cards/{shortLink}
    ?key={TRELLO_API_KEY}&token={TRELLO_TOKEN}
    &fields=name,desc,idBoard,idList,labels
```

### List IDs

Стабильные ID списков хранятся в [`memory-bank/ops/trello-board.md`](../ops/trello-board.md). Использовать их напрямую — динамический поиск не нужен.

### Перемещение карточки

```
PUT https://api.trello.com/1/cards/{shortLink}
    ?key={TRELLO_API_KEY}&token={TRELLO_TOKEN}&idList={targetListId}
```

## Маппинг полей карточки → Feature Package

| Поле карточки | Куда идёт |
|---|---|
| Название | заголовок `feature.md` и `README.md` |
| Описание | основа для `REQ-*` и `NS-*` в `feature.md` |
| Label `feature` | workflow: средняя/большая фича (полный feature package) |
| Label `bug fix` | workflow: баг-фикс (облегчённый FT-пакет) |
| Priority `High` / `Critical` | `CON-*` в `feature.md` |
| Остальные priority | игнорируются |

**Если описание скудное** (недостаточно для `REQ-*` / `NS-*`): агент задаёт уточняющие вопросы пользователю перед созданием Draft.

## Синхронизация колонок

| Trello колонка | Flow событие | Подтверждение |
|---|---|---|
| `TODO` | карточка не взята в обсуждение | — |
| `PLANNING` | агент получил задачу — немедленно, до чтения файлов и до обсуждения | не требуется — автопилот |
| `IN PROGRESS` | Bootstrap — worktree создан (до draft PR) | не требуется — автопилот |
| `DONE` | PR merged (gate Execution → Done пройден) | ✅ требуется |

> **HARD STOP — PLANNING:** карточка переводится в PLANNING **немедленно при получении задачи** — до чтения файлов, до любых вопросов и до обсуждения. Это сигнал на доске: задачей занимаются.

> **HARD STOP — IN PROGRESS:** карточка переводится в IN PROGRESS **сразу после создания worktree** (Bootstrap), до draft PR и до первого коммита с кодом.

Перемещение в DONE выполняется только после явного подтверждения пользователя.

## Правила отката

Если worktree удаляется до завершения задачи:

| Причина удаления | Целевой статус карточки |
|---|---|
| Задача временно приостановлена, вернёмся позже | `IN PROGRESS → PLANNING` |
| Задача отложена или деприоритизирована | любой → `TODO` |

## Lifecycle

```
1. Пользователь передаёт URL карточки
2. Агент читает карточку через API
3. *** HARD STOP *** Переместить карточку TODO → PLANNING (Trello API) — автопилот — немедленно, до чтения файлов
4. Агент читает memory-bank (trello.md, feature-flow.md, git-workflow.md, trello-board.md)
5. [если нужно] Агент задаёт уточняющие вопросы
6. *** Bootstrap *** (до создания feature-файлов):
   a. Создать git worktree: git worktree add ../lfcru_forum-FT-XXX -b feat/FT-XXX-slug
   b. Переместить карточку PLANNING → IN PROGRESS (Trello API) — автопилот
   c. Создать draft PR (до первого коммита с кодом)
   d. Вся дальнейшая работа — внутри worktree
7. Агент создаёт README.md + feature.md draft (внутри worktree)
8. Пользователь ревьюит Draft
9. Агент доводит до Design Ready (feature.md: status active)
10. Агент создаёт implementation-plan.md → Plan Ready
11. После merge PR: агент запрашивает подтверждение → перемещает IN PROGRESS → DONE
```

Trello-интеграция не меняет сам feature flow — только добавляет источник задач и две точки синхронизации статуса.
