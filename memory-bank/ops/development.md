---
title: Development Environment
doc_kind: ops
doc_function: canonical
purpose: Локальная разработка LFC.ru — запуск приложения, тестов и работа с базой данных.
derived_from:
  - ../dna/governance.md
status: active
audience: humans_and_agents
---

# Development Environment

Всё работает в Docker. Go на хосте не нужен.

## Setup

```bash
docker compose -f docker-compose.dev.yml up
```

Postgres поднимается первым (healthcheck), затем app. При первом старте postgres автоматически создаёт тестовую БД `lfcru_test` через `scripts/init-test-db.sql`.

Миграции применяются автоматически при старте app (`goose.Up()` в `main.go`).

## Daily Commands

```bash
# Запуск (dev режим)
docker compose -f docker-compose.dev.yml up

# Запуск в фоне
docker compose -f docker-compose.dev.yml up -d

# Пересборка образа (после изменений в Go-коде)
docker compose -f docker-compose.dev.yml up --build

# Логи приложения
docker compose -f docker-compose.dev.yml logs -f app
```

## Browser Testing

Приложение доступно по адресу: `http://localhost:8080`

## Running Tests

App-контейнер — бинарный образ без Go. Тесты запускаются отдельным golang-контейнером.

Зависимости Go кэшируются в именованном Docker-volume `lfcru_gomod` — скачиваются один раз, переиспользуются во всех запусках.

```bash
# Шаг 0: скачать зависимости в кэш (один раз после клонирования и при изменении go.mod)
docker run --rm \
  -v "$(pwd)":/app -w /app \
  -v lfcru_gomod:/root/go/pkg/mod \
  golang:1.23-alpine \
  go mod download

# Юнит-тесты (без БД)
docker run --rm \
  -v "$(pwd)":/app -w /app \
  -v lfcru_gomod:/root/go/pkg/mod \
  golang:1.23-alpine \
  go test ./...

# Интеграционные тесты (требуется запущенная БД из docker-compose.dev.yml)
docker run --rm \
  -v "$(pwd)":/app -w /app \
  -v lfcru_gomod:/root/go/pkg/mod \
  --network lfcru_forum_default \
  -e DATABASE_URL="postgres://postgres:postgres@postgres:5432/lfcru_test?sslmode=disable" \
  golang:1.23-alpine \
  go test -tags integration -p 1 ./internal/...

# Один пакет (пример)
docker run --rm \
  -v "$(pwd)":/app -w /app \
  -v lfcru_gomod:/root/go/pkg/mod \
  --network lfcru_forum_default \
  -e DATABASE_URL="postgres://postgres:postgres@postgres:5432/lfcru_test?sslmode=disable" \
  golang:1.23-alpine \
  go test -tags integration -v ./internal/auth/...
```

> Флаг `-p 1` обязателен для интеграционных тестов: каждый пакет вызывает `goose.Up()`, параллельный запуск вызывает race condition.

> `vendor/` не хранится в репозитории (добавлен в `.gitignore`). Кэш-volume `lfcru_gomod` заменяет его — персистентен между запусками контейнеров на одной машине.

## E2E Tests (Playwright)

Node.js на хосте не нужен — тесты запускаются в официальном Docker-образе Playwright.
Требуется запущенный dev-стек (`docker-compose.dev.yml`) и e2e-контейнер приложения.

```bash
# Шаг 1: поднять dev-стек (postgres + app на 8080), если ещё не запущен
docker compose -f docker-compose.dev.yml up -d

# Шаг 2: поднять e2e-контейнер (app на 8081 → lfcru_test)
docker compose -f docker-compose.e2e.yml up -d

# Шаг 3: запустить тесты
docker run --rm \
  -v "$(pwd)":/app -w /app \
  --network lfcru_forum_default \
  -e PW_BASE_URL=http://app-e2e:8080 \
  -e PW_DB_HOST=postgres \
  mcr.microsoft.com/playwright:v1.59.1-jammy \
  npx playwright test
```

Playwright-контейнер подключается к сети `lfcru_forum_default` — видит `app-e2e:8080`
и `postgres:5432` напрямую, без проброса портов на хост.

Артефакты (скриншоты при падении, HTML-отчёт) записываются в `e2e/test-results/` и
`e2e/test-report/` на хосте через смонтированный volume (`-v "$(pwd)":/app`).

## Database And Services

| Сервис | Host (из контейнера) | Host (с хоста) | БД |
|--------|---------------------|----------------|-----|
| postgres | `postgres:5432` | `localhost:5432` | `lfcru` (dev), `lfcru_test` (tests) |
| app | — | `localhost:8080` | `lfcru` |
| app-e2e | `app-e2e:8080` | `localhost:8081` | `lfcru_test` |

Docker-сеть: `lfcru_forum_default` (автоматически из имени папки проекта).

Credentials dev: `postgres/postgres`.

Миграции — SQL-файлы в `migrations/`, применяются через goose. Новая миграция: добавить файл `migrations/NNN_description.sql` в формате goose.
