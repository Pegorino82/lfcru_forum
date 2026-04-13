#!/usr/bin/env bash
# SessionEnd hook: парсит JSONL транскрипт и пишет итог сессии
set -euo pipefail

LOG_DIR="${CLAUDE_LOG_DIR:-$HOME/.claude/logs}"

if [[ ! -f "$LOG_DIR/.current_session" ]]; then
  exit 0
fi

SESSION_FILE=$(cat "$LOG_DIR/.current_session")
END_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Получаем путь к транскрипту (сохранён Stop-хуком или в payload)
TRANSCRIPT_PATH=""
payload="$(cat)"

if [[ -f "$LOG_DIR/.transcript_path" ]]; then
  TRANSCRIPT_PATH=$(cat "$LOG_DIR/.transcript_path")
fi

if [[ -z "$TRANSCRIPT_PATH" ]]; then
  TRANSCRIPT_PATH=$(echo "$payload" | jq -r '.transcript_path // ""' 2>/dev/null || true)
fi

# Фолбэк: конструируем путь из session_id
if [[ -z "$TRANSCRIPT_PATH" || ! -f "$TRANSCRIPT_PATH" ]]; then
  SESSION_ID=$(echo "$payload" | jq -r '.session_id // ""' 2>/dev/null || true)
  if [[ -n "$SESSION_ID" ]]; then
    PROJ_DIR="$HOME/.claude/projects/-Users-evgenyshkryabin-study-lfcru-forum"
    TRANSCRIPT_PATH="$PROJ_DIR/${SESSION_ID}.jsonl"
  fi
fi

if [[ -z "$TRANSCRIPT_PATH" || ! -f "$TRANSCRIPT_PATH" ]]; then
  printf "\n## Итог сессии\n\n**Конец:** %s\n*(транскрипт не найден)*\n\n---\n" "$END_TIME" >> "$SESSION_FILE"
  rm -f "$LOG_DIR/.current_session" "$LOG_DIR/.transcript_path"
  exit 0
fi

# --- Парсим транскрипт ---

# Начальный промпт: первое текстовое сообщение пользователя
INITIAL_PROMPT=$(jq -r '
  select(.type == "user" and .isSidechain == false) |
  .message.content |
  if type == "string" then .
  elif type == "array" then
    [.[] | select(.type == "text") | .text] | join(" ")
  else "" end
' "$TRANSCRIPT_PATH" | grep -v '^$' | head -1 | cut -c1-600)

# Продолжительность: первый и последний timestamp в файле
START_TS=$(jq -r 'select(.timestamp != null) | .timestamp' "$TRANSCRIPT_PATH" | head -1)
LAST_TS=$(jq -r 'select(.timestamp != null) | .timestamp' "$TRANSCRIPT_PATH" | tail -1)

# Нормализуем ISO 8601 для macOS date
norm_ts() { echo "$1" | sed 's/\.[0-9]*Z$/Z/'; }

start_epoch=$(date -j -f "%Y-%m-%dT%H:%M:%SZ" "$(norm_ts "$START_TS")" +%s 2>/dev/null || echo "0")
end_epoch=$(date -j -f "%Y-%m-%dT%H:%M:%SZ" "$(norm_ts "$LAST_TS")" +%s 2>/dev/null || echo "0")

if [[ "$start_epoch" -gt 0 && "$end_epoch" -gt "$start_epoch" ]]; then
  DURATION=$(( end_epoch - start_epoch ))
  DURATION_STR="$(( DURATION / 60 ))м $(( DURATION % 60 ))с"
else
  DURATION_STR="n/a"
fi

# Токены: группируем по requestId, берём последнюю запись (финальный стрим)
TOKEN_JSON=$(jq -sc '
  [.[] | select(.type == "assistant" and .requestId != null)] |
  group_by(.requestId) |
  map(last | .message.usage // {}) |
  {
    input:        ([.[].input_tokens               // 0] | add // 0),
    output:       ([.[].output_tokens              // 0] | add // 0),
    cache_create: ([.[].cache_creation_input_tokens // 0] | add // 0),
    cache_read:   ([.[].cache_read_input_tokens    // 0] | add // 0)
  }
' "$TRANSCRIPT_PATH")

INPUT_TOK=$(echo "$TOKEN_JSON"  | jq '.input')
OUTPUT_TOK=$(echo "$TOKEN_JSON" | jq '.output')
CACHE_CR=$(echo  "$TOKEN_JSON"  | jq '.cache_create')
CACHE_RD=$(echo  "$TOKEN_JSON"  | jq '.cache_read')
TOTAL_TOK=$(( INPUT_TOK + OUTPUT_TOK + CACHE_CR + CACHE_RD ))

# Количество ходов
TURN_COUNT=$(jq -r 'select(.type == "user" and .isSidechain == false) | .uuid' "$TRANSCRIPT_PATH" | wc -l | tr -d ' ')

# Ошибки: tool_result с is_error или содержимым-ошибкой
ERRORS=$(jq -r '
  select(.type == "user" and .isSidechain == false) |
  .message.content |
  if type == "array" then
    .[] |
    select(.type == "tool_result") |
    select(.is_error == true) |
    "- " + (
      .content |
      if   type == "array"  then (.[0].text // "" | .[0:300])
      elif type == "string" then .[0:300]
      else (tostring | .[0:300])
      end
    )
  else empty end
' "$TRANSCRIPT_PATH" 2>/dev/null | sort -u | head -15)

# --- Пишем итог в лог-файл ---
{
  echo ""
  echo "## Итог сессии"
  echo ""
  echo "**Конец (UTC):** $END_TIME"
  echo "**Продолжительность:** $DURATION_STR"
  echo "**Ходов диалога:** $TURN_COUNT"
  echo ""
  echo "### Начальный промпт"
  echo '```'
  echo "${INITIAL_PROMPT:-(не определён)}"
  echo '```'
  echo ""
  echo "### Токены"
  echo "| Тип | Количество |"
  echo "|---|---|"
  echo "| Input | $INPUT_TOK |"
  echo "| Output | $OUTPUT_TOK |"
  echo "| Cache create | $CACHE_CR |"
  echo "| Cache read | $CACHE_RD |"
  echo "| **Итого** | **$TOTAL_TOK** |"
  echo ""
  echo "### Ошибки и пути решения"
  if [[ -z "$ERRORS" ]]; then
    echo "Ошибок в инструментах не обнаружено."
  else
    echo "$ERRORS"
  fi
  echo ""
  echo "---"
} >> "$SESSION_FILE"

rm -f "$LOG_DIR/.current_session" "$LOG_DIR/.transcript_path"
