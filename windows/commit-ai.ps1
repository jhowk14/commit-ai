<#
.SYNOPSIS
  commit-ai – AI-powered commit message generator.

.DESCRIPTION
  commit-ai generates conventional commit messages using AI providers
  based on the current staged git changes.

.AUTHOR
  Jonathan Henrique Perozi Lourenço (jhowk14)

.LICENSE
  MIT License

.COPYRIGHT
  Copyright (c) 2026 Jonathan Henrique Perozi Lourenço (jhowk14)
#>

# Licensed under the MIT License

#Requires -Version 5.1
[CmdletBinding()]
param(
    [Alias('e')][switch]$Emoji,
    [Alias('c')][switch]$Conv,
    [Alias('p')][switch]$Preview,
    [Alias('y')][switch]$Yes,
    [Alias('u')][switch]$Undo,
    [Alias('s')][switch]$Setup,
    [Alias('m')][string]$Message,
    [switch]$Config,
    [switch]$EditPrompt,
    [Alias('h')][switch]$Help,
    [Alias('v')][switch]$Version
)

$ErrorActionPreference = 'Stop'

# ================= CONFIG =================
$SCRIPT_VERSION = "1.3.0"
$MAX_CHARS = 14000
$CONFIG_FILE = Join-Path $HOME ".commit-ai.conf"
$CUSTOM_PROMPT_FILE = Join-Path $HOME ".commit-ai-prompt.txt"

# Defaults
$script:PROVIDER = "gemini"
$script:DEFAULT_MODEL = "gemini-3-flash-preview"
$script:EMOJI_MODE = $false
$script:AUTO_YES = $false
$script:ASK_PUSH = $false
# ==========================================

# -------------------------------------------------
# LOAD CONFIG
# -------------------------------------------------
function Load-Config {
    if (Test-Path $CONFIG_FILE) {
        Get-Content $CONFIG_FILE | ForEach-Object {
            $line = $_.Trim()
            if ($line -match '^#' -or [string]::IsNullOrWhiteSpace($line)) { return }
            
            $parts = $line -split '=', 2
            if ($parts.Count -eq 2) {
                $key = $parts[0].Trim()
                $value = $parts[1].Trim()
                
                switch ($key) {
                    'format' { 
                        if ($value -eq 'gitmoji') { $script:EMOJI_MODE = $true }
                    }
                    'auto_confirm' { 
                        if ($value -eq 'true') { $script:AUTO_YES = $true }
                    }
                    'ask_push' { 
                        if ($value -eq 'true') { $script:ASK_PUSH = $true }
                    }
                    'provider' {
                        $script:PROVIDER = $value
                    }
                    'model' { 
                        $script:DEFAULT_MODEL = $value 
                    }
                    'gemini_api_key' { 
                        if ([string]::IsNullOrEmpty($env:GEMINI_API_KEY)) {
                            $env:GEMINI_API_KEY = $value
                        }
                    }
                    'openai_api_key' { 
                        if ([string]::IsNullOrEmpty($env:OPENAI_API_KEY)) {
                            $env:OPENAI_API_KEY = $value
                        }
                    }
                }
            }
        }
    }
}

# -------------------------------------------------
# SAVE CONFIG
# -------------------------------------------------
function Save-Config {
    param(
        [string]$Format,
        [string]$AutoConfirm,
        [string]$AskPush,
        [string]$Provider,
        [string]$Model,
        [string]$GeminiKey,
        [string]$OpenAIKey
    )

    $geminiLine = if ($GeminiKey) { "gemini_api_key=$GeminiKey" } else { "# gemini_api_key=your_key_here" }
    $openaiLine = if ($OpenAIKey) { "openai_api_key=$OpenAIKey" } else { "# openai_api_key=your_key_here" }
    
    $configContent = @"
# commit-ai configuration
# Location: ~/.commit-ai.conf

# Default commit format: conventional | gitmoji
format=$Format

# Auto-confirm commits without prompt: true | false
auto_confirm=$AutoConfirm

# Ask to push after commit: true | false
ask_push=$AskPush

# AI Provider: gemini | openai
provider=$Provider

# Model to use (depends on provider)
model=$Model

# API Keys (optional - can also use environment variables)
$geminiLine
$openaiLine
"@
    
    Set-Content -Path $CONFIG_FILE -Value $configContent -Encoding UTF8
    Write-Host "✅ Configuration saved to $CONFIG_FILE" -ForegroundColor Green
}

