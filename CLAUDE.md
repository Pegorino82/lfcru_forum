# CLAUDE.md

## Проект

**LFC.ru** — русскоязычный фан-сайт и форум болельщиков ФК «Ливерпуль».

| Слой | Технология |
|---|---|
| Backend | Go + Echo |
| Шаблоны | `html/template` (stdlib) |
| Frontend | HTMX + Alpine.js |
| База данных | PostgreSQL (`pgx` драйвер) |
| Миграции | goose (SQL-файлы) |
| Real-time | SSE (stdlib) |
| Сессии | PostgreSQL (httponly + secure cookie) |
| Контейнеры | Docker + docker compose |
| Reverse proxy | nginx |

---

## Общее

- Отвечай на **русском языке**
- Каждый сеанс решает **ровно одну задачу** — не рефакторь без запроса

---

## Memory Bank

Каноничная проектная знания — в `memory-bank/`. Читай нужный раздел вместо того, чтобы угадывать.

| Что нужно | Куда идти |
|---|---|
| Продукт, роли, workflows | `memory-bank/domain/problem.md` |
| Архитектура, слои, модули | `memory-bank/domain/architecture.md` |
| Шаблоны, HTMX, Alpine.js | `memory-bank/domain/frontend.md` |
| Конвенции кода (Go, SQL, Templates) | `memory-bank/engineering/coding-style.md` |
| Тест-политика, sufficient coverage | `memory-bank/engineering/testing-policy.md` |
| Автономия, эскалация, супервизия | `memory-bank/engineering/autonomy-boundaries.md` |
| Запуск, тесты, Docker-команды | `memory-bank/ops/development.md` |
| Env vars, конфигурация, секреты | `memory-bank/ops/config.md` |
| Production, staging, окружения | `memory-bank/ops/stages.md` |
| Релиз, деплой, rollback | `memory-bank/ops/release.md` |
| Маршрутизация задач, типы workflow | `memory-bank/flows/workflows.md` |
| Feature lifecycle, gates | `memory-bank/flows/feature-flow.md` |
| Реализованные фичи | `memory-bank/features/README.md` |

---

## Рабочий процесс

**В начале сеанса:**
1. Прочитай `HANDOFF.md` в корне проекта (если существует) — там контекст от предыдущего агента
2. Прочитай задачу до конца
3. Найди затронутые файлы, прочитай их

**После кода:**
1. Запусти тесты (unit + integration для затронутых пакетов)
2. Убедись, что тесты зелёные
3. **Simplify review** — нет ли premature abstractions, dead code, дублирования логики
4. Сделай коммит (conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`)
5. Обнови `HANDOFF.md` в корне проекта

---

## Handoff

После каждого сеанса обновляй `HANDOFF.md` по шаблону:

```markdown
## Что сделано
- <краткий список>

## Что сделать следующим
- <конкретные шаги>

## Проблемы и решения
- <проблема> → <как решили>
```

Файл предназначен для следующего агента — пиши коротко и конкретно.
