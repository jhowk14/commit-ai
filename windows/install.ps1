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
    [string]$Version = $(if ($env:COMMIT_AI_VERSION) { $env:COMMIT_AI_VERSION } else { "2.0.7" }),
    [string]$InstallDir = $(if ($env:COMMIT_AI_INSTALL_DIR) { $env:COMMIT_AI_INSTALL_DIR } else { Join-Path $HOME "bin" })
)

$ErrorActionPreference = "Stop"
if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
    throw "Git for Windows é necessário. Instale-o em https://git-scm.com/download/win"
}
$architecture = if ($env:PROCESSOR_ARCHITECTURE -match "ARM64") { "arm64" } else { "amd64" }
$asset = "commit-ai-windows-$architecture.exe"
$url = "https://github.com/jhowk14/commit-ai/releases/download/v$Version/$asset"
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
$destination = Join-Path $InstallDir "commit-ai.exe"
Write-Host "Baixando commit-ai v$Version..." -ForegroundColor Cyan
Invoke-WebRequest -Uri $url -OutFile $destination
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$InstallDir", "User")
    $env:Path = "$env:Path;$InstallDir"
}
Write-Host "✅ commit-ai instalado em $destination" -ForegroundColor Green
Write-Host "Abra um novo terminal e execute: commit-ai --setup" -ForegroundColor Gray
