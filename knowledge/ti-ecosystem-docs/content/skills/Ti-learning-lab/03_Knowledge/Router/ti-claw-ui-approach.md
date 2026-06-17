# Ti Claw UI Approach

> **Purpose**: Đề xuất UI approach cho Ti Claw, cá nhân hóa và không phụ thuộc vào bản gốc
> **Date**: 2026-05-04
> **Status**: Design Phase

---

## 1. Ti Router UI Analysis

### Current Stack
```json
{
  "framework": "React 19 + Vite 6",
  "language": "TypeScript",
  "styling": "SCSS Modules",
  "state": "Zustand",
  "routing": "React Router DOM 7 (HashRouter)",
  "icons": "Lucide React",
  "animation": "Motion",
  "charts": "Chart.js + react-chartjs-2",
  "i18n": "i18next + react-i18next"
}
```

### Design System

**Layout**:
- Sidebar navigation (collapsible)
- Header with theme toggle, connection status, logout
- Main content area with Outlet

**Style**:
- Light/Dark theme toggle
- Unicode icons (◈, ⚙, ◉, ◫, ◐)
- Traffic lights (macOS window controls)
- Minimal, modern design
- SCSS modules for styling

**Pages**:
- Dashboard
- Config
- AI Providers
- Auth Files
- System

**Navigation**:
```
/ → Dashboard
/config → Config
/providers → AI Providers
/auth-files → Auth Files
/system → System
```

---

## 2. UI Options for Ti Claw

### Option 1: Reuse Router UI (Add Section)

**Approach**: Thêm Ti Claw section vào Router UI hiện tại

**Pros**:
- ✅ Không cần build UI mới
- ✅ Consistent design
- ✅ Shared state management
- ✅ Fast implementation

**Cons**:
- ❌ Phụ thuộc vào Router UI
- ❌ Không cá nhân hóa
- ❌ Coupling với Router logic
- ❌ Khó mở rộng cho Ti Claw riêng

**Verdict**: ❌ **KHÔNG CHỌN** - User muốn cá nhân hóa, không phụ thuộc

---

### Option 2: Separate UI with Shared Design System

**Approach**: Build UI riêng cho Ti Claw nhưng dùng shared design system với Router

**Pros**:
- ✅ Cá nhân hóa hoàn toàn
- ✅ Không phụ thuộc vào Router UI
- ✅ Có thể mở rộng riêng
- ✅ Consistent design với Router

**Cons**:
- ❌ Cần build UI mới
- ❌ Maintain design system sync
- ❌ Duplicate some code

**Verdict**: ✅ **CHỌN** - Balance giữa cá nhân hóa và consistency

---

### Option 3: Hybrid (Router + Ti Claw UI)

**Approach**: Router UI + Ti Claw UI với shared components library

**Pros**:
- ✅ Cá nhân hóa Ti Claw
- ✅ Shared components (Button, Card, Input, v.v.)
- ✅ Consistent design
- ✅ Mở rộng được

**Cons**:
- ❌ Cần build components library
- ❌ Complex hơn
- ❌ Maintain overhead

**Verdict**: ⚠️ **CONSIDER** - Nếu có time, có thể implement sau

---

## 3. Recommended Approach: Option 2

### Architecture

```
Ti Ecosystem
├── apps/
│   ├── router/
│   │   └── ui/                    # Router UI (existing)
│   │       ├── src/
│   │       ├── package.json
│   │       └── vite.config.ts
│   │
│   └── ticlaw/
│       ├── internal/              # Go backend
│       └── ui/                    # Ti Claw UI (NEW)
│           ├── src/
│           ├── package.json
│           └── vite.config.ts
│
└── packages/
    └── ti-ui-components/          # Shared design system (NEW)
        ├── src/
        │   ├── components/
        │   │   ├── Button.tsx
        │   │   ├── Card.tsx
        │   │   ├── Input.tsx
        │   │   ├── Modal.tsx
        │   │   ├── Sidebar.tsx
        │   │   └── ThemeProvider.tsx
        │   ├── styles/
        │   │   ├── variables.scss
        │   │   ├── mixins.scss
        │   │   └── themes.scss
        │   └── index.ts
        ├── package.json
        └── tsconfig.json
```

