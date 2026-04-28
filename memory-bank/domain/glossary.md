---
title: Glossary
doc_kind: domain
doc_function: canonical
purpose: Единый словарь терминов проекта. Читать при неопределённости в значении термина, аббревиатуры или идентификатора.
derived_from:
  - ../dna/governance.md
status: active
audience: humans_and_agents
canonical_for:
  - term_definitions
  - identifier_vocabulary
---

# Glossary

## Предметная область LFC.ru

| Термин | Определение |
|---|---|
| **Раздел** | Категория форума верхнего уровня (например, «Матчи», «Трансферы»). Может быть открытым (доступен гостям) или закрытым (только для авторизованных). |
| **Тема** | Единица обсуждения внутри раздела. Создаётся авторизованным пользователем. |
| **Сообщение** (пост) | Отдельная публикация в теме. Первое сообщение открывает тему, последующие — ответы. |
| **Гость** | Неавторизованный посетитель. Может читать открытые разделы, не может публиковать. |
| **Пользователь** | Зарегистрированный, вошедший в систему субъект. Может создавать темы и писать сообщения. |
| **Модератор** | Пользователь с расширенными правами: управление темами и сообщениями в подопечных разделах. |
| **Администратор** | Максимальный уровень прав. Управляет структурой форума, пользователями, ролями. |
| **WF-NN** | Идентификатор core workflow из `domain/problem.md` (например, `WF-02` — участие в форуме). |

## Техническая область проекта

| Термин | Определение |
|---|---|
| **SSE** (Server-Sent Events) | Механизм real-time push-обновлений от сервера к браузеру. Используется для появления новых сообщений в форуме без перезагрузки страницы. |
| **LISTEN/NOTIFY** | PostgreSQL-механизм асинхронного уведомления: сервер подписывается (`LISTEN`) на канал и получает события от БД (`NOTIFY`). **Отложен** — применение к SSE запланировано для multi-pod масштабирования. Текущий MVP использует in-process hub (см. `architecture.md`). |
| **Fan-out** | Паттерн рассылки одного события всем активным подписчикам. MVP: `CreatePost` вызывает `hub.Broadcast` напрямую, hub рассылает HTML-фрагмент по каналам каждому SSE-соединению (in-process, без PG). |
| **Cursor-based догонялка** | При реконнекте клиент передаёт `Last-Event-ID`, сервер досылает пропущенные события по cursor из PG. |
| **CSRF-токен** | Защита от Cross-Site Request Forgery. Обязателен для всех POST/PUT/DELETE. Реализован через middleware. |
| **Rate limiting** | Ограничение частоты запросов на `/login` для защиты от брутфорса. Реализован через пакет `ratelimit`. |
| **Sentinel error** | Типизированная ошибка домена (например, `ErrDuplicateEmail`). Объявляется в пакете домена, маппится Repo-слоем из pgx-ошибок, проверяется через `errors.Is` в Handler. |
| **Parametrized query** | SQL-запрос с плейсхолдерами `$1`, `$2`. Единственно допустимый способ формирования SQL; `fmt.Sprintf` в SQL запрещён (`PCON-01`). |
| **goose** | Инструмент SQL-миграций. Файлы в `migrations/`, применяются через `goose.Up()` при старте приложения. |
| **pgxpool** | Пул соединений к PostgreSQL. Конкурентный доступ из горутин безопасен по умолчанию. |
| **Graceful shutdown** | Корректное завершение сервера: дожидается завершения in-flight запросов, отменяет context, останавливает cleanup-горутину. |
| **DI** (Dependency Injection) | Паттерн инициализации зависимостей. В проекте реализован вручную в `cmd/forum/main.go`: config → pool → repos → services → handlers. |
| **PCON-NN** | Идентификатор project constraint из `domain/problem.md` (например, `PCON-01` — запрет `fmt.Sprintf` в SQL). |
| **MET-NN** | Идентификатор outcome-метрики. |

