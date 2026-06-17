# Repository Index - Ti-Learning-Lab

> **Index các repositories đã clone để agents research và học tập**

---

## 📚 Repositories Đã Clone

### RTK (Rust Token Killer)
- **Path**: `../05_Repositories/rtk/`
- **Source**: https://github.com/rtk-ai/rtk
- **License**: MIT
- **Version**: 0.37.2
- **Language**: Rust
- **Mục đích**: CLI proxy giảm token consumption 60-90% cho LLM
- **Status**: ✅ Cloned
- **Docs quan trọng**:
  - `docs/contributing/ARCHITECTURE.md` - System design, module organization
  - `docs/contributing/TECHNICAL.md` - End-to-end flow, hook system
  - `CLAUDE.md` - Guidance cho Claude Code agents
- **Architecture**:
  - Command proxy với Clap CLI
  - Filter modules trong `src/cmds/*/`
  - Core logic trong `src/core/`
  - Filtering strategies trong `src/filters/`
  - SQLite tracking trong `src/core/tracking.rs`
- **Use Cases cho Ti**:
  - Research: Học filtering strategies để áp dụng cho Ti Router
  - Integration: Tích hợp RTK wrapper để giảm tokens cho agent tool outputs
  - Patterns: Command proxy pattern, filter pipeline pattern
- **Key Files**:
  - `src/main.rs` - Entry point, CLI routing
  - `src/cmds/git/` - Git filter modules
  - `src/cmds/cargo/` - Cargo filter modules
  - `src/core/filter.rs` - Core filtering logic
  - `src/hooks/` - Auto-rewrite hook system

---

## 🗂️ Các Folder Khác

### router/
Router-related repositories

### tools/
CLI tools & utilities
- `antigravity-resources/`
- `websocket-proxy-logger/`

### services/
Service-related repositories

---

## 📋 Workflow Clone Repo Mới

### 1. Clone repo
```bash
cd /z/Ti/Ti-learning-lab/05_Repositories
git clone https://github.com/user/repo.git
```

### 2. Categorize (nếu cần)
```bash
mkdir -p category-name
mv repo category-name/
```

### 3. Update INDEX.md
- Thêm entry với đầy đủ thông tin
- Document docs quan trọng
- Note architecture/use cases

### 4. Update .devin context
```bash
# Update .devin/knowledge/repo-index.md
# Update .devin/context/repo-context.md (nếu cần)
```

---

## 🎯 Mục Đích Research

### Cho Ti Router:
- **RTK**: Học filtering strategies để giảm tokens
- **Router repos**: Học architecture patterns
- **Tools repos**: Học CLI tool patterns

### Cho Agent Development:
- **Architecture patterns**: Học cách tổ chức code
- **Filter strategies**: Học cách optimize outputs
- **Hook systems**: Học cách intercept/modify behavior

---

## 📝 Notes

- **Research repos**: Dùng để agents học architecture, patterns
- **Integration repos**: Dùng để tích hợp vào projects active
- **Archive repos**: Lưu trữ để reference sau này

---

## 🔄 Sync Với Index Chính

```bash
# Sync từ 05_Repositories/INDEX.md
cp ../05_Repositories/INDEX.md ./
```

---

**Last Updated**: 2026-04-28
