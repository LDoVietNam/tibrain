# Skyvern Forge App Architecture

**Repository**: Skyvern  
**Location**: `skyvern/forge/forge_app.py` (284 lines)  
**Related Files**: `forge/agent.py`, `forge/agent_functions.py`

---

## Overview

Skyvern's Forge App is a dependency injection container that initializes and manages all shared services for the application. It follows a singleton pattern with comprehensive service configuration including database, storage, cache, LLM handlers, and credential vault services.

---

## ForgeApp Container

**Location**: `forge/forge_app.py` (lines 45-95)

### Core Services

```python
class ForgeApp:
    SETTINGS_MANAGER: Settings
    DATABASE: AgentDB
    REPLICA_DATABASE: AgentDB
    STORAGE: BaseStorage
    CACHE: BaseCache
    ARTIFACT_MANAGER: ArtifactManager
    BROWSER_MANAGER: BrowserManager
    EXPERIMENTATION_PROVIDER: BaseExperimentationProvider
    RATE_LIMITER: RateLimiter
```

### LLM Services

```python
    LLM_API_HANDLER: LLMAPIHandler
    OPENAI_CLIENT: AsyncOpenAI | AsyncAzureOpenAI
    OPENAI_CUA_MODEL: str
    ANTHROPIC_CLIENT: AsyncAnthropic | AsyncAnthropicBedrock
    UI_TARS_CLIENT: AsyncOpenAI | None
    AZURE_CLIENT_FACTORY: AzureClientFactory
```

### Specialized LLM Handlers (20+)

Skyvern uses **20+ specialized LLM handlers** for different purposes:

```python
    SECONDARY_LLM_API_HANDLER: LLMAPIHandler
    SELECT_AGENT_LLM_API_HANDLER: LLMAPIHandler
    NORMAL_SELECT_AGENT_LLM_API_HANDLER: LLMAPIHandler
    CUSTOM_SELECT_AGENT_LLM_API_HANDLER: LLMAPIHandler
    SINGLE_CLICK_AGENT_LLM_API_HANDLER: LLMAPIHandler
    SINGLE_INPUT_AGENT_LLM_API_HANDLER: LLMAPIHandler
    PARSE_SELECT_LLM_API_HANDLER: LLMAPIHandler
    EXTRACTION_LLM_API_HANDLER: LLMAPIHandler
    CHECK_USER_GOAL_LLM_API_HANDLER: LLMAPIHandler
    AUTO_COMPLETION_LLM_API_HANDLER: LLMAPIHandler
    SVG_CSS_CONVERTER_LLM_API_HANDLER: LLMAPIHandler | None
    SCRIPT_GENERATION_LLM_API_HANDLER: LLMAPIHandler
    SCRIPT_REVIEWER_LLM_API_HANDLER: LLMAPIHandler
    ADAPTIVE_SCRIPT_GEN_LLM_API_HANDLER: LLMAPIHandler
    WORKFLOW_COPILOT_AGENT_LLM_API_HANDLER: LLMAPIHandler
    WORKFLOW_COPILOT_FAST_LLM_API_HANDLER: LLMAPIHandler
```

### Workflow Services

```python
    WORKFLOW_CONTEXT_MANAGER: WorkflowContextManager
    WORKFLOW_SERVICE: WorkflowService
    AGENT_FUNCTION: AgentFunction
    PERSISTENT_SESSIONS_MANAGER: PersistentSessionsManager
    BROWSER_SESSION_RECORDING_SERVICE: BrowserSessionRecordingService
```

### Credential Services

```python
    BITWARDEN_CREDENTIAL_VAULT_SERVICE: BitwardenCredentialVaultService
    AZURE_CREDENTIAL_VAULT_SERVICE: AzureCredentialVaultService | None
    CUSTOM_CREDENTIAL_VAULT_SERVICE: CustomCredentialVaultService | None
    CREDENTIAL_VAULT_SERVICES: dict[str, CredentialVaultService | None]
```

### Extension Points

