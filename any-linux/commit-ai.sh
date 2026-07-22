#!/usr/bin/env bash
set -e

# ================= CONFIG =================
VERSION="1.5.0"
MAX_CHARS=14000
CONFIG_FILE="$HOME/.commit-ai.conf"
CUSTOM_PROMPT_FILE="$HOME/.commit-ai-prompt.txt"

# Defaults
DEFAULT_FORMAT="conventional"
DEFAULT_AUTO_CONFIRM="false"
DEFAULT_AUTO_SYNC="false"
DEFAULT_PROVIDER="gemini"
DEFAULT_MODEL="gemini-3-flash-preview"

# Runtime
AUTO_YES=false
SYNC_MODE=false
PREVIEW_ONLY=false
UNDO_LAST=false
EMOJI_MODE=false
CONVENTIONAL_MODE=false
SETUP_MODE=false
SHOW_CONFIG=false
EDIT_PROMPT=false
PROVIDER="gemini"
ASK_PUSH=false
USE_CUSTOM_PROMPT=false
USER_MESSAGE=""
TARGET_BRANCH=""
# =========================================

# -------------------------------------------------
# LOAD CONFIG
# -------------------------------------------------
load_config() {
  if [ -f "$CONFIG_FILE" ]; then
    while IFS='=' read -r key value; do
      [[ "$key" =~ ^#.*$ ]] && continue
      [[ -z "$key" ]] && continue
      
      key=$(echo "$key" | xargs)
      value=$(echo "$value" | xargs)
      
      case "$key" in
        format)
          [[ "$value" == "gitmoji" ]] && EMOJI_MODE=true
          ;;
        auto_confirm)
          [[ "$value" == "true" ]] && AUTO_YES=true
          ;;
        auto_sync)
          [[ "$value" == "true" ]] && SYNC_MODE=true
          ;;
        ask_push)
          [[ "$value" == "true" ]] && ASK_PUSH=true
          ;;
        use_custom_prompt)
          [[ "$value" == "true" ]] && USE_CUSTOM_PROMPT=true
          ;;
        provider)
          PROVIDER="$value"
          ;;
        model)
          DEFAULT_MODEL="$value"
          ;;
        gemini_api_key)
          [[ -z "$GEMINI_API_KEY" ]] && GEMINI_API_KEY="$value"
          ;;
        openai_api_key)
          [[ -z "$OPENAI_API_KEY" ]] && OPENAI_API_KEY="$value"
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
  local ask_push="$3"
  local use_custom_prompt="$4"
  local provider="$5"
  local model="$6"
  local gemini_key="$7"
  local openai_key="$8"

  cat > "$CONFIG_FILE" << EOF
# commit-ai configuration
# Location: ~/.commit-ai.conf

# Default commit format: conventional | gitmoji
format=$format

# Auto-confirm commits without prompt: true | false
auto_confirm=$auto_confirm

# Ask to push after commit: true | false
ask_push=$ask_push

# Use custom prompt file: true | false
use_custom_prompt=$use_custom_prompt

# AI Provider: gemini | openai
provider=$provider

# Model to use (depends on provider)
model=$model

