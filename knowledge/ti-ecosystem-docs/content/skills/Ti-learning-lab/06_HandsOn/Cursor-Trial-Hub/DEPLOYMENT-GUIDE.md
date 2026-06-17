# 🚀 DEPLOYMENT GUIDE - FINAL

**Version:** Final with ALL tokens  
**Date:** 2026-03-27  
**Size:** ~41 KB  
**Status:** ✅ Ready to Deploy

---

## ⚡ **QUICK DEPLOY (5 PHÚT):**

### **BƯỚC 1: COPY ZIP SANG PC PHỤ**

```
Copy: Cursor-Trial-Hub-FINAL.zip
To: PC Phụ
Extract to: C:\Cursor-Trial-Hub
```

---

### **BƯỚC 2: SETUP AI HUB (2 phút)**

```powershell
# Run as Administrator
cd C:\Cursor-Trial-Hub\PC-Phu-AI-Hub

# Setup everything
.\setup-pc-phu-ai-hub.ps1
.\configure-openclaw.ps1
```

**Script sẽ:**
```
✅ Install OpenClaw (C:\openclaw)
✅ Setup CLIProxyAPI (Z:\Router)
✅ Setup My-Router (Z:\My-Router)
✅ Configure auto-start
✅ Create Tools folder
✅ Setup OpenClaw .env with tokens
```

---

### **BƯỚC 3: SETUP TUNNEL (2 phút)**

```powershell
# Quick tunnel setup
.\setup-tunnel-simple.ps1
```

**Chọn Cloudflare:**
```
1. https://dash.teams.cloudflare.com
2. Zero Trust → Access → Tunnels
3. Create tunnel → Web
4. Add hostnames:
   - ai-hub.workers.dev → localhost:8080
   - api-hub.workers.dev → localhost:1810
   - router-hub.workers.dev → localhost:1809
```

---

### **BƯỚC 4: START SERVICES (1 phút)**

```powershell
# Start all services
.\start-ai-hub.ps1

# Or individually
.\start-cursor.ps1
.\start-tunnels.ps1
```

---

## ✅ **VERIFICATION:**

### **CHECK TOKENS:**

```powershell
# View pre-configured tokens
notepad C:\Cursor-Trial-Hub\Tokens\tokens.env

# Should see:
# ✅ CLIPROXY_API_KEY=sk-jarvis-dev
# ✅ MYROUTER_API_KEY=sk-c5d5bc53d5225f9d-slgify-2b6eeca2
# ✅ IFLOW_API_KEY=sk-d9fdcc0b27a39642876651e5bac985a7
# ✅ GEMINI_API_KEY=sk-agents-dev
```

### **CHECK SERVICES:**

```powershell
# Test OpenClaw
Invoke-RestMethod http://localhost:8080

# Test CLIProxyAPI
Invoke-RestMethod http://localhost:1810/v1/models

# Test My-Router
Invoke-RestMethod http://localhost:1809/v1/models
```

### **CHECK TUNNEL:**

```powershell
# Test tunnel access
Invoke-RestMethod https://ai-hub.workers.dev
```

---

## 🌐 **ACCESS:**

### **LOCAL:**

```
http://localhost:8080   # OpenClaw
http://localhost:1810   # CLIProxyAPI
http://localhost:1809   # My-Router
```

### **VIA TUNNEL:**

```
https://ai-hub.workers.dev       # OpenClaw Web
https://api-hub.workers.dev      # API Access
https://router-hub.workers.dev   # Router Admin
```

### **FROM PC CHÍNH:**

```powershell
# Local network
$env:ANTHROPIC_BASE_URL = "http://192.168.1.100:1810"
claude "Hello!"

# Via tunnel
$env:ANTHROPIC_BASE_URL = "https://api-hub.workers.dev"
claude "Hello from anywhere!"
```

---

## 🔄 **TRIAL RESET:**

### **AUTO:**

```
✅ Configured to reset every 14 days
✅ Scheduled task: Cursor-Trial-Reset
✅ Device Spoofer: C:\Tools\Device-Info-21AK22.exe
```

### **MANUAL:**

```powershell
# When trial expires
cd C:\Cursor-Trial-Hub\PC-Phu-AI-Hub
.\auto-reset-pc-phu.ps1 -SpoofExe "C:\Tools\Device-Info-21AK22.exe"
```

---

## 📞 **TELEGRAM (OPTIONAL):**

```powershell
# 1. Get bot token from @BotFather
# 2. Edit: C:\openclaw\.env
# Add:
TELEGRAM_BOT_TOKEN=YOUR_TOKEN
TELEGRAM_CHAT_ID=YOUR_CHAT_ID
```

---

## 🔑 **TOKENS ĐÃ CẤU HÌNH:**

```env
✅ CLIPROXY_API_KEY=sk-jarvis-dev
✅ MYROUTER_API_KEY=sk-c5d5bc53d5225f9d-slgify-2b6eeca2
✅ IFLOW_API_KEY=sk-d9fdcc0b27a39642876651e5bac985a7
✅ GEMINI_API_KEY=sk-agents-dev
```

**TẤT CẢ ĐÃ CÓ SẴN - KHÔNG CẦN CONFIG THÊM!**

---

## 📊 **SERVICES:**

| Service | Port | Token | Status |
|---------|------|-------|--------|
| **OpenClaw** | 8080 | sk-jarvis-dev | ✅ Ready |
| **CLIProxyAPI** | 1810 | sk-jarvis-dev | ✅ Ready |
| **My-Router** | 1809 | sk-c5d5bc53d5225f9d-slgify-2b6eeca2 | ✅ Ready |
| **iFlow API** | External | sk-d9fdcc0b27a39642876651e5bac985a7 | ✅ Ready |
| **Tunnel** | HTTPS | Cloudflare | ⏳ Configure |

---

## 🎯 **SUCCESS CHECKLIST:**

```
✅ ZIP copied to PC Phụ
✅ Extracted to C:\Cursor-Trial-Hub
✅ setup-pc-phu-ai-hub.ps1 run
✅ configure-openclaw.ps1 run
✅ Tunnel configured
✅ Services started
✅ Can access localhost:8080
✅ Can access localhost:1810
✅ Can access localhost:1809
✅ Can access via tunnel
✅ Tokens pre-configured
```

---

## 📝 **LOGS:**

```powershell
# OpenClaw
Get-Content "C:\openclaw\logs\openclaw.log" -Tail 50

# Tunnel
Get-Content "C:\Tools\tunnel.log" -Tail 50

# Tokens
Get-Content "C:\Cursor-Trial-Hub\Tokens\tokens.env"
```

---

## 🚀 **DONE!**

**Everything pre-configured with REAL tokens!**

**Just extract and run!**

---

**LAST UPDATED:** 2026-03-27  
**STATUS:** ✅ Production Ready  
**TOKENS:** ✅ Pre-configured