# -------------------------------------------------
# INTERACTIVE SETUP
# -------------------------------------------------
function Interactive-Setup {
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host "  🤖 commit-ai v$SCRIPT_VERSION - Configuration Setup" -ForegroundColor White
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host

    # Load existing config
    $currentFormat = "conventional"
    $currentAuto = "false"
    $currentPush = "false"
    $currentProvider = "gemini"
    $currentModel = "gemini-3-flash-preview"
    $currentGeminiKey = ""
    $currentOpenAIKey = ""

    if (Test-Path $CONFIG_FILE) {
        Get-Content $CONFIG_FILE | ForEach-Object {
            $line = $_.Trim()
            if ($line -match '^#' -or [string]::IsNullOrWhiteSpace($line)) { return }
            $parts = $line -split '=', 2
            if ($parts.Count -eq 2) {
                $key = $parts[0].Trim()
                $value = $parts[1].Trim()
                switch ($key) {
                    'format' { $currentFormat = $value }
                    'auto_confirm' { $currentAuto = $value }
                    'ask_push' { $currentPush = $value }
                    'provider' { $currentProvider = $value }
                    'model' { $currentModel = $value }
                    'gemini_api_key' { $currentGeminiKey = $value }
                    'openai_api_key' { $currentOpenAIKey = $value }
                }
            }
        }
    }

    # Format selection with validation
    Write-Host "📝 Commit format:" -ForegroundColor Yellow
    Write-Host "   1) conventional (feat:, fix:, etc.)"
    Write-Host "   2) gitmoji (✨, 🐛, etc.)"
    Write-Host
    do {
        $formatChoice = Read-Host "Choose format [current: $currentFormat] (1/2)"
        $valid = $true
        switch ($formatChoice) {
            "1" { $currentFormat = "conventional" }
            "2" { $currentFormat = "gitmoji" }
            "" { } # Keep current
            default { 
                Write-Host "⚠️  Invalid option. Please enter 1 or 2." -ForegroundColor Yellow
                $valid = $false
            }
        }
    } while (-not $valid)

    # Auto-confirm with validation
    Write-Host
    Write-Host "⚡ Auto-confirm commits (skip confirmation prompt)?" -ForegroundColor Yellow
    do {
        $autoChoice = Read-Host "Enable auto-confirm? [current: $currentAuto] (y/n)"
        $valid = $true
        switch ($autoChoice.ToLower()) {
            "y" { $currentAuto = "true" }
            "yes" { $currentAuto = "true" }
            "n" { $currentAuto = "false" }
            "no" { $currentAuto = "false" }
            "" { } # Keep current
            default {
                Write-Host "⚠️  Invalid option. Please enter y or n." -ForegroundColor Yellow
                $valid = $false
            }
        }
    } while (-not $valid)

    # Ask to push after commit
    Write-Host
    Write-Host "🚀 Ask to push after commit?" -ForegroundColor Yellow
    do {
        $pushChoice = Read-Host "Enable push prompt? [current: $currentPush] (y/n)"
        $valid = $true
        switch ($pushChoice.ToLower()) {
            "y" { $currentPush = "true" }
            "yes" { $currentPush = "true" }
            "n" { $currentPush = "false" }
            "no" { $currentPush = "false" }
            "" { } # Keep current
            default {
                Write-Host "⚠️  Invalid option. Please enter y or n." -ForegroundColor Yellow
                $valid = $false
            }
        }
    } while (-not $valid)

    # Provider selection with validation
    $oldProvider = $currentProvider
    Write-Host
    Write-Host "🔌 AI Provider:" -ForegroundColor Yellow
    Write-Host "   1) Gemini (Google)"
    Write-Host "   2) OpenAI (GPT)"
    Write-Host
    do {
        $providerChoice = Read-Host "Choose provider [current: $currentProvider] (1/2)"
        $valid = $true
        switch ($providerChoice) {
            "1" { $currentProvider = "gemini" }
            "2" { $currentProvider = "openai" }
            "" { } # Keep current
            default {
                Write-Host "⚠️  Invalid option. Please enter 1 or 2." -ForegroundColor Yellow
                $valid = $false
            }
        }
    } while (-not $valid)

    # Reset model to default if provider changed
    if ($oldProvider -ne $currentProvider) {
        if ($currentProvider -eq "gemini") {
            $currentModel = "gemini-3-flash-preview"
        } else {
            $currentModel = "gpt-4o-mini"
        }
    }

    # Model selection based on provider
    Write-Host
    Write-Host "🧠 Model selection:" -ForegroundColor Yellow
    if ($currentProvider -eq "gemini") {
        Write-Host "   1) gemini-3-flash-preview (recommended)"
        Write-Host "   2) gemini-2.5-flash"
        Write-Host "   3) gemini-2.0-flash"
        Write-Host "   4) gemini-2.5-pro-preview (advanced)"
        Write-Host "   5) Custom"
    } else {
        Write-Host "   1) gpt-4o-mini (fast, recommended)"
        Write-Host "   2) gpt-4o (advanced)"
        Write-Host "   3) gpt-4-turbo"
        Write-Host "   4) gpt-3.5-turbo (legacy)"
        Write-Host "   5) Custom"
    }
    Write-Host
    do {
        $modelChoice = Read-Host "Choose model [current: $currentModel] (1-5)"
        $valid = $true
        if ($currentProvider -eq "gemini") {
            switch ($modelChoice) {
                "1" { $currentModel = "gemini-3-flash-preview" }
                "2" { $currentModel = "gemini-2.5-flash" }
                "3" { $currentModel = "gemini-2.0-flash" }
                "4" { $currentModel = "gemini-2.5-pro-preview" }
                "5" { $currentModel = Read-Host "Enter custom model name" }
                "" { }
                default {
                    Write-Host "⚠️  Invalid option. Please enter 1-5." -ForegroundColor Yellow
                    $valid = $false
                }
            }
        } else {
            switch ($modelChoice) {
                "1" { $currentModel = "gpt-4o-mini" }
                "2" { $currentModel = "gpt-4o" }
                "3" { $currentModel = "gpt-4-turbo" }
                "4" { $currentModel = "gpt-3.5-turbo" }
                "5" { $currentModel = Read-Host "Enter custom model name" }
                "" { }
                default {
                    Write-Host "⚠️  Invalid option. Please enter 1-5." -ForegroundColor Yellow
                    $valid = $false
                }
            }
        }
    } while (-not $valid)

    # API Key - only ask for the selected provider
    Write-Host
    Write-Host "🔐 API Key configuration:" -ForegroundColor Yellow
    
    if ($currentProvider -eq "gemini") {
        Write-Host
        Write-Host "  Gemini API Key:"
        if ($currentGeminiKey) {
            Write-Host "    Current: ****$($currentGeminiKey.Substring([Math]::Max(0, $currentGeminiKey.Length - 4)))"
        } elseif ($env:GEMINI_API_KEY) {
            Write-Host "    Using environment variable"
        } else {
            Write-Host "    Not configured"
        }
        $newGeminiKey = Read-Host "  Enter Gemini API key (leave empty to keep)"
        if ($newGeminiKey) { $currentGeminiKey = $newGeminiKey }
    } else {
        Write-Host
        Write-Host "  OpenAI API Key:"
        if ($currentOpenAIKey) {
            Write-Host "    Current: ****$($currentOpenAIKey.Substring([Math]::Max(0, $currentOpenAIKey.Length - 4)))"
        } elseif ($env:OPENAI_API_KEY) {
            Write-Host "    Using environment variable"
        } else {
            Write-Host "    Not configured"
        }
        $newOpenAIKey = Read-Host "  Enter OpenAI API key (leave empty to keep)"
        if ($newOpenAIKey) { $currentOpenAIKey = $newOpenAIKey }
    }

    # Save
    Write-Host
    Save-Config -Format $currentFormat -AutoConfirm $currentAuto -AskPush $currentPush -Provider $currentProvider -Model $currentModel -GeminiKey $currentGeminiKey -OpenAIKey $currentOpenAIKey

    Write-Host
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host "  Configuration complete! Run 'commit-ai' to start." -ForegroundColor White
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    exit 0
}