# API Keys (optional - can also use environment variables)
$([ -n "$gemini_key" ] && echo "gemini_api_key=$gemini_key" || echo "# gemini_api_key=your_key_here")
$([ -n "$openai_key" ] && echo "openai_api_key=$openai_key" || echo "# openai_api_key=your_key_here")
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
  local current_push="false"
  local current_custom_prompt="false"
  local current_provider="gemini"
  local current_model="gemini-3-flash-preview"
  local current_gemini_key=""
  local current_openai_key=""

  if [ -f "$CONFIG_FILE" ]; then
    while IFS='=' read -r key value; do
      [[ "$key" =~ ^#.*$ ]] && continue
      [[ -z "$key" ]] && continue
      key=$(echo "$key" | xargs)
      value=$(echo "$value" | xargs)
      case "$key" in
        format) current_format="$value" ;;
        auto_confirm) current_auto="$value" ;;
        ask_push) current_push="$value" ;;
        use_custom_prompt) current_custom_prompt="$value" ;;
        provider) current_provider="$value" ;;
        model) current_model="$value" ;;
        gemini_api_key) current_gemini_key="$value" ;;
        openai_api_key) current_openai_key="$value" ;;
      esac
    done < "$CONFIG_FILE"
  fi

  # Format selection with validation
  echo "📝 Commit format:"
  echo "   1) conventional (feat:, fix:, etc.)"
  echo "   2) gitmoji (✨, 🐛, etc.)"
  echo
  local format_choice
  while true; do
    read -p "Choose format [current: $current_format] (1/2): " format_choice
    case "$format_choice" in
      1) current_format="conventional"; break ;;
      2) current_format="gitmoji"; break ;;
      "") break ;; # Keep current
      *) echo "⚠️  Invalid option. Please enter 1 or 2." ;;
    esac
  done

  # Auto-confirm with validation
  echo
  echo "⚡ Auto-confirm commits (skip confirmation prompt)?"
  local auto_choice
  while true; do
    read -p "Enable auto-confirm? [current: $current_auto] (y/n): " auto_choice
    case "$auto_choice" in
      y|Y|yes|YES) current_auto="true"; break ;;
      n|N|no|NO) current_auto="false"; break ;;
      "") break ;; # Keep current
      *) echo "⚠️  Invalid option. Please enter y or n." ;;
    esac
  done

  # Ask to push after commit
  echo
  echo "🚀 Ask to push after commit?"
  local push_choice
  while true; do
    read -p "Enable push prompt? [current: $current_push] (y/n): " push_choice
    case "$push_choice" in
      y|Y|yes|YES) current_push="true"; break ;;
      n|N|no|NO) current_push="false"; break ;;
      "") break ;; # Keep current
      *) echo "⚠️  Invalid option. Please enter y or n." ;;
    esac
  done

  # Use custom prompt
  echo
  echo "📝 Use custom prompt file (~/.commit-ai-prompt.txt)?"
  local custom_prompt_choice
  while true; do
    read -p "Enable custom prompt? [current: $current_custom_prompt] (y/n): " custom_prompt_choice
    case "$custom_prompt_choice" in
      y|Y|yes|YES) current_custom_prompt="true"; break ;;
      n|N|no|NO) current_custom_prompt="false"; break ;;
      "") break ;; # Keep current
      *) echo "⚠️  Invalid option. Please enter y or n." ;;
    esac
  done

  # Provider selection with validation
  local old_provider="$current_provider"
  echo
  echo "🔌 AI Provider:"
  echo "   1) Gemini (Google)"
  echo "   2) OpenAI (GPT)"
  echo
  local provider_choice
  while true; do
    read -p "Choose provider [current: $current_provider] (1/2): " provider_choice
    case "$provider_choice" in
      1) current_provider="gemini"; break ;;
      2) current_provider="openai"; break ;;
      "") break ;; # Keep current
      *) echo "⚠️  Invalid option. Please enter 1 or 2." ;;
    esac
  done

  # Reset model to default if provider changed
  if [[ "$old_provider" != "$current_provider" ]]; then
    if [[ "$current_provider" == "gemini" ]]; then
      current_model="gemini-3-flash-preview"
    else
      current_model="gpt-4o-mini"
    fi
  fi

  # Model selection based on provider
  echo
  echo "🧠 Model selection:"
  if [[ "$current_provider" == "gemini" ]]; then
    echo "   1) gemini-3-flash-preview (recommended)"
    echo "   2) gemini-2.5-flash"
    echo "   3) gemini-2.0-flash"
    echo "   4) gemini-2.5-pro-preview (advanced)"
    echo "   5) Custom"
    echo
    local model_choice
    while true; do
      read -p "Choose model [current: $current_model] (1-5): " model_choice
      case "$model_choice" in
        1) current_model="gemini-3-flash-preview"; break ;;
        2) current_model="gemini-2.5-flash"; break ;;
        3) current_model="gemini-2.0-flash"; break ;;
        4) current_model="gemini-2.5-pro-preview"; break ;;
        5) 
          read -p "Enter custom model name: " current_model
          break
          ;;
        "") break ;; # Keep current
        *) echo "⚠️  Invalid option. Please enter 1-5." ;;
      esac
    done
  else
    echo "   1) gpt-4o-mini (fast, recommended)"
    echo "   2) gpt-4o (advanced)"
    echo "   3) gpt-4-turbo"
    echo "   4) gpt-3.5-turbo (legacy)"
    echo "   5) Custom"
    echo
    local model_choice
    while true; do
      read -p "Choose model [current: $current_model] (1-5): " model_choice
      case "$model_choice" in
        1) current_model="gpt-4o-mini"; break ;;
        2) current_model="gpt-4o"; break ;;
        3) current_model="gpt-4-turbo"; break ;;
        4) current_model="gpt-3.5-turbo"; break ;;
        5) 
          read -p "Enter custom model name: " current_model
          break
          ;;
        "") break ;; # Keep current
        *) echo "⚠️  Invalid option. Please enter 1-5." ;;
      esac
    done
  fi

  # API Key - only ask for the selected provider
  echo
  echo "🔐 API Key configuration:"
  
  if [[ "$current_provider" == "gemini" ]]; then
    echo
    echo "  Gemini API Key:"
    if [ -n "$current_gemini_key" ]; then
      echo "    Current: ****${current_gemini_key: -4}"
    elif [ -n "$GEMINI_API_KEY" ]; then
      echo "    Using environment variable"
    else
      echo "    Not configured"
    fi
    read -p "  Enter Gemini API key (leave empty to keep): " new_gemini_key
    [ -n "$new_gemini_key" ] && current_gemini_key="$new_gemini_key"
  else
    echo
    echo "  OpenAI API Key:"
    if [ -n "$current_openai_key" ]; then
      echo "    Current: ****${current_openai_key: -4}"
    elif [ -n "$OPENAI_API_KEY" ]; then
      echo "    Using environment variable"
    else
      echo "    Not configured"
    fi
    read -p "  Enter OpenAI API key (leave empty to keep): " new_openai_key
    [ -n "$new_openai_key" ] && current_openai_key="$new_openai_key"
  fi

  # Save
  echo
  save_config "$current_format" "$current_auto" "$current_push" "$current_custom_prompt" "$current_provider" "$current_model" "$current_gemini_key" "$current_openai_key"

  echo
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "  Configuration complete! Run 'commit-ai' to start."
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  exit 0
}

