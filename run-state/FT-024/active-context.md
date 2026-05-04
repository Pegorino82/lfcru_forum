# Active Context: FT-024

**Updated:** 2026-05-04
**Stage:** verification
**Status:** awaiting_review

## Completed

- brief-loop: accept (2026-05-03)
- spec-loop: accept (2026-05-03)
- dr-approval: accept (2026-05-03)
- plan (DR→PR eval): accept (2026-05-03) — EVID-06
- pr-approval: accept (2026-05-03) — human confirm
- impl: done (STEP-00..STEP-10), 2026-05-04
- unit-tests: CI green (25305917589)
- e2e-smoke: CI green 16/16 (25305917589)

## Current

Ожидаем AG-01 — human review PR#15

## Blocked / Pending

—

## Key Decisions

- URL профиля: `/profile/{username}` (username уже UNIQUE в БД)
- Форматы аватара: JPEG, PNG, WebP; макс. 5 МБ; хранить как WebP (ADR-005)
- Клик везде (форум, комментарии, header) → модалка → кнопка «Открыть профиль»
- Header: аватар + имя, кликабельно — поведение как везде
- Удаление аватара вне scope (NS-04); перезапись — единственный способ смены