# -------------------------------------------------
# EDIT CUSTOM PROMPT
# -------------------------------------------------
function Edit-Prompt {
    if (-not (Test-Path $CUSTOM_PROMPT_FILE)) {
        $defaultPrompt = @"
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
"@
        Set-Content -Path $CUSTOM_PROMPT_FILE -Value $defaultPrompt -Encoding UTF8
        Write-Host "📝 Created custom prompt file: $CUSTOM_PROMPT_FILE" -ForegroundColor Green
    }
    
    # Try to open with notepad
    Start-Process notepad.exe -ArgumentList $CUSTOM_PROMPT_FILE -Wait
    exit 0
}

# -------------------------------------------------
# SHOW CONFIG
# -------------------------------------------------
function Show-Config {
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host "  🤖 commit-ai v$SCRIPT_VERSION - Current Config" -ForegroundColor White
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host
    
    if (Test-Path $CONFIG_FILE) {
        Write-Host "📂 Config file: $CONFIG_FILE" -ForegroundColor Gray
        Write-Host
        Get-Content $CONFIG_FILE | ForEach-Object {
            $line = $_.Trim()
            if ($line -match '^#' -or [string]::IsNullOrWhiteSpace($line)) { return }
            $parts = $line -split '=', 2
            if ($parts.Count -eq 2) {
                $key = $parts[0].Trim()
                $value = $parts[1].Trim()
                if ($key -like '*api_key*' -and $value) {
                    Write-Host "   $key = ****$($value.Substring([Math]::Max(0, $value.Length - 4)))"
                } else {
                    Write-Host "   $key = $value"
                }
            }
        }
    } else {
        Write-Host "📂 No config file found at $CONFIG_FILE" -ForegroundColor Yellow
        Write-Host "   Run 'commit-ai -Setup' to create one."
    }
    
    Write-Host
    Write-Host "🔑 Environment variables:" -ForegroundColor Gray
    $geminiStatus = if ($env:GEMINI_API_KEY) { "set" } else { "not set" }
    $openaiStatus = if ($env:OPENAI_API_KEY) { "set" } else { "not set" }
    Write-Host "   GEMINI_API_KEY: $geminiStatus"
    Write-Host "   OPENAI_API_KEY: $openaiStatus"
    Write-Host
    
    if (Test-Path $CUSTOM_PROMPT_FILE) {
        Write-Host "📝 Custom prompt: $CUSTOM_PROMPT_FILE" -ForegroundColor Gray
    }
    exit 0
}

