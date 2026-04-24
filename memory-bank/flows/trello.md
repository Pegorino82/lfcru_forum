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
| `TODO` | карточка не взята в работу | — |
| `IN PROGRESS` | Bootstrap Feature Package создан (README.md + feature.md draft) | ✅ требуется |
| `DONE` | PR merged (gate Execution → Done пройден) | ✅ требуется |

Перемещение выполняется только после явного подтверждения пользователя.

## Lifecycle

```
1. Пользователь передаёт URL карточки
2. Агент читает карточку через API
3. [если нужно] Агент задаёт уточняющие вопросы
4. Агент создаёт Bootstrap Feature Package (README.md + feature.md draft)
5. Агент запрашивает подтверждение → перемещает карточку TODO → IN PROGRESS
6. Пользователь ревьюит Draft
7. Агент доводит до Design Ready (feature.md: status active)
8. Стандартный flow согласно feature-flow.md: Plan Ready → Execution (worktree + PR)
9. После merge PR: агент запрашивает подтверждение → перемещает IN PROGRESS → DONE
```

Trello-интеграция не меняет сам feature flow — только добавляет источник задач и две точки синхронизации статуса.
