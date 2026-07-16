@echo off
title TiBrain Tunnel Restart

echo ========================================
echo TiBrain Tunnel Restart
echo ========================================

REM Kill existing processes
echo Stopping existing processes...
taskkill /IM cloudflared.exe /F 2>nul
taskkill /IM tibrain.exe /F 2>nul

REM Wait for cleanup
timeout /t 2 /nobreak >nul

REM Restart tunnel
cd /d %~dp0
echo Starting tunnel...
cloudflared tunnel run tibrain

pause