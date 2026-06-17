# UI Learning Notes — CLI Proxy API Management Center

> Source: `Z:\Ti\SJ-learning-lab\06_Learning\Cli-Proxy-API-Management-Center-main`
> Purpose: Document UI architecture & API contract to guide Go backend porting

---

## 1. Tech Stack

| Layer | Technology |
|-------|-----------|
| Framework | React 19 + TypeScript |
| Bundler | Vite 7 (`vite-plugin-singlefile` → single-file HTML output) |
| Router | `react-router-dom` (HashRouter) |
| State | Zustand 5 + persist middleware + `obfuscatedStorage` |
| Styling | SCSS Modules + CSS variables for theming |
| i18n | `react-i18next` (language: zh-CN, en, etc.) |
| Charts | `chart.js` + `react-chartjs-2` |
| Editor | CodeMirror (`@uiw/react-codemirror`, `@codemirror/lang-yaml`) |
| HTTP | Axios-based custom `apiClient` |
| Animations | `motion` (Framer Motion successor) |

**Key build config** (`vite.config.ts`):
- `assetsInlineLimit: 100000000` — inline all assets into single HTML
- `cssCodeSplit: false` — bundle all CSS into one file
- `inlineDynamicImports: true` — no lazy chunks

---

## 2. State Management (Zustand)

### Store slices

| Store | Key State | Persistence |
|-------|-----------|-------------|
| `useAuthStore` | `apiBase`, `managementKey`, `isAuthenticated`, `connectionStatus`, `serverVersion`, `serverBuildDate` | ✅ `STORAGE_KEY_AUTH` with `obfuscatedStorage` |
| `useConfigStore` | `config`, loading/error flags | ⚠️ In-memory + manual fetch |
| `useThemeStore` | `theme` (`auto`/`white`/`light`/`dark`) | ✅ localStorage |
| `useLanguageStore` | `language` | ✅ localStorage |
| `useModelsStore` | `models[]`, loading | ⚠️ In-memory |
| `useQuotaStore` | quota data | ⚠️ In-memory |
| `useUsageStatsStore` | usage statistics | ⚠️ In-memory |
| `useNotificationStore` | toast queue | ❌ ephemeral |

### Auth persistence logic (`useAuthStore.ts`)

```
1. On mount: call restoreSession()
2. restoreSession reads obfuscatedStorage for apiBase + managementKey
3. If wasLoggedIn && has credentials → auto-login via checkAuth()
4. checkAuth calls fetchConfig() to verify server is alive
5. On success: set connectionStatus='connected', isAuthenticated=true
6. On failure: set connectionStatus='error', logout()
```

**Go backend implication:** Need `/config` endpoint that returns 200 when auth key is valid.

---

## 3. API Client Contract

### Base configuration
- `apiBase`: dynamic, set via login form or restored from storage
- `managementKey`: sent as `Authorization: Bearer <key>` header
- All endpoints prefixed with `apiBase` (no trailing slash normalization issues handled in client)

### Endpoints consumed by UI

#### Config
- `GET /config` → full config object (normalization applied in UI)
- `PUT /debug` → `{value: boolean}`
- `PUT /proxy-url` → `{value: string}`
- `DELETE /proxy-url`
- `PUT /request-retry` → `{value: number}`
- `PUT /quota-exceeded/switch-project` → `{value: boolean}`
- `PUT /quota-exceeded/switch-preview-model` → `{value: boolean}`
- `PUT /usage-statistics-enabled` → `{value: boolean}`
- `PUT /request-log` → `{value: boolean}`
- `PUT /logging-to-file` → `{value: boolean}`
- `GET /logs-max-total-size-mb` → `{value: number}`
- `PUT /logs-max-total-size-mb` → `{value: number}`
- `PUT /ws-auth` → `{value: boolean}`
- `GET /force-model-prefix` → `{value: boolean}`
- `PUT /force-model-prefix` → `{value: boolean}`
- `GET /routing/strategy` → `{strategy: string}`
- `PUT /routing/strategy` → `{value: string}`

#### Auth / Connection
- `GET /version` → server version string
- Login is purely client-side: store credentials, test with `/config`

#### AI Providers (dashboard stats)
- `GET /api/providers/gemini-keys` → array
- `GET /api/providers/codex-configs` → array
- `GET /api/providers/claude-configs` → array
- `GET /api/providers/openai` → array

