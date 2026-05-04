---
title: "FT-024: Done Eval"
doc_kind: feature
doc_function: gate-eval
ft_id: FT-024
gate: "Execution→Done"
status: closed
date: 2026-05-04
audience: humans_and_agents
---

# FT-024: Done Eval

## Checklist

### CHK-01: CI (Playwright e2e)

- [x] SC-01: Quick-view модалка открывается по клику на имя автора — PASS
- [x] SC-02: Страница профиля отображает данные пользователя — PASS
- [x] SC-03: Загрузка аватара обновляет `#avatar-block` — PASS
- [x] SC-04: Fallback-инициалы при отсутствии аватара — PASS
- [x] SC-05: Чужой профиль — форма загрузки аватара не отображается — PASS
- [x] SC-06: Инициалы на детерминированном цветном фоне — PASS

### CHK-02: E2E — Negative flows

- [x] NEG-01: Файл > 5 МБ → 413, `[data-testid="avatar-error"]` видим — PASS (page.route mock)
- [x] NEG-02: Неподдерживаемый формат → 422, `[data-testid="avatar-error"]` видим — PASS (real text/plain upload)

### CHK-03: Code review

- [x] Handler возвращает `c.HTML()` для error-фрагментов (не `c.String()`), HTMX outerHTML swap работает корректно
- [x] `image.DecodeConfig` с blank imports (`image/jpeg`, `image/png`, `golang.org/x/image/webp`) — валидация формата на уровне binary
- [x] `beforeSwap` в base.html обрабатывает статусы 409, 413, 422
- [x] `storage/avatars/` создаётся через `os.MkdirAll` при старте — нет рантайм-ошибки при первой загрузке

### CHK-04: PR & merge

- [x] PR #15 смержен в main
- [x] Worktree и feature-ветка удалены

## Iterations

| # | Date | Outcome | Findings |
|---|---|---|---|
| 1 | 2026-05-04 | revise | CHK-02: E2E-тесты для NEG-01 и NEG-02 отсутствовали в `profile.spec.ts`; handler возвращал plain text вместо HTML-фрагментов для ошибок |
| 2 | 2026-05-04 | accept | NEG-01 (page.route mock) и NEG-02 (real upload) добавлены; handler исправлен; 413 добавлен в beforeSwap; CI 18/18 |

## Decision

**Outcome:** accept
**Date:** 2026-05-04
**EVID:** EVID-07 (Playwright 18/18 PASS), EVID-08 (PR #15 merged)
