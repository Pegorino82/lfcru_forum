---
title: "FT-012: Навигация «Новости» и 500 на /news"
doc_kind: feature
doc_function: index
purpose: "Облегчённый пакет баг-фикса. Содержит описание бага, репродукцию, root cause и ссылку на фикс."
status: active
delivery_status: done
audience: humans_and_agents
---

# FT-012: Навигация «Новости» и 500 на /news

## Описание бага

Ссылка «Новости» в хэдере не работает, а `/news` падает с 500 для залогиненного пользователя.

## Репродукция

### 1

1. Залогиниться
2. Кликнуть «Новости» в навигации

**Ожидается:** открывается `/news` со списком статей
**Факт:** 500 Internal Server Error

### 2

1. Кликнуть «Новости» в навигации не залогиненным пользователем

**Ожидается:** открывается `/news` со списком статей
**Факт:** 500 Internal Server Error

## Root cause

Два независимых бага:

1. **Навигация**: `templates/layouts/base.html:44` — ссылка «Новости» имела `href="#"` вместо `href="/news"`.
2. **500 для залогиненных**: `internal/news/handler.go` — структура `ListData` не содержала поля `CSRFToken`. Шаблон `base.html` вызывает `{{.CSRFToken}}` внутри `{{if .User}}` (форма выхода); `html/template` возвращает ошибку выполнения `can't evaluate field CSRFToken in type news.ListData` → 500.

## Фикс

`7fbea0f` — fix: news nav link broken and 500 for authenticated users on /news

## Regression-тест

`internal/news/handler_test.go` — `TestShowList_AuthUser_OK`.
