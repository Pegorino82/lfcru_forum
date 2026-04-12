#!/usr/bin/env bash
set -euo pipefail

LOG_DIR="${CLAUDE_LOG_DIR:-$HOME/.claude/logs}"
mkdir -p "$LOG_DIR"

payload="$(cat)"

json_get() {
  if command -v jq >/dev/null 2>&1; then
    jq -r "$1 // \"\"" <<<"$payload"
  else
    # Fallback: no jq, return empty
    echo ""
  fi
}

SESSION_ID="$(json_get '.session_id')"
MODEL="$(json_get '.model')"
PROMPT="$(json_get '.prompt')"
RESPONSE="$(json_get '.response')"
START_TIME_RAW="$(json_get '.start_time')"

# Optional token fields (adjust to actual payload keys once known)
INPUT_TOKENS="$(json_get '.usage.input_tokens')"
OUTPUT_TOKENS="$(json_get '.usage.output_tokens')"
TOTAL_TOKENS="$(json_get '.usage.total_tokens')"

# Default fallbacks
SESSION_ID="${SESSION_ID:-$(date +%s)}"
MODEL="${MODEL:-unknown-model}"

# Ensure session file exists even if SessionStart didn't fire
if [[ -f "$LOG_DIR/.current_session" ]]; then
  SESSION_FILE="$(cat "$LOG_DIR/.current_session")"
else
  TIMESTAMP=$(date -u +"%Y%m%d_%H%M%S")
  SESSION_FILE="$LOG_DIR/${TIMESTAMP}_session_${SESSION_ID}.md"
  cat <<EOF > "$SESSION_FILE"
# Claude Code Session Log

- **Session ID:** $SESSION_ID
- **Model:** $MODEL
- **Start Time (UTC):** $(date -u +"%Y-%m-%dT%H:%M:%SZ")

---

EOF
  echo "$SESSION_FILE" > "$LOG_DIR/.current_session"
fi

TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Compute duration only if START_TIME_RAW is an epoch integer
DURATION=""
if [[ "$START_TIME_RAW" =~ ^[0-9]+$ ]]; then
  END_TIME=$(date +%s)
  DURATION=$((END_TIME - START_TIME_RAW))
fi

cat <<EOF >> "$SESSION_FILE"

## 📝 Prompt — $TIMESTAMP

### 📥 Prompt
\`\`\`
$PROMPT
\`\`\`

### 📤 Response
\`\`\`
$RESPONSE
\`\`\`

### 📊 Metrics
- **Duration:** ${DURATION:-n/a}s
- **Input Tokens:** ${INPUT_TOKENS:-n/a}
- **Output Tokens:** ${OUTPUT_TOKENS:-n/a}
- **Total Tokens:** ${TOTAL_TOKENS:-n/a}

---
EOF
