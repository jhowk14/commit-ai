@echo off
setlocal enabledelayedexpansion

:: ===============================================
:: commit-ai Batch Wrapper for Windows
:: Calls the PowerShell script with all arguments
:: ===============================================

set "SCRIPT_DIR=%~dp0"
powershell -ExecutionPolicy Bypass -File "%SCRIPT_DIR%commit-ai.ps1" %*
exit /b %errorlevel%
