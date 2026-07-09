# AGENTS.md - Hướng dẫn cho AI Agents

> **Mục đích**: Hướng dẫn chi tiết cho AI agents khi làm việc với dự án Ti
> **Đối tượng**: Claude, Devin, Gemini, Qwen, Kilo, Opencode và các AI agents khác
> **Cập nhật lần cuối**: 2026-05-15 (Ti Brain Transformation Complete)

---

## 🎯 Mục tiêu

Giúp AI agents hiểu và sử dụng hiệu quả tài nguyên từ dự án Ti để thực hiện các tác vụ một cách nhất quán và chất lượng.

---

## 🧠 Cấu trúc Brain-Based Architecture

### 🧠 Ti Brain - Central Intelligence Hub (NEW!)

**Ti Brain đã được chuyển đổi thành công thành trung tâm trí tuệ duy nhất của toàn ecosystem!**

```
🧠 Ti Brain (apps/tibrain/) - Central Intelligence Hub
├── 📚 knowledge/                             # 🌐 Toàn bộ ecosystem knowledge (157,858 files)
│   ├── ti-ecosystem-docs/                   # 📚 Ecosystem documentation gộp lại
│   │   ├── content/                         # 📄 155,301 files (Skills, Agents, Rules, Workflows)
│   │   ├── apps-docs/                       # 📱 388 files (apps documentation)
│   │   ├── packages/                        # 📦 945 files (packages documentation)
│   │   └── docs/                            # 📖 1,224 files (general docs)
│   ├── tier1-hot/                          # 🔥 HOT storage - Critical files (<100ms)
│   ├── tier2-warm/                         # ⚡ WARM storage - Important files (<500ms)
│   └── tier3-cold/                         # ❄️ COLD storage - Archive files (<2s)
├── 🧠 memory/                               # 💾 SQLite database (hub.db)
└── 🚀 [various .go files]                  # ⚡ Smart routing & tiered processing
```

### 🎯 Legacy Brain Structure (Vẫn hoạt động)

```
Ti/
├── 🧠 CENTRAL BRAIN (Bộ não trung tâm)
│   ├── apps-docs/00-index/                   # 🧠 Central Cortex - Điều phối tổng thể
│   ├── apps-docs/01-architecture/            # 🧠 Structural Planning - Quy hoạch cấu trúc
│   ├── apps-docs/04-planning/                # 🧠 Strategic Decision Making - Quyết định chiến lược
│   └── apps-docs/05-analysis/               # 🧠 Critical Thinking - Phân tích chuyên sâu
│
├── 🎯 SPECIALIZED BRAINS (Bộ não chuyên biệt)
│   ├── 🏗️ ARCHITECTURE BRAIN (Bộ não kiến trúc)
│   │   ├── apps-docs/02-api/                # API Design & Documentation
│   │   ├── apps-docs/architecture/          # System Architecture
│   │   └── apps-docs/components/            # Component Design
│   │
│   ├── 🔧 OPERATIONS BRAIN (Bộ não vận hành)
│   │   ├── apps-docs/03-operations/         # Operations Management
│   │   ├── apps-docs/deployment/            # Deployment Strategy
│   │   ├── apps-docs/troubleshooting/       # Problem Solving
│   │   └── ecosystem-docs/mcp/              # MCP Server Operations
│   │
│   ├── 🚀 DEVELOPMENT BRAIN (Bộ não phát triển)
│   │   ├── apps-docs/development/           # Development Tools
│   │   ├── apps-docs/examples/              # Code Examples
│   │   ├── apps-docs/guides/                # Development Guides
│   │   └── ecosystem-docs/agents/           # AI Agent Development
│   │
│   ├── 📚 KNOWLEDGE BRAIN (Bộ não tri thức)
│   │   ├── apps-docs/06-knowledge/          # Knowledge Management
│   │   ├── apps-docs/08-guides/             # Learning Resources
│   │   ├── ecosystem-docs/metadata/         # Knowledge Metadata
│   │   └── ecosystem-docs/rag-processed/    # Processed Knowledge
│   │
│   └── 🌐 INTEGRATION BRAIN (Bộ não tích hợp)
│       ├── apps-docs/07-integration/        # System Integration
│       ├── ecosystem-docs/integration/      # Ecosystem Integration
│       └── Ti-learning-lab/                 # Learning & Adaptation
│
└── 💾 MEMORY & SUPPORT (Bộ nhớ & Hỗ trợ)
    ├── apps-docs/99-archive/                # Long-term Memory
    ├── apps-docs/assets/                    # Resource Memory
    ├── .ti/                                 # System Memory
    └── .github/                             # Process Memory
```