# -------------------------------------------------
# HELP
# -------------------------------------------------
function Show-Help {
    @"
commit-ai v$SCRIPT_VERSION - AI-powered Git commit messages

USAGE:
  .\commit-ai.ps1 [OPTIONS]

OPTIONS:
  -Emoji, -e        Use Gitmoji commit format (emoji prefix)
  -Conv, -c         Use Conventional Commits format (overrides config)
  -Message, -m      Provide context/hint for AI (e.g., -m "fix login bug")
  -Preview, -p      Preview commit message only (no commit)
  -Yes, -y          Skip confirmation prompt (auto-commit)
  -Undo, -u         Undo last commit (soft reset, keeps changes staged)
  -Setup, -s        Interactive configuration setup
  -Config           Show current configuration
  -EditPrompt       Edit custom prompt for advanced users
  -Help, -h         Show this help message
  -Version, -v      Show version number

PROVIDERS:
  gemini            Google Gemini (default)
  openai            OpenAI GPT models

EXAMPLES:
  .\commit-ai.ps1                          # Use configured defaults
  .\commit-ai.ps1 -Emoji                   # Gitmoji format
  .\commit-ai.ps1 -c                       # Conventional format
  .\commit-ai.ps1 -m "added user auth"     # AI uses hint for better message
  .\commit-ai.ps1 -e -m "refactored api"   # Gitmoji with context
  .\commit-ai.ps1 -e -p                    # Preview Gitmoji message
  .\commit-ai.ps1 -Setup                   # Configure preferences
  .\commit-ai.ps1 -EditPrompt              # Customize AI prompt

CONFIG FILE:
  Location: ~/.commit-ai.conf

MORE INFO:
  https://github.com/jhowk14/commit-ai
"@
    exit 0
}

