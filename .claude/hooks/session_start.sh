#!/usr/bin/env bash
set -euo pipefail

LOG_DIR="${CLAUDE_LOG_DIR:-$HOME/.claude/logs}"
mkdir -p "$LOG_DIR"

SESSION_ID="${CLAUDE_SESSION_ID:-$(date +%s)}"
TIMESTAMP=$(date -u +"%Y%m%d_%H%M%S")
MODEL="${CLAUDE_MODEL:-unknown-model}"

SESSION_FILE="$LOG_DIR/${TIMESTAMP}_session_${SESSION_ID}.md"

cat <<EOF > "$SESSION_FILE"
# Claude Code Session Log

- **Session ID:** $SESSION_ID
- **Model:** $MODEL
- **Start Time (UTC):** $(date -u +"%Y-%m-%dT%H:%M:%SZ")

---

EOF

echo "$SESSION_FILE" > "$LOG_DIR/.current_session"