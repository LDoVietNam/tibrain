@echo off
REM Ti Brain Quality Check Script
REM Purpose: Check quality of indexed content

echo === 🔍 Ti Brain Quality Check ===
echo Started at: %date% %time%
echo.

REM Check database integrity
echo 🗄️ Database Integrity Check:
if exist "memory\hub.db" (
    echo ✅ Database exists: memory\hub.db
    
    REM Check database size
    for %%F in ("memory\hub.db") do (
        set DB_SIZE=%%~zF
    )
    echo 💾 Database size: %DB_SIZE%
    
    REM Check WAL file size
    if exist "memory\hub.db-wal" (
        for %%F in ("memory\hub.db-wal") do (
            set WAL_SIZE=%%~zF
        )
        echo 📝 WAL file size: %WAL_SIZE%
    )
) else (
    echo ❌ Database not found
)

echo.

REM Check knowledge base structure
echo 📚 Knowledge Base Structure:
if exist "knowledge\ti-ecosystem-docs" (
    echo ✅ Knowledge base exists
    
    REM Count files by category
    echo 📊 File Distribution:
    
    if exist "knowledge\ti-ecosystem-docs\content" (
        for /f %%A in ('dir /s /b "knowledge\ti-ecosystem-docs\content\*" 2^>NUL ^| find /C /V ""') do (
            echo 📄 Content files: %%A
        )
    )
    
    if exist "knowledge\ti-ecosystem-docs\apps-docs" (
        for /f %%A in ('dir /s /b "knowledge\ti-ecosystem-docs\apps-docs\*" 2^>NUL ^| find /C /V ""') do (
            echo 📱 Apps docs: %%A
        )
    )
    
    if exist "knowledge\ti-ecosystem-docs\docs" (
        for /f %%A in ('dir /s /b "knowledge\ti-ecosystem-docs\docs\*" 2^>NUL ^| find /C /V ""') do (
            echo 📖 General docs: %%A
        )
    )
    
    if exist "knowledge\ti-ecosystem-docs\packages" (
        for /f %%A in ('dir /s /b "knowledge\ti-ecosystem-docs\packages\*" 2^>NUL ^| find /C /V ""') do (
            echo 📦 Package docs: %%A
        )
    )
    
    REM Total file count
    for /f %%A in ('dir /s /b "knowledge\ti-ecosystem-docs\*" 2^>NUL ^| find /C /V ""') do (
        echo 📊 Total files: %%A
    )
) else (
    echo ❌ Knowledge base not found
)

echo.

REM Check file types distribution
echo 📁 File Type Distribution:
cd "knowledge\ti-ecosystem-docs"

echo 📄 Markdown files:
dir /s /b "*.md" 2>NUL | find /C /V ""

echo 🐹 Go files:
dir /s /b "*.go" 2>NUL | find /C /V ""

echo 📜 JavaScript files:
dir /s /b "*.js" 2>NUL | find /C /V ""

echo 📘 TypeScript files:
dir /s /b "*.ts" 2>NUL | find /C /V ""

echo 📄 JSON files:
dir /s /b "*.json" 2>NUL | find /C /V ""

echo ⚙️ YAML files:
dir /s /b "*.yaml" 2>NUL | find /C /V ""
dir /s /b "*.yml" 2>NUL | find /C /V ""

echo.

REM Check for potential issues
echo 🔍 Quality Issues Check:

REM Check for empty files
echo 📄 Checking for empty files...
set EMPTY_COUNT=0
for /r %%F in (*) do (
    if %%~zF EQU 0 (
        set /a EMPTY_COUNT+=1
    )
)
if %EMPTY_COUNT% GTR 0 (
    echo ⚠️ Found %EMPTY_COUNT% empty files
) else (
    echo ✅ No empty files found
)

REM Check for very large files
echo 📏 Checking for large files...
set LARGE_COUNT=0
for /r %%F in (*) do (
    if %%~zF GTR 10485760 (
        echo 📊 Large file: %%F (%%~zF bytes)
        set /a LARGE_COUNT+=1
    )
)
if %LARGE_COUNT% EQU 0 (
    echo ✅ No excessively large files found
)

echo.

REM Check indexing progress estimate
echo 📈 Indexing Progress:
if exist "knowledge\ti-ecosystem-docs" (
    for /f %%A in ('dir /s /b "knowledge\ti-ecosystem-docs\*" 2^>NUL ^| find /C /V ""') do (
        set INDEXED_FILES=%%A
    )
    
    echo 📊 Files indexed: %INDEXED_FILES%
    
    REM Estimate progress (assuming we started with ecosystem-docs)
    set ESTIMATED_TOTAL=35000
    set /a PROGRESS=(%INDEXED_FILES% * 100) / %ESTIMATED_TOTAL%
    echo 📈 Estimated progress: %PROGRESS%%%
    
    if %PROGRESS% GTR 90 (
        echo ✅ Indexing nearly complete!
    ) else if %PROGRESS% GTR 70 (
        echo 🚀 Indexing progressing well
    ) else (
        echo 🔄 Indexing in progress
    )
)

echo.
echo === ✅ Quality Check Complete ===
echo.
pause