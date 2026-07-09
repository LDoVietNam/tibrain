Write-Output "=== Starting MCP Hub ==="
$hubExe = "Z:\01_PROJECTS\apps\tibrain\mcp\.runtime\current\mcpproxy.exe"
$hubConfig = "Z:\01_PROJECTS\apps\tibrain\mcp\.runtime\config\mcp_config.json"
$hubData = "Z:\01_PROJECTS\apps\tibrain\mcp\.runtime\data"
$hubLogs = "Z:\01_PROJECTS\apps\tibrain\mcp\.runtime\logs\gateway"

if (Test-Path $hubExe) {
    Write-Output "Hub binary found, starting..."
    $p = Start-Process -FilePath $hubExe -ArgumentList "--config",$hubConfig,"--data-dir",$hubData,"--log-dir",$hubLogs,"--log-to-file","serve","--listen","127.0.0.1:1840" -NoNewWindow -PassThru
    Write-Output "Hub PID: $($p.Id)"
} else {
    Write-Output "Hub binary NOT FOUND at $hubExe"
}

Start-Sleep -Seconds 3

Write-Output "=== Starting TiBrain ==="
$tibrainExe = "Z:\01_PROJECTS\apps\tibrain\tibrain.exe"
if (Test-Path $tibrainExe) {
    Write-Output "TiBrain binary found, starting..."
    $p2 = Start-Process -FilePath $tibrainExe -ArgumentList "--port","1810" -NoNewWindow -PassThru
    Write-Output "TiBrain PID: $($p2.Id)"
} else {
    Write-Output "TiBrain binary NOT FOUND at $tibrainExe"
}

Start-Sleep -Seconds 8

Write-Output "=== Health Check ==="
try {
    $health = Invoke-RestMethod -Uri "http://127.0.0.1:1810/health" -TimeoutSec 5
    Write-Output "TiBrain health: $($health | ConvertTo-Json -Depth 8)"
} catch {
    Write-Output "TiBrain health FAILED: $($_.Exception.Message)"
}

try {
    $key = Get-Content "Z:\01_PROJECTS\apps\tibrain\mcp\.runtime\secrets\api-key.txt" -Raw
    $key = $key.Trim()
    $headers = @{ "X-API-Key" = $key }
    $servers = Invoke-RestMethod -Uri "http://127.0.0.1:1840/api/v1/servers" -Headers $headers -TimeoutSec 5
    Write-Output "Hub servers: $($servers | ConvertTo-Json -Depth 8)"
} catch {
    Write-Output "Hub servers FAILED: $($_.Exception.Message)"
}

Write-Output "=== Starting Cloudflare Tunnel ==="
try {
    $tunnelProcess = Start-Process -FilePath "cloudflared" -ArgumentList "tunnel","--url","http://localhost:1810" -RedirectStandardOutput "Z:\01_PROJECTS\apps\tibrain\tunnel.log" -RedirectStandardError "Z:\01_PROJECTS\apps\tibrain\tunnel.err" -NoNewWindow -PassThru
    Write-Output "Tunnel PID: $($tunnelProcess.Id)"
    Start-Sleep -Seconds 10
    if (Test-Path "Z:\01_PROJECTS\apps\tibrain\tunnel.log") {
        $log = Get-Content "Z:\01_PROJECTS\apps\tibrain\tunnel.log" -Tail 10
        Write-Output "Tunnel log: $log"
    }
} catch {
    Write-Output "Tunnel start FAILED: $($_.Exception.Message)"
}

Write-Output "=== DONE ==="