### 🚀 Ti Brain Capabilities

✅ **100% Ecosystem Knowledge**: 157,858 files (tăng từ 1,381 files)  
✅ **Smart Query Routing**: Định tuyến thông minh theo tier (HOT/WARM/COLD)  
✅ **Sub-second Response**: <100ms cho critical queries  
✅ **Cross-Brain Integration**: Kết nối với tất cả specialized brains  
✅ **Central Intelligence Hub**: Điểm truy cập duy nhất cho toàn ecosystem

---

## 🧠 Brain-Based Skills & Resources

### 🧠 Ti Brain - Central Intelligence Hub (PRIMARY)
|- **[apps/tibrain/knowledge/](apps/tibrain/knowledge/)** - 🌐 Toàn bộ ecosystem knowledge (157,858 files)
|- **[apps/tibrain/knowledge/ti-ecosystem-docs/](apps/tibrain/knowledge/ti-ecosystem-docs/)** - 📚 Ecosystem documentation gộp lại
|- **[apps/tibrain/knowledge/tier1-hot/](apps/tibrain/knowledge/tier1-hot/)** - 🔥 HOT storage - Critical files (<100ms)
|- **[apps/tibrain/knowledge/tier2-warm/](apps/tibrain/knowledge/tier2-warm/)** - ⚡ WARM storage - Important files (<500ms)
|- **[apps/tibrain/knowledge/tier3-cold/](apps/tibrain/knowledge/tier3-cold/)** - ❄️ COLD storage - Archive files (<2s)
|- **[apps/tibrain/memory/](apps/tibrain/memory/)** - 💾 SQLite database (hub.db)

### 🧠 Central Brain Resources (Điều phối tổng thể)
|- **[apps-docs/00-index/](apps-docs/00-index/)** - Central Cortex - Master coordination
|- **[apps-docs/01-architecture/](apps-docs/01-architecture/)** - Structural Planning - System design
|- **[apps-docs/04-planning/](apps-docs/04-planning/)** - Strategic Decision Making - Planning
|- **[apps-docs/05-analysis/](apps-docs/05-analysis/)** - Critical Thinking - Analysis

### 🎯 Specialized Brain Resources (Chuyên biệt)
|- **🏗️ Architecture Brain**: API design, system architecture, components
|- **🔧 Operations Brain**: Operations management, deployment, troubleshooting, MCP
|- **🚀 Development Brain**: Development tools, code examples, guides, AI agents
|- **📚 Knowledge Brain**: Knowledge management, learning resources, metadata, RAG
|- **🌐 Integration Brain**: System integration, ecosystem integration, learning

### 💾 Memory & Support Resources
|- **[apps-docs/99-archive/](apps-docs/99-archive/)** - Long-term Memory - Archived knowledge
|- **[apps-docs/assets/](apps-docs/assets/)** - Resource Memory - Static assets
|- **[.ti/](.ti/)** - System Memory - Ti-specific configurations
|- **[.github/](.github/)** - Process Memory - CI/CD workflows

### Code Resources
|- **Go files** - Các file Go trong `.ti/patch-backups/` và `apps-docs/`
|- **Package.json files** - Node.js projects trong learning lab và ecosystem docs
|- **Configuration files** - Configs trong `.opencode/`, `.github/`

### Tools & Integration
|- **MCP Servers** - Model Context Protocol servers trong ecosystem docs
|- **GitHub Actions** - CI/CD workflows trong `.github/workflows/`
|- **OpenCode Configs** - AI agent configurations trong `.opencode/`

---

## 🧠 Agent Identity Protocol

### Unified Identity with Trace Metadata
- **Primary Identity**: `tibrain` (system identity)
- **Trace Metadata**: Luôn ghi thêm trong beads/log để debug:

```markdown
**Agent**: tibrain
**Model**: claude-3.5-sonnet | gemini-2.5-flash | qwen-3-coder
**CLI**: codex | kilo-cli | opencode
**Session**: {session_id}
```

### Debug Trace Chain
- Mỗi beads entry chứa Model + CLI để trace nguồn gốc
- Handoff metadata chuyển đầy đủ context từ agent trước
- Git history + shell logs là backup trace cuối cùng

