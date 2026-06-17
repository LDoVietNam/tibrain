# Ti Brain Scope Definitions

> **Created**: 2026-05-22
> **Purpose**: Define scopes that group related tags for broader search
> **Status**: 🔄 Draft

---

## 🎯 Scope Philosophy

**Scopes group related tags together for broader search queries.**

- **Tags are specific**: Tags describe specific aspects (e.g., `pattern-auth-pkce`)
- **Scopes are broad**: Scopes group related tags (e.g., `auth` includes all auth-related tags)
- **Multiple scopes per file**: A file can belong to multiple scopes
- **Scope-based search**: Search by scope returns all files with any tag in that scope

---

## 📋 Scope Definitions

### 1. auth

**Description**: All authentication and authorization related knowledge

**Tags Included**:
- `authentication`
- `authorization`
- `oauth`
- `oidc`
- `jwt`
- `api-key`
- `cookie`
- `pattern-auth-pkce`
- `pattern-auth-device-code`
- `pattern-auth-api-key`
- `pattern-auth-cookie`

**Use Cases**:
- Search for all auth patterns
- Find authentication implementations
- Learn about authorization strategies

**Example Files**:
- OAuth PKCE flow implementation
- API key authentication pattern
- Cookie-based authentication for providers

---

### 2. resilience

**Description**: Resilience patterns including rate limiting, retry, caching

**Tags Included**:
- `rate-limiting`
- `retry`
- `caching`
- `circuit-breaker`
- `pattern-retry-exponential`
- `pattern-retry-circuit-breaker`
- `pattern-cache-content`
- `pattern-cache-session`

**Use Cases**:
- Search for resilience patterns
- Find rate limit handling strategies
- Learn about caching strategies

**Example Files**:
- Exponential backoff retry pattern
- Content-based caching implementation
- Circuit breaker pattern

---

### 3. integration

**Description**: Integration patterns including translation, protocols

**Tags Included**:
- `translation`
- `mcp`
- `grpc`
- `sse`
- `websocket`
- `graphql`
- `rest`
- `http`
- `pattern-translation-hub-spoke`
- `pattern-translation-native-passthrough`

**Use Cases**:
- Search for integration patterns
- Find protocol implementations
- Learn about format translation

**Example Files**:
- Hub-and-spoke translation pattern
- MCP integration guide
- gRPC service implementation

---

### 4. observability

**Description**: Observability patterns including monitoring, logging, tracing

**Tags Included**:
- `monitoring`
- `logging`
- `tracing`
- `metrics`
- `pattern-monitoring-metrics`
- `pattern-monitoring-logging`
- `pattern-monitoring-tracing`

**Use Cases**:
- Search for observability patterns
- Find monitoring implementations
- Learn about logging strategies

**Example Files**:
- Structured logging pattern
- Metrics collection implementation
- Distributed tracing setup

---

### 5. infrastructure

**Description**: Infrastructure patterns including deployment, orchestration

**Tags Included**:
- `deployment`
- `docker`
- `kubernetes`
- `terraform`
- `pattern-deployment-blue-green`
- `pattern-deployment-canary`

**Use Cases**:
- Search for infrastructure patterns
- Find deployment strategies
- Learn about orchestration

**Example Files**:
- Docker containerization guide
- Kubernetes deployment pattern
- Terraform IaC patterns

---

### 6. providers

**Description**: All provider-specific knowledge

**Tags Included**:
- `provider-antigravity`
- `provider-openai`
- `provider-claude`
- `provider-gemini`
- `provider-deepseek`
- `provider-groq`
- `provider-openrouter`
- `provider-windsurf`
- `provider-notion`
- `provider-[name]` (all other provider tags)

**Use Cases**:
- Search for all provider knowledge
- Find specific provider implementations
- Compare provider patterns

**Example Files**:
- Antigravity MITM integration
- OpenAI API integration
- Claude OAuth flow

---

### 7. cli

**Description**: CLI-related knowledge including tools and languages

**Tags Included**:
- `cli`
- `cli-tools`
- `go`
- `typescript`
- `javascript`

**Use Cases**:
- Search for CLI patterns
- Find CLI tool implementations
- Learn about CLI languages

**Example Files**:
- CLI plugin architecture
- Go CLI patterns
- TypeScript CLI tools

---

### 8. tibrain

**Description**: TiBrain-related knowledge including memory and RAG

**Tags Included**:
- `tibrain`
- `memory`
- `rag`
- `cache`
- `database`

