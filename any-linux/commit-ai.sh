#!/usr/bin/env bash
set -e

# ================= CONFIG =================
VERSION="1.2.0"
MAX_CHARS=14000
CONFIG_FILE="$HOME/.commit-ai.conf"

# Defaults
DEFAULT_FORMAT="conventional"
DEFAULT_AUTO_CONFIRM="false"
DEFAULT_MODEL="gemini-2.0-flash"

# Runtime
AUTO_YES=false
PREVIEW_ONLY=false
UNDO_LAST=false
EMOJI_MODE=false
SETUP_MODE=false
SHOW_CONFIG=false
# =========================================

# -------------------------------------------------
# LOAD CONFIG
# -------------------------------------------------
load_config() {
  if [ -f "$CONFIG_FILE" ]; then
    while IFS='=' read -r key value; do
      # Skip comments and empty lines
      [[ "$key" =~ ^#.*$ ]] && continue
      [[ -z "$key" ]] && continue
      
      # Trim whitespace
      key=$(echo "$key" | xargs)
      value=$(echo "$value" | xargs)
      
      case "$key" in
        format)
          [[ "$value" == "gitmoji" ]] && EMOJI_MODE=true
          ;;
        auto_confirm)
          [[ "$value" == "true" ]] && AUTO_YES=true
          ;;
        model)
          DEFAULT_MODEL="$value"
          ;;
        api_key)
          [[ -z "$GEMINI_API_KEY" ]] && GEMINI_API_KEY="$value"
          ;;
      esac
    done < "$CONFIG_FILE"
  fi
}

# -------------------------------------------------
# SAVE CONFIG
# -------------------------------------------------
save_config() {
  local format="$1"
  local auto_confirm="$2"
  local model="$3"
  local api_key="$4"

  cat > "$CONFIG_FILE" << EOF
# commit-ai configuration
# Location: ~/.commit-ai.conf

# Default commit format: conventional | gitmoji
format=$format

# Auto-confirm commits without prompt: true | false
auto_confirm=$auto_confirm

# Gemini model to use
model=$model

# API Key (optional - can also use GEMINI_API_KEY environment variable)
$([ -n "$api_key" ] && echo "api_key=$api_key" || echo "# api_key=your_key_here")
EOF
  echo "✅ Configuration saved to $CONFIG_FILE"
}

# -------------------------------------------------
# INTERACTIVE SETUP
# -------------------------------------------------
interactive_setup() {
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "  🤖 commit-ai v$VERSION - Configuration Setup"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo

  # Load existing config for defaults
  local current_format="conventional"
  local current_auto="false"
  local current_model="gemini-2.0-flash"
  local current_key=""

  if [ -f "$CONFIG_FILE" ]; then
    while IFS='=' read -r key value; do
      [[ "$key" =~ ^#.*$ ]] && continue
      [[ -z "$key" ]] && continue
      key=$(echo "$key" | xargs)
      value=$(echo "$value" | xargs)
      case "$key" in
        format) current_format="$value" ;;
        auto_confirm) current_auto="$value" ;;
        model) current_model="$value" ;;
        api_key) current_key="$value" ;;
      esac
    done < "$CONFIG_FILE"
  fi

  # Format selection
  echo "📝 Commit format:"
  echo "   1) conventional (feat:, fix:, etc.)"
  echo "   2) gitmoji (✨, 🐛, etc.)"
  echo
  local format_choice
  read -p "Choose format [current: $current_format] (1/2): " format_choice
  case "$format_choice" in
    1) current_format="conventional" ;;
    2) current_format="gitmoji" ;;
    "") ;; # Keep current
  esac

  # Auto-confirm
  echo
  echo "⚡ Auto-confirm commits (skip confirmation prompt)?"
  local auto_choice
  read -p "Enable auto-confirm? [current: $current_auto] (y/n): " auto_choice
  case "$auto_choice" in
    y|Y|yes) current_auto="true" ;;
    n|N|no) current_auto="false" ;;
    "") ;; # Keep current
  esac

  # Model selection
  echo
  echo "🧠 Gemini model:"
  echo "   1) gemini-2.0-flash (fast, recommended)"
  echo "   2) gemini-2.0-flash-lite (faster, lighter)"
  echo "   3) gemini-2.5-pro-preview (advanced)"
  echo "   4) Custom"
  echo
  local model_choice
  read -p "Choose model [current: $current_model] (1/2/3/4): " model_choice
  case "$model_choice" in
    1) current_model="gemini-2.0-flash" ;;
    2) current_model="gemini-2.0-flash-lite" ;;
    3) current_model="gemini-2.5-pro-preview" ;;
    4) 
      read -p "Enter custom model name: " current_model
      ;;
    "") ;; # Keep current
  esac

  # API Key
  echo
  echo "🔐 API Key configuration:"
  if [ -n "$current_key" ]; then
    echo "   Current: ****${current_key: -4}"
  elif [ -n "$GEMINI_API_KEY" ]; then
    echo "   Using environment variable GEMINI_API_KEY"
  else
    echo "   Not configured"
  fi
  echo
  read -p "Enter new API key (leave empty to keep current): " new_key
  [ -n "$new_key" ] && current_key="$new_key"

  # Save
  echo
  save_config "$current_format" "$current_auto" "$current_model" "$current_key"


  echo
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "  Configuration complete! Run 'commit-ai' to start."
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  exit 0
}

