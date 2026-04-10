---
title: Frontend
doc_kind: domain
doc_function: canonical
purpose: Описание UI-поверхностей, interaction patterns и правил разделения HTMX/Alpine.js. Читать при любых изменениях шаблонов или клиентской интерактивности.
derived_from:
  - ../dna/governance.md
status: active
audience: humans_and_agents
---

# Frontend

## UI Surfaces

Единственная поверхность — **Public Web**: сайт + форум для болельщиков.

| Аспект | Значение |
| --- | --- |
| Стек | Go `html/template` (stdlib) + HTMX + Alpine.js |
| Шаблоны | `templates/` — layouts в `templates/layouts/`, страницы по `templates/<domain>/` |
| Boundary с backend | Echo-хэндлер рендерит полную страницу или HTMX-фрагмент |
| Canonical owner | `memory-bank/domain/frontend.md` |

Нет мобильного приложения, backoffice-интерфейса или отдельного SPA.

## Component And Styling Rules

- Нет отдельной design system или component library.
- Базовый layout: `templates/layouts/base.html` — nav, flash-сообщения, content-блок, подключение HTMX и Alpine.js.
- Новый шаблон создаётся в `templates/<domain>/name.html`; layout парсится в каждый **изолированный `*template.Template` set** — это предотвращает конфликты `{{define}}` между страницами.
- Хэндлер рендерит по полному пути: `"templates/auth/login.html"`.
- Локальные стили допустимы только в рамках конкретной страницы.

## Interaction Patterns

### HTMX — серверные запросы и подмена DOM

Используется для любых запросов к серверу и подмены фрагментов DOM:

- `hx-get`, `hx-post` — инициировать запрос
- `hx-target`, `hx-swap` — указать цель и режим замены
- Все мутирующие действия (создание темы, публикация сообщения, ответ) — HTMX POST

### Alpine.js — клиентский UI-стейт

Используется **только** для состояния на клиенте без обращения к серверу:

- `x-show`, `x-data`, `x-on` — показать/скрыть элементы, счётчики, валидация форм
- Не использовать Alpine.js для запросов к серверу

**Правило**: не ставить `hx-*` и `x-*` на один элемент без явной причины.

### SSE — real-time обновления

- Клиент подписывается через `hx-ext="sse"` или нативный `EventSource`
- nginx обязательно: `proxy_buffering off` и заголовок `X-Accel-Buffering: no` для SSE-эндпоинтов
- Cursor-based догонялка: клиент передаёт `Last-Event-ID`, сервер досылает пропущенные события

### Известная ловушка HTMX + Alpine.js

Alpine.js не обновляет `x-show` реактивно для элементов, вставленных HTMX. Сбрасывать состояние нужно в `htmx:beforeSwap`, **не** в `htmx:afterRequest`.

## Localization

- Только русский язык; i18n-слоя нет и не планируется.
- Все тексты жёстко вшиты в шаблоны на русском.
- PostgreSQL full-text search использует конфигурацию `russian` (`ts_config`) для `tsvector`/`tsquery`.
- Новые ключи: добавлять прямо в шаблон на русском, без отдельного файла переводов.
