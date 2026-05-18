---
title: "FT-028: Eval Strategy"
doc_kind: feature
doc_function: eval-strategy
ft_id: FT-028
status: active
audience: humans_and_agents
---

# FT-028 Eval Strategy

## Gates & Forms

| Gate | Форма | Evaluator |
|---|---|---|
| Draft → Design Ready | brief-loop + spec-loop (evaluator agents) → DR-eval.md | Agent tool |
| Design Ready → Plan Ready | evaluator agent | Agent tool |
| Plan Ready → Execution | human approval | user |
| Execution → Done | hybrid: CI + evaluator agent + human AG-* | Agent tool + AG-* |

> Для `large.md`: evaluator agent обязателен на DR (через brief-loop + spec-loop) и на Done.
> На DR→PR: evaluator agent, если план > 3 STEP-*; self-check, если ≤ 3 STEP-*.

## Risk Areas

- **API dependency:** football-data.org может изменить формат ответа или ограничить доступ — проверять graceful degradation.
- **Idempotency:** генерация тем должна быть безопасна при повторном запуске — дубликаты разделов и тем недопустимы.
- **HTML injection:** карточка игрока содержит данные из внешнего API — sanitization обязательна (bluemonday или template escaping).
- **Rate limiting:** один API-вызов на всю генерацию (CON-01), но при ошибке не должно быть partial state.
