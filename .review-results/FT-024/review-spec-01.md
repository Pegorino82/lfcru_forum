**Loop:** Spec Improve Loop
**Artifact:** memory-bank/features/FT-024/feature.md / ## How + ## Verify
**Date:** 2026-05-03
**Outcome:** revise (итерация 1)

**Details:**

1. [HIGH — C-2/CTR-03] Endpoint-несоответствие: `POST /profile/avatar` не содержит {username}, но FM-05/EC-07/NEG-05 ссылаются на `POST /profile/{other_username}/avatar`. Ownership-check структурно недостижим.
   Исправление (Вариант A): изменить CTR-03 → `POST /profile/{username}/avatar`, handler проверяет session.UserID == target_user.ID.

2. [MEDIUM — B-4] Change Surface не покрывает FileReader JS-превью. Строка «inline стили» не охватывает script-блок.
   Исправление: расширить до «inline стили и скрипты», добавить упоминание FileReader.
