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

# ===============================================
# commit-ai Installer for Windows
# ===============================================

$ErrorActionPreference = 'Stop'
$VERSION = "1.2.0"
$SCRIPT_DIR = Split-Path -Parent $MyInvocation.MyCommand.Path
$INSTALL_DIR = Join-Path $HOME "bin"
$CONFIG_FILE = Join-Path $HOME ".commit-ai.conf"

Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host "  🤖 commit-ai v$VERSION - Windows Installer" -ForegroundColor White
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host

# Check dependencies
Write-Host "📦 Checking dependencies..." -ForegroundColor Yellow
$missing = @()
foreach ($cmd in @('git', 'curl')) {
    if (-not (Get-Command $cmd -ErrorAction SilentlyContinue)) {
        $missing += $cmd
    }
}

if ($missing.Count -gt 0) {
    Write-Host "❌ Missing dependencies: $($missing -join ', ')" -ForegroundColor Red
    Write-Host
    Write-Host "Install Git for Windows from: https://git-scm.com/download/win" -ForegroundColor Gray
    exit 1
}
Write-Host "✅ All dependencies found" -ForegroundColor Green
Write-Host

# Create install directory
if (-not (Test-Path $INSTALL_DIR)) {
    Write-Host "📁 Creating directory: $INSTALL_DIR" -ForegroundColor Yellow
    New-Item -ItemType Directory -Force -Path $INSTALL_DIR | Out-Null
}

# Copy scripts
Write-Host "📥 Installing commit-ai to $INSTALL_DIR..." -ForegroundColor Yellow
Copy-Item -Path (Join-Path $SCRIPT_DIR "commit-ai.ps1") -Destination $INSTALL_DIR -Force
Copy-Item -Path (Join-Path $SCRIPT_DIR "commit-ai.bat") -Destination $INSTALL_DIR -Force
Write-Host "✅ Scripts installed" -ForegroundColor Green
Write-Host

# Add to PATH
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$INSTALL_DIR*") {
    Write-Host "🔧 Adding $INSTALL_DIR to PATH..." -ForegroundColor Yellow
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$INSTALL_DIR", "User")
    $env:Path = "$env:Path;$INSTALL_DIR"
    Write-Host "✅ Added to PATH (restart terminal to apply)" -ForegroundColor Green
} else {
    Write-Host "✅ Already in PATH" -ForegroundColor Green
}
Write-Host

# Create PowerShell alias
Write-Host "🔧 Setting up PowerShell alias..." -ForegroundColor Yellow

# Ensure profile exists
if (-not (Test-Path $PROFILE)) {
    New-Item -ItemType File -Path $PROFILE -Force | Out-Null
}

$aliasLine = "Set-Alias -Name commit-ai -Value `"$INSTALL_DIR\commit-ai.ps1`""
$profileContent = Get-Content $PROFILE -ErrorAction SilentlyContinue

if ($profileContent -notcontains $aliasLine) {
    Add-Content -Path $PROFILE -Value "`n# commit-ai alias"
    Add-Content -Path $PROFILE -Value $aliasLine
    Write-Host "✅ Alias added to PowerShell profile" -ForegroundColor Green
} else {
    Write-Host "✅ Alias already exists" -ForegroundColor Green
}
Write-Host

# Run interactive setup
Write-Host "🔧 Running initial configuration..." -ForegroundColor Yellow
Write-Host

& "$INSTALL_DIR\commit-ai.ps1" -Setup

Write-Host
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host "  ✅ Installation complete!" -ForegroundColor White
Write-Host "  Run 'commit-ai' to generate commit messages" -ForegroundColor Gray
Write-Host "  Run 'commit-ai -Setup' to reconfigure" -ForegroundColor Gray
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
