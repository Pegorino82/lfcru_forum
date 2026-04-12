---
title: "FT-001: Basic Authentication"
doc_kind: feature
doc_function: canonical
purpose: "Регистрация и вход пользователей по email/паролю — фундамент для ролевой модели и авторизованных действий на форуме."
derived_from:
  - ../../domain/problem.md
  - https://github.com/Pegorino82/lfcru_forum/issues/1
status: draft
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
---

# FT-001: Basic Authentication

## What

### Problem

Фундаментальный блок: без аутентификации невозможно развернуть ролевую модель и модерацию — ни одна другая пользовательская фича не реализуется до этой.

Общий контекст: [`../../domain/problem.md`](../../domain/problem.md).

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
|---|---|---|---|---|
| `MET-01` | Пользователи с минимум одним постом | 0 | Рост после запуска | `COUNT` из `users JOIN posts` |

### Scope

- `REQ-1` Гость может зарегистрироваться с email в формате `user@domain.tld` (синтаксическая проверка)
- `REQ-2` Зарегистрированный пользователь может войти и получить доступ к открытым разделам форума
- `REQ-3` Попытка входа с неверным паролем или несуществующим email отклоняется
- `REQ-4` Более 5 неудачных попыток входа с одного IP за 60 секунд → IP-блокировка на 10 минут
- `REQ-5` Сессия пользователя сохраняется не менее 30 дней с момента последней активности

### Non-Scope

- `NS-1` Подтверждение email через письмо (MVP-дефер → ADR-003)
- `NS-2` Социальная аутентификация (OAuth, Google, VK и пр.)
- `NS-3` Восстановление пароля

### Constraints / Assumptions

- `CON-1` Rate-limiting на `/login` реализован на уровне IP через `ratelimit`-пакет → [ADR-002](../../adr/ADR-002-rate-limiting-strategy.md)
- `CON-2` Сессии хранятся в PostgreSQL (httponly + secure cookie), не в Redis и не в JWT → [ADR-001](../../adr/ADR-001-session-storage.md)
- `CON-3` Email-only регистрация без подтверждения email в MVP → [ADR-003](../../adr/ADR-003-email-only-auth-no-confirmation.md)

## How

<!-- Заполнить на Design Ready: solution sketch, change surface, flow -->

### ADR Dependencies

| ADR | decision_status | Used for |
|---|---|---|
| [ADR-001](../../adr/ADR-001-session-storage.md) | `accepted` | Выбор хранилища сессий |
| [ADR-002](../../adr/ADR-002-rate-limiting-strategy.md) | `accepted` | Стратегия rate limiting на login |
| [ADR-003](../../adr/ADR-003-email-only-auth-no-confirmation.md) | `accepted` | Email-only, без подтверждения email в MVP |

## Verify

### Exit Criteria

- `EC-1` Все `REQ-*` покрыты passing `SC-*`
- `EC-2` Rate-limit (`CON-1`) воспроизводится в интеграционном тесте против БД `lfcru_test`

### Acceptance Scenarios

- `SC-1` Гость вводит email `user@domain.tld` и пароль → форма принимает → пользователь создан, сессия открыта
- `SC-2` Пользователь вводит корректные credentials → перенаправляется на главную → видит кнопку «Выйти», может создавать темы
- `SC-3` Пользователь вводит неверный пароль → сервер возвращает ошибку, сессия не создаётся
- `SC-4` IP отправляет 6-ю попытку входа за 60 сек → сервер возвращает 429, IP заблокирован на 10 минут
- `SC-5` Пользователь, активный ≤30 дней назад, открывает сайт → остаётся авторизованным без повторного входа

### Negative / Edge Cases

- `NEG-1` Гость вводит email без `@` или без домена → форма отклоняет, пользователь не создан
- `NEG-2` Пользователь вводит несуществующий email → сервер возвращает обобщённую ошибку (без указания причины)

### Traceability

| REQ | SC | NEG |
|---|---|---|
| `REQ-1` | `SC-1` | `NEG-1` |
| `REQ-2` | `SC-2` | — |
| `REQ-3` | `SC-3` | `NEG-2` |
| `REQ-4` | `SC-4` | — |
| `REQ-5` | `SC-5` | — |
