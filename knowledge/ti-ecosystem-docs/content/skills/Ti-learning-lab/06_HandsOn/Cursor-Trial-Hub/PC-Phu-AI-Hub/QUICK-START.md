# 🚀 QUICK START - CURSOR TRIAL HUB

**Location:** `C:\Cursor-Trial-Hub` (on PC Phụ)

---

## ⚡ **5 PHÚT SETUP:**

### **BƯỚC 1: CHẠY SETUP (1 phút)**

```powershell
# Run as Administrator
cd C:\Cursor-Trial-Hub
.\setup-pc-phu-ai-hub.ps1
```

**Script sẽ:**
```
✅ Install OpenClaw
✅ Setup CLIProxyAPI
✅ Setup My-Router
✅ Configure auto-start
✅ Create Tools folder
```

---

### **BƯỚC 2: CẤU HÌNH OPENCLAW (1 phút)**

```powershell
# Run configuration
.\configure-openclaw.ps1
```

**Script sẽ:**
```
✅ Create optimized .env
✅ Setup logs directory
✅ Create startup script
✅ Configure auto-start
```

---

### **BƯỚC 3: SETUP TUNNEL (2 phút)**

```powershell
# Run tunnel setup
.\setup-tunnel-simple.ps1
```

**Chọn 1 trong 3:**

**A. Cloudflare Tunnel (Recommended):**
```
1. https://dash.teams.cloudflare.com
2. Zero Trust → Access → Tunnels
3. Create tunnel → Web
4. Add hostnames:
   - ai-hub.YOURDOMAIN.com → localhost:8080
   - api-hub.YOURDOMAIN.com → localhost:1810
   - router-hub.YOURDOMAIN.com → localhost:1809
```

**B. Ngrok:**
```
1. Get token: https://dashboard.ngrok.com
2. ngrok config add-authtoken YOUR_TOKEN
3. ngrok http 8080 --subdomain ai-hub
4. ngrok http 1810 --subdomain api-hub
5. ngrok http 1809 --subdomain router-hub
```

**C. LocalXpose:**
```
1. Download: https://localxpose.io
2. loclx tunnel http --to localhost:8080
3. loclx tunnel http --to localhost:1810
4. loclx tunnel http --to localhost:1809
```

---

### **BƯỚC 4: START SERVICES (1 phút)**

```powershell
# Start all services
.\start-ai-hub.ps1

# Or start individually:
.\start-cursor.ps1
.\start-tunnels.ps1
```

---

## ✅ **VERIFICATION:**

### **CHECK SERVICES:**

```powershell
# OpenClaw
Invoke-RestMethod http://localhost:8080

# CLIProxyAPI
Invoke-RestMethod http://localhost:1810/v1/models

# My-Router
Invoke-RestMethod http://localhost:1809/v1/models
```

### **CHECK TUNNEL:**

```powershell
# Check tunnel status
Get-ScheduledTask -TaskName "OpenClaw-Startup"
Get-ScheduledTask -TaskName "Cursor-Trial-Reset"
```

---

## 🌐 **ACCESS FROM ANYWHERE:**

### **LOCAL NETWORK:**

```
http://192.168.1.100:8080   # OpenClaw
http://192.168.1.100:1810   # CLIProxyAPI
http://192.168.1.100:1809   # My-Router
```

### **VIA TUNNEL:**

```
https://ai-hub.YOURDOMAIN.com    # OpenClaw
https://api-hub.YOURDOMAIN.com   # CLIProxyAPI
https://router-hub.YOURDOMAIN.com # My-Router
```

---

## 🔄 **TRIAL RESET:**

### **MANUAL RESET:**

```powershell
# When Cursor Pro trial expires (after 14 days)
.\auto-reset-pc-phu.ps1 -SpoofExe "C:\Tools\Device-Info-21AK22.exe"

# Follow instructions:
# 1. Run Device Spoofer
# 2. Change hardware IDs
# 3. Restart PC
# 4. Start new Pro Trial
```

### **AUTO RESET:**

```powershell
# Already configured to reset every 14 days
# Check scheduled task:
Get-ScheduledTask -TaskName "Cursor-Trial-Reset"
```

---

## 📊 **SERVICES RUNNING:**

| Service | Port | Status |
|---------|------|--------|
| **OpenClaw** | 8080 | ✅ Auto-start |
| **CLIProxyAPI** | 1810 | ✅ Auto-start |
| **My-Router** | 1809 | ✅ Auto-start |
| **Cursor Pro** | N/A | ⏰ Manual trial |
| **Tunnel** | HTTPS | ✅ Configured |

---

## 🔧 **TROUBLESHOOTING:**

### **Services not starting:**

```powershell
# Manual start
cd C:\openclaw
npm start

# Check CLIProxyAPI
Z:\Router\CLIProxyAPI-mainline\CLIProxyAPI.exe

# Check My-Router
cd Z:\My-Router
npm start
```

### **Tunnel not working:**

```powershell
# Check tunnel status
.\start-tunnels.ps1 -Verbose

# Check logs
Get-Content "C:\Tools\tunnel.log" -Tail 50
```

### **Trial reset not working:**

```powershell
# Re-run reset script
.\auto-reset-pc-phu.ps1 -SpoofExe "C:\Tools\Device-Info-21AK22.exe" -Verbose

# Check Device Spoofer
Test-Path "C:\Tools\Device-Info-21AK22.exe"
```

---

## 📝 **LOGS:**

```powershell
# OpenClaw logs
Get-Content "C:\openclaw\logs\openclaw.log" -Tail 50

# Tunnel logs
Get-Content "C:\Tools\tunnel.log" -Tail 50

# Auto-reset logs
Get-Content "C:\Tools\auto-reset.log" -Tail 50
```

---

## 🎯 **DAILY USE:**

### **FROM PC CHÍNH:**

```powershell
# Local network
$env:ANTHROPIC_BASE_URL = "http://192.168.1.100:1810"
claude "Hello!"

# Via tunnel
$env:ANTHROPIC_BASE_URL = "https://api-hub.YOURDOMAIN.com"
claude "Hello from anywhere!"
```

### **FROM MOBILE/REMOTE:**

```
Just use tunnel URLs:
https://ai-hub.YOURDOMAIN.com
```

---

## ✅ **SUCCESS CHECKLIST:**

```
✅ OpenClaw running on :8080
✅ CLIProxyAPI running on :1810
✅ My-Router running on :1809
✅ Tunnel configured
✅ Auto-start configured
✅ Auto-reset configured (14 days)
✅ Device Spoofer ready
✅ Can access from PC Chính
✅ Can access via tunnel
```

---

## 🚀 **NEXT:**

1. ✅ Setup complete
2. ✅ Test from PC Chính
3. ✅ Configure tunnel domain
4. ✅ Enjoy unlimited Pro trial!

---

**LAST UPDATED:** 2026-03-27  
**VERSION:** 1.0  
**STATUS:** ✅ Production Ready
