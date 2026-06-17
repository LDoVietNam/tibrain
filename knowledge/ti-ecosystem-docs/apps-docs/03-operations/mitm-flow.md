# MITM Flow Documentation

Ti Router supports MITM (Man-in-the-Middle) flow for tools that do not natively support custom providers.

## What is MITM Flow?

MITM flow allows Ti Router to intercept and modify API calls from tools that:
- Do not support custom base URLs
- Hard-code API endpoints (e.g., directly call `api.openai.com`)
- Require proxy configuration to redirect traffic

## When to Use MITM Flow

Use MITM flow for tools that:
- **Antigravity** - Requires MITM to intercept API calls
- **GitHub Copilot** - Requires MITM for custom provider support
- **Kiro** - Requires MITM for custom routing
- **Other tools** with hard-coded API endpoints

## How MITM Flow Works

### Architecture

```
Tool → MITM Proxy → Ti Router → Upstream Provider
         (Intercept)    (Route)       (OpenAI, Claude, etc.)
```

### Flow Steps

1. **Tool makes API call** to hard-coded endpoint (e.g., `api.openai.com`)
2. **MITM Proxy intercepts** the call via system proxy configuration
3. **MITM Proxy redirects** to Ti Router (`http://localhost:1807`)
4. **Ti Router routes** to appropriate upstream provider
5. **Response returns** through MITM Proxy back to tool

## Setup MITM Flow

### 1. Configure System Proxy

**Windows:**
```powershell
# Set system proxy
netsh winhttp set proxy 127.0.0.1:8080

# Verify
netsh winhttp show proxy
```

**Linux/Mac:**
```bash
# Set HTTP proxy
export HTTP_PROXY=http://127.0.0.1:8080
export HTTPS_PROXY=http://127.0.0.1:8080
```

### 2. Start MITM Service

Ti Router includes MITM layer at `layers/http/mitm/`:

```bash
cd Z:\Ti\router
.\routerd.exe -config configs/Tiserverrouter.yaml
```

MITM service automatically starts when router initializes.

### 3. Configure Tool

Configure tool to use system proxy or custom proxy settings:

**Antigravity:**
```json
{
  "proxy": {
    "http": "http://127.0.0.1:8080",
    "https": "http://127.0.0.1:8080"
  }
}
```

**GitHub Copilot:**
```json
{
  "proxy": "http://127.0.0.1:8080"
}
```

### 4. Certificate Trust (if needed)

Some tools require MITM certificate to be trusted:

```bash
# Generate MITM certificate
cd Z:\Ti\router
.\bin\generate-cert.exe

# Install certificate (Windows)
certutil -addstore Root mitm-cert.pem

# Install certificate (Linux/Mac)
sudo cp mitm-cert.pem /usr/local/share/ca-certificates/
sudo update-ca-certificates
```

## MITM Layer Architecture

Ti Router's MITM layer (`layers/http/mitm/`) provides:

### Components

- **Proxy Server** - HTTP/HTTPS proxy server
- **Certificate Generator** - Self-signed certificate generation
- **Request Interceptor** - Intercept and modify requests
- **Response Modifier** - Modify responses if needed
- **SSL/TLS Termination** - Terminate SSL for inspection

### Configuration

MITM configuration in `configs/Tiserverrouter.yaml`:

```yaml
mitm:
  enabled: true
  port: 8080
  cert_file: "certs/mitm-cert.pem"
  key_file: "certs/mitm-key.pem"
  intercept_domains:
    - "api.openai.com"
    - "api.anthropic.com"
    - "generativelanguage.googleapis.com"
```

## Security Considerations

### Risks

- **Certificate Trust** - MITM requires certificate installation
- **Traffic Inspection** - All traffic through proxy is visible
- **System-wide Impact** - System proxy affects all applications

### Mitigations

- **Scope Limitation** - Only intercept specific domains
- **Certificate Validation** - Use self-signed certificates only for MITM
- **Local-only Proxy** - Bind to localhost (127.0.0.1) only
- **Disable When Not Needed** - Turn off MITM when not in use

### Best Practices

1. **Use MITM only when necessary** - Prefer native custom provider support
2. **Limit to specific tools** - Don't enable system-wide proxy permanently
3. **Monitor traffic** - Audit MITM logs for suspicious activity
4. **Keep certificates secure** - Don't share MITM certificates
5. **Disable after use** - Turn off MITM proxy when done

## Troubleshooting

### Certificate Errors

**Problem:** `SSL certificate error` or `certificate verify failed`

**Solution:** Install MITM certificate in system trust store

### Connection Refused

**Problem:** `Connection refused` when tool tries to connect

**Solution:** Verify MITM proxy is running on port 8080

### Proxy Not Working

**Problem:** Tool not using proxy despite configuration

**Solution:** 
- Verify system proxy settings
- Check tool-specific proxy configuration
- Restart tool after proxy configuration

### DNS Resolution Issues

**Problem:** Tool can't resolve hostnames through proxy

**Solution:** Configure DNS bypass for local addresses

## Alternatives to MITM

### 1. Native Custom Provider Support

Prefer tools with native custom provider support:
- **Claude Code** - Supports `ANTHROPIC_BASE_URL`
- **OpenCode** - Supports custom provider config
- **Cursor** - Supports custom base URL
- **Cline** - Supports OpenAI-compatible provider

See [Tool Guides](../05-integration/tool-guides/) for native setup.

### 2. Environment Variables

Some tools support environment variables:
```bash
export OPENAI_API_BASE=http://localhost:1807/v1
export ANTHROPIC_BASE_URL=http://localhost:1807/v1
```

### 3. Config File Modification

Modify tool's config file to point to Ti Router:
```json
{
  "api_base": "http://localhost:1807/v1"
}
```

## Additional Resources

- [Tool Guides](../05-integration/tool-guides/)
- [Unified Endpoint Pattern](./unified-endpoint.md)
- [Security Guide](./02-security.md)
- [MITM Layer Code](../../layers/http/mitm/)

## MITM for Specific Tools

### Antigravity

Antigravity requires MITM for custom provider support:

1. Configure Antigravity to use proxy: `http://127.0.0.1:8080`
2. Enable MITM in Ti Router config
3. Restart both Antigravity and Ti Router

### GitHub Copilot

GitHub Copilot requires MITM for custom routing:

1. Configure system proxy to point to Ti Router MITM
2. Install MITM certificate in system trust store
3. Restart GitHub Copilot

### Kiro

Kiro requires MITM for custom provider:

1. Configure Kiro proxy settings
2. Enable domain interception for Kiro's API endpoints
3. Test with simple API call first

## Notes

- MITM is a powerful feature but should be used carefully
- Always disable MITM when not needed
- Monitor MITM logs for security issues
- Keep MITM certificates secure and up-to-date