### Design System Specification

#### Theme Variables
```scss
// packages/ti-ui-components/src/styles/variables.scss

:root {
  // Colors - Light
  --color-bg-primary: #ffffff;
  --color-bg-secondary: #f5f5f5;
  --color-bg-tertiary: #e5e5e5;
  --color-text-primary: #1a1a1a;
  --color-text-secondary: #666666;
  --color-border: #e0e0e0;
  --color-accent: #3b82f6;
  --color-accent-hover: #2563eb;
  --color-success: #10b981;
  --color-warning: #f59e0b;
  --color-error: #ef4444;

  // Spacing
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 16px;
  --spacing-lg: 24px;
  --spacing-xl: 32px;

  // Typography
  --font-size-xs: 12px;
  --font-size-sm: 14px;
  --font-size-md: 16px;
  --font-size-lg: 18px;
  --font-size-xl: 24px;
  --font-size-2xl: 32px;

  // Border Radius
  --radius-sm: 4px;
  --radius-md: 8px;
  --radius-lg: 12px;
  --radius-xl: 16px;

  // Shadows
  --shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
  --shadow-md: 0 4px 6px rgba(0, 0, 0, 0.1);
  --shadow-lg: 0 10px 15px rgba(0, 0, 0, 0.1);
}

[data-theme='dark'] {
  --color-bg-primary: #1a1a1a;
  --color-bg-secondary: #2a2a2a;
  --color-bg-tertiary: #3a3a3a;
  --color-text-primary: #ffffff;
  --color-text-secondary: #a0a0a0;
  --color-border: #404040;
}
```

#### Component Examples

**Button**:
```tsx
// packages/ti-ui-components/src/components/Button.tsx
import styles from './Button.module.scss';

interface ButtonProps {
  variant?: 'primary' | 'secondary' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
  children: React.ReactNode;
  onClick?: () => void;
  disabled?: boolean;
}

export function Button({
  variant = 'primary',
  size = 'md',
  children,
  onClick,
  disabled
}: ButtonProps) {
  return (
    <button
      className={`${styles.button} ${styles[variant]} ${styles[size]}`}
      onClick={onClick}
      disabled={disabled}
    >
      {children}
    </button>
  );
}
```

**Card**:
```tsx
// packages/ti-ui-components/src/components/Card.tsx
import styles from './Card.module.scss';

interface CardProps {
  children: React.ReactNode;
  className?: string;
}

export function Card({ children, className }: CardProps) {
  return (
    <div className={`${styles.card} ${className || ''}`}>
      {children}
    </div>
  );
}
```

---

## 4. Ti Claw UI Structure

### Pages

```
Ti Claw UI
├── src/
│   ├── pages/
│   │   ├── DashboardPage.tsx      # Agent dashboard
│   │   ├── AgentsPage.tsx         # Agent list & management
│   │   ├── AgentDetailPage.tsx    # Agent detail & chat
│   │   ├── ToolsPage.tsx          # Tool registry
│   │   ├── MemoryPage.tsx         # Memory management
│   │   ├── SkillsPage.tsx         # Skills management
│   │   ├── ConfigPage.tsx         # Configuration
│   │   └── SettingsPage.tsx       # Settings
│   │
│   ├── components/
│   │   ├── layout/
│   │   │   ├── MainLayout.tsx     # Main layout (sidebar + header)
│   │   │   └── Sidebar.tsx       # Sidebar navigation
│   │   ├── agent/
│   │   │   ├── AgentCard.tsx      # Agent card component
│   │   │   ├── AgentChat.tsx      # Agent chat interface
│   │   │   ├── AgentRun.tsx       # Agent run interface
│   │   │   └── ToolCall.tsx       # Tool call display
│   │   ├── memory/
│   │   │   ├── MemoryGraph.tsx    # Knowledge graph visualization
│   │   │   └── MemoryTimeline.tsx # Memory timeline
│   │   └── ui/                    # (Imported from ti-ui-components)
│   │
│   ├── stores/
│   │   ├── useAgentStore.ts       # Agent state
│   │   ├── useMemoryStore.ts      # Memory state
│   │   └── useConfigStore.ts      # Config state
│   │
│   ├── services/
│   │   ├── api.ts                 # API client
│   │   ├── agentService.ts        # Agent API
│   │   └── memoryService.ts       # Memory API
│   │
│   └── App.tsx
```

