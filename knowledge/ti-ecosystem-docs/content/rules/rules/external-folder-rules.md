# External Folder Rules

> **Category**: System Rules
> **Priority**: P0 (BẮT BUỘC)
> **Last Updated**: 2026-05-05

---

## 📁 External Folder Structure

```
Z:\02_CORE\_cli\
├── bin\                    # CLI executables (CHỈ tải về đây)
├── .config\                # Config files (single source of truth)
└── ...

Z:\03_DATA\
├── local\                  # .local folder redirected
├── appdata\                # AppData folder redirected
└── ...

C:\Users\MIN\               # User home folder
├── junctions\              # Junctions to CLI tools
├── .local (junction)       # → Z:\03_DATA\local
└── AppData (junction)      # → Z:\03_DATA\appdata
```

---

## 🚫 CLI Installation Rules (BẮT BUỘC)

### Rule 1: CLI Location
**BẮT BUỘC** - Chỉ tải CLI về `Z:\02_CORE\_cli\bin\`

**Yêu cầu:**
|- ✅ Tải CLI về `Z:\02_CORE\_cli\bin\`
|- ❌ KHÔNG tải về vị trí khác (C:\Windows\System32, C:\Program Files, etc.)
|- ❌ KHÔNG tải về project-specific folders

**Reason:**
|- Centralized CLI management
|- Easy update/maintenance
|- Consistent version control
|- Avoids version conflicts

### Rule 2: Config Location
**BẮT BUỘC** - Config files chỉ ở `Z:\02_CORE\_cli\.config\`

**Yêu cầu:**
|- ✅ Config files ở `Z:\02_CORE\_cli\.config\`
|- ❌ KHÔNG config ở các vị trí khác (C:\Users\MIN\.config, project folders)
|- ✅ Single source of truth cho config

**Config Files:**
|- CLI config files (.toml, .yaml, .json)
|- Environment config (.env)
|- Authentication config (tokens, API keys)

### Rule 3: Environment Variable/Junction Setup
**BẮT BUỘC** - Đặt biến môi trường hoặc junction ở `C:\Users\MIN\`

**Yêu cầu:**
|- ✅ Đặt biến môi trường PATH để trỏ đến `Z:\02_CORE\_cli\bin\`
|- ✅ Hoặc tạo junctions ở `C:\Users\MIN\junctions\` trỏ đến `Z:\02_CORE\_cli\bin\`
|- ✅ Để sử dụng CLI ở nhiều vị trí khác

**Option 1: Environment Variable**
```powershell
# Add to PATH
$env:PATH += ";Z:\02_CORE\_cli\bin"

# Permanent (add to User Environment Variables)
[Environment]::SetEnvironmentVariable("PATH", $env:PATH + ";Z:\02_CORE\_cli\bin", "User")
```

**Option 2: Junction**
```powershell
# Create junction directory
New-Item -ItemType Junction -Path "C:\Users\MIN\junctions\cli" -Target "Z:\02_CORE\_cli\bin"

# Create specific junctions
New-Item -ItemType Junction -Path "C:\Users\MIN\junctions\bd" -Target "Z:\02_CORE\_cli\bin\bd.exe"
New-Item -ItemType Junction -Path "C:\Users\MIN\junctions\ti" -Target "Z:\02_CORE\_cli\bin\ti.exe"
```

**When to use:**
|- **Environment Variable**: Cho tất cả CLI, đơn giản setup
|- **Junction**: Cho CLI cụ thể, cần truy cập từ nhiều vị trí

---

## 🔄 System Folder Redirection (BẮT BUỘC)

### Rule 4: .local Folder Redirection
**BẮT BUỘC** - Redirect `C:\Users\MIN\.local` → external drive

**Yêu cầu:**
- ✅ Tạo junction/symlink từ `C:\Users\MIN\.local` → `Z:\03_DATA\local` hoặc `Z:\02_CORE\.local`
- ✅ Hoặc dùng environment variable để redirect
- ✅ Để tối ưu hóa storage trên ổ C

**Option 1: Junction**
```powershell
# Backup existing .local folder
robocopy C:\Users\MIN\.local Z:\03_DATA\local /E /COPYALL

# Remove original .local folder
Remove-Item C:\Users\MIN\.local -Recurse -Force

# Create junction
New-Item -ItemType Junction -Path "C:\Users\MIN\.local" -Target "Z:\03_DATA\local"
```

**Option 2: Environment Variable**
```powershell
# Set XDG_DATA_HOME
[Environment]::SetEnvironmentVariable("XDG_DATA_HOME", "Z:\03_DATA\local", "User")
```

### Rule 5: AppData Folder Redirection
**BẮT BUỘC** - Redirect `C:\Users\MIN\AppData` → external drive

**Yêu cầu:**
- ✅ Tạo junction/symlink từ `C:\Users\MIN\AppData` → `Z:\03_DATA\appdata` hoặc `Z:\02_CORE\appdata`
- ✅ Hoặc redirect từng subfolder (Roaming, Local, LocalLow)
- ✅ Để tối ưu hóa storage trên ổ C

**⚠️ WARNING**: AppData redirect có rủi ro cao, cần backup trước

**Option 1: Full Junction (High Risk)**
```powershell
# Backup existing AppData
robocopy C:\Users\MIN\AppData Z:\03_DATA\appdata /E /COPYALL /R:0 /W:0

# Remove original AppData (DANGEROUS - cần reboot vào safe mode)
# Remove-Item C:\Users\MIN\AppData -Recurse -Force

