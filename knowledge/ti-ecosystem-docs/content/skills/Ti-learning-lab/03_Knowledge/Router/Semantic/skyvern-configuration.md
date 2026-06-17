# Skyvern Configuration

**Repository**: Skyvern  
**Location**: `skyvern/config.py` (697 lines)

---

## Overview

Skyvern's Configuration system uses Pydantic Settings with environment variable support, default values for development, and comprehensive settings for all system components including browser, database, LLM, storage, and performance tuning.

---

## Configuration Pattern

**Location**: `config.py` (lines 56-697)

```python
class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=_DEFAULT_ENV_FILES, extra="ignore")
```

**Environment Files**:
- `.env` - Development
- `.env.staging` - Staging
- `.env.prod` - Production

**Priority**: Environment variables > .env files > defaults

---

## Browser Configuration

```python
BROWSER_TYPE: str = "chromium-headful"
BROWSER_REMOTE_DEBUGGING_URL: str = "http://127.0.0.1:9222"
CHROME_EXECUTABLE_PATH: str | None = None
MAX_SCRAPING_RETRIES: int = 0
VIDEO_PATH: str | None = "./video"
HAR_PATH: str | None = "./har"
LOG_PATH: str = "./log"
TEMP_PATH: str = "./temp"
DOWNLOAD_PATH: str = f"{REPO_ROOT_DIR}/downloads"
```

---

## Timeout Configuration

```python
BROWSER_ACTION_TIMEOUT_MS: int = 5000
CACHED_ACTION_DELAY_SECONDS: float = 1.0
PAGE_READY_NETWORK_IDLE_TIMEOUT_MS: float = 3000
PAGE_READY_LOADING_INDICATOR_TIMEOUT_MS: float = 5000
PAGE_READY_DOM_STABLE_MS: float = 300
PAGE_READY_DOM_STABILITY_TIMEOUT_MS: float = 3000
BROWSER_SCREENSHOT_TIMEOUT_MS: int = 20000
BROWSER_LOADING_TIMEOUT_MS: int = 60000
BROWSER_SCRAPING_BUILDING_ELEMENT_TREE_TIMEOUT_MS: int = 60 * 1000
OPTION_LOADING_TIMEOUT_MS: int = 600000
```

**Purpose**: Fine-grained timeout control for different browser operations

---

## Step and Retry Configuration

```python
MAX_STEPS_PER_RUN: int = 10
MAX_STEPS_PER_TASK_V2: int = 25
MAX_ITERATIONS_PER_TASK_V2: int = 10
MAX_NUM_SCREENSHOTS: int = 10
LONG_RUNNING_TASK_WARNING_RATIO: float = 0.95
MAX_RETRIES_PER_STEP: int = 5
```

**Purpose**: Control execution limits and retry behavior

---

## Database Configuration

```python
DATABASE_STRING: str = Field(default_factory=_default_database_string)
DATABASE_REPLICA_STRING: str | None = None
DATABASE_STATEMENT_TIMEOUT_MS: int = 60000
DISABLE_CONNECTION_POOL: bool = False
```

**Default Database**: SQLite at `~/.skyvern/data.db`

---

## LLM Configuration

```python
LLM_KEY: str
SECONDARY_LLM_KEY: str | None = None
SELECT_AGENT_LLM_KEY: str | None = None
NORMAL_SELECT_AGENT_LLM_KEY: str | None = None
CUSTOM_SELECT_AGENT_LLM_KEY: str | None = None
SINGLE_CLICK_AGENT_LLM_KEY: str | None = None
SINGLE_INPUT_AGENT_LLM_KEY: str | None = None
PARSE_SELECT_LLM_KEY: str | None = None
EXTRACTION_LLM_KEY: str | None = None
CHECK_USER_GOAL_LLM_KEY: str | None = None
AUTO_COMPLETION_LLM_KEY: str | None = None
SCRIPT_GENERATION_LLM_KEY: str | None = None
SCRIPT_REVIEWER_LLM_KEY: str | None = None
ADAPTIVE_SCRIPT_GEN_LLM_KEY: str | None = None
WORKFLOW_COPILOT_AGENT_LLM_KEY: str | None = None
WORKFLOW_COPILOT_FAST_LLM_KEY: str | None = None
```