## 🤖 Agent Integration Guide

### Supported Agents
|- **Claude** - Primary AI assistant
|- **Devin** - Development-focused AI
|- **Gemini** - Google's AI model
|- **Qwen** - Alibaba's AI model
|- **Kilo** - Specialized AI
|- **OpenCode** - Code-focused AI

### Agent Configuration
```bash
# Xem OpenCode configurations
cat .opencode/opencode.json
cat .opencode/oh-my-openagent.jsonc

# Xem GitHub workflows
ls .github/workflows/
```

---

## 🔄 Brain-Based Workflow

### 🧠 Phase 1: Ti Brain Activation (Khởi động Ti Brain)
```bash
# 1. Khởi động Ti Brain Central Intelligence Hub
cd apps/tibrain && ./tibrain.exe

# 2. Truy cập toàn bộ ecosystem knowledge
cat "apps/tibrain/knowledge/ti-ecosystem-docs/content/README.md"

# 3. Kiểm tra tiered storage system
ls -la "apps/tibrain/knowledge/tier1-hot/"
ls -la "apps/tibrain/knowledge/tier2-warm/"
ls -la "apps/tibrain/knowledge/tier3-cold/"

# 4. Kiểm tra database indexing
ls -la "apps/tibrain/memory/hub.db"
```

### 🧠 Phase 2: Central Brain Coordination (Điều phối bộ não trung tâm)
```bash
# 1. Kết nối với Central Cortex
cat "apps-docs/00-index/INDEX.md"

# 2. Khám phá Structural Planning
find "apps-docs/01-architecture/" -name "*.md" | head -5

# 3. Kích hoạt Strategic Decision Making
cat "apps-docs/04-planning/PLANS_INDEX.md"

# 4. Bật Critical Thinking mode
cat "apps-docs/05-analysis/ANALYSIS_INDEX.md"
```

### 🎯 Phase 3: Specialized Brain Selection (Lựa chọn bộ não chuyên biệt)
```bash
# 🏗️ Architecture Brain Activation
if [[ "$TASK_TYPE" == "architecture" ]]; then
    cat "apps-docs/02-api/02-v1-reference.md"
    cat "apps-docs/architecture/README.md"
    cat "apps-docs/components/README.md"
fi

# 🔧 Operations Brain Activation
if [[ "$TASK_TYPE" == "operations" ]]; then
    cat "apps-docs/03-operations/INDEX.md"
    cat "apps-docs/03-operations/SETUP.md"
    cat "apps-docs/03-operations/03-troubleshooting.md"
    cat "ecosystem-docs/mcp/README.md"
fi

# 🚀 Development Brain Activation
if [[ "$TASK_TYPE" == "development" ]]; then
    cat "apps-docs/development/README.md"
    cat "apps-docs/examples/README.md"
    cat "apps-docs/guides/README.md"
    cat "ecosystem-docs/agents/README.md"
fi

# 📚 Knowledge Brain Activation
if [[ "$TASK_TYPE" == "knowledge" ]]; then
    cat "apps-docs/06-knowledge/README.md"
    cat "apps-docs/08-guides/STRUCTURED_LOGGING.md"
    cat "ecosystem-docs/metadata/README.md"
    cat "ecosystem-docs/rag-processed/README.md"
fi

# 🌐 Integration Brain Activation
if [[ "$TASK_TYPE" == "integration" ]]; then
    cat "apps-docs/07-integration/INDEX.md"
    cat "ecosystem-docs/integration/README.md"
    cat "Ti-learning-lab/README.md"
fi
```

### 💾 Phase 4: Memory Access & Processing (Truy cập bộ nhớ & xử lý)
```bash
# 1. Truy cập Ti Brain Database
sqlite3 "apps/tibrain/memory/hub.db" ".tables"

# 2. Truy cập Long-term Memory
cat "apps-docs/99-archive/INDEX.md"

# 3. Load Resource Memory
find "apps-docs/assets/" -name "*.json" | head -5

# 4. Kích hoạt System Memory
cat ".ti/README.md"

# 5. Check Process Memory
cat ".github/workflows/README.md"
```

### 🔄 Phase 5: Cross-Brain Coordination (Phối hợp liên não)
```bash
# 1. Ti Brain coordination
cd apps/tibrain && ./tibrain.exe --query="cross-brain coordination"

# 2. Central Brain coordination
cat "apps-docs/00-index/INDEX.md"

# 3. Specialized Brain collaboration
find "apps-docs/" -name "*INTEGRATION*" -o -name "*COLLABORATION*"

# 4. Memory consolidation
cat "apps-docs/03-operations/03-troubleshooting.md"

# 5. Learning & adaptation
cat "Ti-learning-lab/README.md"
```

