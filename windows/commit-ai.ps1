#Requires -Version 5.1
[CmdletBinding()]
param(
    [Alias('e')][switch]$Emoji,
    [Alias('p')][switch]$Preview,
    [Alias('y')][switch]$Yes,
    [Alias('u')][switch]$Undo,
    [Alias('s')][switch]$Setup,
    [Alias('c')][switch]$Config,
    [Alias('h')][switch]$Help,
    [Alias('v')][switch]$Version
)

$ErrorActionPreference = 'Stop'

# ================= CONFIG =================
$SCRIPT_VERSION = "1.2.0"
$MAX_CHARS = 14000
$CONFIG_FILE = Join-Path $HOME ".commit-ai.conf"

# Defaults
$DEFAULT_FORMAT = "conventional"
$DEFAULT_AUTO_CONFIRM = $false
$DEFAULT_MODEL = "gemini-2.0-flash"
$script:EMOJI_MODE = $false
$script:AUTO_YES = $false
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
                    'model' { 
                        $script:DEFAULT_MODEL = $value 
                    }
                    'api_key' { 
                        if ([string]::IsNullOrEmpty($env:GEMINI_API_KEY)) {
                            $env:GEMINI_API_KEY = $value
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
        [string]$Model,
        [string]$ApiKey
    )

    $apiKeyLine = if ($ApiKey) { "api_key=$ApiKey" } else { "# api_key=your_key_here" }
    
    $configContent = @"
# commit-ai configuration
# Location: ~/.commit-ai.conf

# Default commit format: conventional | gitmoji
format=$Format

# Auto-confirm commits without prompt: true | false
auto_confirm=$AutoConfirm

# Gemini model to use
model=$Model

# API Key (optional - can also use GEMINI_API_KEY environment variable)
$apiKeyLine
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

    # Load existing config for defaults
    $currentFormat = "conventional"
    $currentAuto = "false"
    $currentModel = "gemini-2.0-flash"
    $currentKey = ""

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
                    'model' { $currentModel = $value }
                    'api_key' { $currentKey = $value }
                }
            }
        }
    }

    # Format selection
    Write-Host "📝 Commit format:" -ForegroundColor Yellow
    Write-Host "   1) conventional (feat:, fix:, etc.)"
    Write-Host "   2) gitmoji (✨, 🐛, etc.)"
    Write-Host
    $formatChoice = Read-Host "Choose format [current: $currentFormat] (1/2)"
    switch ($formatChoice) {
        "1" { $currentFormat = "conventional" }
        "2" { $currentFormat = "gitmoji" }
    }

    # Auto-confirm
    Write-Host
    Write-Host "⚡ Auto-confirm commits (skip confirmation prompt)?" -ForegroundColor Yellow
    $autoChoice = Read-Host "Enable auto-confirm? [current: $currentAuto] (y/n)"
    switch ($autoChoice.ToLower()) {
        "y" { $currentAuto = "true" }
        "yes" { $currentAuto = "true" }
        "n" { $currentAuto = "false" }
        "no" { $currentAuto = "false" }
    }

    # Model selection
    Write-Host
    Write-Host "🧠 Gemini model:" -ForegroundColor Yellow
    Write-Host "   1) gemini-2.0-flash (fast, recommended)"
    Write-Host "   2) gemini-2.0-flash-lite (faster, lighter)"
    Write-Host "   3) gemini-2.5-pro-preview (advanced)"
    Write-Host "   4) Custom"
    Write-Host
    $modelChoice = Read-Host "Choose model [current: $currentModel] (1/2/3/4)"
    switch ($modelChoice) {
        "1" { $currentModel = "gemini-2.0-flash" }
        "2" { $currentModel = "gemini-2.0-flash-lite" }
        "3" { $currentModel = "gemini-2.5-pro-preview" }
        "4" { 
            $currentModel = Read-Host "Enter custom model name"
        }
    }

    # API Key
    Write-Host
    Write-Host "🔐 API Key configuration:" -ForegroundColor Yellow
    if ($currentKey) {
        Write-Host "   Current: ****$($currentKey.Substring([Math]::Max(0, $currentKey.Length - 4)))"
    } elseif ($env:GEMINI_API_KEY) {
        Write-Host "   Using environment variable GEMINI_API_KEY"
    } else {
        Write-Host "   Not configured"
    }
    Write-Host
    $newKey = Read-Host "Enter new API key (leave empty to keep current)"
    if ($newKey) { $currentKey = $newKey }

    # Save
    Write-Host
    Save-Config -Format $currentFormat -AutoConfirm $currentAuto -Model $currentModel -ApiKey $currentKey

    Write-Host
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host "  Configuration complete! Run 'commit-ai' to start." -ForegroundColor White
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
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
                if ($key -eq 'api_key' -and $value) {
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
    $envStatus = if ($env:GEMINI_API_KEY) { "set" } else { "not set" }
    Write-Host "🔑 GEMINI_API_KEY env: $envStatus" -ForegroundColor Gray
    Write-Host
    exit 0
}

