param(
    [int]$RefreshSeconds = 3,
    [switch]$NoColor,
    [switch]$Once
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$binary = Join-Path $root "bin\ti-console.exe"

if (-not (Test-Path $binary)) {
    & (Join-Path $PSScriptRoot "build-ti-console.ps1")
}

$arguments = @("--refresh", "${RefreshSeconds}s")
if ($NoColor) { $arguments += "--no-color" }
if ($Once) { $arguments += "--once" }

& $binary @arguments
exit $LASTEXITCODE
