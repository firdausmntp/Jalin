@echo off
setlocal
cd /d "%~dp0"
title Jalin Uninstaller
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0installer.ps1" -Uninstall
if %errorlevel% neq 0 (
    echo.
    echo Uninstallation failed with code %errorlevel%.
    pause
    exit /b %errorlevel%
)
echo.
pause
