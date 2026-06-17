# GRCS Setup Guide

**Location:** `Z:\projects\SJ-learning-lab\04_Projects\active\GRCS`

---

## 🚀 Quick Start

### 1. Setup Virtual Environment

```powershell
cd Z:\projects\SJ-learning-lab\04_Projects\active\GRCS

# Create venv (if not exists)
python -m venv venv

# Activate
.\venv\Scripts\Activate.ps1

# Install dependencies
pip install -r requirements.txt
```

### 2. Start CLIProxyAPI

```powershell
cd Z:\projects\CLIProxyAPI-main
.\start-cli-proxy.bat

# Verify running
curl http://localhost:1810/health
```

### 3. Generate Samples

```powershell
# Using provided script
.\scripts\grcs-generate.ps1 -Prompts 50 -K 3

# Or manually
python -m grcs generate --backend clproxy --k 3 --prompts 50
```

### 4. Label Samples

```powershell
# Auto-label with CLIProxyAPI
.\scripts\grcs-label.ps1

# Or manual UI
python -m grcs judge --port 5000
```

### 5. Build GRCS Map

```powershell
.\scripts\grcs-build.ps1
```

### 6. Test Engine

```powershell
.\scripts\grcs-test.ps1
```

---

## 📁 Folder Structure

```
GRCS/
├── grcs/                    # Source code
│   ├── generator.py         # CLIProxyGenerator added ✅
│   ├── engine.py
│   ├── builder.py
│   ├── judge_ui.py
│   └── ...
├── scripts/                 # PowerShell scripts
│   ├── grcs-generate.ps1   ✅
│   ├── grcs-label.ps1      ✅
│   ├── grcs-build.ps1      ✅
│   └── grcs-test.ps1       ✅
├── data/                    # Generated data
├── maps/                    # GRCS maps
├── docs/                    # Documentation
└── venv/                    # Virtual environment
```

---

## 🔧 Configuration

### CLIProxyAPI Settings

Edit `grcs/generator.py`:

```python
class CLIProxyGenerator(BaseGenerator):
    def __init__(
        self,
        base_url: str = "http://localhost:1810/v1",
        api_key: str = "sk-jarvis-dev",
        model: str = "qwen2.5-coder-7b",  # Change model here
        temperature: float = 0.7,
        max_tokens: int = 8192,
        # ...
    ):
```

### Script Parameters

```powershell
.\scripts\grcs-generate.ps1 `
  -Prompts 50 `
  -K 3 `
  -Model "qwen2.5-coder-7b" `
  -BaseUrl "http://localhost:1810/v1" `
  -ApiKey "sk-jarvis-dev"
```

---

## 📊 Usage Examples

### Generate UI Components

```powershell
# Generate 100 prompts × 3 completions = 300 samples
.\scripts\grcs-generate.ps1 -Prompts 100 -K 3
```

### Custom Model

```powershell
# Use different model from CLIProxyAPI
.\scripts\grcs-generate.ps1 -Model "gpt-4o-mini"
```

### Batch Test

```powershell
# Test with multiple prompts
.\scripts\grcs-test.ps1 -Map "maps/ui_expert.grcs" -K 6
```

---

## 🐛 Troubleshooting

### CLIProxyAPI Not Responding

```powershell
# Check if running
curl http://localhost:1810/health

# If not running, start it
cd Z:\projects\CLIProxyAPI-main
.\start-cli-proxy.bat
```

### Module Not Found

```powershell
# Reinstall dependencies
.\venv\Scripts\Activate.ps1
pip install -r requirements.txt --force-reinstall
```

### Generation Fails

```powershell
# Check CLIProxyAPI logs
Get-Content Z:\projects\CLIProxyAPI-main\logs\cli-proxy.log -Tail 50
```

---

## 📝 Next Steps

1. ✅ Generate samples (Phase 1)
2. ✅ Label samples (Phase 1)
3. ✅ Build GRCS map (Phase 2)
4. ✅ Test engine (Phase 3)
5. ⏳ Create production maps for different domains
6. ⏳ Integrate with Project Z agents

---

## 🔗 Related Docs

- [MASTERPLAN.md](./MASTERPLAN.md) - Full GRCS architecture
- [README.md](./README.md) - Project overview
- [AGENT-WORKFLOW.md](../00_Docs/AGENT-WORKFLOW.md) - Agent workflow guide

---

**Status:** ✅ Setup Complete  
**Last Updated:** 2026-03-31
