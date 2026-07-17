# TiBrain Build Verification Script
# Run this from the project root to verify the build after internal package scaffold

param(
    [switch]$Test
)

Write-Host "🔨 Building TiBrain (go build ./...)" -ForegroundColor Cyan

try {
    $buildOutput = go build ./... 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Build SUCCESS" -ForegroundColor Green
        if ($buildOutput) {
            Write-Host $buildOutput
        }
    } else {
        Write-Host "❌ Build FAILED (exit code: $LASTEXITCODE)" -ForegroundColor Red
        Write-Host $buildOutput
        exit $LASTEXITCODE
    }
} catch {
    Write-Host "❌ Build error: $_" -ForegroundColor Red
    exit 1
}

if ($Test) {
    Write-Host "`n🧪 Running tests..." -ForegroundColor Cyan
    go test -v -race -coverprofile=coverage.out ./...
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Tests PASSED" -ForegroundColor Green
    } else {
        Write-Host "❌ Tests FAILED" -ForegroundColor Red
        exit $LASTEXITCODE
    }
}