# Create junction (cần admin privileges)
# New-Item -ItemType Junction -Path "C:\Users\MIN\AppData" -Target "Z:\03_DATA\appdata"
```

**Option 2: Subfolder Junction (Recommended)**
```powershell
# Redirect AppData\Local
robocopy C:\Users\MIN\AppData\Local Z:\03_DATA\appdata\Local /E /COPYALL
Remove-Item C:\Users\MIN\AppData\Local -Recurse -Force
New-Item -ItemType Junction -Path "C:\Users\MIN\AppData\Local" -Target "Z:\03_DATA\appdata\Local"

# Redirect AppData\Roaming (chọn lọc)
robocopy C:\Users\MIN\AppData\Roaming Z:\03_DATA\appdata\Roaming /E /COPYALL
Remove-Item C:\Users\MIN\AppData\Roaming -Recurse -Force
New-Item -ItemType Junction -Path "C:\Users\MIN\AppData\Roaming" -Target "Z:\03_DATA\appdata\Roaming"
```

**Option 3: Environment Variable (Safe)**
```powershell
# Redirect specific app data
[Environment]::SetEnvironmentVariable("LOCALAPPDATA", "Z:\03_DATA\appdata\Local", "User")
[Environment]::SetEnvironmentVariable("APPDATA", "Z:\03_DATA\appdata\Roaming", "User")
```

### Rule 6: CLI Cache Redirection (Recommended)
**KHUYẾN NGHỊ** - Redirect CLI cache về `Z:\02_CORE\_cli\`

**Yêu cầu:**
- ✅ Tạo cache folders ở `Z:\02_CORE\_cli\`
- ✅ Đặt biến môi trường để CLI dùng folder này
- ✅ Để tối ưu hóa storage trên ổ C

**Cache Folders:**
```
Z:\02_CORE\_cli\
├── .npm-cache              # npm cache
├── .cargo                  # Cargo cache
├── .pip-cache              # pip cache
├── .go-cache               # Go cache
├── .yarn-cache             # Yarn cache
└── .pnpm-store             # pnpm store
```

**Environment Variables:**
```powershell
# npm
[Environment]::SetEnvironmentVariable("npm_config_cache", "Z:\02_CORE\_cli\.npm-cache", "User")

# Cargo
[Environment]::SetEnvironmentVariable("CARGO_HOME", "Z:\02_CORE\_cli\.cargo", "User")

# pip
[Environment]::SetEnvironmentVariable("PIP_CACHE_DIR", "Z:\02_CORE\_cli\.pip-cache", "User")

# Go
[Environment]::SetEnvironmentVariable("GOCACHE", "Z:\02_CORE\_cli\.go-cache", "User")

# Yarn
[Environment]::SetEnvironmentVariable("YARN_CACHE_FOLDER", "Z:\02_CORE\_cli\.yarn-cache", "User")

# pnpm
[Environment]::SetEnvironmentVariable("PNPM_STORE_DIR", "Z:\02_CORE\_cli\.pnpm-store", "User")
```

**Priority**
- **P0 (Cần thiết)**: .local redirection
- **P1 (Nên có)**: CLI cache redirection (npm, cargo, pip, go)
- **P2 (Cẩn trọng)**: AppData\Local redirection (có thể break apps)
- **P3 (Rất cẩn trọng)**: AppData\Roaming redirection (có thể break apps)

### Backup Before Redirection
```powershell
# Backup toàn bộ
robocopy C:\Users\MIN\.local Z:\BACKUP\.local /E /COPYALL
robocopy C:\Users\MIN\AppData Z:\BACKUP\AppData /E /COPYALL
```

---

## 📋 CLI Installation Checklist

Trước khi tải CLI mới:

- [ ] Check CLI đã tồn tại ở `Z:\02_CORE\_cli\bin\` chưa
- [ ] Nếu chưa, tải về `Z:\02_CORE\_cli\bin\`
- [ ] Config file đặt ở `Z:\02_CORE\_cli\.config\`
- [ ] Update PATH hoặc tạo junction
- [ ] Test CLI từ nhiều vị trí khác
- [ ] Log installation với BD tool

---

## 📋 System Folder Redirection Checklist

Trước khi redirect system folder:

- [ ] Backup folder gốc (C:\Users\MIN\.local, C:\Users\MIN\AppData)
- [ ] Create target folder trên external drive (Z:\03_DATA\local, Z:\03_DATA\appdata)
- [ ] Copy data sang target folder
- [ ] Test junction/environment variable
- [ ] Verify apps hoạt động sau redirect
- [ ] Log với BD tool

---

## 🔍 Enforcement

### Violation Detection
- Tải CLI về vị trí khác → Log violation
- Config file ở vị trí khác → Log violation
- Không setup PATH/junction → Log violation
- Redirect system folder không backup → Log violation

### Violation Logging
```bash
bd log --task="External folder rule violation: {violation_type} - {details}" \
       --agent=devin \
       --status=failed \
       --type=violation \
       --domain=cli
```

### Escalation
- Repeated violations (≥3 times) → Escalate to human review
- Critical violations (tải CLI về system folders) → Immediate escalation
- Critical violations (redirect AppData không backup) → Immediate escalation

---

## 📚 Reference

|- **CLI Location**: `Z:\02_CORE\_cli\bin\`
|- **Config Location**: `Z:\02_CORE\_cli\.config\`
|- **User Home**: `C:\Users\MIN\`
|- **Junction Location**: `C:\Users\MIN\junctions\`
|- **.local Redirect**: `C:\Users\MIN\.local` → `Z:\03_DATA\local`
|- **AppData Redirect**: `C:\Users\MIN\AppData` → `Z:\03_DATA\appdata`
