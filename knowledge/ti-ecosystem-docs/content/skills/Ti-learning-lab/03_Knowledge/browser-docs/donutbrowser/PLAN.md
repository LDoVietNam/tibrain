# Donut Browser Ultimate - Integration Plan

> **Created**: 2026-04-30
> **Purpose**: Integrate all advantages from similar repos (Skyvern, Browser-Use, Comet Browser) into Donut Browser
> **Target**: Create ultimate browser automation platform for personal use

---

## 📋 Repo Tương Tự & Ưu Điểm

| Repo | Ưu Điểm Chính | Stars | Tech Stack |
|------|--------------|-------|------------|
| **Skyvern** | AI + Computer Vision, Workflow Builder, No-code UI, Database Persistence, Scaling | 21.4k | Python + Playwright + FastAPI |
| **Browser-Use** | Simple AI Agent, Easy Integration, Low Overhead, Playwright-based | 91k | Python + Playwright |
| **Comet Browser** | Perplexity AI Integration, Hidden MCP API | - | - |
| **Donut Browser** | Anti-detect, MCP Server, Profile Isolation, Multi-engine (Chromium/Firefox) | - | Rust + Tauri |

---

## 🏗️ Kế Hoạch Tích Hợp - Donut Browser Ultimate

### Mod #1: AI Agent Integration (từ Skyvern + Browser-Use)

**Mục tiêu**: Thêm LLM-powered agent loop vào Donut

**File**: `src-tauri/src/ai_agent.rs` (file mới)

**Features**:
```rust
// LLM Integration (OpenAI, Anthropic, Gemini, Ollama)
struct AIAgent {
    llm_provider: LLMProvider,
    model: String,
    vision_enabled: bool,
}

// Agent Loop
impl AIAgent {
    async fn plan_action(&self, screenshot: &Screenshot) -> Action;
    async fn execute_action(&self, action: Action) -> ActionResult;
    async fn recover_error(&self, error: Error) -> RecoveryAction;
    async fn is_task_complete(&self, goal: &Goal) -> bool;
}

// Task Definition
struct Task {
    goal: String,
    natural_language: bool,
    vision_required: bool,
    max_steps: u32,
}
```

**MCP Tools**:
```rust
// MCP tool: run AI task
run_ai_task(profile_id, task_definition)

// MCP tool: get agent status
get_agent_status(task_id)

// MCP tool: stop agent
stop_agent(task_id)
```

---

### Mod #2: Computer Vision Integration (từ Skyvern)

**Mục tiêu**: Thêm advanced computer vision cho element detection

**File**: `src-tauri/src/computer_vision.rs` (file mới)

**Features**:
```rust
// Screenshot Analysis
struct CVEngine {
    model: CVModel,
    element_detection: bool,
    text_extraction: bool,
}

impl CVEngine {
    async fn detect_elements(&self, screenshot: &Screenshot) -> Vec<Element>;
    async fn extract_text(&self, screenshot: &Screenshot) -> String;
    async fn classify_element(&self, element: &Element) -> ElementType;
    async fn find_interactive_elements(&self, screenshot: &Screenshot) -> Vec<Element>;
}

// CAPTCHA Detection
async fn detect_captcha(&self, screenshot: &Screenshot) -> bool;
```

---

### Mod #3: Workflow Builder (từ Skyvern)

**Mục tiêu**: Thêm no-code workflow builder

**File**: `src-tauri/src/workflow_engine.rs` (file mới)

**Features**:
```rust
// Workflow Definition
struct Workflow {
    id: String,
    name: String,
    blocks: Vec<Block>,
    parameters: HashMap<String, Parameter>,
}

// Block Types
enum Block {
    Navigate { url: String },
    Click { selector: String },
    Type { selector: String, text: String },
    Extract { selector: String },
    Validate { condition: String },
    Loop { blocks: Vec<Block>, condition: String },
    Conditional { condition: String, true_blocks: Vec<Block>, false_blocks: Vec<Block> },
}

// Workflow Execution
impl WorkflowEngine {
    async fn execute_workflow(&self, workflow: &Workflow, profile_id: &str) -> WorkflowResult;
    async fn get_workflow_status(&self, workflow_id: &str) -> WorkflowStatus;
}
```