**Use Cases**:
- Search for TiBrain patterns
- Find memory implementations
- Learn about RAG systems

**Example Files**:
- TiBrain architecture
- Memory management patterns
- RAG implementation guide

---

### 9. ticrew

**Description**: Ticrew-related knowledge including agents and workflows

**Tags Included**:
- `ticrew`
- `agent`
- `workflow`
- `messaging`

**Use Cases**:
- Search for Ticrew patterns
- Find agent implementations
- Learn about workflow orchestration

**Example Files**:
- Ticrew agent framework
- Workflow orchestration pattern
- Agent communication patterns

---

### 10. web

**Description**: Web-related knowledge including automation and networking

**Tags Included**:
- `web-automation`
- `browser-automation`
- `networking`

**Use Cases**:
- Search for web patterns
- Find automation implementations
- Learn about networking

**Example Files**:
- Web automation patterns
- Browser automation guide
- Network configuration

---

### 11. code

**Description**: Code-related knowledge including generation and analysis

**Tags Included**:
- `code-generation`
- `code-analysis`
- `testing`

**Use Cases**:
- Search for code patterns
- Find generation implementations
- Learn about testing strategies

**Example Files**:
- Code generation patterns
- Static code analysis
- Testing strategies

---

## 🎯 Scope Usage

### Adding Scopes to Files

When adding scopes to a file, consider:

1. **Primary scope**: What is the main domain? (e.g., `auth`, `resilience`)
2. **Secondary scopes**: What other domains are relevant? (e.g., `cli`, `providers`)
3. **Technology scope**: What technology is used? (e.g., `go`, `typescript`)

**Example**:
```yaml
---
title: "OAuth PKCE Flow for CLI"
tags: ["authentication", "oauth", "pkce", "go", "pattern-auth-pkce"]
scope: ["auth", "cli"]  # Primary: auth, Secondary: cli
---
```

### Searching by Scope

When searching by scope, you'll get all files with any tag in that scope.

**Example**:
- Search scope `auth` → Returns all files with authentication, oauth, jwt, etc.
- Search scope `resilience` → Returns all files with rate-limiting, retry, caching, etc.

---

## 📝 Scope Naming Convention

### Rules

1. **Use lowercase**: All scopes should be in lowercase (e.g., `auth`, not `Auth`)
2. **Use single words**: Scopes should be single words (e.g., `auth`, not `authentication-authorization`)
3. **Be broad**: Scopes should be broad categories, not specific patterns
4. **Avoid abbreviations**: Use full words (e.g., `infrastructure`, not `infra`)
5. **Keep scopes short**: Scopes should be concise (ideally < 15 characters)

### Examples

✅ **Correct**:
- `auth`
- `resilience`
- `integration`
- `observability`

❌ **Wrong**:
- `Auth` (capitalized)
- `authentication-authorization` (too long, hyphenated)
- `infra` (abbreviation)
- `auth-patterns` (too specific)

---

## 🔄 Scope Maintenance

### Adding New Scopes

1. Check if scope already exists
2. If not, define scope with included tags
3. Update TAGS.md if new tags are needed
4. Communicate change to team

### Deprecating Scopes

1. Mark scope as deprecated
2. Find all files using deprecated scope
3. Replace with new scope(s)
4. Remove from this document

### Scope Review Schedule

- **Quarterly**: Review scope definitions for relevance
- **As needed**: Add new scopes for new domains
- **As needed**: Deprecate unused scopes

---

## 📊 Scope Statistics

| Metric | Current | Target |
|--------|---------|--------|
| Total scopes defined | 11 | 15-20 |
| Scopes in use | 0 | 100% of defined scopes |
| Average scopes per file | 0 | 2-3 |
| Tags per scope | 3-11 | 3-10 |

---

## 🔗 Related Documents

- <ref_file file="Z:\10_WORKPLACE\Ti\Ti-learning-lab\04_Planning\TAGS.md" /> - Tag taxonomy
- <ref_file file="Z:\10_WORKPLACE\Ti\Ti-learning-lab\04_Planning\TAG_NAMING_CONVENTION.md" /> - Tag naming rules
- <ref_file file="Z:\10_WORKPLACE\Ti\Ti-learning-lab\04_Planning\TAG_MAPPING.md" /> - File → Tags mapping
- <ref_file file="Z:\10_WORKPLACE\Ti\Ti-learning-lab\04_Planning\tibrain-pattern-organization.md" /> - Organization plan

---

*Last Updated: 2026-05-22*
*Status: Draft - Ready for review*
