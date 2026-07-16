@echo off
title TiBrain Tunnel Stop

echo ========================================
echo TiBrain Tunnel Stop
echo ========================================

echo Stopping TiBrain server...
taskkill /IM tibrain.exe /F 2>nul

echo Stopping Cloudflare tunnel...
taskkill /IM cloudflared.exe /F 2>nul

echo All processes stopped.
echo You can now run start-tunnel.bat or restart-tunnel.bat

pause