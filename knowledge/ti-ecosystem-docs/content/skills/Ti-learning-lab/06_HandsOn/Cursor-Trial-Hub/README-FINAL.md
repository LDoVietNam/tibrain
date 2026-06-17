# 🚀 CURSOR TRIAL HUB - FINAL PACKAGE

**Version:** Final  
**Date:** 2026-03-27  
**Size:** ~36 KB  
**Status:** ✅ Ready to Deploy

---

## 📦 **CONTENTS:**

```
Cursor-Trial-Hub-FINAL.zip
├── PC-Phu-AI-Hub/           (12 scripts + configs)
│   ├── setup-pc-phu-ai-hub.ps1
│   ├── configure-openclaw.ps1
│   ├── auto-reset-pc-phu.ps1
│   ├── setup-tunnel-simple.ps1
│   ├── start-tunnels.ps1
│   └── ...
├── PC-Chinh/                (Connection guide)
│   └── README.md
├── Tools/                   (Device Spoofer + logs)
└── README-FINAL.md          (This file)
```

---

## 🚀 **DEPLOYMENT:**

### **STEP 1: COPY TO PC PHỤ**

```powershell
# Copy ZIP to PC Phụ
# Extract to: C:\Cursor-Trial-Hub
```

### **STEP 2: RUN SETUP**

```powershell
# Run as Administrator
cd C:\Cursor-Trial-Hub\PC-Phu-AI-Hub

# 1. Setup AI Hub
.\setup-pc-phu-ai-hub.ps1

# 2. Configure OpenClaw
.\configure-openclaw.ps1

# 3. Setup Tunnel
.\setup-tunnel-simple.ps1
```

### **STEP 3: START SERVICES**

```powershell
# Start all services
.\start-ai-hub.ps1

# Or individually:
.\start-cursor.ps1
.\start-tunnels.ps1
```

---

## 🌐 **CONFIGURE DOMAIN:**

### **CLOUDFLARE TUNNEL:**

1. **Get Domain:**
   - Free: `YOUR-NAME.workers.dev`
   - Or use custom domain

2. **Setup Tunnel:**
   ```
   https://dash.teams.cloudflare.com
   → Zero Trust → Access → Tunnels
   → Create Tunnel → Web
   ```

3. **Add Hostnames:**
   ```
   ai-hub.YOURDOMAIN.com     → localhost:8080
   api-hub.YOURDOMAIN.com    → localhost:1810
   router-hub.YOURDOMAIN.com → localhost:1809
   ```

4. **Update .env:**
   ```powershell
   # Edit: C:\openclaw\.env
   TUNNEL_DOMAIN=ai-hub.YOURDOMAIN.com
   ```

---

## 📞 **CONFIGURE TELEGRAM:**

### **GET BOT TOKEN:**

1. **Message @BotFather on Telegram**
2. **Send:** `/newbot`
3. **Follow instructions**
4. **Copy token**

### **GET CHAT ID:**

1. **Message your new bot**
2. **Visit:** `https://api.telegram.org/botYOUR_TOKEN/getUpdates`
3. **Copy:** `chat` → `id`

### **UPDATE CONFIG:**

```powershell
# Edit: C:\openclaw\.env
TELEGRAM_BOT_TOKEN=YOUR_TOKEN
TELEGRAM_CHAT_ID=YOUR_CHAT_ID
```

---

## 🌐 **ACCESS:**

### **LOCAL NETWORK:**

```
http://192.168.1.100:8080   # OpenClaw
http://192.168.1.100:1810   # CLIProxyAPI
http://192.168.1.100:1809   # My-Router
```

### **VIA TUNNEL:**

```
https://ai-hub.YOURDOMAIN.com    # OpenClaw Web UI
https://api-hub.YOURDOMAIN.com   # API Access
https://router-hub.YOURDOMAIN.com # Router Admin
```

---

## 🔄 **TRIAL RESET:**

### **AUTO RESET:**

```
✅ Configured to reset every 14 days
✅ Scheduled task: Cursor-Trial-Reset
✅ Device Spoofer: C:\Tools\Device-Info-21AK22.exe
```

### **MANUAL RESET:**

```powershell
# When trial expires
cd C:\Cursor-Trial-Hub\PC-Phu-AI-Hub
.\auto-reset-pc-phu.ps1 -SpoofExe "C:\Tools\Device-Info-21AK22.exe"
```

---

## 📊 **SERVICES:**

| Service | Port | Auto-Start |
|---------|------|------------|
| **OpenClaw** | 8080 | ✅ Yes |
| **CLIProxyAPI** | 1810 | ✅ Yes |
| **My-Router** | 1809 | ✅ Yes |
| **Tunnel** | HTTPS | ✅ Yes |
| **Cursor Pro** | N/A | ⏰ Manual |

---

## 🔧 **TROUBLESHOOTING:**

### **Services not starting:**

```powershell
# Check status
Get-ScheduledTask -TaskName "OpenClaw-Startup"
Get-ScheduledTask -TaskName "Cursor-Trial-Reset"

# Manual start
cd C:\openclaw
npm start
```

### **Tunnel not working:**

```powershell
# Restart tunnel
.\start-tunnels.ps1

# Check logs
Get-Content "C:\Tools\tunnel.log" -Tail 50
```

---

## ✅ **SUCCESS CHECKLIST:**

```
✅ OpenClaw running on :8080
✅ CLIProxyAPI running on :1810
✅ My-Router running on :1809
✅ Tunnel configured
✅ Domain configured
✅ Telegram configured (optional)
✅ Auto-start working
✅ Auto-reset configured
✅ Can access from PC Chính
✅ Can access via tunnel
```

---

## 📞 **TELEGRAM NOTIFICATIONS:**

```powershell
# Test Telegram
cd C:\Cursor-Trial-Hub\Tools
.\send-telegram.ps1 -Message "AI Hub online!"
```

---

## 📝 **LOGS:**

```powershell
# OpenClaw
Get-Content "C:\openclaw\logs\openclaw.log" -Tail 50

# Tunnel
Get-Content "C:\Tools\tunnel.log" -Tail 50

# Auto-reset
Get-Content "C:\Tools\auto-reset.log" -Tail 50
```

---

## 🎯 **QUICK COMMANDS:**

```powershell
# Start all services
.\start-ai-hub.ps1

# Stop all services
Get-Process | Where-Object {$_.ProcessName -like '*openclaw*' -or $_.ProcessName -like '*CLIProxy*'} | Stop-Process -Force

# Check status
netstat -ano | findstr :8080
netstat -ano | findstr :1810
netstat -ano | findstr :1809
```

---

## 📦 **PACKAGE INFO:**

```
File: Cursor-Trial-Hub-FINAL.zip
Size: ~36 KB
Scripts: 12
Configs: Pre-configured
Documentation: Complete
```

---

## 🚀 **NEXT STEPS:**

1. ✅ Copy ZIP to PC Phụ
2. ✅ Extract to C:\Cursor-Trial-Hub
3. ✅ Run setup scripts
4. ✅ Configure domain
5. ✅ Configure Telegram (optional)
6. ✅ Start services
7. ✅ Enjoy unlimited Pro trial!

---

**LAST UPDATED:** 2026-03-27  
**VERSION:** Final  
**STATUS:** ✅ Production Ready

---

## 📧 **SUPPORT:**

**Files included:**
- ✅ Setup scripts
- ✅ Configuration files
- ✅ Documentation
- ✅ Telegram integration
- ✅ Tunnel setup
- ✅ Auto-reset system

**Everything you need in one ZIP!**