```python
    scrape_exclude: ScrapeExcludeFunc | None
    authentication_function: Callable[[str], Awaitable[Organization]] | None
    authenticate_user_function: Callable[[str], Awaitable[str | None]] | None
    setup_api_app: Callable[[FastAPI], None] | None
    api_app_startup_event: Callable[[FastAPI], Awaitable[None]] | None
    api_app_shutdown_event: Callable[[], Awaitable[None]] | None
    agent: ForgeAgent
```

---

## Initialization Pattern

**Location**: `forge/forge_app.py` (lines 97-284)

### create_forge_app()

```python
def create_forge_app() -> ForgeApp:
    """Create and initialize a ForgeApp instance with all services"""
    settings: Settings = SettingsManager.get_settings()
    app = ForgeApp()
    
    # Initialize core services
    app.DATABASE = AgentDB(settings.DATABASE_STRING, debug_enabled=settings.DEBUG_MODE)
    app.REPLICA_DATABASE = AgentDB(settings.DATABASE_REPLICA_STRING, ...)
    
    # Storage backend selection
    if settings.SKYVERN_STORAGE_TYPE == "s3":
        StorageFactory.set_storage(S3Storage())
    elif settings.SKYVERN_STORAGE_TYPE == "azureblob":
        StorageFactory.set_storage(AzureStorage())
    app.STORAGE = StorageFactory.get_storage()
    app.CACHE = CacheFactory.get_cache()
    
    # Artifact and browser management
    app.ARTIFACT_MANAGER = ArtifactManager()
    app.BROWSER_MANAGER = RealBrowserManager()
    
    # LLM handlers with fallback logic
    app.LLM_API_HANDLER = LLMAPIHandlerFactory.get_llm_api_handler(settings.LLM_KEY)
    app.SECONDARY_LLM_API_HANDLER = LLMAPIHandlerFactory.get_llm_api_handler(
        settings.SECONDARY_LLM_KEY or settings.LLM_KEY
    )
    
    # 20+ specialized LLM handlers with fallback logic
    app.SELECT_AGENT_LLM_API_HANDLER = LLMAPIHandlerFactory.get_llm_api_handler(
        settings.SELECT_AGENT_LLM_KEY or settings.SECONDARY_LLM_KEY or settings.LLM_KEY
    )
    
    # ... (similar pattern for all 20+ handlers)
    
    # Workflow services
    app.WORKFLOW_CONTEXT_MANAGER = WorkflowContextManager()
    app.WORKFLOW_SERVICE = WorkflowService()
    app.AGENT_FUNCTION = AgentFunction()
    
    # Persistent sessions
    app.PERSISTENT_SESSIONS_MANAGER = DefaultPersistentSessionsManager(database=app.DATABASE)
    app.PERSISTENT_SESSIONS_MANAGER.watch_session_pool()
    
    # Credential vault services
    app.BITWARDEN_CREDENTIAL_VAULT_SERVICE = BitwardenCredentialVaultService()
    
    # Azure with workload identity support
    if settings.AZURE_CREDENTIAL_VAULT:
        if settings.AZURE_CLIENT_SECRET:
            azure_vault_client = app.AZURE_CLIENT_FACTORY.create_from_client_secret(...)
        else:
            azure_vault_client = app.AZURE_CLIENT_FACTORY.create_default()
        app.AZURE_CREDENTIAL_VAULT_SERVICE = AzureCredentialVaultService(...)
    
    app.agent = ForgeAgent()
    
    return app
```

---

## Key Patterns

### 1. Multi-LLM Handler Pattern

**Pattern**: 20+ specialized LLM handlers for different purposes

**Benefits**:
- **Cost optimization**: Use cheaper models for simple tasks
- **Performance**: Use faster models for latency-sensitive tasks
- **Quality**: Use best models for critical tasks
- **Flexibility**: Easy to swap models per use case