### Navigation

```
/ → Dashboard
/agents → Agents
/agents/:id → Agent Detail
/tools → Tools
/memory → Memory
/skills → Skills
/config → Config
/settings → Settings
```

### Key Features

**Dashboard**:
- Agent status overview
- Recent agent runs
- Memory usage stats
- Quick actions

**Agents Page**:
- Agent list with cards
- Create new agent
- Edit agent config
- Delete agent

**Agent Detail Page**:
- Agent chat interface
- Tool execution history
- Memory graph
- Agent settings

**Tools Page**:
- Tool registry
- Tool status
- Tool configuration

**Memory Page**:
- Knowledge graph visualization
- Memory search
- Memory management

**Skills Page**:
- Skill list
- Skill search (BM25)
- Skill management

**Config Page**:
- LLM provider config
- Memory config
- Tool config
- Bootstrap config

**Settings Page**:
- Theme settings
- Language settings
- User preferences

---

## 5. Implementation Plan

### Phase 1: Design System (Week 1)

**Tasks**:
1. Create `packages/ti-ui-components/`
2. Define theme variables (SCSS)
3. Implement core components:
   - Button, Card, Input, Modal
   - Sidebar, Header, Layout
   - ThemeProvider
4. Document components

**Deliverables**:
- Shared design system
- Component library
- Documentation

---

### Phase 2: Ti Claw UI Skeleton (Week 2)

**Tasks**:
1. Create `apps/ticlaw/ui/`
2. Set up React + Vite + TypeScript
3. Implement MainLayout with Sidebar
4. Implement navigation
5. Integrate ti-ui-components
6. Set up API client

**Deliverables**:
- Ti Claw UI skeleton
- Navigation working
- API integration

---

### Phase 3: Core Pages (Week 3-4)

**Tasks**:
1. Implement DashboardPage
2. Implement AgentsPage
3. Implement AgentDetailPage
4. Implement ToolsPage
5. Implement MemoryPage

**Deliverables**:
- 5 core pages
- Agent management UI
- Tool registry UI
- Memory visualization

---

### Phase 4: Additional Pages (Week 5)

**Tasks**:
1. Implement SkillsPage
2. Implement ConfigPage
3. Implement SettingsPage
4. Add i18n support (Vietnamese)
5. Add theme toggle

**Deliverables**:
- 3 additional pages
- Vietnamese i18n
- Theme toggle

---

### Phase 5: Polish & Integration (Week 6)

**Tasks**:
1. Polish UI/UX
2. Add animations (Motion)
3. Add charts (Chart.js)
4. Integrate with Ti Claw backend
5. Test and fix bugs
6. Documentation

**Deliverables**:
- Production-ready UI
- Backend integration
- Complete documentation

---

## 6. Design Decisions

### Color Palette

**Ti Brand Colors**:
- Primary: `#3b82f6` (Blue)
- Secondary: `#8b5cf6` (Purple)
- Accent: `#10b981` (Green)
- Warning: `#f59e0b` (Orange)
- Error: `#ef4444` (Red)

**Dark Theme**:
- Background: `#1a1a1a`
- Surface: `#2a2a2a`
- Border: `#404040`

