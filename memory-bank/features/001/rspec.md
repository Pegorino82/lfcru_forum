# 01_auth_basic — Спецификация (SDD)

## 1. Контекст

Первая функция MVP форума LFC.ru. Без аутентификации невозможна ролевая модель, модерация и любое пользовательское взаимодействие. Задача охватывает регистрацию, вход, выход, поддержание сессии и защиту от брутфорса.

**Ссылка на бриф:** [brief.md](brief.md)
**GitHub Issue:** [#1](https://github.com/Pegorino82/lfcru_forum/issues/1)

---

## 2. Акторы

| Актор | Описание |
|---|---|
| Гость | Незарегистрированный посетитель. Может просматривать открытые разделы |
| Пользователь | Зарегистрированный аккаунт. Может создавать темы и сообщения |

---

## 3. Сценарии использования (Use Cases)

### UC-1: Регистрация

**Актор:** Гость

**Предусловие:** Гость не авторизован.

**Основной поток:**

1. Гость открывает страницу `/register`
2. Система отображает форму: username, email, пароль, подтверждение пароля
3. Гость заполняет форму и отправляет
4. Система валидирует данные (см. [§4 Валидация](#4-правила-валидации))
5. Система создаёт запись пользователя с хешированным паролем
6. Система создаёт сессию и устанавливает cookie
7. Система устанавливает flash-сообщение «Регистрация прошла успешно»
8. Система перенаправляет на главную страницу форума

**Альтернативные потоки:**

- **3a.** Email уже занят → ошибка «Пользователь с таким email уже зарегистрирован»
- **3b.** Username уже занят → ошибка «Это имя уже занято»
- **3c.** Невалидный email → ошибка «Введите корректный email»
- **3d.** Пароль не соответствует требованиям → ошибка (см. §4)
- **3e.** Пароли не совпадают → ошибка «Пароли не совпадают»
- **3f.** Одновременная регистрация с тем же email/username → при нарушении UNIQUE constraint БД вернуть 409 (обрабатывать duplicate key error, не предварительный SELECT)

### UC-2: Вход

**Актор:** Гость (с существующим аккаунтом)

**Предусловие:** Гость не авторизован. Аккаунт существует.

**Основной поток:**

1. Гость открывает страницу `/login`
2. Система отображает форму: email, пароль
3. Гость заполняет форму и отправляет
4. Система проверяет rate-limit (см. [§6 Rate-limiting](#6-rate-limiting))
5. Система проверяет email + пароль
6. Система создаёт сессию и устанавливает cookie
7. Система устанавливает flash-сообщение «Вы вошли в систему»
8. Система перенаправляет на URL из query-параметра `?next=/path` (если валидный относительный путь) или на главную

**Альтернативные потоки:**

- **4a.** IP заблокирован по rate-limit → HTTP 429, сообщение «Слишком много попыток входа. Попробуйте через N минут»
- **5a.** Неверный email или пароль → ошибка «Неверный email или пароль» (единое сообщение, без раскрытия, что именно неверно). При несуществующем email — выполняется dummy `bcrypt.CompareHashAndPassword` для выравнивания времени ответа (защита от timing attack)

### UC-3: Выход

**Актор:** Пользователь

**Предусловие:** Пользователь авторизован.

**Основной поток:**

1. Пользователь нажимает «Выход»
2. Система удаляет сессию из БД
3. Система удаляет cookie
4. Система перенаправляет на главную

### UC-4: Поддержание сессии

**Актор:** Пользователь

**Предусловие:** Пользователь авторизован, cookie с session_id установлен.

**Основной поток:**

1. Пользователь делает любой запрос
2. Middleware читает session_id из cookie
3. Middleware находит сессию в БД и проверяет `expires_at > now()`
4. Middleware обновляет `expires_at` на `now() + 30 дней` (grace period: не чаще раза в 5 минут — если `expires_at - now() > 29 дней 23 часа 55 минут`, UPDATE пропускается)
5. Запрос обрабатывается с контекстом пользователя

**Альтернативные потоки:**

- **2a.** Cookie отсутствует → запрос обрабатывается как гостевой
- **3a.** Сессия не найдена или истекла → cookie удаляется, запрос гостевой

---

## 4. Правила валидации

| Поле | Правило | Сообщение об ошибке |
|---|---|---|
| username | Длина 3–30 символов, только `[a-zA-Z0-9_-]` | «Имя должно содержать от 3 до 30 символов (латиница, цифры, _ и -)» |
| username | Уникальность в таблице `users` (case-insensitive) | «Это имя уже занято» |
| email | Нормализация: `strings.TrimSpace()` + `strings.ToLower()` перед валидацией и сохранением | — |
| email | Формат `user@domain.tld` (Go `net/mail.ParseAddress` или regexp `^[^@\s]+@[^@\s]+\.[^@\s]+$`) | «Введите корректный email» |
| email | Уникальность в таблице `users` (через `lower(email)` индекс) | «Пользователь с таким email уже зарегистрирован» |
| email | Длина ≤ 254 символа (RFC 5321) | «Email слишком длинный» |
| password | Длина ≥ 8 символов | «Пароль должен содержать не менее 8 символов» |
| password | Длина ≤ 72 байт (лимит bcrypt) | «Пароль слишком длинный» |
| password_confirm | Совпадает с password | «Пароли не совпадают» |

---

## 5. Модель данных

### Таблица `users`

```sql
CREATE TABLE users (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username   VARCHAR(30)  NOT NULL,
    email      VARCHAR(254) NOT NULL,
    pass_hash  BYTEA        NOT NULL,
    role       VARCHAR(20)  NOT NULL DEFAULT 'user',
    is_active  BOOLEAN      NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_users_email ON users (lower(email));
CREATE UNIQUE INDEX idx_users_username ON users (lower(username));
```

> **Примечание:** UNIQUE constraint убран с колонки email — уникальность обеспечивается только через case-insensitive индекс `idx_users_email`. Поле `role` заложено для будущих ролей (модератор, администратор), логика ролей — в отдельной задаче. Поле `is_active` — soft delete: деактивированный пользователь не может войти, но данные сохраняются в БД.

### Таблица `sessions`

```sql
CREATE TABLE sessions (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ip_addr    INET         NOT NULL,
    user_agent TEXT         NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ  NOT NULL
);

CREATE INDEX idx_sessions_user_id ON sessions (user_id);
CREATE INDEX idx_sessions_expires_at ON sessions (expires_at);
```

**Очистка истекших сессий:** фоновая горутина с `time.Ticker` каждый час выполняет `DELETE FROM sessions WHERE expires_at < now()`. Индекс `idx_sessions_expires_at` обеспечивает эффективное удаление.

**Политика множественных сессий:** разрешено до 10 активных сессий на пользователя. При создании новой сессии, если у пользователя уже 10 активных — удаляется самая старая (по `created_at`).

### Таблица `login_attempts`

```sql
CREATE TABLE login_attempts (
    id         BIGINT       GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    ip_addr    INET         NOT NULL,
    attempted_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_login_attempts_ip_time ON login_attempts (ip_addr, attempted_at);
```

---

## 6. Rate-limiting

| Параметр | Значение |
|---|---|
| Окно | 10 минут (sliding window) |
| Максимум неудачных попыток с одного IP | 5 за окно |
| Блокировка | до тех пор, пока в окне 10 минут есть ≥ 5 записей |

**Алгоритм (на каждый POST `/login`):**

1. **Rate-limit проверяется ПЕРЕД проверкой credentials.** Если IP заблокирован — HTTP 429 независимо от правильности пароля.
2. `SELECT count(*) FROM login_attempts WHERE ip_addr = $1 AND attempted_at > now() - interval '10 minutes'`
3. Если count ≥ 5 → HTTP 429, **не проверять пароль**
4. При неудачном входе → `INSERT INTO login_attempts`
5. При успешном входе → попытки не записываются (естественно истекают через 10 минут)
6. Периодическая очистка: горутина с `time.Ticker` каждые 10 минут — `DELETE FROM login_attempts WHERE attempted_at < now() - interval '1 hour'`

**Rate-limit на POST `/register`:**

Аналогичный лимит: 5 попыток регистрации с одного IP за 10 минут. Использует ту же таблицу `login_attempts` (или общий per-IP rate-limiter). При превышении — HTTP 429.

---

## 7. API-контракты

Все эндпоинты возвращают HTML (server-side rendering). Сервер проверяет заголовок `HX-Request`: если присутствует — возвращает HTML-фрагмент (без `<html>`, `<head>`, `<body>`), если нет — полную страницу. Это правило действует для всех status-кодов (200, 422, 409, 429).

**Формат ошибок валидации:** при 422/409 сервер возвращает ре-рендер формы. Ошибки отображаются inline под соответствующим полем (`<span class="field-error">текст</span>`). Введённые данные сохраняются в полях (кроме паролей). Поле с ошибкой помечается `aria-invalid="true"`, ошибка связана через `aria-describedby`.

**Accessibility форм:**
- Каждое поле формы имеет `<label>` с атрибутом `for`, ссылающимся на `id` поля
- Ошибки валидации связаны с полем через `aria-describedby`
- Поле с ошибкой помечается `aria-invalid="true"`
- При ре-рендере формы с ошибками — автофокус (`autofocus`) на первое поле с ошибкой

**Flash-сообщения:** после успешной регистрации/входа сервер устанавливает flash-сообщение в cookie (одноразовое, удаляется после прочтения). Шаблон `base.html` проверяет наличие flash и отображает его в блоке уведомлений.

### POST /register

**Запрос:** `application/x-www-form-urlencoded`

| Параметр | Тип | Обязательно |
|---|---|---|
| username | string | да |
| email | string | да |
| password | string | да |
| password_confirm | string | да |
| _csrf | string | да |

**Ответы:**

| Код | Условие | Действие |
|---|---|---|
| 303 | Успех | Redirect на `/` с Set-Cookie |
| 422 | Ошибка валидации | Повторный рендер формы с ошибками (inline под полями, данные полей сохраняются кроме паролей) |
| 409 | Email/username занят | Повторный рендер формы с ошибкой |
| 429 | Rate-limit | Страница с сообщением о блокировке |

### POST /login

**Запрос:** `application/x-www-form-urlencoded`

| Параметр | Тип | Обязательно |
|---|---|---|
| email | string | да |
| password | string | да |
| _csrf | string | да |

**Ответы:**

| Код | Условие | Действие |
|---|---|---|
| 303 | Успех | Redirect на `?next=/path` (если валидный относительный путь) или `/` с Set-Cookie |
| 422 | Неверные credentials | Повторный рендер формы с ошибкой (inline, данные email сохраняются) |
| 429 | Rate-limit | Страница с сообщением о блокировке |

**Redirect после входа:** передаётся через query-параметр `?next=/path`. Валидация: только относительные пути (начинаются с `/`), не `/login`, не `/register`, не `/logout`. При невалидном значении — redirect на `/`.

### POST /logout

**Запрос:** `application/x-www-form-urlencoded`

| Параметр | Тип | Обязательно |
|---|---|---|
| _csrf | string | да |

**Ответы:**

| Код | Условие | Действие |
|---|---|---|
| 303 | Успех | Redirect на `/`, удаление cookie |

### GET /register

Отображает форму регистрации. Если пользователь авторизован — redirect на `/`.

### GET /login

Отображает форму входа. Если пользователь авторизован — redirect на `/`.

---

## 8. Cookie и безопасность сессии

| Параметр | Значение |
|---|---|
| Имя cookie | `session_id` |
| Значение | UUID сессии |
| HttpOnly | `true` |
| Secure | `true` (в prod) |
| SameSite | `Lax` |
| Path | `/` |
| Max-Age | 30 дней (2 592 000 секунд) |

**Дополнительно:**

- Пароли хешируются через `bcrypt` с cost = 12
- Email нормализуется (trim + lower) при регистрации и входе, хранится в нижнем регистре
- CSRF-токен проверяется middleware для всех POST/PUT/DELETE
- Session ID — криптографически случайный UUID v4 (генерируется PostgreSQL)
- nginx обязан передавать `proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for` и `X-Real-IP $remote_addr`. Go-приложение использует Echo `RealIP()` middleware для получения реального IP (необходимо для корректной работы rate-limiting)
- Защита от timing attack: при несуществующем email выполняется dummy `bcrypt.CompareHashAndPassword` с заранее сгенерированным хешем, чтобы время ответа не отличалось от случая с существующим пользователем

---

## 9. Переменные окружения

| Переменная | Описание | Значение по умолчанию |
|---|---|---|
| `DATABASE_URL` | DSN подключения к PostgreSQL | — (обязательная) |
| `APP_PORT` | Порт HTTP-сервера | `8080` |
| `COOKIE_SECURE` | Флаг `Secure` для cookie | `true` |
| `SESSION_LIFETIME` | Время жизни сессии | `720h` (30 дней) |
| `BCRYPT_COST` | Cost-параметр bcrypt | `12` |
| `RATE_LIMIT_WINDOW` | Окно rate-limiting | `10m` |
| `RATE_LIMIT_MAX` | Максимум попыток за окно | `5` |
| `SESSION_GRACE_PERIOD` | Grace period обновления expires_at | `5m` |
| `MAX_SESSIONS_PER_USER` | Максимум активных сессий | `10` |

---

## 10. Структура файлов (целевая)

```
internal/
├── auth/
│   ├── handler.go        # HTTP-хендлеры: Register, Login, Logout
│   ├── handler_test.go
│   ├── service.go        # Бизнес-логика: CreateUser, Authenticate, CreateSession
│   ├── service_test.go
│   └── middleware.go      # SessionMiddleware — загрузка пользователя из cookie
├── user/
│   ├── model.go          # Структура User
│   └── repo.go           # UserRepo: Create, GetByEmail, GetByID
├── session/
│   ├── model.go          # Структура Session
│   └── repo.go           # SessionRepo: Create, GetByID, Delete, Touch
└── ratelimit/
    ├── repo.go           # LoginAttemptRepo: Record, Count, Cleanup
    └── repo_test.go

migrations/
├── 001_create_users.sql
├── 002_create_sessions.sql
└── 003_create_login_attempts.sql

templates/
├── auth/
│   ├── login.html
│   └── register.html
└── layouts/
    └── base.html
```

---

## 11. Тест-план

### Юнит-тесты

| Компонент | Что проверяем |
|---|---|
| Валидация email | Корректные и некорректные форматы |
| Валидация пароля | Граничные значения длины (7, 8, 72, 73 байт) |
| Хеширование | bcrypt hash создаётся и верифицируется |
| Rate-limit логика | Счётчик попыток, блокировка, сброс после окна |

### Интеграционные тесты (через testcontainers или docker)

| Сценарий | Ожидаемый результат |
|---|---|
| Регистрация с валидными данными | 303 + cookie + запись в `users` |
| Регистрация с занятым email | 409 + сообщение об ошибке |
| Вход с верными credentials | 303 + cookie + запись в `sessions` |
| Вход с неверным паролем | 422 + ошибка + запись в `login_attempts` |
| 6 неудачных попыток за 10 мин | 429 на 6-й попытке |
| Выход | 303 + cookie удалён + сессия удалена из БД |
| Запрос с валидной сессией | Пользователь доступен в контексте |
| Запрос с истёкшей сессией | Обрабатывается как гостевой |

---

## 12. Не входит в scope

- Подтверждение email через письмо
- Восстановление пароля
- OAuth / социальные логины
- Двухфакторная аутентификация
- Роли модератор / администратор (отдельная задача)
- UI-дизайн (используется минимальная стилизация)

---

_Spec Review v1.11.0 | 2026-04-02_
