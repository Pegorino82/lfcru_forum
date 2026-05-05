# Active Context: FT-025

**Updated:** 2026-05-05
**Stage:** impl
**Status:** in_progress

## Completed

- bootstrap: worktree создан, Trello IN PROGRESS, Draft PR #16
- brief-loop: accept (2026-05-04)
- spec-loop: accept (2026-05-04, итерация 2)
- dr-approval: approved (2026-05-04)
- plan: accept (2026-05-05, итерация 2)
- pr-approval: approved (2026-05-05)

## Current

STEP-01..11 — имплементация

## Blocked / Pending

—

## Key Decisions

- `PostView` и `CommentView` расширяются полем `AuthorAvatarURL *string`
- Header: использует уже имеющийся `.User.AvatarURL` — только шаблонное изменение
- Fallback: `avatarColor`/`avatarInitials` — те же функции FuncMap что на странице профиля
- SSE-фрагмент `renderPostFragment` — нужна отдельная проверка
