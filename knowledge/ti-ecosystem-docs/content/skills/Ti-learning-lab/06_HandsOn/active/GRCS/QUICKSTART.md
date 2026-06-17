# GRCS - Quick Start Guide

**TL;DR:** Generate UI components với GRCS + CLIProxyAPI

---

## ⚡ 5-Minute Setup

```powershell
# 1. Move to GRCS folder
cd Z:\projects\SJ-learning-lab\04_Projects\active\GRCS

# 2. Activate venv
.\venv\Scripts\Activate.ps1

# 3. Start CLIProxyAPI (new terminal)
cd Z:\projects\CLIProxyAPI-main
.\start-cli-proxy.bat

# 4. Generate samples
cd Z:\projects\SJ-learning-lab\04_Projects\active\GRCS
.\scripts\grcs-generate.ps1 -Prompts 10 -K 3

# 5. Label samples
.\scripts\grcs-label.ps1

# 6. Build map
.\scripts\grcs-build.ps1

# 7. Test
.\scripts\grcs-test.ps1
```

---

## 📊 Expected Output

```
=== GRCS Generation with CLIProxyAPI ===
Prompts: 10
K (completions per prompt): 3
Model: qwen2.5-coder-7b
Output: data/samples.jsonl

[1/2] Checking CLIProxyAPI status...
✓ CLIProxyAPI is running

[2/2] Generating samples...
✓ Created 10 prompts

Starting generation...
[CLIProxy] Dispatching 3 parallel requests (workers=3)...
  [+] Completion 1 received (2048 chars cleaned)
  [+] Completion 2 received (2156 chars cleaned)
  [+] Completion 3 received (1987 chars cleaned)

✓ Generation complete!
Total samples: 30
Output file: data/samples.jsonl

Next step: python -m grcs judge --port 5000
```

---

## 🎯 Commands Cheat Sheet

| Command | Purpose |
|---------|---------|
| `.\scripts\grcs-generate.ps1` | Generate samples |
| `.\scripts\grcs-label.ps1` | Auto-label samples |
| `.\scripts\grcs-build.ps1` | Build GRCS map |
| `.\scripts\grcs-test.ps1` | Test engine |
| `python -m grcs judge` | Manual labeling UI |

---

## 📝 Full documentation: [SETUP.md](./SETUP.md)