function Show-Version {
    Write-Host "commit-ai v$SCRIPT_VERSION"
    exit 0
}

# Load config first
Load-Config

# Handle flags
if ($Help) { Show-Help }
if ($Version) { Show-Version }
if ($Setup) { Interactive-Setup }
if ($Config) { Show-Config }
if ($EditPrompt) { Edit-Prompt }
if ($Emoji) { $script:EMOJI_MODE = $true }
if ($Conv) { $script:EMOJI_MODE = $false }  # Force conventional
if ($Yes) { $script:AUTO_YES = $true }

# -------------------------------------------------
# UNDO LAST COMMIT
# -------------------------------------------------
if ($Undo) {
    $lastMsg = git log -1 --pretty=%B
    git reset --soft HEAD~1
    Write-Host "↩️  Undone last commit:" -ForegroundColor Yellow
    Write-Host $lastMsg
    exit 0
}

# -------------------------------------------------
# DEPENDENCIES
# -------------------------------------------------
if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Missing dependency: git" -ForegroundColor Red
    exit 1
}

# Check API key based on provider
if ($script:PROVIDER -eq "openai") {
    if ([string]::IsNullOrEmpty($env:OPENAI_API_KEY)) {
        Write-Host "❌ OPENAI_API_KEY is not set." -ForegroundColor Red
        Write-Host "Run 'commit-ai -Setup' to configure." -ForegroundColor Gray
        exit 1
    }
    $apiKey = $env:OPENAI_API_KEY
} else {
    if ([string]::IsNullOrEmpty($env:GEMINI_API_KEY)) {
        Write-Host "❌ GEMINI_API_KEY is not set." -ForegroundColor Red
        Write-Host "Run 'commit-ai -Setup' to configure." -ForegroundColor Gray
        exit 1
    }
    $apiKey = $env:GEMINI_API_KEY
}

# -------------------------------------------------
# STAGING CHECK
# -------------------------------------------------
git diff --cached --quiet 2>$null
if ($LASTEXITCODE -eq 0) {
    Write-Host "❌ No staged changes found." -ForegroundColor Red
    Write-Host "ℹ️ Run: git add <files>" -ForegroundColor Cyan
    exit 1
}

# Get diff
$rawDiff = git diff --cached --unified=3
$filteredDiff = $rawDiff | Select-String -Pattern '^\+|^-|^@@|^diff --git' | ForEach-Object { $_.Line }
$diffText = ($filteredDiff -join "`n")
$DIFF = $diffText.Substring(0, [Math]::Min($MAX_CHARS, $diffText.Length))

$FILES = (git diff --cached --name-only) -join "`n"
$HISTORY = (git log --oneline -n 20) -join "`n"

# -------------------------------------------------
# PROMPT SELECTION
# -------------------------------------------------
if (Test-Path $CUSTOM_PROMPT_FILE) {
    $PROMPT = (Get-Content $CUSTOM_PROMPT_FILE | Where-Object { $_ -notmatch '^#' }) -join "`n"
    $PROMPT = $PROMPT -replace '\{HISTORY\}', $HISTORY
    $PROMPT = $PROMPT -replace '\{FILES\}', $FILES
    $PROMPT = $PROMPT -replace '\{DIFF\}', $DIFF
} elseif ($script:EMOJI_MODE) {
    $USER_HINT = ""
    if (-not [string]::IsNullOrWhiteSpace($Message)) {
        $USER_HINT = @"

User context/hint for this commit:
$Message

Use this hint to better understand the intent and generate a more accurate commit message.

"@
    }
    $PROMPT = @"
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
"@
} else {
    $USER_HINT = ""
    if (-not [string]::IsNullOrWhiteSpace($Message)) {
        $USER_HINT = @"

User context/hint for this commit:
$Message

Use this hint to better understand the intent and generate a more accurate commit message.

"@
    }
    $PROMPT = @"
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
"@
}

