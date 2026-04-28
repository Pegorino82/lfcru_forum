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
| `IN PROGRESS` | gate Plan Ready → Execution — перед первым коммитом с кодом (worktree и draft PR уже существуют с Bootstrap) | не требуется — пользователь уже дал go-ahead |
| `DONE` | PR merged (gate Execution → Done пройден) | ✅ требуется |

> **HARD STOP — IN PROGRESS:** карточка обязана быть переведена в IN PROGRESS **до первого коммита с кодом** (gate Plan Ready → Execution). Worktree и draft PR к этому моменту уже существуют (Bootstrap). Пропускать нельзя.

Перемещение в DONE выполняется только после явного подтверждения пользователя.

## Lifecycle

```
1. Пользователь передаёт URL карточки
2. Агент читает карточку через API
3. [если нужно] Агент задаёт уточняющие вопросы
4. *** Bootstrap *** (до создания feature-файлов):
   a. Создать git worktree: git worktree add ../lfcru_forum-FT-XXX -b feat/FT-XXX-slug
   b. Создать draft PR (до первого коммита с кодом)
   c. Вся дальнейшая работа — внутри worktree
5. Агент создаёт README.md + feature.md draft (внутри worktree)
6. Пользователь ревьюит Draft
7. Агент доводит до Design Ready (feature.md: status active)
8. Агент создаёт implementation-plan.md → Plan Ready
9. *** HARD STOP *** Перед первым коммитом с кодом:
   a. Переместить карточку TODO → IN PROGRESS (Trello API)
   b. Убедиться, что worktree и draft PR уже созданы (Bootstrap, шаг 4)
10. После merge PR: агент запрашивает подтверждение → перемещает IN PROGRESS → DONE
```

Trello-интеграция не меняет сам feature flow — только добавляет источник задач и две точки синхронизации статуса.