**Light Theme**:
- Background: `#ffffff`
- Surface: `#f5f5f5`
- Border: `#e0e0e0`

### Typography

**Font Family**: System fonts (San Francisco, Segoe UI, Roboto)

**Font Sizes**:
- Body: 14px (sm), 16px (md)
- Headings: 18px (lg), 24px (xl), 32px (2xl)
- Captions: 12px (xs)

### Spacing

**Scale**: 4px base
- xs: 4px
- sm: 8px
- md: 16px
- lg: 24px
- xl: 32px

### Border Radius

**Scale**:
- sm: 4px (buttons, inputs)
- md: 8px (cards)
- lg: 12px (modals)
- xl: 16px (containers)

### Shadows

**Scale**:
- sm: 0 1px 2px rgba(0, 0, 0, 0.05)
- md: 0 4px 6px rgba(0, 0, 0, 0.1)
- lg: 0 10px 15px rgba(0, 0, 0, 0.1)

---

## 7. Tech Stack

### Frontend
```json
{
  "framework": "React 19 + Vite 6",
  "language": "TypeScript",
  "styling": "SCSS Modules",
  "state": "Zustand",
  "routing": "React Router DOM 7 (HashRouter)",
  "icons": "Lucide React",
  "animation": "Motion",
  "charts": "Chart.js + react-chartjs-2",
  "i18n": "i18next + react-i18next"
}
```

### Backend API
```json
{
  "protocol": "HTTP + WebSocket",
  "format": "JSON",
  "authentication": "API Key (from Z:\\00_SECRET\\ticlaw.env)",
  "baseURL": "http://localhost:8080"
}
```

---

## 8. Integration with Ti Router UI

### Shared Components

Both Router UI and Ti Claw UI will use `packages/ti-ui-components/` for:
- Button, Card, Input, Modal
- Sidebar, Header, Layout
- ThemeProvider
- Utility components

### Consistent Design

Both UIs will share:
- Color palette
- Typography
- Spacing scale
- Border radius
- Shadows
- Theme system

### Separate Deployments

- Router UI: `http://localhost:5173` (existing)
- Ti Claw UI: `http://localhost:5174` (new)

---

## 9. Migration from Router UI

### Phase 1: Extract Design System
- Extract theme variables from Router UI
- Create shared components library
- Document design system

### Phase 2: Refactor Router UI
- Replace custom components with ti-ui-components
- Ensure no breaking changes
- Test thoroughly

### Phase 3: Build Ti Claw UI
- Use ti-ui-components
- Implement Ti Claw-specific pages
- Integrate with backend

---

## 10. Success Criteria

- [x] Design system defined
- [x] UI approach selected (Option 2)
- [ ] Shared components library created
- [ ] Ti Claw UI skeleton built
- [ ] Core pages implemented
- [ ] Backend integration working
- [ ] Theme toggle working
- [ ] Vietnamese i18n working
- [ ] Documentation complete
- [ ] Testing complete

---

## 11. Risks & Mitigations

### Risk 1: Design System Complexity
**Mitigation**: Start with minimal components, add incrementally

### Risk 2: UI/UX Inconsistency
**Mitigation**: Strict design system documentation, design reviews

### Risk 3: Backend Integration Issues
**Mitigation**: Mock API for development, integrate later

### Risk 4: Performance Issues
**Mitigation**: Code splitting, lazy loading, optimization

---

## 12. Next Steps

1. **Create design system** - `packages/ti-ui-components/`
2. **Extract Router UI components** - Move to shared library
3. **Build Ti Claw UI skeleton** - `apps/ticlaw/ui/`
4. **Implement core pages** - Dashboard, Agents, Tools, Memory
5. **Integrate with backend** - API integration
6. **Polish & test** - UI/UX polish, testing

---

**Status**: Design Complete ✅
**Next**: Implement Phase 1 (Design System)
