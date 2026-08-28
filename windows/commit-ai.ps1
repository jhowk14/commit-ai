<#
.SYNOPSIS
  commit-ai – AI-powered commit message generator.

.DESCRIPTION
  commit-ai generates conventional commit messages using AI providers (Gemini, OpenAI, or any OpenAI-compatible provider)
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
    [Alias("e")][switch]$Emoji,
    [Alias("c")][switch]$Conv,
    [Alias("p")][switch]$Preview,
    [Alias("y")][switch]$Yes,
    [Alias("u")][switch]$Undo,
    [Alias("s")][switch]$Setup,
    [Alias("m")][string]$Message,
    [Alias("b")][string]$Branch,
    [Alias("B")][string]$BaseUrl,
    [switch]$Config,
    [switch]$EditPrompt,
    [Alias("h")][switch]$Help,
    [Alias("v")][switch]$Version
)

$ErrorActionPreference = "Stop"

# ================= CONFIG =================
$SCRIPT_VERSION = "1.6.0"
$MAX_CHARS = 14000
$CONFIG_FILE = Join-Path $HOME ".commit-ai.conf"
$CUSTOM_PROMPT_FILE = Join-Path $HOME ".commit-ai-prompt.txt"

# Defaults
$script:PROVIDER = "gemini"
$script:DEFAULT_MODEL = "gemini-3-flash-preview"
$script:OPENAI_BASE_URL = "https://api.openai.com/v1"
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
            if ($line -match "^#" -or [string]::IsNullOrWhiteSpace($line)) { return }
            
            $parts = $line -split "=", 2
            if ($parts.Count -eq 2) {
                $key = $parts[0].Trim()
                $value = $parts[1].Trim()
                
                switch ($key) {
                    "format" { 
                        if ($value -eq "gitmoji") { $script:EMOJI_MODE = $true }
                    }
                    "auto_confirm" { 
                        if ($value -eq "true") { $script:AUTO_YES = $true }
                    }
                    "ask_push" { 
                        if ($value -eq "true") { $script:ASK_PUSH = $true }
                    }
                    "provider" {
                        $script:PROVIDER = $value
                    }
                    "model" { 
                        $script:DEFAULT_MODEL = $value 
                    }
                    "openai_base_url" {
                        $script:OPENAI_BASE_URL = $value
                    }
                    "base_url" {
                        $script:OPENAI_BASE_URL = $value
                    }
                    "gemini_api_key" { 
                        if ([string]::IsNullOrEmpty($env:GEMINI_API_KEY)) {
                            $env:GEMINI_API_KEY = $value
                        }
                    }
                    "openai_api_key" { 
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
        [string]$OpenAIKey,
        [string]$OpenAIBaseUrl
    )

    $geminiLine = if ($GeminiKey) { "gemini_api_key=$GeminiKey" } else { "# gemini_api_key=your_key_here" }
    $openaiLine = if ($OpenAIKey) { "openai_api_key=$OpenAIKey" } else { "# openai_api_key=your_key_here" }
    $baseUrlVal = if ($OpenAIBaseUrl) { $OpenAIBaseUrl } else { "https://api.openai.com/v1" }
    
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

# Base URL for OpenAI-compatible providers
openai_base_url=$baseUrlVal

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
    $currentOpenAIBaseUrl = "https://api.openai.com/v1"
    $currentGeminiKey = ""
    $currentOpenAIKey = ""

    if (Test-Path $CONFIG_FILE) {
        Get-Content $CONFIG_FILE | ForEach-Object {
            $line = $_.Trim()
            if ($line -match "^#" -or [string]::IsNullOrWhiteSpace($line)) { return }
            $parts = $line -split "=", 2
            if ($parts.Count -eq 2) {
                $key = $parts[0].Trim()
                $value = $parts[1].Trim()
                switch ($key) {
                    "format" { $currentFormat = $value }
                    "auto_confirm" { $currentAuto = $value }
                    "ask_push" { $currentPush = $value }
                    "provider" { $currentProvider = $value }
                    "model" { $currentModel = $value }
                    "openai_base_url" { $currentOpenAIBaseUrl = $value }
                    "base_url" { $currentOpenAIBaseUrl = $value }
                    "gemini_api_key" { $currentGeminiKey = $value }
                    "openai_api_key" { $currentOpenAIKey = $value }
                }
            }
        }
    }

    # Format selection with validation
    Write-Host "📝 Commit format:" -ForegroundColor Yellow
    Write-Host "   1) conventional (feat:, fix:, etc.)"
    Write-Host "   2) gitmoji (emoji prefix)"
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
    Write-Host "   2) OpenAI / OpenAI-Compatible (OpenAI, OpenRouter, Groq, DeepSeek, Ollama, Cerebras, LM Studio, etc.)"
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

    # Base URL selection if OpenAI
    $chosenPreset = "openai"
    if ($currentProvider -eq "openai") {
        Write-Host
        Write-Host "🌐 OpenAI Endpoint / Base URL:" -ForegroundColor Yellow
        Write-Host "   1) OpenAI Official (https://api.openai.com/v1)"
        Write-Host "   2) OpenRouter (https://openrouter.ai/api/v1)"
        Write-Host "   3) Groq (https://api.groq.com/openai/v1)"
        Write-Host "   4) DeepSeek (https://api.deepseek.com/v1)"
        Write-Host "   5) Ollama Local (http://localhost:11434/v1)"
        Write-Host "   6) LM Studio / LocalAI (http://localhost:1234/v1)"
        Write-Host "   7) Cerebras (https://api.cerebras.ai/v1)"
        Write-Host "   8) Custom Base URL"
        Write-Host
        do {
            $urlChoice = Read-Host "Choose endpoint [current: $currentOpenAIBaseUrl] (1-8)"
            $valid = $true
            switch ($urlChoice) {
                "1" { $currentOpenAIBaseUrl = "https://api.openai.com/v1"; $chosenPreset = "openai" }
                "2" { $currentOpenAIBaseUrl = "https://openrouter.ai/api/v1"; $chosenPreset = "openrouter" }
                "3" { $currentOpenAIBaseUrl = "https://api.groq.com/openai/v1"; $chosenPreset = "groq" }
                "4" { $currentOpenAIBaseUrl = "https://api.deepseek.com/v1"; $chosenPreset = "deepseek" }
                "5" { $currentOpenAIBaseUrl = "http://localhost:11434/v1"; $chosenPreset = "ollama" }
                "6" { $currentOpenAIBaseUrl = "http://localhost:1234/v1"; $chosenPreset = "lmstudio" }
                "7" { $currentOpenAIBaseUrl = "https://api.cerebras.ai/v1"; $chosenPreset = "cerebras" }
                "8" { 
                    $customUrl = Read-Host "Enter custom Base URL"
                    if ($customUrl) { $currentOpenAIBaseUrl = $customUrl }
                    $chosenPreset = "custom"
                }
                "" { }
                default {
                    Write-Host "⚠️  Invalid option. Please enter 1-8." -ForegroundColor Yellow
                    $valid = $false
                }
            }
        } while (-not $valid)
    }

    # Reset model to default if provider changed
    if ($oldProvider -ne $currentProvider) {
        if ($currentProvider -eq "gemini") {
            $currentModel = "gemini-3-flash-preview"
        } else {
            switch ($chosenPreset) {
                "openrouter" { $currentModel = "meta-llama/llama-3.3-70b-instruct" }
                "groq" { $currentModel = "llama-3.3-70b-versatile" }
                "deepseek" { $currentModel = "deepseek-chat" }
                "ollama" { $currentModel = "llama3.2" }
                "cerebras" { $currentModel = "llama-3.3-70b" }
                default { $currentModel = "gpt-4o-mini" }
            }
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
        Write-Host
        do {
            $modelChoice = Read-Host "Choose model [current: $currentModel] (1-5)"
            $valid = $true
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
        } while (-not $valid)
    } else {
        Write-Host "   1) gpt-4o-mini (fast, recommended)"
        Write-Host "   2) gpt-4o (advanced)"
        Write-Host "   3) Custom model name"
        Write-Host
        do {
            $modelChoice = Read-Host "Choose model [current: $currentModel] (1-3)"
            $valid = $true
            switch ($modelChoice) {
                "1" { $currentModel = "gpt-4o-mini" }
                "2" { $currentModel = "gpt-4o" }
                "3" { $currentModel = Read-Host "Enter custom model name" }
                "" { }
                default {
                    Write-Host "⚠️  Invalid option. Please enter 1-3." -ForegroundColor Yellow
                    $valid = $false
                }
            }
        } while (-not $valid)
    }

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
        Write-Host "  API Key for OpenAI / Compatible Provider ($currentOpenAIBaseUrl):"
        if ($currentOpenAIKey) {
            Write-Host "    Current: ****$($currentOpenAIKey.Substring([Math]::Max(0, $currentOpenAIKey.Length - 4)))"
        } elseif ($env:OPENAI_API_KEY) {
            Write-Host "    Using environment variable"
        } else {
            Write-Host "    Not configured"
        }
        $newOpenAIKey = Read-Host "  Enter API key (optional for local servers, leave empty to keep/skip)"
        if ($newOpenAIKey) { $currentOpenAIKey = $newOpenAIKey }
    }

    # Save
    Write-Host
    Save-Config -Format $currentFormat -AutoConfirm $currentAuto -AskPush $currentPush -Provider $currentProvider -Model $currentModel -GeminiKey $currentGeminiKey -OpenAIKey $currentOpenAIKey -OpenAIBaseUrl $currentOpenAIBaseUrl

    Write-Host
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host "  Configuration complete! Run commit-ai to start." -ForegroundColor White
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    exit 0
}

# -------------------------------------------------
# SHOW CONFIG
# -------------------------------------------------
function Show-Config {
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host "  🤖 commit-ai v$SCRIPT_VERSION - Current Config" -ForegroundColor White
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host
    
    if (Test-Path $CONFIG_FILE) {
        Write-Host "📂 Config file: $CONFIG_FILE" -ForegroundColor Gray
        Write-Host
        Get-Content $CONFIG_FILE | ForEach-Object {
            $line = $_.Trim()
            if ($line -match "^#" -or [string]::IsNullOrWhiteSpace($line)) { return }
            $parts = $line -split "=", 2
            if ($parts.Count -eq 2) {
                $key = $parts[0].Trim()
                $value = $parts[1].Trim()
                if ($key -like "*api_key*" -and $value) {
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
    Write-Host "   OPENAI_BASE_URL: $($env:OPENAI_BASE_URL)"
    Write-Host
    exit 0
}

# Load config
Load-Config

if ($BaseUrl) {
    $script:OPENAI_BASE_URL = $BaseUrl
}

if ($Config) { Show-Config }
if ($Setup) { Interactive-Setup }

# API Endpoint Calculation
if ($script:PROVIDER -eq "openai") {
    $baseUrl = if ($script:OPENAI_BASE_URL) { $script:OPENAI_BASE_URL.TrimEnd("/") } else { "https://api.openai.com/v1" }
    $endpointUrl = if ($baseUrl.EndsWith("/chat/completions")) { $baseUrl } else { "$baseUrl/chat/completions" }
    $apiKey = if ($env:OPENAI_API_KEY) { $env:OPENAI_API_KEY } else { "none" }
    
    $headers = @{ "Content-Type" = "application/json" }
    if ($apiKey -ne "none" -and -not [string]::IsNullOrWhiteSpace($apiKey)) {
        $headers["Authorization"] = "Bearer $apiKey"
    }
    if ($baseUrl -like "*openrouter.ai*") {
        $headers["HTTP-Referer"] = "https://github.com/jhowk14/commit-ai"
        $headers["X-Title"] = "commit-ai"
    }
}
