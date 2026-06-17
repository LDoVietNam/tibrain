@echo off
REM Ti Brain Indexing Monitor Script (Windows)
REM Purpose: Real-time monitoring of indexing operations

echo === 🚀 Ti Brain Indexing Monitor ===
echo Started at: %date% %time%
echo.

REM Check if indexing is running
echo 🔍 Checking indexing status...

REM Check for active kb-builder processes
tasklist /FI "IMAGENAME eq kb-builder.exe" 2>NUL | find /I "kb-builder.exe" >NUL
if %ERRORLEVEL% EQU 0 (
    echo ✅ Indexing is running
    echo 📊 Process details:
    tasklist /FI "IMAGENAME eq kb-builder.exe" /FO TABLE
) else (
    echo ❌ No indexing process found
)

REM Monitor indexing progress
echo.
echo 📈 Indexing Progress:

REM Check knowledge base size
if exist "knowledge\ti-ecosystem-docs" (
    echo 📄 Checking files in knowledge base...
    dir /s /b "knowledge\ti-ecosystem-docs\*" 2>NUL | find /C /V "" > temp_count.txt
    set /p FILE_COUNT=<temp_count.txt
    echo 📄 Files indexed: %FILE_COUNT%
    del temp_count.txt
    
    REM Calculate directory size
    for /f "tokens=3" %%A in ('dir "knowledge\ti-ecosystem-docs" /s ^| find "bytes"') do (
        set SIZE=%%A
    )
    echo 💾 Knowledge base size: %SIZE% bytes
)

REM Check database size
if exist "memory\hub.db" (
    for %%F in ("memory\hub.db") do (
        set DB_SIZE=%%~zF
    )
    echo 🗄️ Database size: %DB_SIZE%
)

REM Check Ti Brain logs
if exist "logs\tibrain.log" (
    echo 📊 Recent log entries:
    powershell "Get-Content 'logs\tibrain.log' | Select-Object -Last 10 | Where-Object { $_ -match 'Processed.*files' }"
)

echo.
echo 🔄 Monitoring complete!
echo.
pause