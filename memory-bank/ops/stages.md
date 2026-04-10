---
title: Stages And Non-Local Environments
doc_kind: engineering
doc_function: canonical
purpose: Описание non-local окружений LFC.ru — только production на VPS.
derived_from:
  - ../dna/governance.md
status: active
audience: humans_and_agents
---

# Stages And Non-Local Environments

У проекта два окружения: `local` (docker compose dev) и `production` (VPS).
Staging/sandbox окружений нет.

## Environment Inventory

| Environment | Purpose | Access |
|---|---|---|
| `local` | Разработка и тесты | `localhost:8080`, docker compose dev |
| `production` | Живые пользователи | VPS + docker compose + nginx reverse proxy |

## Production Stack

```
nginx (reverse proxy, TLS) → app container (:8080) → postgres container
```

- nginx пробрасывает запросы на app, обеспечивает TLS.
- SSE-эндпоинты требуют `proxy_buffering off` и `X-Accel-Buffering: no` в nginx конфиге.
- Сессии хранятся в PostgreSQL (не в памяти) — рестарт контейнера не сбрасывает сессии.

## Common Operations

```bash
# Логи приложения (на VPS)
docker compose logs -f app

# Статус контейнеров
docker compose ps

# Подключение к prod БД (только read-only операции без явной необходимости)
psql "$DATABASE_URL"

# Health check
curl -fsS http://localhost:8080/health
```

## Credentials And Access

- `DATABASE_URL` с prod credentials — `.env` файл на VPS, не в репозитории.
- `COOKIE_SECURE=true` — обязательно в prod.
- SSH-доступ к VPS — через ключи, не пароли.
- Не запускать мутирующие DB-операции без бэкапа.

## Version And Health Checks

```bash
# Health endpoint (если реализован)
curl -fsS https://<domain>/health

# Текущая версия образа
docker compose images app

# Логи последних 50 строк
docker compose logs --tail=50 app
```

## Logs And Observability

- Логи приложения: stdout/stderr контейнера → `docker compose logs`.
- Метрики и трейсы: не реализованы (проект в стадии разработки).
- Error tracking: не реализован.