#### API Keys
- `GET /api/keys` → array of key objects

#### Auth Files
- `GET /api/auth-files` → `{files: [...]}`

#### Models
- `GET /api/models` → array (used for dashboard "available models" count)

#### Usage
- `GET /api/usage` → usage statistics

#### Logs
- `GET /api/logs` → log entries (with pagination/search)

#### OAuth
- `GET /api/oauth/...` → OAuth flows

---

## 4. Page Structure

| Route | Page | Features |
|-------|------|----------|
| `/login` | `LoginPage` | API base + management key input, remember me |
| `/` (dashboard) | `DashboardPage` | Quick stats, greeting, config pills, connection status |
| `/config` | `ConfigPage` | All settings: debug, proxy, retry, routing strategy, toggles |
| `/ai-providers` | `AiProvidersPage` | Provider list, add/edit/delete |
| `/ai-providers/gemini` | `AiProvidersGeminiEditPage` | Gemini-specific config |
| `/ai-providers/codex` | `AiProvidersCodexEditPage` | Codex config |
| `/ai-providers/claude` | `AiProvidersClaudeEditPage` | Claude config |
| `/ai-providers/openai` | `AiProvidersOpenAIEditPage` | OpenAI config |
| `/ai-providers/vertex` | `AiProvidersVertexEditPage` | Vertex AI config |
| `/ai-providers/ampcode` | `AiProvidersAmpcodeEditPage` | Ampcode config |
| `/auth-files` | `AuthFilesPage` | Cookie/auth file management |
| `/auth-files/oauth-excluded` | `AuthFilesOAuthExcludedEditPage` | OAuth exclusion rules |
| `/auth-files/oauth-model-alias` | `AuthFilesOAuthModelAliasEditPage` | Model aliasing |
| `/oauth` | `OAuthPage` | OAuth callback/import flows |
| `/quota` | `QuotaPage` | Quota management |
| `/usage` | `UsagePage` | Usage statistics with charts |
| `/logs` | `LogsPage` | Log viewer (conditional: only if `config.loggingToFile`) |
| `/system` | `SystemPage` | System info, available models |

**Conditional nav item:** `/logs` only appears in sidebar when `config.loggingToFile === true`.

---

## 5. Layout Architecture (`MainLayout.tsx`)

### Structure
```
AppShell
├── TopGradientBlur (decorative)
├── MainHeader
│   ├── SidebarToggle (collapsible)
│   ├── MobileMenuBtn
│   ├── RefreshAllBtn
│   ├── LanguageMenu (dropdown with Escape/click-outside)
│   ├── ThemeMenu (visual theme cards: auto/white/light/dark)
│   └── LogoutBtn
├── MainBody
│   ├── SidebarBackdrop (mobile overlay)
│   ├── Sidebar
│   │   ├── Brand (logo + abbr name)
│   │   └── NavSection (NavLink items)
│   └── Content
│       └── MainContent
│           └── PageTransition + MainRoutes
```

### Responsive behavior
- Desktop: sidebar can be collapsed (chevron toggle), labels hide when collapsed
- Mobile: sidebar is overlay drawer, hamburger menu button
- `ResizeObserver` on header → writes `--header-height` CSS variable
- `ResizeObserver` on content → writes `--content-center-x` CSS variable (used by floating panels)

### Page transitions
- `PageTransition` component receives `getRouteOrder` and `getTransitionVariant`
- Route order determines slide direction (`vertical` or `ios`)
- `ios` transition used for nested routes under `/ai-providers` and `/auth-files`

---

## 6. Theming System

Four themes defined in `MainLayout.tsx`:

| Theme | Background | Card | Border | Text |
|-------|-----------|------|--------|------|
| `auto` | Split gradient white/black | Split gradient | `#bdbdbd` | `#2d2a26` |
| `white` | `#ffffff` | `#ffffff` | `#e5e5e5` | `#2d2a26` |
| `light` | `#faf9f5` | `#f0eee8` | `#e3e1db` | `#2d2a26` |
| `dark` | `#151412` | `#1d1b18` | `#3a3530` | `#f6f4f1` |

Theme applied via CSS classes on app-shell; SCSS variables consumed by all pages.

---

## 7. Dashboard Patterns (`DashboardPage.tsx`)

