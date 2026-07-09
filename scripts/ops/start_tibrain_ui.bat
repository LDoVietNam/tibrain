@echo off
setlocal

set "ROOT=%~dp0"
set "PORT=1810"
set "URL=http://localhost:%PORT%/overview"
set "BIN=%ROOT%tibrain.exe~"
if exist "%ROOT%tibrain.exe" set "BIN=%ROOT%tibrain.exe"
if exist "%ROOT%tibrain.next.exe" set "BIN=%ROOT%tibrain.next.exe"

pushd "%ROOT%" >nul

"%SystemRoot%\System32\WindowsPowerShell\v1.0\powershell.exe" -NoProfile -ExecutionPolicy Bypass -Command "Get-NetTCPConnection -LocalPort %PORT% -State Listen -ErrorAction SilentlyContinue | Select-Object -ExpandProperty OwningProcess -Unique | ForEach-Object { Stop-Process -Id $_ -Force -ErrorAction SilentlyContinue }"

if exist "%BIN%" (
    start "" "%BIN%" --port %PORT%
) else (
    where go >nul 2>&1
    if errorlevel 1 (
        echo [ERROR] No TiBrain binary found and go.exe is not on PATH.
        popd >nul
        exit /b 1
    )
    start "" cmd /c "go run . --port %PORT%"
)

"%SystemRoot%\System32\WindowsPowerShell\v1.0\powershell.exe" -NoProfile -ExecutionPolicy Bypass -Command "$url = '%URL%'; for ($i = 0; $i -lt 30; $i++) { try { Invoke-RestMethod -Uri $url -TimeoutSec 2 -ErrorAction Stop | Out-Null; break } catch { Start-Sleep -Seconds 1 } }; Start-Process $url"

popd >nul
endlocal
