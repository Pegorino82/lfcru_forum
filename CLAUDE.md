# CLAUDE.md

## Проект

**LFC.ru** — русскоязычный фан-сайт и форум болельщиков ФК «Ливерпуль».

Две части:
- **Сайт**: главная, новости, трансферы, турнирная таблица, календарь, архив статей
- **Форум**: структура «разделы → темы → сообщения»; гости читают открытые разделы, авторизованные — пишут

Роли пользователей: Гость → Пользователь → Модератор → Администратор

Аутентификация: регистрация по email/паролю, вход/выход, восстановление пароля

---

## Стек

| Слой | Технология |
|---|---|
| Backend | Go + Echo |
| Шаблоны | `html/template` (stdlib) |
| Frontend | HTMX + Alpine.js |
| База данных | PostgreSQL (`pgx` драйвер) |
| Миграции | goose (SQL-файлы) |
| Real-time | SSE (stdlib) + cursor-based догонялка из PG |
| Сессии | PostgreSQL (httponly + secure cookie) |
| Поиск | PostgreSQL `tsvector` + GIN-индекс |
| Контейнеры | Docker + docker compose |
| Reverse proxy | nginx |

**Ключевые принципы:**
- SQL только через параметризованные запросы — никогда `fmt.Sprintf` в SQL
- `html/template` с ручным экранированием — все пользовательские данные через `{{.}}` (автоэкранирование в HTML-контексте)
- CSRF-токен для всех POST/PUT/DELETE через middleware
- Rate-limiting на `/login` против брутфорса

### HTMX vs Alpine.js — разделение ответственности

- **HTMX** — запросы к серверу и подмена фрагментов DOM (`hx-get`, `hx-post`, `hx-swap`, `hx-target`)
- **Alpine.js** — только клиентский UI-стейт (показать/скрыть, счётчики, валидация форм, `x-show`, `x-data`, `x-on`)
- Не смешивать `hx-*` и `x-*` на одном элементе без явной причины

### SSE — заметки по реализации

- Fan-out broadcaster: одна горутина слушает PG (`LISTEN/NOTIFY`), рассылает по каналам всем подписчикам
- Cursor-based догонялка при реконнекте: клиент передаёт `Last-Event-ID`, сервер досылает пропущенные события из PG
- `LISTEN/NOTIFY` payload ≤ 8000 байт — передавать только ID события, данные клиент получит через cursor
- nginx: обязательно `proxy_buffering off;` и `X-Accel-Buffering: no` для SSE-эндпоинтов
- Cleanup: периодическая очистка старых событий из таблицы (cron / pg_cron)

### Полнотекстовый поиск

- Колонка `search_vector tsvector` в таблицах сообщений и статей
- GIN-индекс по `search_vector`
- Триггер на INSERT/UPDATE для автоматического обновления вектора
- Конфигурация: `russian` для ts_config

---

## Архитектура

### Структура каталогов

```
cmd/forum/main.go          — точка входа: конфиг, pgxpool, миграции, DI, Echo, graceful shutdown

internal/
├── config/config.go       — загрузка конфигурации из env-переменных (с дефолтами)
├── auth/
│   ├── handler.go         — HTTP-хэндлеры (Register, Login, Logout, ShowRegister, ShowLogin)
│   ├── service.go         — бизнес-логика: валидация, bcrypt, rate-limit, создание сессий
│   ├── errors.go          — sentinel-ошибки (ErrDuplicateEmail, ErrRateLimited и т.д.)
│   ├── middleware.go       — LoadSession (cookie → user в контексте), RequireAuth, UserFromContext
│   ├── service_test.go    — юнит-тесты сервиса (моки репозиториев)
│   └── integration_test.go — интеграционные тесты (testcontainers + PostgreSQL)
├── user/
│   ├── model.go           — структура User
│   └── repo.go            — UserRepo (Create, GetByEmail, GetByID) + маппинг unique violation
├── session/
│   ├── model.go           — структура Session (UUID, UserID, IP, UA, ExpiresAt)
│   └── repo.go            — SessionRepo (Create, GetByID, Delete, Touch, CountByUser, DeleteOldestByUser, DeleteExpired)
├── ratelimit/
│   ├── repo.go            — LoginAttemptRepo (Record, Count, Cleanup)
│   └── repo_test.go       — интеграционные тесты rate-limit
├── cleanup/cleanup.go     — фоновая горутина: очистка истёкших сессий (1ч) и старых login_attempts (10м)
├── middleware/csrf.go     — CSRF-middleware (токен в cookie + проверка POST/PUT/DELETE)
└── tmpl/renderer.go       — кастомный echo.Renderer: изолированный template set на каждую страницу

migrations/                — SQL-миграции (goose): 001_users, 002_sessions, 003_login_attempts

templates/
├── layouts/base.html      — базовый layout: nav, flash, content-блок, HTMX + Alpine.js
└── auth/
    ├── login.html         — форма входа
    └── register.html      — форма регистрации
```

