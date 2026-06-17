# 💻 PC CHÍNH - CONNECT TO AI HUB

**Location:** `Z:\projects\SJ-learning-lab\04_Projects\Cursor-Trial-Hub\PC-Chinh\`

**Purpose:** Connect từ PC Chính tới AI Hub trên PC Phụ

---

## 📁 **FILES:**

| File | Purpose |
|------|---------|
| `README.md` | This guide |

---

## 🚀 **SETUP PC CHÍNH:**

### **BƯỚC 1: TÌM IP PC PHỤ:**

```powershell
# Trên PC Phụ, check IP
ipconfig

# Example: 192.168.1.100
```

---

### **BƯỚC 2: CẤU HÌNH ENVIRONMENT:**

```powershell
# Set environment variables
$env:ANTHROPIC_BASE_URL = "http://192.168.1.100:1810"
$env:ANTHROPIC_AUTH_TOKEN = "sk-jarvis-dev"

# Add to $PROFILE for persistence:
Add-Content $PROFILE @"
# AI Hub Connection
`$env:ANTHROPIC_BASE_URL = "http://192.168.1.100:1810"
`$env:ANTHROPIC_AUTH_TOKEN = "sk-jarvis-dev"
"@

# Reload profile
. $PROFILE
```

---

### **BƯỚC 3: TEST CONNECTION:**

```powershell
# Test CLIProxyAPI
Invoke-RestMethod "http://192.168.1.100:1810/v1/models" | Select-Object -ExpandProperty data

# Test OpenClaw
Invoke-RestMethod "http://192.168.1.100:8080"

# Test My-Router
Invoke-RestMethod "http://192.168.1.100:1809/v1/models"
```

---

### **BƯỚC 4: SỬ DỤNG:**

```powershell
# Use claude CLI
claude "Hello from PC chinh!"

# Or use Cursor IDE
# Settings → AI → Add Provider
#   Name: AI Hub
#   Base URL: http://192.168.1.100:1810
#   API Key: sk-jarvis-dev
```

---

## 🌐 **NETWORK ACCESS:**

```
PC Chính (192.168.1.x)
      ↕
PC Phụ (192.168.1.100) - AI Hub
  ├── OpenClaw:      http://192.168.1.100:8080
  ├── CLIProxyAPI:   http://192.168.1.100:1810
  └── My-Router:     http://192.168.1.100:1809
```

---

## 🔧 **TROUBLESHOOTING:**

### **Cannot connect:**

```powershell
# Check PC Phụ is on
Test-Connection 192.168.1.100

# Check ports
Test-NetConnection 192.168.1.100 -Port 1810
Test-NetConnection 192.168.1.100 -Port 1809
Test-NetConnection 192.168.1.100 -Port 8080

# Check firewall on PC Phụ
# Allow ports: 1809, 1810, 8080
```

### **Services not running:**

```powershell
# On PC Phụ, restart services
# Or reboot PC Phụ
```

---

## ✅ **SUCCESS CHECKLIST:**

```
✅ Can ping PC Phụ
✅ Can access http://192.168.1.100:1810
✅ Can access http://192.168.1.100:1809
✅ Can access http://192.168.1.100:8080
✅ claude CLI works
✅ Cursor IDE connected
```

---

## 🎯 **DAILY USE:**

```powershell
# Just use AI services
claude "Write a function to..."

# Or via OpenClaw UI
# http://192.168.1.100:8080
```

---

**LAST UPDATED:** 2026-03-27  
**VERSION:** 1.0  
**STATUS:** ✅ Ready to Connect
