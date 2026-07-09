$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

Write-Host "[Ti Console] Formatting Go sources..." -ForegroundColor Cyan
$sourceFiles = @(
    Get-ChildItem -Path ".\internal\ticonsole" -Filter "*.go" -File
    Get-ChildItem -Path ".\cmd\ti-console" -Filter "*.go" -File
) | ForEach-Object { $_.FullName }

if ($sourceFiles.Count -gt 0) {
    & gofmt -w @sourceFiles
}

Write-Host "[Ti Console] Running focused tests..." -ForegroundColor Cyan
& go test ./internal/ticonsole ./cmd/ti-console
if ($LASTEXITCODE -ne 0) {
    throw "Ti Console tests failed with exit code $LASTEXITCODE"
}

$binDir = Join-Path $root "bin"
New-Item -ItemType Directory -Force -Path $binDir | Out-Null
$output = Join-Path $binDir "ti-console.exe"

Write-Host "[Ti Console] Building $output..." -ForegroundColor Cyan
& go build -trimpath -o $output ./cmd/ti-console
if ($LASTEXITCODE -ne 0) {
    throw "Ti Console build failed with exit code $LASTEXITCODE"
}

Write-Host "[Ti Console] Build completed: $output" -ForegroundColor Green
