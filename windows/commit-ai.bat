@echo off
REM commit-ai
REM Copyright (c) 2026 Jonathan Henrique Perozi Lourenço (jhowk14)
REM Licensed under the MIT License

setlocal enabledelayedexpansion

REM ===============================================
REM commit-ai Batch Wrapper for Windows
REM Calls the PowerShell script with all arguments
REM ===============================================

set "SCRIPT_DIR=%~dp0"
powershell -ExecutionPolicy Bypass -File "%SCRIPT_DIR%commit-ai.ps1" %*
exit /b %errorlevel%