## Memory-bank шаблон

| Термин | Определение |
|---|---|
| **SSoT** (Single Source of Truth) | Принцип: каждый факт имеет ровно одного canonical owner. Дублирование факта в двух местах — дефект документации. |
| **Governed document** | Markdown-файл в `memory-bank/` с валидным YAML frontmatter. Только такие документы участвуют в governance. |
| **canonical_for** | Поле frontmatter. Перечисляет факты, которыми владеет данный документ. При конфликте побеждает документ с явным `canonical_for`. |
| **derived_from** | Поле frontmatter. Указывает прямые upstream-документы. Authority течёт upstream → downstream; циклические зависимости запрещены. |
| **doc_kind** | Тип документа: `governance`, `domain`, `engineering`, `ops`, `feature`, `prd`, `use_case`, `adr`. |
| **doc_function** | Роль документа: `canonical` (owner факта), `index` (навигация), `template` (шаблон для инстанцирования). |
| **ADR** (Architecture Decision Record) | Документ, фиксирующий архитектурное решение: контекст, варианты, выбор и последствия. Живёт в `memory-bank/adr/`. |
| **Feature package** | Папка `memory-bank/features/FT-XXX/` со всеми документами одной фичи: `README.md`, `feature.md`, опционально `implementation-plan.md`. |
| **Vertical slice** | Единица пользовательской ценности, пронизывающая все затронутые слои (UI, API, storage, infra). Предпочтительный scope фичи. |
| **Grounding** | Обязательный этап перед составлением `implementation-plan.md`: агент исследует текущее состояние системы (relevant paths, patterns, dependencies) и фиксирует в discovery context. |
| **Delivery status** | Статус исполнения фичи: `planned` → `in_progress` → `done` / `cancelled`. Отделён от publication status (`status` в frontmatter). |
| **HANDOFF.md** | Файл в корне проекта. Передаёт контекст от одного агента к следующему. Обновляется после каждой сессии по шаблону из `CLAUDE.md`. |
| **PRD** | Product Requirements Document. Описывает отдельную продуктовую инициативу. Живёт в `memory-bank/prd/`. Не заменяет `domain/problem.md`. |
| **Use case** (`UC-*`) | Устойчивый пользовательский или операционный сценарий уровня проекта. Живёт в `memory-bank/use-cases/`. |

## Стабильные идентификаторы feature.md

| Prefix | Значение |
|---|---|
| `REQ-*` | Scope и обязательные capability |
| `NS-*` | Non-scope — явно исключённое из фичи |
| `ASM-*` | Assumptions и рабочие предпосылки |
| `CON-*` | Ограничения |
| `DEC-*` | Blocking decisions |
| `NT-*` | Do-not-touch / explicit change boundaries |
| `INV-*` | Инварианты |
| `CTR-*` | Контракты |
| `FM-*` | Failure modes |
| `RB-*` | Rollout / backout stages |
| `EC-*` | Exit criteria |
| `SC-*` | Acceptance scenarios |
| `NEG-*` | Negative / edge test cases |
| `CHK-*` | Проверки (в `feature.md` — acceptance-level; в `implementation-plan.md` — execution-level) |
| `EVID-*` | Evidence-артефакты (path к файлу, CI run, screenshot) |
| `RJ-*` | Rejection rules |

## Стабильные идентификаторы implementation-plan.md

| Prefix | Значение |
|---|---|
| `PRE-*` | Preconditions |
| `OQ-*` | Unresolved questions / ambiguities |
| `WS-*` | Workstreams |
| `AG-*` | Approval gates для рискованных действий |
| `STEP-*` | Атомарные шаги исполнения |
| `PAR-*` | Параллелизуемые блоки |
| `CP-*` | Checkpoints |
| `ER-*` | Execution risks |
| `STOP-*` | Stop conditions / fallback |
