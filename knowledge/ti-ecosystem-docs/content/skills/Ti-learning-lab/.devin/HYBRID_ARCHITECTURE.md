# Hybrid Memory Architecture Proposal

> **Local .devin memory + Ti Brain centralized - Hybrid approach**

---

## 🎯 Mục Tiêu

Kết hợp **local .devin memory** (nhanh, context-specific) với **Ti Brain centralized** (chia sẻ, system-wide) để có:
- Speed của local memory
- Sharing của centralized brain
- Context-specific knowledge
- System-wide patterns

---

## 📊 Architecture So Sánh

### Local .devin Only:
**Ưu điểm:**
- ✅ Nhanh (local access)
- ✅ Context-specific cho từng project
- ✅ Không phụ thuộc external system
- ✅ Dễ quản lý per-project

**Nhược điểm:**
- ❌ Không chia sẻ knowledge giữa projects
- ❌ Phải duplicate patterns
- ❌ Không có system-wide view

### Ti Brain Only:
**Ưu điểm:**
- ✅ Centralized knowledge
- ✅ Chia sẻ giữa projects
- ✅ System-wide view
- ✅ Smart Auto-Update

**Nhược điểm:**
- ❌ Chậm hơn (query centralized system)
- ❌ Không context-specific
- ❌ Phụ thuộc external system

### Hybrid (Both):
**Ưu điểm:**
- ✅ Nhanh (local access)
- ✅ Context-specific
- ✅ Chia sẻ (sync với Ti Brain)
- ✅ System-wide view
- ✅ Redundancy (2 sources)

**Nhược điểm:**
- ⚠️ Cần sync mechanism
- ⚠️ Complexity cao hơn
- ⚠️ Cần maintain consistency

---

## 🏗️ Hybrid Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Ti Brain (Centralized)                     │
│  Z:\Ti\brain\                                                     │
│  ├── system-overview.md                                         │
│  ├── workflow-automation.md                                     │
│  ├── memory/indexes/ (beads, tasks, lessons)                 │
│  ├── providers/                                                 │
│  ├── projects/                                                 │
│  └── protocols/                                                │
└──────────────────────┬──────────────────────────────────────┘
                       │ Sync (bi-directional)
┌──────────────────────▼──────────────────────────────────────┐
│              .devin (Local - Ti-Learning-Lab)                 │
│  Z:\Ti\Ti-learning-lab\.devin\                                  │
│  ├── context/project-context.md                                   │
│  ├── knowledge/repo-index.md                                    │
│  ├── knowledge/patterns.md                                       │
│  ├── knowledge/lessons.md                                        │
│  └── memory/session-memory.md                                   │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔄 Sync Strategy

### 1. **Push Sync** (.devin → Ti Brain)
Khi học được pattern/lesson mới:

```bash
# 1. Agent học pattern từ repo
# 2. Document vào .devin/knowledge/patterns.md

# 3. Sync lên Ti Brain
rsync -av .devin/knowledge/patterns.md Z:\Ti\brain\memory\lessons\patterns-ti-learning-lab.md
rsync -av .devin/knowledge/lessons.md Z:\Ti\brain\memory\lessons\lessons-ti-learning-lab.md
```

### 2. **Pull Sync** (Ti Brain → .devin)
Khi cần system-wide patterns:

```bash
# 1. Pull từ Ti Brain
rsync -av Z:\Ti\brain\memory\lessons\lessons-index.md .devin/knowledge/

# 2. Merge với local knowledge
# (manual review và merge)
```

### 3. **Bi-directional Sync** (Full Sync)
Đồng bộ cả 2 chiều:

```bash
# 1. Push local changes lên Ti Brain
rsync -av .devin/knowledge/ Z:\Ti\brain\memory\lessons\ti-learning-lab/

# 2. Pull system-wide patterns từ Ti Brain
rsync -av Z:\Ti\brain\memory\lessons\ .devin/knowledge/

# 3. Resolve conflicts (manual)
```

---

## 📋 Sync Scenarios

### Scenario 1: Learn New Pattern from RTK
```
1. Agent research RTK
2. Extract pattern (Command Proxy Pattern)
3. Document vào .devin/knowledge/patterns.md
4. Push sync lên Ti Brain
```

