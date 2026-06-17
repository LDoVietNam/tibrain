# Ti Brain Tag Taxonomy

> **Created**: 2026-05-22
> **Purpose**: Define all tags for Ti Brain knowledge organization
> **Status**: 🔄 Draft

---

## 🎯 Tag Philosophy

**Tags are the primary organization mechanism for Ti Brain knowledge.**

- **Flat structure**: Files in `knowledge/` directory, organized by tags (not folders)
- **Multiple tags per file**: Each file can have multiple tags
- **Scopes group tags**: Related tags are grouped into scopes for broader search
- **Tag taxonomy**: Defined here to ensure consistency

---

## 📋 Tag Categories

### 1. Core Tags (High-Level)

Tags that describe the primary domain or concern of the knowledge.

| Tag | Description | Examples |
|-----|-------------|----------|
| `authentication` | Authentication mechanisms and flows | OAuth, API key, Cookie, JWT |
| `authorization` | Authorization and permissions | RBAC, scopes, permissions |
| `rate-limiting` | Rate limit handling and strategies | 429, retry-after, backoff |
| `caching` | Caching strategies and implementations | Content cache, session cache |
| `translation` | Format translation between AI providers | OpenAI ↔ Claude, Gemini ↔ OpenAI |
| `retry` | Retry logic and patterns | Exponential backoff, circuit breaker |
| `monitoring` | Monitoring and observability | Metrics, logging, tracing |
| `deployment` | Deployment patterns and strategies | Docker, K8s, bare metal |
| `troubleshooting` | Troubleshooting guides and common issues | Debug steps, error resolution |
| `security` | Security patterns and best practices | Input validation, secrets management |
| `performance` | Performance optimization patterns | Latency reduction, throughput |
| `testing` | Testing strategies and patterns | Unit tests, integration tests |
| `documentation` | Documentation patterns and practices | API docs, architecture docs |

---

### 2. Provider Tags

Tags specific to AI model providers.

| Tag | Description |
|-----|-------------|
| `provider-antigravity` | Antigravity provider specific knowledge |
| `provider-openai` | OpenAI provider specific knowledge |
| `provider-claude` | Claude (Anthropic) provider specific knowledge |
| `provider-gemini` | Gemini (Google) provider specific knowledge |
| `provider-deepseek` | DeepSeek provider specific knowledge |
| `provider-groq` | Groq provider specific knowledge |
| `provider-openrouter` | OpenRouter provider specific knowledge |
| `provider-windsurf` | Windsurf provider specific knowledge |
| `provider-notion` | Notion provider specific knowledge |
| `provider-[name]` | Other providers (follow pattern: provider-{name}) |

---

### 3. Technology Tags

Tags that describe the technology stack or language.

| Tag | Description |
|-----|-------------|
| `go` | Go language |
| `typescript` | TypeScript/JavaScript |
| `javascript` | JavaScript |
| `python` | Python |
| `rust` | Rust |
| `java` | Java |
| `mcp` | Model Context Protocol |
| `oauth` | OAuth protocol |
| `oidc` | OpenID Connect |
| `jwt` | JSON Web Tokens |
| `sse` | Server-Sent Events |
| `grpc` | gRPC protocol |
| `http` | HTTP protocol |
| `websocket` | WebSocket protocol |
| `graphql` | GraphQL |
| `rest` | REST API |
| `sql` | SQL databases |
| `nosql` | NoSQL databases |
| `docker` | Docker containers |
| `kubernetes` | Kubernetes orchestration |
| `terraform` | Terraform IaC |

---

### 4. Component Tags

Tags that describe which Ti component the knowledge relates to.

| Tag | Description |
|-----|-------------|
| `router` | Router component |
| `cli` | CLI component |
| `tibrain` | TiBrain component |
| `ticrew` | Ticrew component |
| `mcp-server` | MCP server |
| `provider` | Provider system |
| `plugin` | Plugin system |
| `skill` | Skill system |
| `workflow` | Workflow system |
| `agent` | Agent system |
| `dashboard` | Dashboard UI |
| `api` | API layer |
| `database` | Database layer |
| `cache` | Cache layer |

---

### 5. Pattern Tags

Tags that describe specific patterns or best practices.

| Tag | Description |
|-----|-------------|
| `pattern-auth-pkce` | PKCE OAuth pattern |
| `pattern-auth-device-code` | Device code OAuth pattern |
| `pattern-auth-api-key` | API key authentication pattern |
| `pattern-auth-cookie` | Cookie-based authentication pattern |
| `pattern-retry-exponential` | Exponential backoff retry pattern |
| `pattern-retry-circuit-breaker` | Circuit breaker pattern |
| `pattern-cache-content` | Content-based caching pattern |
| `pattern-cache-session` | Session caching pattern |
| `pattern-translation-hub-spoke` | Hub-and-spoke translation pattern |
| `pattern-translation-native-passthrough` | Native passthrough pattern |
| `pattern-monitoring-metrics` | Metrics collection pattern |
| `pattern-monitoring-logging` | Structured logging pattern |
| `pattern-monitoring-tracing` | Distributed tracing pattern |
| `pattern-deployment-blue-green` | Blue-green deployment pattern |
| `pattern-deployment-canary` | Canary deployment pattern |
| `pattern-testing-tdd` | Test-driven development pattern |
| `pattern-testing-bdd` | Behavior-driven development pattern |

