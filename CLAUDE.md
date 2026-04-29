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

Каноничные проектные знания — в `memory-bank/`. Читай нужный раздел вместо того, чтобы угадывать.

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
| Маршрутизация задач, типы workflow, session protocol | `memory-bank/flows/workflows.md` |
| Feature lifecycle, gates | `memory-bank/flows/feature-flow.md` |
| Eval-чеклисты, evaluator agent, gate eval | `memory-bank/flows/eval.md` |
| Trello: workflow trigger, маппинг, lifecycle | `memory-bank/flows/trello.md` |
| Trello: board ID, list IDs (TODO/IN PROGRESS/DONE) | `memory-bank/ops/trello-board.md` |
| Git workflow, коммиты, PR | `memory-bank/engineering/git-workflow.md` |
| Реализованные фичи | `memory-bank/features/README.md` |

---

## Рабочий процесс

→ [`memory-bank/flows/workflows.md`](memory-bank/flows/workflows.md) — session protocol (начало и конец сеанса), маршрутизация задач по типам, шаблон HANDOFF.md.

<!-- rtk-instructions v2 -->
# RTK (Rust Token Killer) - Token-Optimized Commands

## RTK Commands by Workflow

### Build & Compile (80-90% savings)
```bash
rtk cargo clippy        # Clippy warnings grouped by file (80%)
rtk tsc                 # TypeScript errors grouped by file/code (83%)
rtk lint                # ESLint/Biome violations grouped (84%)
```

### Test (60-99% savings)
```bash
rtk cargo test          # Cargo test failures only (90%)
rtk go test             # Go test failures only (90%)
rtk jest                # Jest failures only (99.5%)
rtk playwright test     # Playwright failures only (94%)
rtk test <cmd>          # Generic test wrapper - failures only
```

### Analysis & Debug (70-90% savings)
```bash
rtk err <cmd>           # Filter errors only from any command
rtk log <file>          # Deduplicated logs with counts
rtk summary <cmd>       # Smart summary of command output
```

### Meta Commands
```bash
rtk gain                # View token savings statistics
rtk gain --history      # View command history with savings
rtk discover            # Analyze Claude Code sessions for missed RTK usage
rtk proxy <cmd>         # Run command without filtering (for debugging)
rtk init                # Add RTK instructions to CLAUDE.md
rtk init --global       # Add RTK to ~/.claude/CLAUDE.md
```
<!-- /rtk-instructions -->