#!/usr/bin/env bash
set -euo pipefail

LOG_DIR="${CLAUDE_LOG_DIR:-$HOME/.claude/logs}"
mkdir -p "$LOG_DIR"

payload="$(cat)"

SESSION_ID=$(echo "$payload" | jq -r '.session_id // ""' 2>/dev/null || true)
SESSION_ID="${SESSION_ID:-$(date +%s)}"

TIMESTAMP=$(date -u +"%Y%m%d_%H%M%S")
SESSION_FILE="$LOG_DIR/${TIMESTAMP}_session_${SESSION_ID}.md"

cat <<EOF > "$SESSION_FILE"
# Claude Code Session Log

- **Session ID:** $SESSION_ID
- **Start Time (UTC):** $(date -u +"%Y-%m-%dT%H:%M:%SZ")

---

EOF

echo "$SESSION_FILE" > "$LOG_DIR/.current_session"