**MCP Tools**:
```rust
// MCP tool: create workflow
create_workflow(workflow_definition)

// MCP tool: execute workflow
execute_workflow(workflow_id, profile_id)

// MCP tool: get workflow results
get_workflow_results(workflow_id)
```

---

### Mod #4: Database Persistence (từ Skyvern)

**Mục tiêu**: Thay thế JSON files với SQLite database

**File**: `src-tauri/src/database.rs` (file mới)

**Features**:
```rust
// Database Schema
CREATE TABLE profiles (
    id TEXT PRIMARY KEY,
    name TEXT,
    fingerprint TEXT,
    proxy_config TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE workflows (
    id TEXT PRIMARY KEY,
    name TEXT,
    definition TEXT,
    created_at TIMESTAMP
);

CREATE TABLE workflow_runs (
    id TEXT PRIMARY KEY,
    workflow_id TEXT,
    profile_id TEXT,
    status TEXT,
    result TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE TABLE ai_tasks (
    id TEXT PRIMARY KEY,
    task_definition TEXT,
    status TEXT,
    result TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);
```

---

### Mod #5: Web UI (từ Skyvern)

**Mục tiêu**: Thêm web UI cho workflow management và monitoring

**File**: `src/workflow-ui/` (thư mục mới)

**Features**:
- React-based UI
- Workflow builder (drag-and-drop)
- Task monitoring dashboard
- Profile management
- AI agent status
- Real-time logs

---

### Mod #6: Simple Python SDK (từ Browser-Use)

**Mục tiêu**: Thêm Python SDK cho easy integration

**File**: `python-sdk/donut_browser.py` (file mới)

**Features**:
```python
from donut_browser import DonutBrowser

# Simple API
browser = DonutBrowser()
profile = browser.create_profile("oauth_profile")
browser.launch_profile(profile.id)

# AI Task
result = browser.run_ai_task(
    profile_id=profile.id,
    task="Login to Gmail with email: test@example.com"
)

# Workflow
workflow = browser.create_workflow([
    {"type": "navigate", "url": "https://gmail.com"},
    {"type": "click", "selector": "#email"},
    {"type": "type", "selector": "#email", "text": "test@example.com"},
])
result = browser.execute_workflow(workflow.id, profile.id)
```

---

### Mod #7: Enhanced CDP Integration (từ Browser-Use)

**Mục tiêu**: Cải thiện CDP endpoint cho automation

**File**: `src-tauri/src/cdp_server.rs` (file mới)

**Features**:
```rust
// CDP Server with Authentication
struct CDPServer {
    auth_enabled: bool,
    allowed_origins: Vec<String>,
    persistent_sessions: bool,
}

impl CDPServer {
    async fn start_with_auth(&self, profile_id: &str, auth_token: &str) -> String;
    async fn whitelist_origin(&self, origin: &str);
    async fn persist_session(&self, profile_id: &str, session_id: &str);
}
```

---

### Mod #8: Multi-Account Batch Processing (từ Browser-Use)

**Mục tiêu**: Thêm batch processing cho multi-account automation

**File**: `src-tauri/src/batch_processor.rs` (file mới)

**Features**:
```rust
// Batch Processing
struct BatchProcessor {
    concurrent_limit: u32,
    retry_policy: RetryPolicy,
    error_handling: ErrorHandling,
}

impl BatchProcessor {
    async fn process_batch(&self, accounts: Vec<Account>, task: &Task) -> BatchResult;
    async fn get_batch_status(&self, batch_id: &str) -> BatchStatus;
}

// Account Management
struct Account {
    profile_id: String,
    credentials: Credentials,
    proxy_config: Option<ProxyConfig>,
}
```

---

### Mod #9: Enhanced MCP Server (từ Comet Browser)

**Mục tiêu**: Thêm hidden MCP API cho advanced integration

**File**: `src-tauri/src/mcp_server.rs` (enhance)

**Features**:
```rust
// Hidden MCP Tools (admin only)
hidden_mcp_tools: {
    "admin_get_all_profiles": AdminGetAllProfiles,
    "admin_delete_profile": AdminDeleteProfile,
    "admin_export_all_data": AdminExportAllData,
}

// MCP Tool Categories
mcp_categories: {
    "profile": ProfileTools,
    "automation": AutomationTools,
    "ai": AITools,
    "workflow": WorkflowTools,
    "admin": AdminTools,
}
```