### Слои и зависимости

```
Handler (HTTP) → Service (бизнес-логика) → Repo (SQL/pgx) → PostgreSQL
```

- **Handler** — парсит запрос, вызывает Service, рендерит шаблон или редирект. Зависит только от Service.
- **Service** — принимает интерфейсы репозиториев (`UserRepo`, `SessionRepo`, `AttemptRepo`). Содержит валидацию, хеширование, rate-limit логику, политику сессий.
- **Repo** — прямые SQL-запросы через `pgxpool.Pool`. Маппит PG-ошибки в доменные sentinel-ошибки.
- **Middleware** — `LoadSession` (загрузка пользователя из cookie), `RequireAuth` (редирект на /login), `CSRF`.

### DI и инициализация (cmd/forum/main.go)

1. `config.Load()` — env-переменные с дефолтами
2. `pgxpool.New()` — пул соединений к PostgreSQL
3. `goose.Up()` — автоматические миграции при старте
4. Создание репозиториев → Service → Handler
5. Echo: middleware (Logger, Recover, CSRF, LoadSession) → маршруты
6. `cleanup.Run()` — фоновая горутина очистки
7. Graceful shutdown по SIGINT/SIGTERM

### Шаблонизация

Кастомный `tmpl.Renderer` создаёт **изолированный `*template.Template` на каждый page-файл** (layouts парсятся в каждый set). Это предотвращает конфликты `{{define}}` блоков между страницами. Хэндлер рендерит по полному пути: `"templates/auth/login.html"`.

### Реализованные фичи

| # | Фича | Spec | Статус |
|---|---|---|---|
| 001 | Аутентификация (регистрация, вход, выход, сессии, rate-limit) | `memory-bank/features/001/rspec.md` | Реализовано |
| 002 | Базовый layout (хэдер с навигацией, футер, sticky footer, skip-link, a11y) | `memory-bank/features/002/rspec.md` | В работе |

---

## Общее
- Отвечай на **русском языке**
- Каждый сеанс решает **ровно одну задачу** — не рефакторь без запроса

## Окружение

Всё запускается в **Docker**. На хосте — только редактор и `docker` / `docker compose`.

| Режим      | Команда                                        |
|------------|------------------------------------------------|
| Разработка | `docker compose -f docker-compose.dev.yml up`  |
| Продакшн   | `docker compose -f docker-compose.prod.yml up` |

Деплой: VPS + docker compose + nginx (reverse proxy).

### Запуск тестов

App-контейнер — бинарный образ без Go. Тесты запускаются через отдельный `golang`-контейнер с монтированием исходников.

**Юнит-тесты** (без БД):
```bash
docker run --rm \
  -v "$(pwd)":/app -w /app \
  golang:1.23-alpine \
  go test ./...
```

**Интеграционные тесты** (нужна запущенная БД из `docker-compose.dev.yml`):
```bash
docker run --rm \
  -v "$(pwd)":/app -w /app \
  --network lfcru_forum_default \
  -e DATABASE_URL="postgres://postgres:postgres@postgres:5432/lfcru?sslmode=disable" \
  golang:1.23-alpine \
  go test -tags integration ./internal/...
```

**Только один пакет** (пример — layout):
```bash
docker run --rm \
  -v "$(pwd)":/app -w /app \
  --network lfcru_forum_default \
  -e DATABASE_URL="postgres://postgres:postgres@postgres:5432/lfcru?sslmode=disable" \
  golang:1.23-alpine \
  go test -tags integration -v ./internal/layout/...
```

> Имя сети `lfcru_forum_default` формируется автоматически из имени директории проекта (`lfcru_forum`) и суффикса `_default`.

## Рабочий процесс

**В начале сеанса:**
1. Прочитай `CLAUDE.md` и `PROJECT.md`
2. Прочитай `HANDOFF.md` в корне проекта (если существует) — там контекст от предыдущего агента
3. Прочитай задачу до конца
4. Найди затронутые файлы, прочитай их

**После кода:**
1. Запусти тесты внутри контейнера
2. Убедись, что тесты зелёные
3. Сделай коммит
4. Обнови `HANDOFF.md` в корне проекта (создай, если нет)

## Файл передачи HANDOFF.md

После каждого сеанса обновляй `HANDOFF.md` в корне проекта по шаблону:

```markdown
## Что сделано
- <краткий список выполненных изменений>

## Что сделать следующим
- <ссылка на spec>
- <конкретные задачи или шаги>

## Проблемы и решения
- <проблема> → <как решили>
```

Файл предназначен для следующего агента — пиши коротко и конкретно.
