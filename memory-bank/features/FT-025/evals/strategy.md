---
title: "FT-025: Eval Strategy"
doc_kind: feature
doc_function: eval-strategy
ft_id: FT-025
status: active
audience: humans_and_agents
---

# FT-025 Eval Strategy

## Gates & Forms

| Gate | Форма | Evaluator |
|---|---|---|
| Draft → Design Ready | evaluator agent (large.md) | Agent tool |
| Design Ready → Plan Ready | evaluator agent (large.md) | Agent tool |
| Plan Ready → Execution | human approval | human |
| Execution → Done | hybrid: CI + evaluator agent + human AG-* | Agent tool + AG-* |

## Risk Areas

- Изменение SQL-запросов в двух репозиториях (forum, comment) — риск сломать существующие сканы
- SSE-фрагмент для форума (`renderPostFragment`) — рендерит PostView напрямую, нужно убедиться что поле `AuthorAvatarURL` передаётся корректно
- Шаблон `post.html` используется и для первичного рендера страницы, и для SSE-вставки — изменения должны работать в обоих контекстах
- Playwright-тесты обязательны (UI-изменения по testing-policy.md)