# -------------------------------------------------
# SHOW CONFIG
# -------------------------------------------------
show_config() {
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "  🤖 commit-ai v$VERSION - Current Config"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo
  if [ -f "$CONFIG_FILE" ]; then
    echo "📂 Config file: $CONFIG_FILE"
    echo
    while IFS='=' read -r key value; do
      [[ "$key" =~ ^#.*$ ]] && continue
      [[ -z "$key" ]] && continue
      key=$(echo "$key" | xargs)
      value=$(echo "$value" | xargs)
      case "$key" in
        api_key)
          [ -n "$value" ] && echo "   $key = ****${value: -4}" || echo "   $key = (not set)"
          ;;
        *)
          echo "   $key = $value"
          ;;
      esac
    done < "$CONFIG_FILE"
  else
    echo "📂 No config file found at $CONFIG_FILE"
    echo "   Run 'commit-ai --setup' to create one."
  fi
  echo
  echo "🔑 GEMINI_API_KEY env: $([ -n "$GEMINI_API_KEY" ] && echo "set" || echo "not set")"
  echo
  exit 0
}

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
  -s, --setup     Interactive configuration setup
  -c, --config    Show current configuration
  -h, --help      Show this help message
  -v, --version   Show version number

EXAMPLES:
  commit-ai              # Conventional Commits format
  commit-ai -e           # Gitmoji format
  commit-ai -e -p        # Preview Gitmoji message
  commit-ai -y           # Auto-commit without confirmation
  commit-ai -u           # Undo last commit
  commit-ai --setup      # Configure preferences

CONFIG FILE:
  Location: ~/.commit-ai.conf
  
  Available settings:
    format=conventional|gitmoji
    auto_confirm=true|false
    model=gemini-2.0-flash
    api_key=your_key_here

ENVIRONMENT:
  GEMINI_API_KEY         Your Google Gemini API key (or set in config)

MORE INFO:
  https://github.com/jhowk14/commit-ai
EOF
  exit 0
}

show_version() {
  echo "commit-ai v$VERSION"
  exit 0
}

# Load config first
load_config

# Parse flags (override config)
for arg in "$@"; do
  case $arg in
    --help|-h) show_help ;;
    --version|-v) show_version ;;
    --yes|-y) AUTO_YES=true ;;
    --preview|-p) PREVIEW_ONLY=true ;;
    --undo|-u) UNDO_LAST=true ;;
    --emoji|-e) EMOJI_MODE=true ;;
    --setup|-s) SETUP_MODE=true ;;
    --config|-c) SHOW_CONFIG=true ;;
  esac
done

# Handle special modes
$SETUP_MODE && interactive_setup
$SHOW_CONFIG && show_config

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
  echo "ℹ️  Run 'commit-ai --setup' to configure, or:"
  echo "   export GEMINI_API_KEY=\"your_api_key\""
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
  "https://generativelanguage.googleapis.com/v1beta/models/$DEFAULT_MODEL:generateContent" \
  -H "Content-Type: application/json" \
  -H "x-goog-api-key: $GEMINI_API_KEY" \
  -d "$JSON"
)

COMMIT_MSG=$(echo "$RESPONSE" | jq -r '.candidates[0].content.parts[0].text')

if [ -z "$COMMIT_MSG" ] || [ "$COMMIT_MSG" = "null" ]; then
  echo "❌ Failed to generate commit message."
  echo "$RESPONSE" | jq -r '.error.message // empty'
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
