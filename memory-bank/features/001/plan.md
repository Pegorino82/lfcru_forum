# 01_auth_basic — Implementation Plan

**Источник:** [rspec.md](rspec.md) | **Issue:** [#1](https://github.com/Pegorino82/lfcru_forum/issues/1)

---

## Обзор

Проект greenfield — нет ни Go-кода, ни Docker-файлов. План охватывает полную настройку проекта + реализацию auth (регистрация, вход, выход, сессии, rate-limiting).

**Итого: 9 шагов, ~25 файлов.**

---

## Шаг 0: Скаффолдинг проекта

**Цель:** Рабочий Go-сервер в Docker, подключённый к PostgreSQL.

**Файлы:**

```
go.mod
go.sum
cmd/forum/main.go              # точка входа: Echo server, подключение к PG, маршруты
internal/config/config.go       # загрузка конфигурации из ENV
Dockerfile                      # multi-stage: build + runtime
docker-compose.dev.yml          # app + postgres
.env.example                    # DATABASE_URL, APP_PORT, COOKIE_SECURE, ...
```

**Действия:**
1. `go mod init github.com/Pegorino82/lfcru_forum`
2. Зависимости: `echo/v4`, `pgx/v5`, `golang.org/x/crypto` (bcrypt), `google/uuid`
3. `cmd/forum/main.go` — инициализация Echo, подключение к PG через `pgxpool`, graceful shutdown
4. `internal/config/config.go` — структура `Config` с полями из ENV:
   - `DATABASE_URL`, `APP_PORT` (default 8080), `COOKIE_SECURE` (bool), `SESSION_LIFETIME` (default 720h), `BCRYPT_COST` (default 12), `RATE_LIMIT_WINDOW` (default 10m), `RATE_LIMIT_MAX` (default 5), `SESSION_GRACE_PERIOD` (default 5m), `MAX_SESSIONS_PER_USER` (default 10)
5. `Dockerfile` — multi-stage build
6. `docker-compose.dev.yml` — services: `app` (hot-reload через air или rebuild), `postgres` (volume для данных)

**Definition of Done:** `docker compose -f docker-compose.dev.yml up` — сервер стартует, отвечает 200 на `GET /health`

---

## Шаг 1: Миграции БД

**Цель:** Таблицы `users`, `sessions`, `login_attempts` готовы.

**Файлы:**

```
migrations/001_create_users.sql
migrations/002_create_sessions.sql
migrations/003_create_login_attempts.sql
```

**Действия:**
1. Установить goose в Dockerfile (или Go-вызов из `main.go`)
2. Миграции запускаются автоматически при старте приложения (goose up)
3. SQL — точно по спеке (§5 Модель данных):
   - `users`: id (BIGINT IDENTITY), username, email, pass_hash (BYTEA), role (default 'user'), is_active (BOOLEAN default true), created_at, updated_at
   - Индексы: `idx_users_email` (lower(email) UNIQUE), `idx_users_username` (lower(username) UNIQUE)
   - `sessions`: id (UUID, gen_random_uuid()), user_id (FK → users ON DELETE CASCADE), ip_addr (INET), user_agent (TEXT, DEFAULT ''), created_at, expires_at
   - Индексы: `idx_sessions_user_id`, `idx_sessions_expires_at`
   - `login_attempts`: id (BIGINT IDENTITY), ip_addr (INET), attempted_at (default now())
   - Индекс: `idx_login_attempts_ip_time` (ip_addr, attempted_at)

**Definition of Done:** После `goose up` — все таблицы и индексы существуют. `goose down` — откатываются.

---

## Шаг 2: Модели и репозитории

**Цель:** CRUD-операции для users, sessions, login_attempts.

**Файлы:**

```
internal/user/model.go
internal/user/repo.go
internal/session/model.go
internal/session/repo.go
internal/ratelimit/repo.go
```

### internal/user/model.go

```go
type User struct {
    ID        int64
    Username  string
    Email     string
    PassHash  []byte
    Role      string
    IsActive  bool
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### internal/user/repo.go — `UserRepo`

| Метод | SQL | Заметки |
|---|---|---|
| `Create(ctx, user) (int64, error)` | `INSERT INTO users (...) VALUES (...) RETURNING id` | Обработать duplicate key → конкретная ошибка (email/username) |
| `GetByEmail(ctx, email) (*User, error)` | `SELECT ... WHERE lower(email) = lower($1)` | |
| `GetByID(ctx, id) (*User, error)` | `SELECT ... WHERE id = $1` | |

**Обработка UNIQUE violation:** парсить `pgconn.PgError` (code `23505`), по имени constraint определять, какое поле дублируется. Определить кастомные ошибки:
- `ErrDuplicateEmail`
- `ErrDuplicateUsername`

### internal/session/model.go

```go
type Session struct {
    ID        uuid.UUID
    UserID    int64
    IPAddr    string
    UserAgent string
    CreatedAt time.Time
    ExpiresAt time.Time
}
```

### internal/session/repo.go — `SessionRepo`

| Метод | SQL |
|---|---|
| `Create(ctx, session) (uuid.UUID, error)` | `INSERT ... RETURNING id` |
| `GetByID(ctx, id) (*Session, error)` | `SELECT ... WHERE id = $1 AND expires_at > now()` |
| `Delete(ctx, id) error` | `DELETE FROM sessions WHERE id = $1` |
| `Touch(ctx, id, newExpiry) error` | `UPDATE sessions SET expires_at = $1 WHERE id = $2` |
| `CountByUser(ctx, userID) (int, error)` | `SELECT count(*) FROM sessions WHERE user_id = $1 AND expires_at > now()` |
| `DeleteOldestByUser(ctx, userID) error` | `DELETE FROM sessions WHERE id = (SELECT id FROM sessions WHERE user_id = $1 ORDER BY created_at ASC LIMIT 1)` |
| `DeleteExpired(ctx) (int64, error)` | `DELETE FROM sessions WHERE expires_at < now()` |

### internal/ratelimit/repo.go — `LoginAttemptRepo`

| Метод | SQL |
|---|---|
| `Record(ctx, ip) error` | `INSERT INTO login_attempts (ip_addr) VALUES ($1)` |
| `Count(ctx, ip, window) (int, error)` | `SELECT count(*) ... WHERE ip_addr = $1 AND attempted_at > now() - $2` |
| `Cleanup(ctx) (int64, error)` | `DELETE ... WHERE attempted_at < now() - interval '1 hour'` |

**Definition of Done:** Все методы покрыты интеграционными тестами (testcontainers + PostgreSQL).

---

## Шаг 3: Сервис аутентификации

**Цель:** Бизнес-логика регистрации, входа, выхода.

**Файлы:**

```
internal/auth/service.go
internal/auth/service_test.go
internal/auth/errors.go
```

### internal/auth/service.go — `AuthService`

**Зависимости (интерфейсы):**
- `UserRepo` (Create, GetByEmail, GetByID)
- `SessionRepo` (Create, Delete, Touch)
- `LoginAttemptRepo` (Record, Count)
- `Config` (BcryptCost, SessionLifetime)

**Методы:**

#### `Register(ctx, input, ip) (*User, *Session, error)`

1. Rate-limit check: `attemptRepo.Count(ip, window)` → если ≥ max, вернуть `ErrRateLimited`
2. Валидация (вынесена в `validateRegisterInput`):
   - username: 3–30 символов, regex `^[a-zA-Z0-9_-]+$`
   - email: trim + lower, формат, длина ≤ 254
   - password: ≥ 8 символов, ≤ 72 байт
   - password_confirm == password
3. `bcrypt.GenerateFromPassword(password, cost)`
4. `userRepo.Create(...)` — при duplicate key → маппинг на понятную ошибку
5. Политика множественных сессий: `sessionRepo.CountByUser(userID)` → если ≥ 10, `sessionRepo.DeleteOldestByUser(userID)`
6. `sessionRepo.Create(...)` — сессия с expires_at = now() + 30d
7. Вернуть user + session

#### `Login(ctx, email, password, ip) (*User, *Session, error)`

1. Rate-limit check: `attemptRepo.Count(ip, 10min)` → если ≥ 5, вернуть `ErrRateLimited`
2. `userRepo.GetByEmail(email)` → если не найден:
   - **Dummy bcrypt compare** (timing attack protection)
   - `attemptRepo.Record(ip)`
   - Вернуть `ErrInvalidCredentials`
3. `bcrypt.CompareHashAndPassword(user.PassHash, password)` → при ошибке:
   - `attemptRepo.Record(ip)`
   - Вернуть `ErrInvalidCredentials`
4. Политика множественных сессий: `sessionRepo.CountByUser(userID)` → если ≥ 10, `sessionRepo.DeleteOldestByUser(userID)`
5. `sessionRepo.Create(...)` — сессия
6. Вернуть user + session

#### `Logout(ctx, sessionID) error`

1. `sessionRepo.Delete(sessionID)`

#### `GetSession(ctx, sessionID) (*User, *Session, error)`

1. `sessionRepo.GetByID(sessionID)` → если не найдена/истекла → `ErrSessionNotFound`
2. `userRepo.GetByID(session.UserID)`
3. `sessionRepo.Touch(sessionID, now() + 30d)` — **с grace period**: обновлять только если до expiry < 29d 23h 55min (≈ раз в 5 минут)
4. Вернуть user + session

### internal/auth/errors.go

```go
var (
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrRateLimited        = errors.New("rate limited")
    ErrDuplicateEmail     = errors.New("duplicate email")
    ErrDuplicateUsername  = errors.New("duplicate username")
    ErrSessionNotFound    = errors.New("session not found")
)

type ValidationErrors map[string]string  // field → message
```

**Definition of Done:** Юнит-тесты с моками репозиториев. Проверки: валидация, bcrypt, rate-limit (login + register), timing attack (dummy hash), grace period (5 мин), политика множественных сессий (max 10).

---

## Шаг 4: CSRF Middleware

**Цель:** Защита POST/PUT/DELETE от CSRF.

**Файлы:**

```
internal/middleware/csrf.go
```

**Реализация:**
1. Генерация CSRF-токена (crypto/rand, 32 байта, base64)
2. Токен хранится в cookie (`_csrf`, HttpOnly=false, SameSite=Strict)
3. Для POST/PUT/DELETE — проверка `_csrf` из form data vs cookie
4. При несовпадении — 403 Forbidden
5. Токен доступен в шаблонах через `c.Get("csrf_token")`

**Вариант:** использовать `echo.middleware.CSRF()` (встроенный в Echo). Проверить, подходит ли — если да, использовать его вместо кастомного.

**Definition of Done:** POST без валидного CSRF → 403. POST с валидным → проходит.

---

## Шаг 5: Session Middleware

**Цель:** Загрузка пользователя из cookie на каждый запрос.

**Файлы:**

```
internal/auth/middleware.go
```

**Реализация:**

1. Читает cookie `session_id`
2. Если нет → `c.Set("user", nil)`, next
3. `authService.GetSession(ctx, sessionID)`:
   - Успех → `c.Set("user", user)`, next
   - Ошибка → удалить cookie, `c.Set("user", nil)`, next
4. Хелпер `UserFromContext(c) *User` — достаёт user из контекста (или nil для гостя)
5. Middleware `RequireAuth` — если user == nil → redirect на `/login?next=<current_path>`

**Definition of Done:** Запрос с валидной сессией — user в контексте. Истёкшая сессия — cookie удалён, гостевой запрос.

---

## Шаг 6: HTTP Handlers

**Цель:** Эндпоинты регистрации, входа, выхода.

**Файлы:**

```
internal/auth/handler.go
internal/auth/handler_test.go
```

### Маршруты

```go
e.GET("/register", h.ShowRegister)
e.POST("/register", h.Register)
e.GET("/login", h.ShowLogin)
e.POST("/login", h.Login)
e.POST("/logout", h.Logout)
```

### handler.go — `AuthHandler`

**Зависимости:** `AuthService`, `template.Templates` (рендерер)

#### `ShowRegister` / `ShowLogin`

- Если пользователь авторизован → redirect на `/`
- Иначе → рендер формы
- HX-Request check: фрагмент vs полная страница

#### `Register` (POST)

1. Парсинг формы (username, email, password, password_confirm)
2. Rate-limit check (по IP, через сервис)
3. `authService.Register(...)`:
   - `ValidationErrors` → 422, ре-рендер формы с ошибками inline, данные полей сохранены (кроме паролей)
   - `ErrDuplicateEmail` / `ErrDuplicateUsername` → 409, ре-рендер с ошибкой
   - `ErrRateLimited` → 429
   - Успех → Set-Cookie `session_id`, flash «Регистрация прошла успешно», redirect 303 → `/`

#### `Login` (POST)

1. Парсинг формы (email, password)
2. `authService.Login(ctx, email, password, ip)`:
   - `ErrRateLimited` → 429
   - `ErrInvalidCredentials` → 422, ре-рендер формы, email сохранён, сообщение «Неверный email или пароль»
   - Успех → Set-Cookie, flash «Вы вошли в систему», redirect 303 → `?next` (валидация: относительный путь, не /login, /register, /logout) или `/`

#### `Logout` (POST)

1. `authService.Logout(ctx, sessionID)`
2. Удалить cookie (MaxAge = -1)
3. Redirect 303 → `/`

### Set-Cookie helper

```go
func setSessionCookie(c echo.Context, sessionID string, secure bool) {
    c.SetCookie(&http.Cookie{
        Name:     "session_id",
        Value:    sessionID,
        Path:     "/",
        MaxAge:   2592000, // 30 дней
        HttpOnly: true,
        Secure:   secure,
        SameSite: http.SameSiteLaxMode,
    })
}
```

**Definition of Done:** Интеграционные тесты через `httptest` + testcontainers. Все сценарии из §10 спеки.

---

## Шаг 7: Шаблоны

**Цель:** HTML-формы для регистрации и входа.

**Файлы:**

```
templates/layouts/base.html         # + блок flash-сообщений + nav (гость: Вход/Регистрация, юзер: username/Выход)
templates/auth/register.html
templates/auth/login.html
```

### base.html

- `<html>`, `<head>` (HTMX CDN, Alpine.js CDN, базовые стили), `<body>`
- `{{template "nav" .}}`
- Блок flash-сообщений: проверка наличия flash-cookie, отображение и удаление (одноразовое)
- `{{block "content" .}}{{end}}`

### register.html

- Форма: username, email, password, password_confirm, `_csrf` (hidden)
- `hx-post="/register"` + `hx-target` на контейнер формы (для подмены при ошибке)
- Ошибки inline: `<span class="field-error">` под каждым полем
- `aria-invalid="true"` + `aria-describedby` для доступности
- `label` + `for` для каждого поля
- При ре-рендере с ошибками — `autofocus` на первое поле с ошибкой

### login.html

- Форма: email, password, `_csrf` (hidden)
- `hx-post="/login"` + `hx-target`
- Ошибка «Неверный email или пароль» — общая, над формой

### HX-Request логика

В handler: если заголовок `HX-Request` присутствует — рендерить только `content` block (фрагмент), иначе — полную страницу через `base.html`.

**Definition of Done:** Формы рендерятся, ошибки отображаются inline, HTMX-запросы получают фрагменты.

---

## Шаг 8: Фоновые задачи

**Цель:** Очистка истёкших сессий и старых login_attempts.

**Файлы:**

```
internal/cleanup/cleanup.go
```

**Реализация:**

1. Горутина с `time.Ticker`:
   - Каждый **час** → `sessionRepo.DeleteExpired()`
   - Каждые **10 минут** → `attemptRepo.Cleanup()`
2. Запускается из `main.go`, принимает `context.Context` для graceful shutdown
3. Логирование количества удалённых записей

**Definition of Done:** Истёкшие сессии и старые попытки удаляются автоматически.

---

## Шаг 9: Интеграционные тесты

**Цель:** End-to-end проверка всех сценариев из §10 спеки.

**Файлы:**

```
internal/auth/integration_test.go
internal/ratelimit/repo_test.go
```

**Инфраструктура:** testcontainers-go (PostgreSQL контейнер для тестов)

**Тест-кейсы (из спеки §10):**

| # | Сценарий | Ожидание |
|---|---|---|
| 1 | Регистрация с валидными данными | 303, cookie `session_id`, запись в `users` |
| 2 | Регистрация с занятым email | 409, сообщение «Пользователь с таким email уже зарегистрирован» |
| 3 | Регистрация с занятым username | 409, «Это имя уже занято» |
| 4 | Регистрация с невалидными данными | 422, ошибки inline, данные полей сохранены |
| 5 | Вход с верными credentials | 303, cookie, запись в `sessions` |
| 6 | Вход с неверным паролем | 422, «Неверный email или пароль», запись в `login_attempts` |
| 7 | 6 неудачных попыток входа за 10 мин | 429 на 6-й попытке |
| 8 | Выход | 303, cookie удалён, сессия удалена из БД |
| 9 | Запрос с валидной сессией | user доступен в контексте |
| 10 | Запрос с истёкшей сессией | гостевой запрос, cookie удалён |
| 11 | GET /register будучи авторизованным | redirect → `/` |
| 12 | GET /login будучи авторизованным | redirect → `/` |
| 13 | Login redirect: ?next=/topics | redirect → `/topics` |
| 14 | Login redirect: ?next=http://evil.com | redirect → `/` (игнор) |
| 15 | CSRF: POST без токена | 403 |
| 16 | Rate-limit на регистрацию | 429 после 5 попыток |

**Юнит-тесты:**

| # | Компонент | Что проверяем |
|---|---|---|
| 1 | Валидация email | Корректные/некорректные форматы |
| 2 | Валидация пароля | Граничные: 7, 8, 72, 73 байт |
| 3 | Валидация username | Граничные: 2, 3, 30, 31 символ, спецсимволы |
| 4 | bcrypt | Hash создаётся и верифицируется |
| 5 | Redirect validation | Относительные пути, абсолютные URL, /login, /register |

**Definition of Done:** Все тесты зелёные в CI (docker compose run tests).

---

## Порядок реализации и зависимости

```
Шаг 0: Скаффолдинг ──→ Шаг 1: Миграции ──→ Шаг 2: Репозитории ──→ Шаг 3: Сервис
                                                                         │
                                                                         ▼
                        Шаг 7: Шаблоны ◄── Шаг 6: Handlers ◄── Шаг 5: Session MW
                                                                    │
                                                                    ▼
                                                              Шаг 4: CSRF MW
                                                                    │
                                                                    ▼
                                                        Шаг 8: Фоновые задачи
                                                                    │
                                                                    ▼
                                                     Шаг 9: Интеграционные тесты
```

**Рекомендация:** реализовывать строго по шагам 0→9. Каждый шаг — отдельный коммит.

---

## Полный список файлов

```
cmd/forum/main.go
internal/config/config.go
internal/user/model.go
internal/user/repo.go
internal/session/model.go
internal/session/repo.go
internal/ratelimit/repo.go
internal/ratelimit/repo_test.go
internal/auth/service.go
internal/auth/service_test.go
internal/auth/errors.go
internal/auth/handler.go
internal/auth/handler_test.go
internal/auth/middleware.go
internal/auth/integration_test.go
internal/middleware/csrf.go
internal/cleanup/cleanup.go
migrations/001_create_users.sql
migrations/002_create_sessions.sql
migrations/003_create_login_attempts.sql
templates/layouts/base.html
templates/auth/register.html
templates/auth/login.html
Dockerfile
docker-compose.dev.yml
.env.example
go.mod
```

**Итого: 26 файлов, 9 шагов.**