**Handler Categories**:
- **Selection**: SELECT_AGENT, NORMAL_SELECT_AGENT, CUSTOM_SELECT_AGENT
- **Single Action**: SINGLE_CLICK_AGENT, SINGLE_INPUT_AGENT
- **Extraction**: EXTRACTION_LLM_API_HANDLER
- **Verification**: CHECK_USER_GOAL_LLM_API_HANDLER
- **Auto-completion**: AUTO_COMPLETION_LLM_API_HANDLER
- **Script Generation**: SCRIPT_GENERATION_LLM_API_HANDLER, ADAPTIVE_SCRIPT_GEN_LLM_API_HANDLER
- **Script Review**: SCRIPT_REVIEWER_LLM_API_HANDLER
- **Copilot**: WORKFLOW_COPILOT_AGENT_LLM_API_HANDLER, WORKFLOW_COPILOT_FAST_LLM_API_HANDLER

### 2. Fallback Logic Pattern

**Pattern**: Chain of fallback for each handler

```python
app.SELECT_AGENT_LLM_API_HANDLER = LLMAPIHandlerFactory.get_llm_api_handler(
    settings.SELECT_AGENT_LLM_KEY or settings.SECONDARY_LLM_KEY or settings.LLM_KEY
)
```

**Benefits**:
- **Reliability**: Fallback if primary model unavailable
- **Cost optimization**: Use cheaper secondary model
- **Flexibility**: Easy to reconfigure

### 3. Storage Backend Factory Pattern

**Pattern**: Factory pattern for storage backend selection

```python
if settings.SKYVERN_STORAGE_TYPE == "s3":
    StorageFactory.set_storage(S3Storage())
elif settings.SKYVERN_STORAGE_TYPE == "azureblob":
    StorageFactory.set_storage(AzureStorage())
app.STORAGE = StorageFactory.get_storage()
```

**Benefits**:
- **Backend flexibility**: Easy to switch storage backends
- **Configuration-driven**: Environment-based selection
- **Consistent API**: Same interface regardless of backend

### 4. Credential Vault Service Pattern

**Pattern**: Multiple credential vault services with unified interface

```python
app.CREDENTIAL_VAULT_SERVICES = {
    CredentialVaultType.BITWARDEN: app.BITWARDEN_CREDENTIAL_VAULT_SERVICE,
    CredentialVaultType.AZURE_VAULT: app.AZURE_CREDENTIAL_VAULT_SERVICE,
    CredentialVaultType.CUSTOM: app.CUSTOM_CREDENTIAL_VAULT_SERVICE,
}
```

**Benefits**:
- **Multi-vault support**: Bitwarden, Azure, custom
- **Unified interface**: Consistent API across vaults
- **Extensibility**: Easy to add new vault types

### 5. Azure Workload Identity Pattern

**Pattern**: Support both explicit credentials and workload identity

```python
if settings.AZURE_CLIENT_SECRET:
    azure_vault_client = app.AZURE_CLIENT_FACTORY.create_from_client_secret(...)
else:
    azure_vault_client = app.AZURE_CLIENT_FACTORY.create_default()
```

**Benefits**:
- **Cloud-native**: Workload identity for AKS/Azure VMs
- **Flexibility**: Explicit credentials for development
- **Security**: No secrets in code for production

---

## Extension Points

### Authentication Extension

```python
authentication_function: Callable[[str], Awaitable[Organization]] | None
authenticate_user_function: Callable[[str], Awaitable[str | None]] | None
```

**Purpose**: Allow custom authentication logic

### API App Extension

```python
setup_api_app: Callable[[FastAPI], None] | None
api_app_startup_event: Callable[[FastAPI], Awaitable[None]] | None
api_app_shutdown_event: Callable[[], Awaitable[None]] | None
```

**Purpose**: Allow custom FastAPI setup and lifecycle hooks

### Scrape Exclusion Extension

```python
scrape_exclude: ScrapeExcludeFunc | None
```

**Purpose**: Allow custom element exclusion logic

---

## Testing Considerations

### Test Scenarios

1. **Service initialization** - Verify all services initialize correctly
2. **LLM handler fallback** - Verify fallback logic works
3. **Storage backend selection** - Verify factory pattern
4. **Credential vault services** - Verify multi-vault support
5. **Extension points** - Verify custom extensions work

---

## References

- **Forge App**: `forge/forge_app.py` (284 lines)
- **Agent**: `forge/agent.py` (5,581 lines)
- **Agent Functions**: `forge/agent_functions.py`
- **Settings**: `config.py` (697 lines)
