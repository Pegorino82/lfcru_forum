---
title: "FT-017: fix: позиция цитирующего комментария на форуме"
doc_kind: feature
doc_function: canonical
purpose: "Исправляет баг, при котором ответ с цитатой появляется под цитируемым постом вместо конца списка. Добавляет переход к оригинальному посту по клику на цитату."
derived_from:
  - ../../domain/problem.md
status: active
delivery_status: done
audience: humans_and_agents
trello_card: https://trello.com/c/Pwgd2nnJ
must_not_define:
  - implementation_sequence
---

# FT-017: fix: позиция цитирующего комментария на форуме

## What

### Problem

При ответе на комментарий (нажатие «Ответить» + отправка) новый пост появляется не в конце списка, а сразу после цитируемого поста. Причина — намеренная группировка в `ORDER BY` (`COALESCE(p.parent_id, p.id)`), которая нарушает ожидаемое хронологическое поведение форума.

Дополнительно: блок цитаты в посте не даёт перейти к оригинальному сообщению.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Ответ с цитатой отображается в конце списка | Всегда под цитируемым | Всегда в конце (хронологически) | Ручная проверка SC-01 |
| `MET-02` | Клик на цитату переходит к оригиналу | Отсутствует | Anchor-переход `#post-{id}` | Ручная проверка SC-02 |

### Scope

- `REQ-01` Ответ с цитатой (post с `parent_id != NULL`) отображается в конце списка всех комментариев темы, отсортированных хронологически по `created_at`.
- `REQ-02` Блок цитаты в посте является кликабельным и выполняет anchor-переход к оригинальному посту (`#post-{parent_id}`). Если родительский пост удалён (`parent_id IS NULL`, но снапшоты сохранены), цитата отображается без ссылки.

### Non-Scope

- `NS-01` Переработка формы ответа (inline → single bottom-form) — отдельная фича, обсуждается отдельно.
- `NS-02` Изменение схемы БД, новые миграции.
- `NS-03` Изменение `ListPostsAfter` (SSE catch-up) — он уже сортирует `ORDER BY p.id ASC`.
- `NS-04` Ограничение глубины вложенности ответов (уже реализовано в `ErrReplyToReply`).

### Constraints / Assumptions

- `ASM-01` Все существующие ответы в БД имеют корректный `parent_id`; переупорядочивание SQL не требует миграции данных.
- `CON-01` Якорная навигация `href="#post-{id}"` — нативный anchor без JS-скролла; достаточно для первой итерации.

## How

### Solution

Заменить `ORDER BY` в `ListPostsByTopic` с группирующего по `COALESCE(parent_id, id)` на простой хронологический `ORDER BY p.created_at ASC, p.id ASC`. В шаблонах: убрать CSS-отступ `.post.reply`, заменить `<div class="post-quote">` на `<a class="post-quote">` с условным `href`.

### Change Surface

| Surface | Type | Why it changes |
| --- | --- | --- |
| `internal/forum/repo.go:116` | code | Корневая причина бага — ORDER BY |
| `templates/forum/topic.html` | code | CSS `.post.reply`, класс div, блок цитаты |
| `templates/forum/partials/post.html` | code | Класс article, блок цитаты (SSE partial) |

### Flow

1. Пользователь А нажимает «Ответить» на посте пользователя Б, вводит текст, отправляет форму.
2. Сервер создаёт пост с `parent_id = B.id`, возвращает re-rendered `#posts-list`.
3. Список отсортирован `ORDER BY created_at ASC` → новый пост появляется последним.
4. Внутри поста блок цитаты отрендерен как `<a href="#post-{B.id}">` — клик прокручивает к посту Б.

### Failure Modes

- `FM-01` `parent_id IS NOT NULL`, но пост удалён → `ParentID = nil`, снапшоты сохранены. Цитата показывается без ссылки (атрибут `href` не рендерится).

## Verify

### Exit Criteria

- `EC-01` После отправки ответа с цитатой он появляется последним в списке (не под цитируемым постом).
- `EC-02` Клик на блок цитаты в посте прокручивает страницу к оригинальному посту.

### Traceability matrix

| Requirement ID | Design refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-01`, `CON-01`, `FM-01` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `CON-01`, `FM-01` | `EC-02`, `SC-02` | `CHK-02` | `EVID-02` |

### Acceptance Scenarios

- `SC-01` Пользователь А нажимает «Ответить» на посте Б (не последнем), вводит текст, отправляет. Ответ появляется **внизу списка** (после всех текущих постов), содержит блок цитаты с автором и фрагментом текста Б.
- `SC-02` Пользователь кликает на блок цитаты внутри ответа. Страница прокручивается (anchor) к оригинальному посту пользователя Б. Если пост Б удалён — цитата отображается, клик не выполняется (нет href).

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `SC-01` | Создать тему, написать 2 поста, ответить на первый; проверить порядок в DOM | Ответ — последний `#post-*` в `#posts-list` | `artifacts/ft-017/verify/chk-01/` |
| `CHK-02` | `EC-02`, `SC-02` | Кликнуть на `.post-quote` внутри ответа; проверить URL hash | URL содержит `#post-{id}` родительского поста | `artifacts/ft-017/verify/chk-02/` |

### Test matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | `artifacts/ft-017/verify/chk-01/` |
| `CHK-02` | `EVID-02` | `artifacts/ft-017/verify/chk-02/` |

### Evidence

- `EVID-01` Скриншот или описание: ответ с цитатой отображается последним в списке.
- `EVID-02` Скриншот или описание: URL hash `#post-{id}` после клика на цитату.

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Скриншот страницы с постами | human | `artifacts/ft-017/verify/chk-01/` | `CHK-01` |
| `EVID-02` | Скриншот URL с hash или описание | human | `artifacts/ft-017/verify/chk-02/` | `CHK-02` |
