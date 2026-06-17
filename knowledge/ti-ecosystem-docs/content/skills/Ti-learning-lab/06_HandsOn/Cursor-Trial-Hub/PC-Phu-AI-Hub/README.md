# 🖥️ PC PHỤ - AI HUB SETUP

**Location:** `Z:\projects\SJ-learning-lab\04_Projects\Cursor-Trial-Hub\PC-Phu-AI-Hub\`

**Purpose:** Setup PC Phụ thành AI Hub với OpenClaw, CLIProxyAPI, My-Router, và Auto Reset

---

## 📁 **FILES TRONG FOLDER NÀY:**

| File | Purpose | When to Use |
|------|---------|-------------|
| `setup-pc-phu-ai-hub.ps1` | ⭐ **MAIN** - Setup toàn bộ AI Hub | First time setup |
| `auto-reset-pc-phu.ps1` | Auto reset trial với Device Spoofer | Khi hết trial |
| `setup-auto-reset-pc-phu.ps1` | Setup scheduled task auto reset | Optional - auto reset |
| `reset-cursordata.ps1` | Manual reset Cursor data | Quick manual reset |
| `start-cursor.ps1` | Launch Cursor | Daily use |
| `setup-scheduled-task.ps1` | Setup auto-start services | Auto-start on boot |

---

## 🚀 **SETUP PC PHỤ:**

### **BƯỚC 1: CHUẨN BỊ:**

```powershell
# 1. Copy folder này sang PC Phụ
xcopy "Z:\projects\SJ-learning-lab\04_Projects\Cursor-Trial-Hub\PC-Phu-AI-Hub" "C:\Cursor-Trial-Hub" /E /I

# 2. Download Device-Info-21AK22.exe
# Copy vào: C:\Tools\Device-Info-21AK22.exe
```

---

### **BƯỚC 2: CHẠY SETUP:**

```powershell
# Run as Administrator
cd C:\Cursor-Trial-Hub
.\setup-pc-phu-ai-hub.ps1
```

**Script sẽ:**
```
✅ Install OpenClaw (C:\openclaw)
✅ Setup CLIProxyAPI (Z:\Router\CLIProxyAPI-mainline)
✅ Setup My-Router (Z:\My-Router)
✅ Create Tools folder (C:\Tools)
✅ Copy auto-reset script
✅ Configure auto-start services
✅ Setup scheduled tasks
```

---

### **BƯỚC 3: KIỂM TRA:**

```powershell
# Check services running
Get-Process | Where-Object {$_.ProcessName -like '*cursor*' -or $_.ProcessName -like '*CLIProxy*'}

# Check scheduled tasks
Get-ScheduledTask -TaskName "AI-Hub-Startup"
Get-ScheduledTask -TaskName "Cursor-Trial-Reset"
```

---

### **BƯỚC 4: CẤU HÌNH NETWORK:**

```powershell
# Set static IP (example: 192.168.1.100)
# Control Panel → Network → Adapter Settings
# Properties → IPv4 → Use following IP:
#   IP: 192.168.1.100
#   Subnet: 255.255.255.0
#   Gateway: 192.168.1.1
#   DNS: 8.8.8.8
```

---

## 🔄 **RESET TRIAL KHI HẾT:**

### **MANUAL RESET:**

```powershell
cd C:\Cursor-Trial-Hub
.\auto-reset-pc-phu.ps1 -SpoofExe "C:\Tools\Device-Info-21AK22.exe"

# Follow instructions:
# 1. Run Device-Info-21AK22.exe
# 2. Change: BIOS Serial, MAC Address, Disk Serial, CPU ID
# 3. Restart PC
# 4. Open Cursor → Start NEW Pro Trial
```

### **AUTO RESET (OPTIONAL):**

```powershell
# Setup auto reset every 14 days
.\setup-auto-reset-pc-phu.ps1

# Check scheduled task
Get-ScheduledTask -TaskName "Cursor-Trial-Reset"
```

---

## 📊 **SERVICES RUNNING:**

| Service | Port | Status Check |
|---------|------|--------------|
| **OpenClaw** | 8080 | http://localhost:8080 |
| **CLIProxyAPI** | 1810 | http://localhost:1810/v1/models |
| **My-Router** | 1809 | http://localhost:1809/v1/models |
| **Cursor Pro** | N/A | GUI Application |

---

## 🔧 **TROUBLESHOOTING:**

### **Services not starting:**

```powershell
# Manual start
Start-Process "C:\openclaw\node_modules\.bin\npm" -ArgumentList "start" -WorkingDirectory "C:\openclaw"
Start-Process "Z:\Router\CLIProxyAPI-mainline\CLIProxyAPI.exe"
Start-Process "node" -ArgumentList "Z:\My-Router\index.js"
```

### **Trial reset not working:**

```powershell
# Check Device Spoofer
Test-Path "C:\Tools\Device-Info-21AK22.exe"

# Re-run reset script
.\auto-reset-pc-phu.ps1 -SpoofExe "C:\Tools\Device-Info-21AK22.exe" -Verbose
```

### **Network not accessible:**

```powershell
# Check firewall
netsh advfirewall firewall add rule name="AI Hub" dir=in action=allow protocol=TCP localport=1809,1810,8080

# Check IP config
ipconfig /all
```

---

## 📝 **LOGS:**

```powershell
# View auto-reset log
Get-Content "C:\Tools\auto-reset.log" -Tail 50

# View scheduled task history
Get-ScheduledTask -TaskName "Cursor-Trial-Reset" | Get-ScheduledTaskInfo
```

---

## ✅ **SUCCESS CHECKLIST:**

```
✅ OpenClaw running on :8080
✅ CLIProxyAPI running on :1810
✅ My-Router running on :1809
✅ Cursor Pro trial active
✅ Auto-reset scheduled task created
✅ Static IP configured
✅ Firewall rules added
```

---

## 🎯 **NEXT STEPS:**

1. ✅ Setup complete
2. ✅ Test from PC Chính
3. ✅ Configure auto-reset
4. ✅ Enjoy unlimited Pro trial!

---

**LAST UPDATED:** 2026-03-27  
**VERSION:** 1.0  
**STATUS:** ✅ Production Ready
