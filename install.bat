@echo off
setlocal
cd /d "%~dp0"
title Jalin Installer
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0installer.ps1" -Install
if %errorlevel% neq 0 (
    echo.
    echo Installation failed with code %errorlevel%.
    pause
    exit /b %errorlevel%
)
echo.
pause
