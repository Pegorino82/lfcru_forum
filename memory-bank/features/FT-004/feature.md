---
title: "FT-004: Article Page"
doc_kind: feature
doc_function: canonical
purpose: "Страница просмотра полного текста статьи по URL /news/{id}, доступная гостям без авторизации."
derived_from:
  - ../../domain/problem.md
  - https://github.com/Pegorino82/lfcru_forum/issues/4
status: draft
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
---

# FT-004: Article Page

## What

### Problem

Сайт показывает список новостей на главной, но клик на заголовок ведёт в никуда — роута `/news/{id}` не существует. Пользователь видит заголовок и дату, но не может прочитать полный текст статьи.

Модель `News` с полем `Content` уже существует в codebase; задача — добавить роут и шаблон.

Общий контекст: [`../../domain/problem.md`](../../domain/problem.md).

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
|---|---|---|---|---|
| `MET-01` | 404-ошибки для существующих опубликованных статей | >0 | 0 | Логи HTTP-ответов |

### Scope

- `REQ-1` GET `/news/{id}` для существующей опубликованной статьи возвращает страницу с заголовком, датой публикации и полным текстом
- `REQ-2` Ссылки в блоке «Последние новости» на главной ведут на соответствующие `/news/{id}`
- `REQ-3` Страница доступна для неавторизованных гостей

### Non-Scope

- `NS-1` Редактирование и удаление статей (admin UI)
- `NS-2` Комментарии к статьям
- `NS-3` Шаринг в социальных сетях

### Constraints / Assumptions

- `CON-1` GET `/news/{id}` для несуществующего или неопубликованного id возвращает 404 (не 403)
- `ASM-1` Модель `News` с полем `Content` уже существует в codebase

## How

<!-- Заполнить на Design Ready: solution sketch, change surface, flow -->

## Verify

### Exit Criteria

- `EC-1` Все `REQ-*` покрыты passing `SC-*`
- `EC-2` `CON-1` проверен: неопубликованный id возвращает 404, а не 200 или 403

### Acceptance Scenarios

- `SC-1` Гость кликает на заголовок новости на главной → переходит на `/news/{id}` → видит заголовок, дату публикации и полный текст
- `SC-2` Гость открывает `/news/99999` (несуществующий id) → получает 404-страницу

### Negative / Edge Cases

- `NEG-1` GET `/news/{id}` для неопубликованной статьи → 404 (не 200, не 403)

### Traceability

| REQ | SC | NEG |
|---|---|---|
| `REQ-1` | `SC-1` | — |
| `REQ-2` | `SC-1` | — |
| `REQ-3` | `SC-1` | — |
