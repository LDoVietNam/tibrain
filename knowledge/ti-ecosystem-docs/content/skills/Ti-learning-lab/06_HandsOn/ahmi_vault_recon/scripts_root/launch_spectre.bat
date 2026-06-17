@echo off
taskkill /F /IM comet.exe >nul 2>&1
echo Launching Comet with Spectre Assistant...
start "" "C:\Users\MIN\AppData\Local\Perplexity\Comet\Application\comet.exe" --load-extension="Z:\SnJ\browser\browser-assistant-extension" --user-data-dir="C:\Users\MIN\AppData\Local\Perplexity\Comet\User Data"
exit
