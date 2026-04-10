---
title: Release And Deployment
doc_kind: ops
doc_function: canonical
purpose: Релизный процесс LFC.ru — сборка Docker-образа, деплой на VPS, rollback.
derived_from:
  - ../dna/governance.md
status: active
audience: humans_and_agents
---

# Release And Deployment

## Release Flow

1. Убедиться, что тесты зелёные (unit + integration).
2. Смержить feature-ветку в `main`.
3. Собрать Docker-образ на VPS (или локально и передать через registry).
4. Перезапустить контейнер — миграции применятся автоматически при старте.
5. Проверить логи и health.

## Build And Deploy

```bash
# На VPS: пересборка и рестарт (downtime ~секунды)
docker compose -f docker-compose.prod.yml up --build -d

# Только рестарт без пересборки
docker compose -f docker-compose.prod.yml restart app

# Проверить что поднялось
docker compose -f docker-compose.prod.yml ps
docker compose -f docker-compose.prod.yml logs --tail=50 app
```

> Миграции (`goose.Up()`) применяются автоматически при каждом старте app. Новая миграция применится сама при деплое.

## Required Env Vars In Production

| Переменная | Требование |
|---|---|
| `DATABASE_URL` | Обязательна, содержит credentials |
| `COOKIE_SECURE` | Должна быть `true` (HTTPS) |
| `BCRYPT_COST` | Рекомендуется `12` (default) |

Остальные переменные — опциональны, используются defaults из `config.go`.

## Release Test Plan

После каждого деплоя проверить:

- [ ] Главная страница открывается (`/`)
- [ ] Форум открывается, разделы и темы отображаются
- [ ] Регистрация нового пользователя работает
- [ ] Вход и выход из аккаунта работают
- [ ] Логи app не содержат panic/fatal

## Rollback

Rollback unit — предыдущий Docker-образ.

```bash
# Откат к предыдущему образу (если тег сохранён)
docker compose -f docker-compose.prod.yml stop app
docker tag <previous-image> forum:latest
docker compose -f docker-compose.prod.yml up -d app
```

**Важно по миграциям:**
- goose применяет миграции только вперёд при старте.
- Обратный откат миграции (`goose down`) — ручная операция с потенциальной потерей данных.
- Если новая миграция несовместима с предыдущим кодом — rollback невозможен без `goose down`.
- Проектируй миграции как backward-compatible где возможно (additive changes).