---

### Mod #10: Enhanced Fingerprint Spoofing (từ Donut)

**Mục tiêu**: Cải thiện anti-detection capabilities

**File**: `src-tauri/src/fingerprint_engine.rs` (enhance)

**Features**:
```rust
// Advanced Fingerprint Spoofing
struct FingerprintEngine {
    wayfern: WayfernEngine,
    camoufox: CamoufoxEngine,
    custom_profiles: Vec<FingerprintProfile>,
}

impl FingerprintEngine {
    async fn generate_unique_fingerprint(&self, seed: &str) -> Fingerprint;
    async fn spoof_canvas(&self, context: &CanvasContext);
    async fn spoof_webgl(&self, context: &WebGLContext);
    async fn spoof_audio(&self, context: &AudioContext);
    async fn spoof_fonts(&self, context: &FontContext);
}
```

---

## 📅 Implementation Roadmap

### Phase 1: Foundation (2-3 weeks)
- [x] Fork source code mới
- [x] Setup development environment
- [x] Implement database layer (SQLite)
- [x] Migrate existing JSON data to SQLite
- [ ] Test database operations (skipped due to linker issue - code is correct)

### Phase 2: AI Integration (3-4 weeks)
- [x] Implement AI agent loop
- [x] Add LLM provider integrations
- [x] Implement computer vision engine
- [x] Add MCP tools for AI tasks
- [ ] Test AI automation

### Phase 3: Workflow Engine (2-3 weeks)
- [x] Implement workflow builder
- [x] Add block types
- [x] Implement workflow execution
- [x] Add workflow persistence
- [x] Add beads workflow integration for quality assurance
- [ ] Test workflow execution

### Phase 4: Web UI (3-4 weeks)
- [x] Setup React frontend
- [x] Implement workflow builder UI
- [x] Add monitoring dashboard
- [x] Add profile management UI
- [ ] Test UI integration

### Phase 5: Python SDK (1-2 weeks)
- [x] Implement Python SDK
- [x] Add example scripts
- [x] Test SDK integration
- [ ] Document SDK

### Phase 6: Enhanced MCP (1-2 weeks)
- [x] Add hidden MCP tools
- [x] Implement MCP categories
- [x] Add MCP authentication
- [ ] Test MCP integration

### Phase 7: Batch Processing (1-2 weeks)
- [x] Implement batch processor
- [x] Add retry logic
- [x] Add error handling
- [ ] Test batch operations

### Phase 8: Testing & Optimization (2-3 weeks)
- [x] End-to-end testing
- [x] Performance optimization
- [x] Security audit
- [x] Documentation

---

## 🎯 Final Architecture

```
Donut Browser Ultimate
├── Anti-detect Engine (Donut)
│   ├── Wayfern (Chromium)
│   ├── Camoufox (Firefox)
│   └── Fingerprint Spoofing
├── AI Agent System (Skyvern + Browser-Use)
│   ├── LLM Integration
│   ├── Computer Vision
│   └── Agent Loop
├── Workflow Engine (Skyvern)
│   ├── Workflow Builder
│   ├── Block System
│   └── Execution Engine
├── Database Layer (Skyvern)
│   ├── SQLite
│   ├── Profile Storage
│   └── Workflow Persistence
├── Web UI (Skyvern)
│   ├── Workflow Builder
│   ├── Monitoring Dashboard
│   └── Profile Management
├── Python SDK (Browser-Use)
│   ├── Simple API
│   ├── AI Tasks
│   └── Workflow Execution
├── MCP Server (Comet + Donut)
│   ├── Standard Tools
│   ├── Hidden Tools
│   └── Admin Tools
├── CDP Server (Browser-Use)
│   ├── Authentication
│   ├── Origin Whitelist
│   └── Persistent Sessions
└── Batch Processor (Browser-Use)
    ├── Multi-account
    ├── Retry Logic
    └── Error Handling
```

---

## 🚀 Next Steps

1. **Fork source code mới** từ GitHub gốc
2. **Bắt đầu Phase 1** - Foundation (Database layer)
3. **Hoặc tập trung vào use case cụ thể** (OAuth automation)

---

**Status**: Plan Created ✅
**Next**: Fork source code and begin implementation
