**Loop:** Spec Improve Loop
**Artifact:** memory-bank/features/FT-024/feature.md / ## How + ## Verify
**Date:** 2026-05-03
**Outcome:** revise (итерация 2)

**Details:**

1. [BLOCKER — B-4] internal/tmpl/renderer.go отсутствует в Change Surface. Файл — единственный владелец FuncMap; avatarColor/avatarInitials и relativeTime требуют его изменения.
   Исправление: добавить строку в Change Surface.

2. [HIGH — F-2/D-3] FM-06 не покрыт NEG-* и CHK-*. Playwright может детерминированно тестировать через page.route().
   Исправление: добавить NEG-06, обновить traceability REQ-01 и CHK-01.

3. [MEDIUM — B-1] Placeholder 0XX → 013 (последняя миграция 012).

4. [MEDIUM — B-4/A-3] user/service.go и profile/service.go дублировали "логику получения профиля". По architecture.md user-модуль ownes только модель/поиск.
   Исправление: user/repo.go содержит GetByUsername+Scan; profile/service.go — агрегация через ForumRepo/CommentRepo.