# -------------------------------------------------
# EDIT CUSTOM PROMPT
# -------------------------------------------------
edit_prompt() {
  if [ ! -f "$CUSTOM_PROMPT_FILE" ]; then
    cat > "$CUSTOM_PROMPT_FILE" << 'EOF'
# Custom commit-ai prompt
# Available variables: {HISTORY}, {FILES}, {DIFF}
# Delete this file to use default prompts

You are a senior Git expert.

Recent commit history:
{HISTORY}

Staged files:
{FILES}

Relevant diff:
{DIFF}

MANDATORY RULES:
- Return ONLY the commit message
- Use imperative mood
- Max length: 72 characters
- No trailing period
EOF
    echo "📝 Created custom prompt file: $CUSTOM_PROMPT_FILE"
  fi
  
  # Try to open with default editor
  if [ -n "$EDITOR" ]; then
    "$EDITOR" "$CUSTOM_PROMPT_FILE"
  elif command -v nano &> /dev/null; then
    nano "$CUSTOM_PROMPT_FILE"
  elif command -v vim &> /dev/null; then
    vim "$CUSTOM_PROMPT_FILE"
  else
    echo "📂 Custom prompt file: $CUSTOM_PROMPT_FILE"
    echo "ℹ️  Edit this file with your preferred editor"
  fi
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
        *api_key*)
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
  echo "🔑 Environment variables:"
  echo "   GEMINI_API_KEY: $([ -n "$GEMINI_API_KEY" ] && echo "set" || echo "not set")"
  echo "   OPENAI_API_KEY: $([ -n "$OPENAI_API_KEY" ] && echo "set" || echo "not set")"
  echo
  if [ -f "$CUSTOM_PROMPT_FILE" ]; then
    echo "📝 Custom prompt: $CUSTOM_PROMPT_FILE"
  fi
  exit 0
}