### QuickStats (Bento Grid)
4 cards linking to respective pages:
1. **Management Keys** → `/config` (count from `apiKeysApi.list()`)
2. **AI Providers** → `/ai-providers` (aggregated gemini+codex+claude+openai counts)
3. **Auth Files** → `/auth-files` (count from `authFilesApi.list()`)
4. **Available Models** → `/system` (count from `modelsStore`)

### Provider Stats Logic
```typescript
Promise.allSettled([
  apiKeysApi.list(),
  authFilesApi.list(),
  providersApi.getGeminiKeys(),
  providersApi.getCodexConfigs(),
  providersApi.getClaudeConfigs(),
  providersApi.getOpenAIProviders()
])
```

Each call is independent; failures show `-` instead of crashing.

### Greeting System
```typescript
type TimeOfDay = 'morning' | 'afternoon' | 'evening' | 'night';
// Updates every 60 seconds via setInterval
// i18n keys: dashboard.greeting_morning, dashboard.caring_morning, etc.
```

### Connection Pill
- Shows server version (stripped leading `v/V`)
- Shows build date (localized)
- Status dot color: green=connected, blue=connecting, red=error

---

## 8. Auth Flow

```
[Login Page]
  |
  v
Enter apiBase + managementKey
  |
  v
POST (test) → fetchConfig(force=true)
  |
  v
Success → isAuthenticated=true, connectionStatus='connected'
  |         → persist to obfuscatedStorage (if rememberPassword)
  |
  v
[MainLayout + Dashboard]
  |
  v
On page load → restoreSession()
  → Read from obfuscatedStorage
  → If credentials exist → checkAuth() → fetchConfig()
  → If valid → stay logged in
  → If invalid → redirect /login
```

**Go backend needs:**
- `GET /config` must be protected (return 401 if key invalid)
- `GET /version` can be public (used to display server version even before full auth)

---

## 9. Code Quality Patterns

- **SCSS Modules**: Every page has `PageName.module.scss` (no global CSS conflicts)
- **CSS Variables**: `--header-height`, `--content-center-x` for responsive calculations
- **Icon System**: Custom SVG icons as React components, not icon font/library
- **Notification Toast**: Global store (`useNotificationStore`) with auto-dismiss
- **Draft Stores**: Separate Zustand stores for edit drafts (`useClaudeEditDraftStore`, `useOpenAIEditDraftStore`) to avoid polluting main config store
- **Transformers Layer**: `transformers.ts` normalizes API responses (snake_case ↔ camelCase, field aliases)
- **Type Safety**: Full TypeScript coverage, `unknown` parsing with runtime validation

---

## 10. Implications for Go Backend

### Required endpoints (minimum viable)
| Priority | Endpoint | Purpose |
|----------|----------|---------|
| P0 | `GET /config` | Auth validation + all settings |
| P0 | `GET /version` | Server version display |
| P0 | `PUT /config/*` | Toggle settings |
| P1 | `GET /api/keys` | Dashboard stat |
| P1 | `GET /api/auth-files` | Dashboard stat |
| P1 | `GET /api/providers/*` | Provider stats |
| P1 | `GET /api/models` | Model list |
| P2 | `GET /api/usage` | Usage page |
| P2 | `GET /api/logs` | Logs page |
| P2 | `GET /api/oauth/*` | OAuth flows |

### Data shape expectations
All JSON responses should be objects (not top-level arrays for security). UI expects:
```json
{
  "value": <primitive>,           // for toggle endpoints
  "strategy": "round-robin",     // for routing strategy
  "files": [...],                // for auth-files
  "models": [...]                // for models list
}
```

The UI applies `normalizeConfigResponse()` to handle legacy field names and case inconsistencies. Go backend can emit clean camelCase; UI will adapt.

---

## 11. Comparison with `noderouter.disabled`

| Aspect | noderouter (Node.js) | This UI |
|--------|---------------------|---------|
| Router | Next.js App Router | React HashRouter |
| State | Redux / Context | Zustand |
| Styling | Tailwind + CSS-in-JS | SCSS Modules |
| Build | Next.js static export | Vite single-file |
| Auth | Session/cookie | API key (Bearer) |

The UI is **completely decoupled** from backend tech. It only needs a REST API. This makes Go porting straightforward — we don't need to match Node.js internals, just implement the API contract.

---

*Last updated: 2026-04-26*