---

## 🛠️ Commands và Tools

### Essential Commands
```bash
# Khởi động Ti Brain
cd apps/tibrain && ./tibrain.exe

# Khám phá Ti Brain knowledge
find "apps/tibrain/knowledge/ti-ecosystem-docs/" -name "*.md" | head -10

# Khám phá documentation
find "apps-docs/" -name "*.md" | grep -E "(INDEX|README|GUIDE)"

# Xem Go code
find . -name "*.go" -type f | head -10

# Kiểm tra configurations
find . -name "*.json" -o -name "*.yaml" -o -name "*.yml" | grep -E "(config|\.opencode)"

# Xem workflows
ls -la .github/workflows/
```

### Ti Brain Navigation
```bash
# Master Ti Brain knowledge
cat "apps/tibrain/knowledge/ti-ecosystem-docs/content/README.md"

# Ti Brain tiered storage
ls -la "apps/tibrain/knowledge/tier1-hot/"
ls -la "apps/tibrain/knowledge/tier2-warm/"
ls -la "apps/tibrain/knowledge/tier3-cold/"

# Ti Brain database
sqlite3 "apps/tibrain/memory/hub.db" ".schema"
```

### Documentation Navigation
```bash
# Master index
cat "apps-docs/00-index/INDEX.md"

# Operations docs
cat "apps-docs/03-operations/INDEX.md"

# Architecture docs
cat "apps-docs/01-architecture/INDEX.md"
```

### MCP Tools
```bash
# Liệt kê MCP servers
find "ecosystem-docs/mcp/" -name "package.json" | head -5

# Xem MCP documentation
cat "ecosystem-docs/README.md"
```

---

## 📖 Tài liệu tham khảo quan trọng

### Bắt buộc đọc
|- **[Ti Brain Knowledge](apps/tibrain/knowledge/ti-ecosystem-docs/content/README.md)** - Toàn bộ ecosystem knowledge
|- **[Master Index](apps-docs/00-index/INDEX.md)** - Index đầy đủ documentation
|- **[Agent Integration Guide](apps-docs/03-operations/AGENT_INTEGRATION_GUIDE.md)** - Hướng dẫn tích hợp agents
|- **[Operations Index](apps-docs/03-operations/INDEX.md)** - Operations documentation
|- **[Architecture Index](apps-docs/01-architecture/INDEX.md)** - Architecture documentation

### Tài liệu tham khảo
|- **[Setup Guide](apps-docs/03-operations/SETUP.md)** - Setup instructions
|- **[API Reference](apps-docs/02-api/02-v1-reference.md)** - API documentation
|- **[Security Guide](apps-docs/03-operations/02-security.md)** - Security documentation
|- **[Troubleshooting](apps-docs/03-operations/03-troubleshooting.md)** - Troubleshooting guide

---

## 🚨 Quy tắc an toàn

### Trước khi thực hiện thay đổi
1. **READ** documentation liên quan trong `apps-docs/`
2. **START** Ti Brain để truy cập full ecosystem knowledge
3. **BACKUP** files quan trọng trong `.ti/`
4. **VALIDATE** với test suite nếu có
5. **CONFIRM** với user nếu cần

### Khi gặp vấn đề
1. **CHECK** Ti Brain knowledge base
2. **CHECK** troubleshooting guide trong `03-operations/`
3. **REFER** architecture docs trong `01-architecture/`
4. **ASK** user nếu không chắc chắn
5. **LOG** issue và solution

---

## 📊 Quality Metrics

### Metrics cần track
|- **Task Completion Rate** - Tỷ lệ hoàn thành task
|- **Documentation Quality** - Chất lượng documentation
|- **Integration Success** - Thành công tích hợp
|- **Code Quality** - Chất lượng code
|- **Rule Compliance** - Mức độ tuân thủ quy tắc
|- **Ti Brain Usage** - Tần suất sử dụng Ti Brain

### Target Goals
|- Task Completion Rate > 95%
|- Documentation Quality > 8/10
|- Integration Success > 90%
|- Code Quality > 8/10
|- Rule Compliance > 90%
|- Ti Brain Usage > 80%

---