# -------------------------------------------------
# HELP
# -------------------------------------------------
show_help() {
  cat << EOF
commit-ai v$VERSION — AI-powered Git commit messages

USAGE:
  commit-ai [OPTIONS]

OPTIONS:
  -e, --emoji       Use Gitmoji commit format (emoji prefix)
  -c, --conv        Use Conventional Commits format (overrides config)
  -C, --custom      Use custom prompt file (~/.commit-ai-prompt.txt)
  -m, --message     Provide context/hint for AI (e.g., -m "fix login bug")
  -b, --branch      Create to branch and deploy (e.g. -b feature-name)
  -p, --preview     Preview commit message only (no commit)
  -y, --yes         Skip confirmation prompt (auto-commit)
  -s, -S, --sync    Auto sync remote (git add -> stash -> pull -> stash pop -> add) before commit
  -u, --undo        Undo last commit (soft reset, keeps changes staged)
  --setup           Interactive configuration setup
  --config          Show current configuration
  --edit-prompt     Edit custom prompt for advanced users
  -h, --help        Show this help message
  -v, --version     Show version number

PROVIDERS:
  gemini            Google Gemini (default)
  openai            OpenAI GPT models

EXAMPLES:
  commit-ai                          # Use configured defaults
  commit-ai -e                       # Gitmoji format
  commit-ai -b "new-feature"         # Create/switch to new-feature, commit, and push
  commit-ai -m "added user auth"     # AI uses hint for better message
  commit-ai -e -m "refactored api"   # Gitmoji with context
  commit-ai -e -p                    # Preview Gitmoji message
  commit-ai -y                       # Auto-commit without confirmation

CONFIG FILE:
  Location: ~/.commit-ai.conf
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
while [[ $# -gt 0 ]]; do
  case $1 in
    --help|-h) show_help ;;
    --version|-v) show_version ;;
    --yes|-y) AUTO_YES=true ;;
    --sync|-s|-S) SYNC_MODE=true ;;
    --preview|-p) PREVIEW_ONLY=true ;;
    --undo|-u) UNDO_LAST=true ;;
    --emoji|-e) EMOJI_MODE=true ;;
    --conv|-c) CONVENTIONAL_MODE=true ;;
    --custom|-C) USE_CUSTOM_PROMPT=true ;;
    --setup) SETUP_MODE=true ;;
    --config) SHOW_CONFIG=true ;;
    --edit-prompt) EDIT_PROMPT=true ;;
    --branch|-b)
      shift
      TARGET_BRANCH="$1"
      ;;
    --message|-m)
      shift
      USER_MESSAGE="$1"
      ;;
  esac
  shift
done

# Handle mode flags (--conv overrides gitmoji default)
if $CONVENTIONAL_MODE; then
  EMOJI_MODE=false
fi

# Handle special modes
$SETUP_MODE && interactive_setup
$SHOW_CONFIG && show_config
$EDIT_PROMPT && edit_prompt

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

# Check API key based on provider
if [[ "$PROVIDER" == "openai" ]]; then
  if [ -z "$OPENAI_API_KEY" ]; then
    echo "❌ OPENAI_API_KEY is not set."
    echo "ℹ️  Run 'commit-ai --setup' to configure, or:"
    echo "   export OPENAI_API_KEY=\"your_api_key\""
    exit 1
  fi
else
  if [ -z "$GEMINI_API_KEY" ]; then
    echo "❌ GEMINI_API_KEY is not set."
    echo "ℹ️  Run 'commit-ai --setup' to configure, or:"
    echo "   export GEMINI_API_KEY=\"your_api_key\""
    exit 1
  fi
fi

# -------------------------------------------------
# CHECKOUT/CREATE TARGET BRANCH & AUTO-DETECT CURRENT
# -------------------------------------------------
ACTIVE_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "main")
if [ -z "$TARGET_BRANCH" ]; then
  TARGET_BRANCH="$ACTIVE_BRANCH"
else
  if git show-ref --verify --quiet "refs/heads/$TARGET_BRANCH"; then
    echo "🔀 Mudando para a branch existente '$TARGET_BRANCH'..."
    git checkout "$TARGET_BRANCH"
  else
    echo "🌱 Criando e mudando para a nova branch '$TARGET_BRANCH'..."
    git checkout -b "$TARGET_BRANCH"
  fi
  echo
fi

# -------------------------------------------------
# AUTO SYNC REMOTE (ADD -> STASH -> PULL -> POP -> ADD)
# -------------------------------------------------
if $SYNC_MODE; then
  CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "main")
  echo "🔄 Auto-sync ativado. Sincronizando com a branch remota origin/$CURRENT_BRANCH..."

  # Adiciona tudo para verificar modificações
  git add .

  HAD_STASH=false
  if ! git diff --cached --quiet || ! git diff --quiet; then
    echo "📦 Guardando alterações locais temporariamente (git stash)..."
    git stash push -u -m "commit-ai-auto-stash" >/dev/null 2>&1 && HAD_STASH=true || HAD_STASH=false
  fi

  echo "⬇️ Baixando atualizações remotas (git pull origin $CURRENT_BRANCH)..."
  git pull origin "$CURRENT_BRANCH" --rebase >/dev/null 2>&1 || git pull origin "$CURRENT_BRANCH" >/dev/null 2>&1 || true

  if $HAD_STASH; then
    echo "📂 Restaurando alterações salvas (git stash pop)..."
    git stash pop >/dev/null 2>&1 || true
  fi

  echo "➕ Adicionando arquivos para staging (git add .)..."
  git add .
  echo
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
if $USE_CUSTOM_PROMPT && [ -f "$CUSTOM_PROMPT_FILE" ]; then
  PROMPT=$(cat "$CUSTOM_PROMPT_FILE" | grep -v '^#')
  PROMPT="${PROMPT//\{HISTORY\}/$HISTORY}"
  PROMPT="${PROMPT//\{FILES\}/$FILES}"
  PROMPT="${PROMPT//\{DIFF\}/$DIFF}"
