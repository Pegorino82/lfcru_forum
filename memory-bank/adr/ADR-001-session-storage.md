---
title: "ADR-001: Session Storage in PostgreSQL"
doc_kind: adr
doc_function: canonical
purpose: "Фиксирует выбор PostgreSQL как хранилища пользовательских сессий вместо Redis или JWT."
derived_from:
  - ../features/FT-001/feature.md
status: active
decision_status: accepted
date: 2024-01-01
audience: humans_and_agents
must_not_define:
  - current_system_state
  - implementation_plan
---

# ADR-001: Session Storage in PostgreSQL

## Контекст

Для авторизации пользователей (FT-001) необходимо хранить сессии. Нужно выбрать хранилище с учётом уже принятого технологического стека (Go + PostgreSQL, Docker, без внешних зависимостей кроме уже задействованных).

## Драйверы решения

- Минимизация инфраструктурных зависимостей (Docker Compose уже включает только PostgreSQL)
- Сессии должны переживать рестарт приложения
- Проект находится на стадии MVP — сложность инфры должна быть минимальной
- httponly + secure cookie обязательны (PCON-02 из `domain/problem.md`)

## Рассмотренные варианты

| Вариант | Плюсы | Минусы | Статус |
|---|---|---|---|
| PostgreSQL | Уже в стеке, транзакционность, простой старт | Дополнительная нагрузка на hot-path запросы к БД | **Выбрано** |
| Redis | Быстрый in-memory lookup, TTL из коробки | Новая инфра-зависимость, усложняет Docker Compose | Отклонено |
| JWT (stateless) | Нет хранилища, горизонтальное масштабирование | Нет инвалидации без blacklist, сложнее logout | Отклонено |

## Решение

Сессии хранятся в таблице PostgreSQL. Cookie — httponly, secure, SameSite. Время жизни сессии — не менее 30 дней с момента последней активности (REQ-5 из FT-001). Сессия инвалидируется при явном logout.

## Последствия

### Положительные

- Нет новых инфраструктурных компонентов
- Сессии консистентны с остальными данными (транзакционная БД)
- Простой механизм инвалидации: DELETE из таблицы сессий

### Отрицательные

- Каждый HTTP-запрос включает lookup сессии в PostgreSQL — дополнительная нагрузка на БД
- При высокой нагрузке потребуется индекс по session token

### Нейтральные / организационные

- Необходима миграция: таблица `sessions` (или аналогичная)
- `PCON-05` из `domain/problem.md`: интеграционные тесты сессий используют `lfcru_test`, не `lfcru`

## Риски и mitigation

| Риск | Mitigation |
|---|---|
| Нагрузка на БД при многих одновременных сессиях | Индекс по `session_token`; в случае роста — возможен дефер миграции на Redis |
| Утечка session token | httponly cookie + secure flag закрывают XSS-вектор |

## Follow-up

- Создать миграцию таблицы сессий
- Добавить индекс по `session_token`
- Реализовать middleware валидации сессии

## Связанные ссылки

- [FT-001: Basic Authentication](../features/FT-001/feature.md)
- `PCON-02`, `PCON-05` в [`../domain/problem.md`](../domain/problem.md)