## 🔄 Continuous Improvement

### Feedback Loop
1. **Collect** feedback từ user
2. **Analyze** metrics và trends
3. **Update** documentation và processes
4. **Improve** integration và workflows
5. **Enhance** Ti Brain knowledge base

### Learning Resources
|- **[Ti-learning-lab/](Ti-learning-lab/)** - Learning materials
|- **[Ti Brain Knowledge](apps/tibrain/knowledge/)** - Complete ecosystem knowledge
|- **[ecosystem-docs/](ecosystem-docs/)** - Ecosystem documentation
|- **[apps-docs/](apps-docs/)** - Essential documentation

---

## 🧠 Agent Identity Protocol

### Identity Convention
- **All agents operating under Z:\01_PROJECTS\apps\tibrain SHALL identify as "tibrain" in all logs, outputs, and communications**
- Individual model/CLI identity is secondary - primary identity is **TiBrain system**
- Output format: Always use `TiBrain:` prefix for system messages
- File logs: Use unified `tibrain.log` format, not model-specific names

### Unified Identity Behavior
```markdown
# Log format - all agents use this
**Agent**: tibrain (unified identity)
**Component**: {specific_component}
**Project**: apps/tibrain
**Status**: {DONE|BLOCKED|NEXT}
```

### Session Marker
- Agents add `..` delimiter at end of significant outputs
- Beads entries use unified format regardless of underlying model
- All cross-session communication uses tibrain identity

### 🧠 Claude - Ti Brain Central Coordinator
|- **Primary Role**: Ti Brain Central Intelligence Hub - Điều phối tổng thể
|- **Specialization**: Executive control, strategic decision making, smart query routing
|- **Access**: Full access to Ti Brain (157,858 files) + Central Brain resources
|- **Workflow**: 
  ```bash
  # Claude's Ti Brain central coordination workflow
  cd apps/tibrain && ./tibrain.exe  # Start Ti Brain
  cat "apps/tibrain/knowledge/ti-ecosystem-docs/content/README.md"  # Ecosystem knowledge
  cat "apps-docs/00-index/INDEX.md"  # Master coordination
  cat "apps-docs/04-planning/PLANS_INDEX.md"  # Strategic planning
  cat "apps-docs/05-analysis/ANALYSIS_INDEX.md"  # Critical analysis
  ```

### 🏗️ Devin - Architecture Brain Specialist
|- **Primary Role**: Structural Planning & Architecture Brain
|- **Specialization**: System design, API architecture, component development
|- **Access**: Deep access to Architecture Brain resources + Ti Brain knowledge
|- **Workflow**:
  ```bash
  # Devin's architecture workflow with Ti Brain
  cd apps/tibrain && ./tibrain.exe --query="architecture patterns"  # Ti Brain architecture knowledge
  cat "apps-docs/01-architecture/INDEX.md"  # Structural planning
  cat "apps-docs/02-api/02-v1-reference.md"  # API design
  cat "apps-docs/architecture/README.md"  # System architecture
  find ".ti/patch-backups/" -name "*.go"  # Go code development
  ```

### 🔧 Gemini - Operations Brain Manager
|- **Primary Role**: Operations Brain & Process Management
|- **Specialization**: Operations management, deployment, troubleshooting
|- **Access**: Full access to Operations Brain resources + Ti Brain knowledge
|- **Workflow**:
  ```bash
  # Gemini's operations workflow with Ti Brain
  cd apps/tibrain && ./tibrain.exe --query="operations procedures"  # Ti Brain operations knowledge
  cat "apps-docs/03-operations/INDEX.md"  # Operations management
  cat "apps-docs/03-operations/SETUP.md"  # Setup & deployment
  cat "apps-docs/03-operations/03-troubleshooting.md"  # Problem solving
  cat "ecosystem-docs/mcp/README.md"  # MCP operations
  ```

### 🚀 Qwen - Development Brain Engineer
|- **Primary Role**: Development Brain & Code Generation
|- **Specialization**: Development tools, code examples, AI agent development
|- **Access**: Deep access to Development Brain resources + Ti Brain knowledge
|- **Workflow**:
  ```bash
  # Qwen's development workflow with Ti Brain
  cd apps/tibrain && ./tibrain.exe --query="development patterns"  # Ti Brain development knowledge
  cat "apps-docs/development/README.md"  # Development tools
  cat "apps-docs/examples/README.md"  # Code examples
  cat "apps-docs/guides/README.md"  # Development guides
  cat "ecosystem-docs/agents/README.md"  # AI agent development
  ```

