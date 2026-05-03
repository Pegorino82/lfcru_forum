**Loop:** Brief Improve Loop
**Artifact:** memory-bank/features/FT-024/feature.md / ## What
**Date:** 2026-05-03
**Outcome:** revise (итерация 1)

**Details:**

1. [MEDIUM — A-4] REQ-04 содержит implementation detail: «Файл конвертируется в WebP и сохраняется по ADR-005» — это как, а не что. CON-03 является canonical owner этой детали.
   Исправление: убрать фразу из REQ-04.

2. [MEDIUM — E-3] Молчаливое допущение об auth middleware: CON-02/NEG-03 подразумевают session middleware, но ASM-* отсутствовал.
   Исправление: добавить ASM-04 о session middleware для /profile/*.