### Scenario 2: Need System-Wide Pattern
```
1. Agent cần pattern từ project khác
2. Pull từ Ti Brain memory/lessons/
3. Apply vào local project
4. Document lessons learned
```

### Scenario 3: Daily/Weekly Sync
```
1. Push local changes lên Ti Brain
2. Pull system-wide updates từ Ti Brain
3. Resolve conflicts
4. Update local context
```

---

## 🛠️ Implementation

### Phase 1: Setup Sync Script
```bash
# scripts/sync-brain.sh
#!/bin/bash
# Sync .devin with Ti Brain

DEVIN_PATH="/z/Ti/Ti-learning-lab/.devin"
BRAIN_PATH="/z/Ti/brain"
PROJECT="ti-learning-lab"

# Push local knowledge to Ti Brain
rsync -av "$DEVIN_PATH/knowledge/" "$BRAIN_PATH/memory/lessons/$PROJECT/"

# Pull system-wide patterns from Ti Brain
rsync -av "$BRAIN_PATH/memory/lessons/" "$DEVIN_PATH/knowledge/"

echo "Sync completed"
```

### Phase 2: Add to .devin/config/settings.json
```json
{
  "sync": {
    "enabled": true,
    "brain_path": "/z/Ti/brain",
    "sync_on": ["task_complete", "daily", "weekly"],
    "auto_push": true,
    "auto_pull": false
  }
}
```

### Phase 3: Integrate into Agent Workflow
```
Khi agent hoàn thành task:
1. Update .devin/memory/session-memory.md
2. Update .devin/knowledge/ (nếu có pattern/lesson mới)
3. Run sync script (nếu enabled)
4. Log vào Ti Brain beads
```

---

## 🎯 Benefits

### 1. **Speed**
- Local memory access: <10ms
- Không cần query centralized system

### 2. **Context-Specific**
- Knowledge riêng cho Ti-Learning-Lab
- Không bị noise từ projects khác

### 3. **Sharing**
- Patterns learned có thể chia sẻ qua Ti Brain
- Lessons learned có thể benefit các projects khác

### 4. **Redundancy**
- 2 sources of knowledge
- Nếu một fails, có backup

### 5. **Flexibility**
- Có thể work offline với local memory
- Có thể sync khi online

---

## 📝 Workflow Cho Agents

### Trước khi bắt đầu task:
1. Đọc `.devin/context/project-context.md`
2. Đọc `.devin/knowledge/repo-index.md`
3. (Optional) Đọc Ti Brain system overview

### Trong khi làm task:
1. Update `.devin/memory/session-memory.md`
2. Document patterns vào `.devin/knowledge/patterns.md`
3. Document lessons vào `.devin/knowledge/lessons.md`

### Sau khi hoàn thành task:
1. Run sync script (nếu enabled)
2. Log beads vào Ti Brain
3. Update Ti Brain nếu cần

---

## 🔧 Configuration

### .devin/config/settings.json
```json
{
  "sync": {
    "enabled": true,
    "brain_path": "/z/Ti/brain",
    "sync_on": ["task_complete"],
    "auto_push": true,
    "auto_pull": false,
    "conflict_resolution": "manual"
  }
}
```

### Sync Modes:
- **task_complete**: Sync sau mỗi task
- **daily**: Sync mỗi ngày
- **weekly**: Sync mỗi tuần
- **manual**: Sync khi agent yêu cầu

---

## 🚀 Next Steps

### Phase 1: Setup Sync Script (Tuần này)
- [ ] Create `scripts/sync-brain.sh`
- [ ] Test sync .devin → Ti Brain
- [ ] Test sync Ti Brain → .devin
- [ ] Add to .devin/config/settings.json

### Phase 2: Integrate into Agent Workflow (Tuần sau)
- [ ] Modify agent workflow để auto-sync
- [ ] Add sync step sau task completion
- [ ] Test end-to-end

### Phase 3: Optimize (Tháng sau)
- [ ] Implement conflict resolution
- [ ] Add incremental sync
- [ ] Add sync scheduling

---

## 📊 Comparison Summary

| Aspect | Local Only | Central Only | Hybrid |
|--------|------------|--------------|---------|
| **Speed** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Sharing** | ⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Context** | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Redundancy** | ⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Complexity** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ |

---

**Recommendation**: **Hybrid approach** - best của cả 2 worlds!

**Last Updated**: 2026-04-28
