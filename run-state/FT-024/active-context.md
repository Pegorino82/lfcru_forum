# Active Context: FT-024

**Updated:** 2026-05-03
**Stage:** dr-approval
**Status:** awaiting-human

## Completed

<!-- Список пройденных этапов с датой и outcome -->

## Current

Ожидание подтверждения Design Ready от человека

## Blocked / Pending

—

## Key Decisions

- URL профиля: `/profile/{username}` (username уже UNIQUE в БД)
- Форматы аватара: JPEG, PNG, WebP; макс. 5 МБ; хранить как WebP (ADR-005)
- Клик везде (форум, комментарии, header) → модалка → кнопка «Открыть профиль»
- Header: аватар + имя, кликабельно — поведение как везде
- Удаление аватара вне scope (NS-04); перезапись — единственный способ смены
