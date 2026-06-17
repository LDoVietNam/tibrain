# Ti-Router Dual Implementation

## Setup Node.js Version (Active)

```bash
cd Z:\Ti\router\noderouter

# 1. Install dependencies
npm install

# 2. Configure environment
# Copy .env.example to .env.local and fill in your values:
cp .env.example .env.local
# Then edit .env.local with your API keys

# 3. Setup database (if not exists)
# Database will auto-initialize on first run
# Location: Z:\Ti\router\noderouter\data\router.db

# 4. Start development server
npm run dev
# Open http://localhost:3100
```

---

## **Ports**

| Service | URL |
|---------|-----|
| Web UI | http://localhost:3100 |
| API | http://localhost:1816 |
| SSE Stream | http://localhost:3003 |
| MCP | http://localhost:3101 |

---

## **Go Version (Disabled)**

Go implementation at `layers.disabled/` is **experimental** and not ready.

To enable (future):
```bash
cd Z:\Ti\router\layers.disabled
go build -o spectre.exe .
spectre.exe --port 1807
```

---

## **Project Structure**

```
noderouter/
├── .env.local              # Your secrets (copy from .env.example)
├── package.json
├── src/
│   ├── app/               # Next.js App Router
│   │   ├── api/           # API endpoints (100+ routes)
│   │   ├── page.tsx       # Main UI page
│   │   └── layout.tsx     # Root layout
│   ├── components/        # React components
│   │   ├── ui/           # ShadCN UI components
│   │   └── ...
│   ├── lib/              # Core libraries
│   │   ├── db/           # SQLite client
│   │   ├── providers/    # AI provider implementations
│   │   ├── oauth/        # OAuth flows
│   │   ├── usage/        # Usage tracking
│   │   └── ...
│   └── ...
├── data/                  # SQLite database
├── logs/                  # Application logs
├── cookies/               # Browser cookie exports
└── open-sse/              # SSE streaming server

layers.disabled/           # Go implementation (backup)
├── layers/
│   ├── http/
│   ├── authentication/
│   ├── provider/
│   ├── routing/
│   ├── db/
│   └── main.go
└── go.mod
```

---

## **API Quick Reference**

### **OAuth Import (Windsurf)**
```bash
# GET instructions
curl http://localhost:1816/api/oauth/windsurf/import

# POST token import
curl -X POST http://localhost:1816/api/oauth/windsurf/import \
  -H "Content-Type: application/json" \
  -d '{"accessToken":"ott$..."}'
```

### **Chat Completions (OpenAI-compatible)**
```bash
curl http://localhost:1816/v1/chat/completions \
  -H "Authorization: Bearer sk-jarvis-dev" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "windsurf/claude-4.5-sonnet-thinking",
    "messages": [{"role":"user","content":"Hello"}]
  }'
```

### **List Models**
```bash
curl http://localhost:1816/v1/models \
  -H "Authorization: Bearer sk-jarvis-dev"
```

---

## **Troubleshooting**

### **Port already in use**
```bash
# Kill process on port 3100 (Windows)
netstat -ano | findstr :3100
taskkill /PID <PID> /F

# Or change port in .env.local:
# PORT=3101
```

### **Database errors**
```bash
# Delete database to reset (WARNING: loses data)
rm Z:\Ti\router\noderouter\data\router.db

# Restart server - DB will recreate
```

### **Module not found**
```bash
cd Z:\Ti\router\noderouter
rm -rf node_modules package-lock.json
npm install
```

---

## **Deployment**

### **Development**
```bash
npm run dev
```

### **Production Build**
```bash
npm run build
npm start
```

### **Docker**
```bash
docker-compose up -d
```

---

## **Next Steps**

1. ✅ Setup Node.js version (above)
2. ⏭️ Import Windsurf token via UI or API
3. ⏭️ Test chat completion
4. ⏭️ Explore management panel at http://localhost:3100
5. ⏭️ (Optional) Try Go version when ready

---

## **Support**

- Issues: https://github.com/anomalyco/opencode/issues
- Docs: `Z:\Ti\router\noderouter\docs\`
- CLI tools: `Z:\Ti\router\noderouter\bin\`
