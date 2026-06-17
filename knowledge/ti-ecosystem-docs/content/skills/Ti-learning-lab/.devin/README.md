# Ti-Learning-Lab - Devin Memory

> **Local memory/context cho Ti-Learning-Lab project**  
> **Purpose**: Agents có context về dự án, repositories, patterns, và workflow

---

## 📁 Cấu Trúc

```
.devin/
├── README.md                # File này
├── context/                 # Context files
│   ├── project-context.md   # Context về dự án Ti-Learning-Lab
│   ├── repo-context.md      # Context về từng repo
│   └── task-context.md      # Context về tasks hiện tại
├── knowledge/                # Knowledge base
│   ├── repo-index.md         # Index các repositories
│   ├── patterns.md           # Design patterns đã học
│   └── lessons.md            # Lessons learned
├── memory/                   # Memory storage
│   ├── session-memory.md     # Session memory (tạm thời)
│   └── long-term-memory.md   # Long-term memory (lưu trữ)
├── skills/                   # Agent skills
│   └── [skill-name]/         # Các skills cho agents
└── config/                   # Configuration
    └── settings.json         # Cấu hình agents
```

---

## 🎯 Mục Đích

### 1. **Context** (context/)
Agents hiểu về:
- Dự án là gì (Ti-Learning-Lab)
- Cấu trúc project
- Workflow chuẩn
- Quy tắc coding

### 2. **Knowledge** (knowledge/)
Agents học từ:
- Repositories đã clone (rtk, router, tools, etc.)
- Design patterns
- Lessons learned từ các projects

### 3. **Memory** (memory/)
Agents nhớ:
- Session hiện tại (task đang làm)
- Long-term knowledge (lessons, patterns)

### 4. **Skills** (skills/)
Agents có khả năng:
- Clone repo mới
- Analyze architecture
- Extract patterns
- Generate documentation

---

## 📋 Workflow Cho Agents

### Khi làm việc với Ti-Learning-Lab:

1. **Load context**:
   ```
   Đọc: .devin/context/project-context.md
   Đọc: .devin/knowledge/repo-index.md
   ```

2. **Execute task**:
   ```
   Clone repo → Analyze → Extract patterns → Document
   ```

3. **Update knowledge**:
   ```
   Lưu patterns mới vào .devin/knowledge/patterns.md
   Lưu lessons vào .devin/knowledge/lessons.md
   ```

4. **Update memory**:
   ```
   Lưu session progress vào .devin/memory/session-memory.md
   ```

---

## 🔗 Liên Kết Với Ti Brain (Hybrid Approach)

**Hybrid approach**: Local .devin memory + Ti Brain centralized

**Sync script**: `.devin/scripts/sync-brain.sh`

**Usage**:
```bash
# Push local knowledge to Ti Brain
./sync-brain.sh push

# Pull system-wide patterns from Ti Brain
./sync-brain.sh pull

# Full sync (bi-directional)
./sync-brain.sh full

# Check sync status
./sync-brain.sh status
```

**Lợi ích**:
- ✅ Local memory cho speed
- ✅ Central brain cho sharing
- ✅ Context-specific knowledge
- ✅ System-wide patterns
- ✅ Redundancy

**Chi tiết**: Xem `.devin/HYBRID_ARCHITECTURE.md`

**Lưu ý**: Cần rsync để sync script hoạt động. Nếu không có rsync, có thể:
- Cài đặt rsync (Linux/Mac: package manager, Windows: WSL hoặc cygwin)
- Hoặc copy manual (không tự động)

---

## 📝 Quy Tắc

1. **Luôn đọc context trước khi làm task**
2. **Update knowledge sau khi học được pattern mới**
3. **Document lessons learned** sau mỗi project
4. **Giữ context updated** khi project thay đổi

---

**Last Updated**: 2026-04-28