---

### 6. Domain Tags

Tags that describe the business domain or use case.

| Tag | Description |
|-----|-------------|
| `cli-tools` | CLI tools and utilities |
| `web-automation` | Web automation and scraping |
| `browser-automation` | Browser automation |
| `code-generation` | Code generation patterns |
| `code-analysis` | Code analysis and linting |
| `documentation` | Documentation generation |
| `integration` | System integration patterns |
| `messaging` | Messaging and queues |
| `storage` | Storage patterns |
| `networking` | Networking patterns |

---

## 🎯 Scope Definitions

**Scopes group related tags together for broader search.**

| Scope | Tags Included | Description |
|-------|---------------|-------------|
| `auth` | `authentication`, `authorization`, `oauth`, `oidc`, `jwt`, `api-key`, `cookie`, `pattern-auth-*` | All authentication and authorization |
| `resilience` | `rate-limiting`, `retry`, `caching`, `circuit-breaker`, `pattern-retry-*`, `pattern-cache-*` | Resilience patterns |
| `integration` | `translation`, `mcp`, `grpc`, `sse`, `websocket`, `graphql`, `rest`, `pattern-translation-*` | Integration patterns |
| `observability` | `monitoring`, `logging`, `tracing`, `metrics`, `pattern-monitoring-*` | Observability patterns |
| `infrastructure` | `deployment`, `docker`, `kubernetes`, `terraform`, `pattern-deployment-*` | Infrastructure patterns |
| `providers` | `provider-*` (all provider tags) | All provider-specific knowledge |
| `cli` | `cli`, `cli-tools`, `go`, `typescript`, `javascript` | CLI-related knowledge |
| `tibrain` | `tibrain`, `memory`, `rag`, `cache`, `database` | TiBrain-related knowledge |
| `ticrew` | `ticrew`, `agent`, `workflow`, `messaging` | Ticrew-related knowledge |
| `web` | `web-automation`, `browser-automation`, `networking` | Web-related knowledge |
| `code` | `code-generation`, `code-analysis`, `testing` | Code-related knowledge |

---

## 📝 Tag Naming Convention

### Rules

1. **Use kebab-case**: All tags should be in kebab-case (e.g., `rate-limiting`, not `rateLimiting`)
2. **Be specific**: Use specific tags instead of generic ones (e.g., `pattern-auth-pkce` instead of just `auth`)
3. **Use prefixes for pattern tags**: Pattern tags should start with `pattern-` (e.g., `pattern-auth-pkce`)
4. **Use prefixes for provider tags**: Provider tags should start with `provider-` (e.g., `provider-antigravity`)
5. **Avoid abbreviations**: Use full words instead of abbreviations (e.g., `authentication` instead of `auth`)
6. **Keep tags short**: Tags should be concise but descriptive (ideally < 20 characters)

### Examples

✅ **Correct**:
- `rate-limiting`
- `pattern-auth-pkce`
- `provider-antigravity`
- `caching-strategy`

❌ **Wrong**:
- `rateLimiting` (camelCase)
- `pkce` (too generic)
- `antigravity` (missing provider- prefix)
- `auth` (abbreviation, too generic)

---

## 🔄 Tag Maintenance

### Adding New Tags

1. Check if tag already exists in this taxonomy
2. If not, add to appropriate category
3. Update scope definitions if needed
4. Update TAG_MAPPING.md
5. Communicate change to team

### Deprecating Tags

1. Mark tag as deprecated in this taxonomy
2. Find all files using deprecated tag
3. Replace with new tag(s)
4. Update TAG_MAPPING.md
4. Remove from scope definitions

### Tag Review Schedule

- **Quarterly**: Review tag taxonomy for relevance
- **As needed**: Add new tags for new domains
- **As needed**: Deprecate unused tags

---

## 📊 Tag Statistics

| Metric | Current | Target |
|--------|---------|--------|
| Total tags defined | 90 | 100+ |
| Tags in use | 34 | 100% of defined tags |
| Files with tags | 92 (8.8%) | 100% |
| Average tags per file | 4.9 | 3-5 |

---

## 🔗 Related Documents

- <ref_file file="Z:\10_WORKPLACE\Ti\Ti-learning-lab\04_Planning\SCOPES.md" /> - Scope definitions
- <ref_file file="Z:\10_WORKPLACE\Ti\Ti-learning-lab\04_Planning\TAG_NAMING_CONVENTION.md" /> - Tag naming rules
- <ref_file file="Z:\10_WORKPLACE\Ti\Ti-learning-lab\04_Planning\TAG_MAPPING.md" /> - File → Tags mapping
- <ref_file file="Z:\10_WORKPLACE\Ti\Ti-learning-lab\04_Planning\tibrain-pattern-organization.md" /> - Organization plan

---

*Last Updated: 2026-05-22*
*Status: Draft - Ready for review*