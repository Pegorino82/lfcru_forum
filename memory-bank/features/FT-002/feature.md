---
title: "FT-002: Base Layout"
doc_kind: feature
doc_function: canonical
purpose: "Единый header и footer на всех страницах сайта и форума — визуальная целостность и стабильная навигация."
derived_from:
  - ../../domain/problem.md
  - https://github.com/Pegorino82/lfcru_forum/issues/2
status: draft
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
---

# FT-002: Base Layout

## What

### Problem

На разных страницах сайта отсутствуют header или footer. Пользователь не понимает, что находится на одном ресурсе, и теряет навигацию при переходе между разделами.

Общий контекст: [`../../domain/problem.md`](../../domain/problem.md).

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
|---|---|---|---|---|
| `MET-01` | Страницы без единого header и footer | >0 | 0 | Визуальный обход всех роутов |

### Scope

- `REQ-1` Header отображается на всех страницах сайта и форума (содержит логотип и навигацию)
- `REQ-2` Footer отображается на всех страницах сайта и форума (содержит ссылки и copyright)

### Non-Scope

- `NS-1` Мобильная адаптация header/footer (вне MVP)
- `NS-2` Авторизационные элементы в header («Войти» / «Выйти») — зависят от FT-001

### Constraints

- `CON-1` Шаблоны рендерятся через `html/template` stdlib; layout реализуется через template composition

## How

### Solution

Base-шаблон определяет общий layout; страницы встраиваются через `{{template "content" .}}`. Header и footer — отдельные partial-шаблоны, подключаемые в base.

### Change Surface

| Surface | Why |
|---|---|
| `templates/base.html` | Новый base-шаблон с header и footer |
| `templates/partials/header.html` | Partial: логотип + навигация |
| `templates/partials/footer.html` | Partial: ссылки + copyright |

### Flow

1. HTTP-запрос поступает на любой роут.
2. Handler рендерит шаблон, унаследованный от `base.html`.
3. Браузер получает страницу с единым header и footer.

## Verify

### Exit Criteria

- `EC-1` Все `REQ-*` покрыты passing `SC-*`

### Acceptance Scenarios

- `SC-1` Гость открывает любую страницу сайта → header (логотип + навигация) виден вверху, footer (ссылки + copyright) — внизу
- `SC-2` Пользователь переходит между несколькими разделами → header и footer остаются неизменными

### Traceability

| REQ | SC | NEG |
|---|---|---|
| `REQ-1` | `SC-1`, `SC-2` | — |
| `REQ-2` | `SC-1`, `SC-2` | — |
