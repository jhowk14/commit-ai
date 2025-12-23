#!/usr/bin/env bash
set -e

# ================= CONFIG =================
VERSION="1.1.0"
MAX_CHARS=14000
AUTO_YES=false
PREVIEW_ONLY=false
UNDO_LAST=false
EMOJI_MODE=false
# =========================================

# -------------------------------------------------
# HELP
# -------------------------------------------------
show_help() {
  cat << EOF
commit-ai v$VERSION — AI-powered Git commit messages using Gemini

USAGE:
  commit-ai [OPTIONS]

OPTIONS:
  -e, --emoji     Use Gitmoji commit format (emoji prefix)
  -p, --preview   Preview commit message only (no commit)
  -y, --yes       Skip confirmation prompt (auto-commit)
  -u, --undo      Undo last commit (soft reset, keeps changes staged)
  -h, --help      Show this help message
  -v, --version   Show version number

EXAMPLES:
  commit-ai              # Conventional Commits format
  commit-ai -e           # Gitmoji format
  commit-ai -e -p        # Preview Gitmoji message
  commit-ai -y           # Auto-commit without confirmation
  commit-ai -u           # Undo last commit

ENVIRONMENT:
  GEMINI_API_KEY         Required. Your Google Gemini API key.

MORE INFO:
  https://github.com/jhowk14/commit-ai
EOF
  exit 0
}

show_version() {
  echo "commit-ai v$VERSION"
  exit 0
}

# Flags
for arg in "$@"; do
  case $arg in
    --help|-h) show_help ;;
    --version|-v) show_version ;;
    --yes|-y) AUTO_YES=true ;;
    --preview|-p) PREVIEW_ONLY=true ;;
    --undo|-u) UNDO_LAST=true ;;
    --emoji|-e) EMOJI_MODE=true ;;
  esac
done

# -------------------------------------------------
# UNDO LAST COMMIT
# -------------------------------------------------
if $UNDO_LAST; then
  LAST_MSG=$(git log -1 --pretty=%B)
  git reset --soft HEAD~1
  echo "↩️  Undone last commit:"
  echo "$LAST_MSG"
  exit 0
fi

# -------------------------------------------------
# DEPENDENCIES
# -------------------------------------------------
for cmd in git jq curl; do
  command -v "$cmd" >/dev/null 2>&1 || {
    echo "❌ Missing dependency: $cmd"
    exit 1
  }
done

if [ -z "$GEMINI_API_KEY" ]; then
  echo "❌ GEMINI_API_KEY is not set."
  exit 1
fi

# -------------------------------------------------
# STAGING CHECK
# -------------------------------------------------
if git diff --cached --quiet; then
  echo "❌ No staged changes found."
  echo "ℹ️ Run: git add <files>"
  exit 1
fi

DIFF=$(git diff --cached --unified=3 \
  | grep -E '^\+|^-|^@@|^diff --git' \
  | head -c "$MAX_CHARS")

FILES=$(git diff --cached --name-only)
HISTORY=$(git log --oneline -n 20)

# -------------------------------------------------
# PROMPT SELECTION
# -------------------------------------------------
if $EMOJI_MODE; then
  PROMPT=$(cat <<EOF
You are a senior Git and Gitmoji expert.

Recent commit history:
$HISTORY

Staged files:
$FILES

Relevant diff:
$DIFF

MANDATORY RULES:
- Choose ONLY ONE gitmoji (emoji)
- EXACT format: <emoji><space><Message>
- First letter MUST be CAPITALIZED
- Message in English
- Use imperative mood
- Max length: 72 characters
- No trailing period
- Return ONLY the final commit message
EOF
)
else
  PROMPT=$(cat <<EOF
You are a senior Git and Conventional Commits expert.

Recent commit history:
$HISTORY

Staged files:
$FILES

Relevant diff:
$DIFF

MANDATORY RULES:
- Choose ONE action type
- EXACT format: <type>: <message>
- Message starts with lowercase
- Message in English
- Use imperative mood
- Max length: 72 characters
- No trailing period
- Return ONLY the final commit message

AVAILABLE TYPES:
feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert
EOF
)
fi

# -------------------------------------------------
# GEMINI REQUEST
# -------------------------------------------------
JSON=$(jq -n --arg text "$PROMPT" '{
  contents: [{ parts: [{ text: $text }] }]
}')

RESPONSE=$(curl -s \
  -X POST \
  "https://generativelanguage.googleapis.com/v1beta/models/gemini-3-flash-preview:generateContent" \
  -H "Content-Type: application/json" \
  -H "x-goog-api-key: $GEMINI_API_KEY" \
  -d "$JSON"
)

COMMIT_MSG=$(echo "$RESPONSE" | jq -r '.candidates[0].content.parts[0].text')

if [ -z "$COMMIT_MSG" ] || [ "$COMMIT_MSG" = "null" ]; then
  echo "❌ Failed to generate commit message."
  exit 1
fi

# -------------------------------------------------
# NORMALIZATION
# -------------------------------------------------
COMMIT_MSG=$(echo "$COMMIT_MSG" | tr -d '\n' | xargs)

if $EMOJI_MODE; then
  # ensure emoji + space
  COMMIT_MSG=$(echo "$COMMIT_MSG" | sed -E 's/^(\X)([^ ])/\1 \2/')
  COMMIT_MSG=$(echo "$COMMIT_MSG" | sed -E 's/^(\X )([a-z])/\1\u\2/')
else
  # ensure lowercase after colon
  COMMIT_MSG=$(echo "$COMMIT_MSG" | sed -E 's/^([a-z]+): ([A-Z])/\1: \L\2/')
fi

# -------------------------------------------------
# OUTPUT
# -------------------------------------------------
if $PREVIEW_ONLY; then
  echo "ℹ️ Preview:"
  echo "$COMMIT_MSG"
  exit 0
fi

if ! $AUTO_YES; then
  echo
  read -e -i "$COMMIT_MSG" EDITED_MSG
  [ -n "$EDITED_MSG" ] && COMMIT_MSG="$EDITED_MSG"
fi

git commit -m "$COMMIT_MSG"
echo "✅ Commit created!"