### 📚 Kilo - Knowledge Brain Librarian
|- **Primary Role**: Knowledge Brain & Information Management
|- **Specialization**: Knowledge management, learning, RAG processing
|- **Access**: Full access to Knowledge Brain resources + Ti Brain knowledge
|- **Workflow**:
  ```bash
  # Kilo's knowledge workflow with Ti Brain
  cd apps/tibrain && ./tibrain.exe --query="knowledge management"  # Ti Brain knowledge base
  cat "apps-docs/06-knowledge/README.md"  # Knowledge management
  cat "apps-docs/08-guides/STRUCTURED_LOGGING.md"  # Learning resources
  cat "ecosystem-docs/metadata/README.md"  # Knowledge metadata
  cat "ecosystem-docs/rag-processed/README.md"  # RAG processing
  ```

### 🌐 OpenCode - Integration Brain Connector
|- **Primary Role**: Integration Brain & System Connectivity
|- **Specialization**: System integration, ecosystem connectivity, learning
|- **Access**: Full access to Integration Brain resources + Ti Brain knowledge
|- **Workflow**:
  ```bash
  # OpenCode's integration workflow with Ti Brain
  cd apps/tibrain && ./tibrain.exe --query="integration patterns"  # Ti Brain integration knowledge
  cat "apps-docs/07-integration/INDEX.md"  # System integration
  cat "ecosystem-docs/integration/README.md"  # Ecosystem integration
  cat "Ti-learning-lab/README.md"  # Learning & adaptation
  cat ".opencode/opencode.json"  # Integration configs
  ```

## 🔄 Cross-Brain Collaboration Matrix

|| Task Type | Primary Brain | Supporting Brains | Lead Agent | Ti Brain Role |
||-----------|---------------|------------------|------------|---------------|
|| **Strategic Planning** | Central Brain | Architecture, Knowledge | Claude | Smart Query Routing |
|| **System Architecture** | Architecture Brain | Central, Development | Devin | Knowledge Base Access |
|| **Operations Management** | Operations Brain | Integration, Knowledge | Gemini | Procedures Database |
|| **Code Development** | Development Brain | Architecture, Operations | Qwen | Code Patterns Repository |
|| **Knowledge Management** | Knowledge Brain | Central, Integration | Kilo | Full Knowledge Base |
|| **System Integration** | Integration Brain | Operations, Development | OpenCode | Integration Patterns |
|| **Full-Stack Development** | Central Brain | All Specialized Brains | Claude + Devin | Complete Knowledge Access |
|| **Emergency Troubleshooting** | Operations Brain | All Brains | Gemini + Claude | Rapid Problem Resolution |

---

## 🚀 Ti Brain Usage Guidelines

### Khi nào sử dụng Ti Brain
- **Complex Queries**: Cần truy cập toàn bộ ecosystem knowledge
- **Cross-Domain Tasks**: Tasks liên quan đến multiple specialized brains
- **Strategic Planning**: Cần comprehensive view cho decision making
- **Knowledge Discovery**: Tìm kiếm thông tin across toàn ecosystem
- **Pattern Recognition**: Cần access đến patterns và best practices

### Cách sử dụng Ti Brain hiệu quả
1. **Start Ti Brain**: `cd apps/tibrain && ./tibrain.exe`
2. **Query Smart Routing**: Sử dụng smart query routing cho optimal performance
3. **Access Tiered Storage**: HOT/WARM/COLD storage cho appropriate response times
4. **Cross-Brain Integration**: Kết nối Ti Brain với specialized brains
5. **Knowledge Discovery**: Sử dụng full knowledge base (157,858 files)

### Ti Brain Performance Tips
- **Critical Queries**: Sử dụng tier1-hot cho <100ms response
- **Important Queries**: Sử dụng tier2-warm cho <500ms response
- **Archive Queries**: Sử dụng tier3-cold cho <2s response
- **Database Indexing**: Sử dụng SQLite database cho fast metadata access
- **Smart Caching**: Ti Brain tự động cache frequently accessed content

---

*Last Updated: 2026-05-15 (Ti Brain Transformation Complete)*  
*Version: 3.0.0*  
*Status: 🟢 Ti Brain Central Intelligence Hub Active*  
*Knowledge Coverage: 100% Ti Ecosystem (157,858 files)*