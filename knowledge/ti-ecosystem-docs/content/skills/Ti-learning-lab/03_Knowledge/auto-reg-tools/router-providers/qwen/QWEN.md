# QWEN.md - Project Context & Instructions

## 🔧 GO INSTALLATION (CRITICAL)

**Go Location:** `Z:\09_TOOLS\go`

**IMPORTANT:**
- Go installation status needs verification
- If `go version` fails, reinstall from https://go.dev/dl/
- Install to `Z:\09_TOOLS\go` and add to PATH

**Environment Setup:**
```powershell
$env:GOROOT = "Z:\09_TOOLS\go"
$env:GOPATH = "Z:\go"
$env:PATH = "$env:GOROOT\bin;$env:PATH"
```

---

## 📊 ACTIVE PROJECTS

### 1. Ticlaw (Go CLI Gateway)
- **Path:** `Z:\01_PROJECTS\ticlaw\`
- **Type:** AI CLI Gateway (100% Free Models)
- **Language:** Go 1.26.1
- **Module:** `github.com/nextlevelbuilder/ticlaw`
- **Status:** ✅ Active development

### 2. Auto Reg Main (Python Auto-Registration)
- **Path:** `Z:\01_PROJECTS\ticlaw\auto_reg-main\`
- **Type:** Multi-platform AI account registration
- **Language:** Python 3.12+ (FastAPI + React)
- **Status:** ✅ Active development
- **Master Plan:** `docs/MASTER_PLAN.md` (95 tasks, 14 phases)

---

## 🔐 SECRETS MANAGEMENT

**All credentials stored in `Z:\00_SECRET\` - SINGLE SOURCE OF TRUTH.**

```
Z:\00_SECRET\
├── ticlaw.env              → Ticlaw Go CLI keys
├── auto_reg-main.env       → Auto Reg Python config
├── README.md               → Secrets guide
└── INDEX.md                → Credentials index
```

**Rules:**
- ❌ NEVER store secrets in project folders
- ❌ NEVER commit `.env` files with secrets
- ✅ ONLY store secrets in `Z:\00_SECRET\`
- ✅ Code auto-loads from `Z:\00_SECRET\` (with fallback)

**17+ credentials already available** from `Z:\02_CORE\06_database\config\`:
- Google OAuth, Mail72H, Browserbase, n8n, Tailscale
- Exa AI (6 keys), ElevenLabs (3 keys), DeepSeek, Modal, GitHub, Orbit

**Missing (need manual setup):**
- Azure AD: Tenant ID, Client ID, Client Secret
- Azure Domains: 2 custom domains from M365

---

## 📁 KEY FILES

| File | Purpose |
|------|---------|
| `Z:\README.md` | Workspace overview + active projects |
| `Z:\AGENTS.md` | CLI automation guide + secrets rules |
| `Z:\00_SECRET\README.md` | Secrets management guide |
| `Z:\01_PROJECTS\ticlaw\auto_reg-main\docs\MASTER_PLAN.md` | Full implementation plan (95 tasks) |
| `Z:\01_PROJECTS\ticlaw\auto_reg-main\docs\INTEGRATION_PLAN.md` | GitHub repos integration |
| `Z:\01_PROJECTS\ticlaw\auto_reg-main\docs\PLAN_CHANGES_SUMMARY.md` | Plan change log |

---

## 🛠️ BUILD & TEST

### Go (Ticlaw)
```bash
cd Z:\01_PROJECTS\ticlaw
go build ./...
go test ./... -v
```

### Python (Auto Reg Main)
```bash
cd Z:\01_PROJECTS\ticlaw\auto_reg-main
py -m py_compile core/*.py services/*.py api/*.py
py -m pytest tests/ -v
```

---

## 📋 DEVELOPMENT RULES

1. **Always run** `go build ./...` after Go code changes
2. **Always run** `py -m py_compile` after Python code changes
3. **Always run** tests after feature additions
4. **Use table-driven tests** in Go
5. **Keep functions < 100 lines** when possible
6. **Add tests immediately** for new code
7. **NEVER hardcode secrets** - use `Z:\00_SECRET\`
8. **Update MASTER_PLAN.md** when starting new tasks

---

## 🎯 CURRENT PRIORITIES

### Option B: Balanced (46 tasks, ~88h) ⭐ RECOMMENDED
- **Phase P1:** Critical Security (4 tasks)
- **Phase P2:** Performance (6 tasks)
- **Phase P3:** Security Hardening (5 tasks)
- **Phase P4:** Code Quality (7 tasks)
- **Phase P8:** Enterprise Integration (5 tasks)
- **Phase AZ:** Azure Integration (7 tasks)
- **Phase ENV:** Existing Credentials (7 HIGH tasks)

### Next Immediate Actions:
1. Get Azure credentials from Azure Portal
2. Import all credentials to `Z:\00_SECRET\auto_reg-main.env`
3. Start with P1-01: Password hashing SHA-256 → bcrypt

---

**Last Updated:** 2026-04-07  
**Status:** Active Development  
**Next Step:** Azure credentials setup → Begin Option B plan