# -------------------------------------------------
# HELP
# -------------------------------------------------
function Show-Help {
    @"
commit-ai v$SCRIPT_VERSION - AI-powered Git commit messages using Gemini

USAGE:
  .\commit-ai.ps1 [OPTIONS]

OPTIONS:
  -Emoji, -e      Use Gitmoji commit format (emoji prefix)
  -Preview, -p    Preview commit message only (no commit)
  -Yes, -y        Skip confirmation prompt (auto-commit)
  -Undo, -u       Undo last commit (soft reset, keeps changes staged)
  -Setup, -s      Interactive configuration setup
  -Config, -c     Show current configuration
  -Help, -h       Show this help message
  -Version, -v    Show version number

EXAMPLES:
  .\commit-ai.ps1              # Conventional Commits format
  .\commit-ai.ps1 -Emoji       # Gitmoji format
  .\commit-ai.ps1 -e -p        # Preview Gitmoji message
  .\commit-ai.ps1 -y           # Auto-commit without confirmation
  .\commit-ai.ps1 -u           # Undo last commit
  .\commit-ai.ps1 -Setup       # Configure preferences

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
if ($Emoji) { $script:EMOJI_MODE = $true }
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
$deps = @('git', 'curl')
foreach ($cmd in $deps) {
    if (-not (Get-Command $cmd -ErrorAction SilentlyContinue)) {
        Write-Host "❌ Missing dependency: $cmd" -ForegroundColor Red
        exit 1
    }
}

$apiKey = $env:GEMINI_API_KEY
if ([string]::IsNullOrEmpty($apiKey)) {
    Write-Host "❌ GEMINI_API_KEY is not set." -ForegroundColor Red
    Write-Host "Run 'commit-ai -Setup' to configure, or:" -ForegroundColor Gray
    Write-Host '  $env:GEMINI_API_KEY = "your_key"' -ForegroundColor Gray
    exit 1
}

# -------------------------------------------------
# STAGING CHECK
# -------------------------------------------------
$stagedDiff = git diff --cached --quiet 2>$null
if ($LASTEXITCODE -eq 0) {
    Write-Host "❌ No staged changes found." -ForegroundColor Red
    Write-Host "ℹ️ Run: git add <files>" -ForegroundColor Cyan
    exit 1
}

# Get diff (filtered and limited)
$rawDiff = git diff --cached --unified=3
$filteredDiff = $rawDiff | Select-String -Pattern '^\+|^-|^@@|^diff --git' | ForEach-Object { $_.Line }
$diffText = ($filteredDiff -join "`n")
$DIFF = $diffText.Substring(0, [Math]::Min($MAX_CHARS, $diffText.Length))

$FILES = git diff --cached --name-only
$HISTORY = git log --oneline -n 20

# -------------------------------------------------
# PROMPT SELECTION
# -------------------------------------------------
if ($script:EMOJI_MODE) {
    $PROMPT = @"
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
"@
} else {
    $PROMPT = @"
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
"@
}

# -------------------------------------------------
# GEMINI REQUEST
# -------------------------------------------------
$body = @{
    contents = @(
        @{
            parts = @(
                @{ text = $PROMPT }
            )
        }
    )
} | ConvertTo-Json -Depth 5 -Compress

$uri = "https://generativelanguage.googleapis.com/v1beta/models/$($script:DEFAULT_MODEL):generateContent"

try {
    $response = Invoke-RestMethod -Uri $uri -Method Post -Body $body -ContentType 'application/json' -Headers @{
        'x-goog-api-key' = $apiKey
    }
    $COMMIT_MSG = $response.candidates[0].content.parts[0].text
} catch {
    Write-Host "❌ Failed to call Gemini API: $_" -ForegroundColor Red
    exit 1
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
    # Ensure space after emoji and capitalize first letter
    if ($COMMIT_MSG -match '^(\p{So}|\p{Cs}{2})(\S)') {
        $COMMIT_MSG = $COMMIT_MSG -replace '^(\p{So}|\p{Cs}{2})(\S)', '$1 $2'
    }
} else {
    # Ensure lowercase after colon for conventional commits
    if ($COMMIT_MSG -match '^([a-z]+): ([A-Z])') {
        $lower = $Matches[2].ToLower()
        $COMMIT_MSG = $COMMIT_MSG -replace '^([a-z]+): [A-Z]', "`$1: $lower"
    }
}

# -------------------------------------------------
# OUTPUT
# -------------------------------------------------
if ($Preview) {
    Write-Host "`nℹ️ Preview:" -ForegroundColor Cyan
    Write-Host $COMMIT_MSG -ForegroundColor Green
    exit 0
}

if (-not $script:AUTO_YES) {
    Write-Host "`n📝 Generated commit message:" -ForegroundColor Cyan
    Write-Host $COMMIT_MSG -ForegroundColor Green
    Write-Host "`nPress Enter to confirm, or type a new message:" -ForegroundColor Yellow
    $edited = Read-Host
    if (-not [string]::IsNullOrWhiteSpace($edited)) {
        $COMMIT_MSG = $edited
    }
}

git commit -m $COMMIT_MSG
Write-Host "✅ Commit created!" -ForegroundColor Green
