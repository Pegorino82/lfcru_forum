---
title: "FT-027: Eval Strategy"
doc_kind: feature
doc_function: eval-strategy
ft_id: FT-027
status: active
audience: humans_and_agents
---

# FT-027 Eval Strategy

## Gates & Forms

| Gate | Форма | Evaluator |
|---|---|---|
| Draft → Design Ready | brief-loop + spec-loop (evaluator agents) | Agent tool |
| Design Ready → Plan Ready | self-check (план ≤ 3 STEP-* допустим) | Agent tool |
| Plan Ready → Execution | human approval | человек |
| Execution → Done | hybrid: CI + evaluator agent + human AG-* | Agent tool + AG-* |

## Risk Areas

- UI-only фича: основной риск — визуальные регрессии на существующих страницах (главная, тема форума)
- Адаптивная сетка: breakpoints могут работать некорректно на реальных устройствах — Playwright покрывает только фиксированные viewport
- Пагинация новостей уже существует в handler — изменение page size или формата может сломать существующую навигацию
- Данные для карточек новостей (изображение, анонс) могут отсутствовать в текущем News query — потребуется проверка struct/repo
