---
title: "FT-014: Дублирование ошибок при логине (HTMX outerHTML swap)"
doc_kind: feature
doc_function: index
purpose: "Облегчённый пакет баг-фикса. Содержит описание бага, репродукцию, root cause и ссылку на фикс."
status: active
delivery_status: planned
audience: humans_and_agents
---

# FT-014: Дублирование ошибок при логине (HTMX outerHTML swap)

## Описание бага

При повторных неудачных попытках войти форма логина вкладывается сама в себя — каждая ошибка добавляет новый слой DOM вместо замены предыдущего.

## Репродукция

1. Открыть `/login`
2. Ввести неверный пароль → отобразилась ошибка
3. Снова ввести неверный пароль

**Ожидается:** ошибка заменяется новой
**Факт:** форма с ошибкой вкладывается внутрь предыдущей формы, DOM растёт с каждой попыткой

## Root cause

`hx-target="#login-form"` + `hx-swap="outerHTML"`: при ошибке сервер возвращает полный блок `content` (outer `<div>` + `<h1>` + `<form id="login-form">`). HTMX делает outerHTML-swap — заменяет `<form>` на весь ответ, создавая вложение. При следующей ошибке находит вложенную форму и снова заменяет — DOM растёт с каждой попыткой.

## Фикс

Коммит `b32cb8c`:
- `templates/auth/login.html` — добавлен `id="login-wrapper"` на внешний `<div>`, `hx-target` изменён на `#login-wrapper`
- `templates/auth/register.html` — аналогично (`id="register-wrapper"`)

Теперь target совпадает с корневым элементом partial-ответа — swap заменяет wrapper на wrapper без вложения.

## Regression-тест

`internal/auth/handler_integration_test.go` → `TestLogin_HTMX_InvalidCredentials_NoNestedForm` (коммит `b32cb8c`): два последовательных HTMX POST /login с неверными данными; проверяет, что каждый ответ содержит ровно один `id="login-wrapper"` и ровно один `id="login-form"`.