elif $EMOJI_MODE; then
  USER_HINT=""
  if [ -n "$USER_MESSAGE" ]; then
    USER_HINT="
User context/hint for this commit:
$USER_MESSAGE

Use this hint to better understand the intent and generate a more accurate commit message.
"
  fi
  PROMPT=$(cat <<EOF
You are a senior Git and Gitmoji expert.
$USER_HINT
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
  USER_HINT=""
  if [ -n "$USER_MESSAGE" ]; then
    USER_HINT="
User context/hint for this commit:
$USER_MESSAGE

Use this hint to better understand the intent and generate a more accurate commit message.
"
  fi
  PROMPT=$(cat <<EOF
You are a senior Git and Conventional Commits expert.
$USER_HINT
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
# API REQUEST
# -------------------------------------------------
if [[ "$PROVIDER" == "openai" ]]; then
  # OpenAI API
  JSON=$(jq -n --arg text "$PROMPT" '{
    model: "'"$DEFAULT_MODEL"'",
    messages: [{ role: "user", content: $text }],
    max_tokens: 100
  }')

  RESPONSE=$(curl -s \
    -X POST \
    "https://api.openai.com/v1/chat/completions" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $OPENAI_API_KEY" \
    -d "$JSON"
  )

  COMMIT_MSG=$(echo "$RESPONSE" | jq -r '.choices[0].message.content')
else
  # Gemini API
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
fi

if [ -z "$COMMIT_MSG" ] || [ "$COMMIT_MSG" = "null" ]; then
  echo "❌ Failed to generate commit message."
  if [[ "$PROVIDER" == "openai" ]]; then
    echo "$RESPONSE" | jq -r '.error.message // empty'
  else
    echo "$RESPONSE" | jq -r '.error.message // empty'
  fi
  exit 1
fi

# -------------------------------------------------
# NORMALIZATION
# -------------------------------------------------
# Remove newlines and trim whitespace (without xargs to avoid quote issues)
COMMIT_MSG=$(echo "$COMMIT_MSG" | tr -d '\n' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')

if $EMOJI_MODE; then
  COMMIT_MSG=$(echo "$COMMIT_MSG" | sed -E 's/^(\X)([^ ])/\1 \2/')
  COMMIT_MSG=$(echo "$COMMIT_MSG" | sed -E 's/^(\X )([a-z])/\1\u\2/')
else
  COMMIT_MSG=$(echo "$COMMIT_MSG" | sed -E 's/^([a-z]+): ([A-Z])/\1: \L\2/')
fi

# -------------------------------------------------
# OUTPUT
# -------------------------------------------------
if $PREVIEW_ONLY; then
  echo
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "  ℹ️  Preview Mode"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo
  echo "  $COMMIT_MSG"
  echo
  exit 0
fi

if ! $AUTO_YES; then
  echo
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "  📝 Generated Commit Message"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo
  read -e -i "$COMMIT_MSG" EDITED_MSG
  [ -n "$EDITED_MSG" ] && COMMIT_MSG="$EDITED_MSG"
  echo
fi

git commit -m "$COMMIT_MSG"
echo
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ✅ Commit created successfully!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Auto push to target/current branch
if [ -n "$TARGET_BRANCH" ]; then
  echo
  echo "🚀 Enviando alterações para origin/$TARGET_BRANCH..."
  git push -u origin "$TARGET_BRANCH" || git push origin "$TARGET_BRANCH"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "  🚀 Enviado para origin/$TARGET_BRANCH com sucesso!"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
elif $ASK_PUSH; then
  echo
  read -p "🚀 Push para o repositório remoto? (y/n): " push_choice
  case "$push_choice" in
    y|Y|yes|YES)
      echo
      git push
      echo
      echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
      echo "  🚀 Pushed para o repositório remoto com sucesso!"
      echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
      ;;
  esac
fi
