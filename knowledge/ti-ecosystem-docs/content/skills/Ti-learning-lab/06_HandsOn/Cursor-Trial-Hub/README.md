# 🚀 CURSOR TRIAL HUB - COMPLETE SETUP

**Location:** `Z:\projects\SJ-learning-lab\04_Projects\Cursor-Trial-Hub`

**Purpose:** Unlimited Cursor Pro Trial với PC Phụ + Auto Reset

---

## 📁 **FILES:**

| File | Purpose | When to Use |
|------|---------|-------------|
| `setup-pc-phu-ai-hub.ps1` | ⭐ **MAIN** - Setup PC Phụ as AI Hub | First time setup |
| `auto-reset-pc-phu.ps1` | Auto reset trial với Device Spoofer | When trial expires |
| `setup-auto-reset-pc-phu.ps1` | Setup scheduled task for auto reset | Optional - auto reset |
| `reset-cursordata.ps1` | Manual reset Cursor data | Quick manual reset |
| `start-cursor.ps1` | Launch Cursor | Daily use |
| `setup-cursor-trial-vm.ps1` | Create Hyper-V VM (alternative) | If not using PC Phụ |
| `reset-vm.ps1` | Reset Hyper-V VM | VM trial reset |
| `setup-vm-scheduled-task.ps1` | VM auto reset schedule | VM auto reset |
| `setup-scheduled-task.ps1` | Legacy scheduled task | Legacy use |
| `simple-reset.ps1` | Simple manual reset | Quick test |
| `fix-script.ps1` | Fix script syntax errors | Debugging |
| `fix-all.ps1` | Fix all script errors | Debugging |
| `CursorHelper.ps1` | Original helper script | Legacy |
| `reset-z-cursordata.ps1` | Reset Z:\CursorData | Legacy |

---

## 🏆 **RECOMMENDED WORKFLOW:**

### **PC PHỤ (AI Hub):**

```powershell
# 1. First time setup
.\setup-pc-phu-ai-hub.ps1

# This will:
# - Install OpenClaw
# - Setup CLIProxyAPI
# - Setup My-Router
# - Configure auto-start
# - Copy Device Spoofer
# - Create Tools folder

# 2. After setup, services auto-start on boot
# 3. When trial expires (14 days):
.\auto-reset-pc-phu.ps1 -SpoofExe "C:\Tools\Device-Info-21AK22.exe"
```

---

### **PC CHÍNH (Development):**

```powershell
# Setup environment to connect to PC Phụ
$env:ANTHROPIC_BASE_URL = "http://192.168.1.100:1810"
$env:ANTHROPIC_AUTH_TOKEN = "sk-jarvis-dev"

# Add to $PROFILE for persistence:
Add-Content $PROFILE @"
`$env:ANTHROPIC_BASE_URL = "http://192.168.1.100:1810"
`$env:ANTHROPIC_AUTH_TOKEN = "sk-jarvis-dev"
"@

# Use AI services
claude "Hello from PC chinh!"
```

---

## 📋 **NETWORK CONFIG:**

```
PC Chính (192.168.1.x)
      ↕
PC Phụ (192.168.1.100) - AI Hub
  ├── OpenClaw:      http://192.168.1.100:8080
  ├── CLIProxyAPI:   http://192.168.1.100:1810
  └── My-Router:     http://192.168.1.100:1809
```

---

## 🔄 **TRIAL RESET CYCLE:**

```
Day 0: Install Cursor → Start Pro Trial
Day 1-14: Use Pro features
Day 14: Trial expires
        → Run: .\auto-reset-pc-phu.ps1
        → Device Spoofer changes HWID
        → Restart PC
        → Start NEW Pro Trial
        → Repeat from Day 1
```

---

## ⚡ **QUICK START:**

### **Option 1: PC Phụ (Recommended)**

```powershell
# On PC Phụ
cd Z:\projects\SJ-learning-lab\04_Projects\Cursor-Trial-Hub
.\setup-pc-phu-ai-hub.ps1

# After setup, services auto-start
# Reset when trial expires:
.\auto-reset-pc-phu.ps1 -SpoofExe "C:\Tools\Device-Info-21AK22.exe"
```

### **Option 2: Hyper-V VM (Alternative)**

```powershell
# On PC Chính
cd Z:\projects\SJ-learning-lab\04_Projects\Cursor-Trial-Hub
.\setup-cursor-trial-vm.ps1 -ISOPath "C:\ISOs\Windows11.iso"

# After VM setup:
.\reset-vm.ps1 -HyperV -VMName "Cursor-Trial" -SnapshotName "Fresh-Trial"
```

---

## 🛠️ **REQUIREMENTS:**

### **PC Phụ:**

```
✅ Windows 10/11
✅ Hyper-V enabled (for VM option)
✅ Device-Info-21AK22.exe
✅ Network connection
✅ 60GB free disk
```

### **PC Chính:**

```
✅ Any Windows
✅ Network access to PC Phụ
✅ PowerShell
```

---

## 📊 **COMPARISON:**

| Method | PC Phụ | Hyper-V VM |
|--------|--------|------------|
| **Performance** | ⭐⭐⭐⭐⭐ Native | ⭐⭐⭐ VM overhead |
| **Setup Time** | ⭐⭐⭐ 30min | ⭐⭐ 1hr |
| **Reset Speed** | ⭐⭐⭐⭐ Fast | ⭐⭐⭐ Medium |
| **Resource** | ⭐⭐⭐⭐ Light | ⭐⭐ Heavy |
| **Recommended** | ✅ YES | ⚠️ Alternative |

---

## 🎯 **TROUBLESHOOTING:**

### **Issue: Services not starting**

```powershell
# Manual start
Start-Process "Z:\Router\CLIProxyAPI-mainline\CLIProxyAPI.exe"
Start-Process "node" -ArgumentList "Z:\My-Router\index.js"
Start-Process "npm" -ArgumentList "start" -WorkingDirectory "C:\openclaw"
```

### **Issue: Device Spoofer not found**

```powershell
# Update path in auto-reset script
.\auto-reset-pc-phu.ps1 -SpoofExe "C:\Your\Path\Device-Info-21AK22.exe"
```

### **Issue: Network not connecting**

```powershell
# Check firewall
# Allow ports: 1809, 1810, 8080
netsh advfirewall firewall add rule name="AI Hub" dir=in action=allow protocol=TCP localport=1809,1810,8080
```

---

## 📝 **LOGS & MONITORING:**

```powershell
# Check scheduled tasks
Get-ScheduledTask -TaskName "AI-Hub-Startup"
Get-ScheduledTask -TaskName "Cursor-Trial-Reset"

# View logs
Get-Content "C:\Tools\auto-reset.log" -Tail 50
```

---

## ✅ **SUCCESS CRITERIA:**

```
✅ PC Phụ: Services running
✅ PC Chính: Can connect to AI Hub
✅ Cursor: Pro trial active
✅ Auto-reset: Scheduled every 14 days
✅ Network: All ports accessible
```

---

## 🚀 **NEXT STEPS:**

1. ✅ Run `setup-pc-phu-ai-hub.ps1` on PC Phụ
2. ✅ Wait for OpenClaw install to complete
3. ✅ Configure network (static IP)
4. ✅ Test connection from PC Chính
5. ✅ Setup auto-reset scheduled task
6. ✅ Enjoy unlimited Pro trial!

---

**LAST UPDATED:** 2026-03-27  
**VERSION:** 1.0  
**STATUS:** ✅ Production Ready
