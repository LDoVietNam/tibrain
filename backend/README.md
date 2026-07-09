# TiBrain Cloudflare Backend

Wrapper backend for Cloudflare tunnel and DNS operations.

## Run

```powershell
cd Z:\01_PROJECTS\apps\tibrain
$env:CLOUDFLARE_API_TOKEN="..."
$env:CLOUDFLARE_ACCOUNT_ID="..."
$env:CLOUDFLARE_ZONE_ID="..."
$env:TIBRAIN_PUBLIC_HOSTNAME="tibrain.trepremium.online"
$env:TIBRAIN_TUNNEL_NAME="tibrain-fixed"
$env:TIBRAIN_ORIGIN_URL="http://127.0.0.1:1810"
go run ./backend
```

## Endpoints

- `GET /health`
- `GET /zones`
- `GET /zones/{zone_id}/dns-records`
- `GET /tunnels`
- `GET /preflight/scope-check`
- `POST /deploy/fixed-tunnel`

## Notes

- The backend keeps the Cloudflare API token server-side.
- A fixed tunnel requires `Cloudflare Tunnel Edit` and `DNS Edit` permissions on the token.
- `GET /preflight/scope-check` runs a safe permission probe:
  - `DNS Edit` is checked with a `PUT` against a fake DNS record id.
  - `Tunnel Edit` is checked with a `PUT .../configurations?validate_only=true` against a fake tunnel id.
- `POST /deploy/fixed-tunnel` runs the same preflight first and returns `403` with the probe details if a required write scope is missing.
- `validate_only=true` still skips resource creation and DNS writes; it only validates against an existing tunnel when one is already present.
