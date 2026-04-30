---
title: "FT-015: UX редактора статей (превью + фидбек сохранения)"
doc_kind: feature
doc_function: index
purpose: "Облегчённый пакет баг-фикса. Содержит описание бага, репродукцию, root cause и ссылку на фикс."
status: active
delivery_status: done
audience: humans_and_agents
---

# FT-015: UX редактора статей (превью + фидбек сохранения)

## Описание бага

Два UX-дефекта в редакторе статей админки:
1. При превью статьи не видно, что это превью — нет баннера и кнопки «Назад к редактору»
2. После сохранения статьи нет визуального подтверждения успеха

## Репродукция

**Баг 1:**
1. Открыть редактор статьи в админке
2. Нажать «Превью»

**Ожидается:** на странице виден баннер «Режим превью» и кнопка возврата к редактированию
**Факт:** страница выглядит как обычная публичная статья без каких-либо индикаторов

**Баг 2:**
1. Открыть редактор статьи, внести изменения
2. Нажать «Сохранить»

**Ожидается:** появляется уведомление об успешном сохранении
**Факт:** страница просто остаётся без каких-либо изменений

## Root cause

**Баг 1 (Preview):** `Preview` handler рендерил публичный `templates/news/article.html` с анонимной struct, в которой не было флага `IsPreview`. Шаблон не имел условного баннера, поэтому страница выглядела идентично обычной статье.

**Баг 2 (Save):** `Update` handler после успешного сохранения делал `Redirect` без параметра, `Edit` handler не передавал никакого сигнала об успехе в шаблон, а `articleEditData` не имела поля `Saved`.

## Фикс

Commit: `94f21a2`

- `articlePreviewData` получила поле `IsPreview bool`; `Preview` handler использует эту struct с `IsPreview: true` вместо анонимной; в `templates/news/article.html` добавлен баннер «Режим превью» + ссылка «Назад к редактору» при `{{if .IsPreview}}`.
- `articleEditData` получила поле `Saved bool`; `Update` redirects to `?saved=1`; `Edit` передаёт `Saved: c.QueryParam("saved") == "1"`; `templates/admin/articles/edit.html` показывает success-баннер при `{{if .Saved}}`.

## Regression-тест

Commit: `94f21a2` — `internal/admin/articles_handler_test.go`

- `TestAdminArticles_Preview_HasPreviewBanner` — GET `/admin/articles/:id/preview` → 200, тело содержит «Режим превью» и «Назад к редактору».
- `TestAdminArticles_Update_ShowsSavedConfirmation` — POST update → 303 Location содержит `saved=1`; GET edit?saved=1 → 200, тело содержит «Статья сохранена».
