# Ti Router Performance Optimization

## Build & Run

```bash
cd Z:\Ti\router

# Build
go build -o bin/routerd.exe ./cmd/routerd

# Run
./bin/routerd.exe -config configs/Tiserverrouter.yaml
# → http://localhost:1807
```

## Memory & Performance

Go runtime tự quản lý memory tốt. Không cần cấu hình thêm.

## Database Optimization

```bash
# Vacuum SQLite database
sqlite3 data/router.db "VACUUM; ANALYZE;"
```

## Common Issues

### Router won't start
- Check port 1807: `netstat -ano | findstr :1807`
- Verify config file syntax

### API returns 401
- Verify API key in config
- Use header: `Authorization: Bearer <key>`

## Quick Test

```bash
# Health check
curl http://localhost:1807/health

# List models
curl -H "Authorization: Bearer sk-jarvis-dev" http://localhost:1807/v1/models
```

---

Last Updated: 2026-04-27