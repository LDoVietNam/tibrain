# Management Panel Guide

## Access Management Panel

The CLIProxyAPI management panel provides a web UI for monitoring and configuration.

### URL
```
http://localhost:1807/v0/management/index.html
```

### Authentication
Add this header to your request:
```
X-Management-Key: $2a$10$3r42boO4dJZVP.cFkcCdJeH.2CtLkDtOJXZkS2fssZirl8TUSyHy
```

### Access via Browser

Since browsers don't support custom headers easily, use one of these methods:

#### Method 1: Use curl
```bash
curl -H "X-Management-Key: $2a$10$3r42boO4dJZVP.cFkcCdJeH.2CtLkDtOJXZkS2fssZirl8TUSyHy" http://localhost:1807/v0/management/index.html
```

#### Method 2: Use PowerShell
```powershell
$headers = @{"X-Management-Key" = "$2a$10$3r42boO4dJZVP.cFkcCdJeH.2CtLkDtOJXZkS2fssZirl8TUSyHy"}
Invoke-WebRequest -Uri "http://localhost:1807/v0/management/index.html" -Headers $headers -OutFile "management.html"
```

#### Method 3: Use browser extension
Install a browser extension that allows adding custom headers (like "ModHeader").

## Features

Once accessed, the management panel provides:
- API key management
- Provider status monitoring
- Usage statistics
- Configuration overview
- Log viewer

## Configuration

Management panel settings in `Tiserverrouter.yaml`:
```yaml
remote-management:
  allow-remote: true
  secret-key: "$2a$10$3r42boO4dJZVP.cFkcCdJeH.2CtLkDtOJXZkS2fssZirl8TUSyHy"
  disable-control-panel: false
  panel-github-repository: "https://github.com/router-for-me/Cli-Proxy-API-Management-Center"
```

## Troubleshooting

If the management panel doesn't load:
1. Check if router is running: `.\check-status.bat`
2. Verify management key is correct in config
3. Check firewall settings
4. Ensure port 1807 is accessible