# -------------------------------------------------
# API REQUEST
# -------------------------------------------------
if ($script:PROVIDER -eq "openai") {
    $body = @{
        model = $script:DEFAULT_MODEL
        messages = @(
            @{ role = "user"; content = $PROMPT }
        )
        max_tokens = 100
    } | ConvertTo-Json -Depth 5 -Compress

    try {
        $response = Invoke-RestMethod -Uri "https://api.openai.com/v1/chat/completions" -Method Post -Body $body -ContentType 'application/json' -Headers @{
            'Authorization' = "Bearer $apiKey"
        }
        $COMMIT_MSG = $response.choices[0].message.content
    } catch {
        Write-Host "❌ Failed to call OpenAI API: $_" -ForegroundColor Red
        exit 1
    }
} else {
    $body = @{
        contents = @(
            @{ parts = @( @{ text = $PROMPT } ) }
        )
    } | ConvertTo-Json -Depth 5 -Compress

    try {
        $response = Invoke-RestMethod -Uri "https://generativelanguage.googleapis.com/v1beta/models/$($script:DEFAULT_MODEL):generateContent" -Method Post -Body $body -ContentType 'application/json' -Headers @{
            'x-goog-api-key' = $apiKey
        }
        $COMMIT_MSG = $response.candidates[0].content.parts[0].text
    } catch {
        Write-Host "❌ Failed to call Gemini API: $_" -ForegroundColor Red
        exit 1
    }
}

if ([string]::IsNullOrWhiteSpace($COMMIT_MSG)) {
    Write-Host "❌ Failed to generate commit message." -ForegroundColor Red
    exit 1
}

# -------------------------------------------------
# NORMALIZATION
# -------------------------------------------------
$COMMIT_MSG = $COMMIT_MSG.Trim() -replace "`n", " " -replace "\s+", " "

if ($script:EMOJI_MODE) {
    if ($COMMIT_MSG -match '^(\p{So}|\p{Cs}{2})(\S)') {
        $COMMIT_MSG = $COMMIT_MSG -replace '^(\p{So}|\p{Cs}{2})(\S)', '$1 $2'
    }
} else {
    if ($COMMIT_MSG -match '^([a-z]+): ([A-Z])') {
        $lower = $Matches[2].ToLower()
        $COMMIT_MSG = $COMMIT_MSG -replace '^([a-z]+): [A-Z]', "`$1: $lower"
    }
}

# -------------------------------------------------
# OUTPUT
# -------------------------------------------------
if ($Preview) {
    Write-Host
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host "  ℹ️  Preview Mode" -ForegroundColor White
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host
    Write-Host "  $COMMIT_MSG" -ForegroundColor Green
    Write-Host
    exit 0
}

if (-not $script:AUTO_YES) {
    Write-Host
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host "  📝 Generated Commit Message" -ForegroundColor White
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host
    Write-Host "  $COMMIT_MSG" -ForegroundColor Green
    Write-Host
    Write-Host "Press Enter to confirm, or type a new message:" -ForegroundColor Yellow
    $edited = Read-Host
    if (-not [string]::IsNullOrWhiteSpace($edited)) {
        $COMMIT_MSG = $edited
    }
    Write-Host
}

git commit -m $COMMIT_MSG
Write-Host
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host "  ✅ Commit created successfully!" -ForegroundColor Green
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan

# Ask to push if configured
if ($script:ASK_PUSH) {
    Write-Host
    $pushChoice = Read-Host "🚀 Push to remote? (y/n)"
    switch ($pushChoice.ToLower()) {
        "y" { 
            Write-Host
            git push
            Write-Host
            Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
            Write-Host "  🚀 Pushed to remote!" -ForegroundColor Green
            Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
        }
        "yes" { 
            Write-Host
            git push
            Write-Host
            Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
            Write-Host "  🚀 Pushed to remote!" -ForegroundColor Green
            Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
        }
    }
}
