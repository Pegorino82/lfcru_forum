---
title: Configuration Guide
doc_kind: engineering
doc_function: canonical
purpose: Конфигурация LFC.ru — env vars, defaults и правила секретов.
derived_from:
  - ../dna/governance.md
status: active
audience: humans_and_agents
---

# Configuration Guide

Конфигурация загружается из переменных окружения в `internal/config/config.go`. Нет YAML, нет Helm — только env vars с defaults.

## Configuration Architecture

Единственная точка входа — `config.Load()`. Возвращает структуру `Config` с полями. Defaults заданы прямо в `getEnv / getInt / getBool / getDuration` в `config.go`.

В dev переменные передаются через `docker-compose.dev.yml` (секция `environment`) и опциональный `.env` файл (`env_file: .env`).

## Env Vars

| Переменная | Тип | Default | Описание |
|---|---|---|---|
| `DATABASE_URL` | string | `postgres://postgres:postgres@postgres:5432/lfcru?sslmode=disable` | Подключение к PostgreSQL |
| `APP_PORT` | string | `8080` | HTTP порт приложения |
| `COOKIE_SECURE` | bool | `false` | Флаг `Secure` для session cookie. `false` в dev (HTTP), `true` в prod (HTTPS) |
| `SESSION_LIFETIME` | duration | `720h` (30 дней) | TTL сессии |
| `BCRYPT_COST` | int | `12` | Стоимость bcrypt. В dev compose задан `10` для скорости |
| `RATE_LIMIT_WINDOW` | duration | `10m` | Окно rate-limit для `/login` |
| `RATE_LIMIT_MAX` | int | `5` | Макс. попыток входа в окне |
| `SESSION_GRACE_PERIOD` | duration | `5m` | Grace period при обновлении сессии (Touch) |
| `MAX_SESSIONS_PER_USER` | int | `10` | Макс. активных сессий на пользователя |

Duration формат — Go: `10m`, `1h`, `720h`.

## Naming Convention

Без префикса. Плоские имена в UPPER_SNAKE_CASE.

## Secrets

- `DATABASE_URL` содержит credentials — в prod задаётся через `.env` файл на сервере или секреты CI/CD.
- Не коммитить `.env` с prod-значениями в репозиторий.
- Ротация: пересоздать пользователя в PG, обновить `DATABASE_URL` на VPS, рестартовать контейнер.

## Dev Overrides

В `docker-compose.dev.yml` заданы dev-специфичные значения:

```yaml
COOKIE_SECURE: "false"   # HTTP в dev
BCRYPT_COST: "10"        # быстрее для разработки
```