**Purpose**: 20+ LLM keys for different handlers (matches ForgeApp)

---

## OpenAI Configuration

```python
OPENAI_API_KEY: str
OPENAI_CUA_MODEL: str
ENABLE_AZURE_CUA: bool = False
AZURE_CUA_API_KEY: str | None = None
AZURE_CUA_API_VERSION: str | None = None
AZURE_CUA_ENDPOINT: str | None = None
AZURE_CUA_DEPLOYMENT: str | None = None
```

---

## Anthropic Configuration

```python
ANTHROPIC_API_KEY: str
ENABLE_BEDROCK_ANTHROPIC: bool = False
```

---

## Thinking Budget Configuration

```python
EXTRACT_ACTION_THINKING_BUDGET: int
DEFAULT_THINKING_BUDGET: int
```

**Purpose**: Control Claude thinking token usage

---

## Storage Configuration

```python
SKYVERN_STORAGE_TYPE: str  # "s3" or "azureblob"
AWS_S3_BUCKET: str | None = None
AWS_S3_REGION: str | None = None
AWS_ACCESS_KEY_ID: str | None = None
AWS_SECRET_ACCESS_KEY: str | None = None
AZURE_STORAGE_ACCOUNT_NAME: str | None = None
AZURE_STORAGE_CONTAINER_NAME: str | None = None
AZURE_STORAGE_ACCESS_KEY: str | None = None
```

---

## Credential Vault Configuration

```python
AZURE_CREDENTIAL_VAULT: str | None = None
AZURE_TENANT_ID: str | None = None
AZURE_CLIENT_ID: str | None = None
AZURE_CLIENT_SECRET: str | None = None
CUSTOM_CREDENTIAL_API_BASE_URL: str | None = None
CUSTOM_CREDENTIAL_API_TOKEN: str | None = None
```

---

## Proxy Configuration

```python
BROWSER_PROXY_URL: str | None = None
BROWSER_PROXY_USERNAME: str | None = None
BROWSER_PROXY_PASSWORD: str | None = None
```

---

## Experimentation Configuration

```python
ENABLE_EXP_ALL_TEXTUAL_ELEMENTS_INTERACTABLE: bool = False
```

---

## Script Reviewer Configuration

```python
SCRIPT_REVIEW_DAILY_CAP: int = 5  # Max script reviews per wpid per day
```

---

## Key Patterns

### 1. Environment File Priority

**Pattern**: Environment variables > .env files > defaults

**Benefits**:
- **Flexibility**: Easy to override settings
- **Environment-specific**: Different configs per environment
- **Security**: Sensitive data in environment variables

### 2. Default SQLite Database

**Pattern**: Use SQLite by default for development

**Benefits**:
- **Zero setup**: Works out of the box
- **Easy testing**: No database required
- **Production-ready**: Easy to switch to PostgreSQL

### 3. 20+ LLM Keys

**Pattern**: Separate LLM keys for different handlers

**Benefits**:
- **Cost optimization**: Use different models per use case
- **Flexibility**: Easy to reconfigure
- **Reliability**: Fallback between models

### 4. Fine-Grained Timeouts

**Pattern**: Separate timeouts for different operations

**Benefits**:
- **Performance optimization**: Appropriate timeouts per operation
- **Reliability**: Prevent hangs on slow operations
- **Debugging**: Easier to identify slow operations

### 5. Azure Workload Identity Support

**Pattern**: Support both explicit credentials and workload identity

**Benefits**:
- **Security**: No secrets in code for production
- **Cloud-native**: Workload identity for AKS/Azure VMs
- **Flexibility**: Explicit credentials for development

---

## Testing Considerations

### Test Scenarios

1. **Environment file loading** - Verify priority order
2. **Default values** - Verify defaults work
3. **Validation** - Verify invalid values are rejected
4. **Type conversion** - Verify correct type handling

---

## References

- **Configuration**: `config.py` (697 lines)
