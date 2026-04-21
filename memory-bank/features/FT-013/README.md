---
title: "FT-013: Форум не отображает залогиненного пользователя в навигации"
doc_kind: feature
doc_function: index
purpose: "Облегчённый пакет баг-фикса. Содержит описание бага, репродукцию, root cause и ссылку на фикс."
status: active
delivery_status: planned
audience: humans_and_agents
---

# FT-013: Форум не отображает залогиненного пользователя в навигации

## Описание бага

Все страницы форума показывают «Войти / Регистрация» даже для залогиненного пользователя.

## Репродукция

1. Залогиниться
2. Перейти на любую страницу форума

**Ожидается:** в навигации отображается имя пользователя
**Факт:** навигация показывает «Войти / Регистрация»

## Root cause

`forum/handler.go` не передаёт `User` в data map при рендере шаблонов.

## Фикс

Commit `481eb97` — добавлен ключ `"User": auth.UserFromContext(c)` во все `data map` в `internal/forum/handler.go`.

## Regression-тест

`TestIndex_AuthUser_ShowsUsername` в `internal/forum/handler_test.go`.
