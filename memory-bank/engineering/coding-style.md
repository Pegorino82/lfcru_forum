---
title: Coding Style
doc_kind: engineering
doc_function: canonical
purpose: Конвенции оформления кода проекта LFC.ru forum.
derived_from:
  - ../dna/governance.md
status: active
audience: humans_and_agents
---

# Coding Style

## Go (Backend)

- Форматтер: `gofmt` — обязателен, без исключений
- Ошибки — через sentinel-переменные: `var ErrXxx = errors.New(...)`; не используй `fmt.Errorf` как единственный механизм
- Имена файлов и пакетов — `snake_case`, структуры и интерфейсы — `CamelCase`
- Комментарии добавляются только там, где логика не очевидна (why или boundary condition)

## SQL

- Только параметризованные запросы (`$1`, `$2`); **никогда** `fmt.Sprintf` в SQL-строках
- Миграции — SQL-файлы через goose; именование `NNN_description.sql`

## Templates (`html/template`)

- Все пользовательские данные через `{{.}}` — автоэкранирование в HTML-контексте
- Не доверяй raw-строкам от пользователя; не использовать `template.HTML()` без явной проверки

## Frontend (HTMX + Alpine.js)

- **HTMX** — только запросы к серверу и подмена фрагментов DOM (`hx-get`, `hx-post`, `hx-swap`, `hx-target`)
- **Alpine.js** — только клиентский UI-стейт (`x-show`, `x-data`, `x-on`); без серверных вызовов
- Не смешивать `hx-*` и `x-*` на одном элементе без явной причины
- При HTMX outerHTML swap Alpine не обновляет реактивно — сбрасывай состояние в `htmx:beforeSwap`, не в `htmx:afterRequest`

## Change Discipline

- Не переписывай несвязанный код ради единообразия, если задача этого не требует
- При touch-up следуй локальному стилю файла, если нет явного конфликта с canonical rule
- Не добавляй docstrings, комментарии или type annotations к коду, который не менял
