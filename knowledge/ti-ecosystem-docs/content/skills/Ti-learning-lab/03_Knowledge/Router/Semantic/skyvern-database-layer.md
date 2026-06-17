# Skyvern Database Layer

**Repository**: Skyvern  
**Location**: `skyvern/forge/sdk/db/`  
**Key Files**:
- `models.py` (1,348 lines) - Database models
- `agent_db.py` - Database connection and session management
- `base_alchemy_db.py` - Base SQLAlchemy database class
- `base_repository.py` - Base repository pattern
- `repositories/` - Individual repository implementations
- `enums.py` - Database enums
- `id.py` - Custom ID generation

---

## Overview

Skyvern's Database Layer uses SQLAlchemy with async support, custom ID generation, soft delete mixin, and repository pattern for data access. It supports both SQLite (development) and PostgreSQL (production) with replica support for read scaling.

---

## Database Models

**Location**: `models.py` (1,348 lines)

### Custom ID Generation

**Location**: `id.py`

Skyvern uses **custom ID generation** for all entities:

```python
generate_task_id()
generate_workflow_id()
generate_workflow_run_id()
generate_workflow_permanent_id()
generate_artifact_id()
generate_script_id()
generate_script_revision_id()
generate_script_block_id()
generate_credential_id()
generate_browser_profile_id()
generate_persistent_browser_session_id()
# ... 40+ ID generators
```

**Benefits**:
- **Globally unique IDs**: No collisions across entities
- **URL-safe IDs**: Safe for URLs and APIs
- **Predictable format**: Consistent ID structure
- **No database dependency**: Can generate IDs before DB insert

### Base Model

```python
class Base(AsyncAttrs, DeclarativeBase):
    pass
```

### Task Model

```python
class TaskModel(Base):
    __tablename__ = "tasks"
    __table_args__ = (Index("idx_tasks_org_created", "organization_id", "created_at"),)
    
    task_id = Column(String, primary_key=True, default=generate_task_id)
    organization_id = Column(String, ForeignKey("organizations.organization_id"))
    browser_session_id = Column(String, nullable=True, index=True)
    status = Column(String, index=True)
    task_type = Column(String, default=TaskType.general)
    url = Column(String)
    navigation_goal = Column(String)
    data_extraction_goal = Column(String)
    navigation_payload = Column(JSON)
    extracted_information = Column(JSON)
    # ... more fields
```

### Key Tables

- **organizations** - Organization data
- **tasks** - Task definitions
- **actions** - Action executions
- **steps** - Step executions
- **artifacts** - Artifact storage
- **workflows** - Workflow definitions
- **workflow_runs** - Workflow run executions
- **workflow_run_blocks** - Block executions
- **workflow_parameters** - Workflow parameters
- **scripts** - Generated scripts
- **script_revisions** - Script versions
- **script_blocks** - Script block metadata
- **credentials** - Credential storage
- **browser_sessions** - Browser sessions
- **persistent_browser_sessions** - Persistent session tracking
- **schedules** - Workflow schedules
- **folders** - Organization folders

---

## Soft Delete Mixin

**Location**: `_soft_delete.py`

```python
class SoftDeleteMixin:
    deleted_at = Column(DateTime, nullable=True, index=True)
    
    def is_deleted(self) -> bool:
        return self.deleted_at is not None
```

**Benefits**:
- **Data retention**: Keep deleted data for audit
- **Recovery**: Can restore deleted records
- **Compliance**: Meet data retention requirements

---

## Database Connection

**Location**: `agent_db.py`

### AgentDB

```python
class AgentDB:
    def __init__(self, database_string: str, debug_enabled: bool = False):
        self.database_string = database_string
        self.debug_enabled = debug_enabled
        self.engine = create_async_engine(database_string, ...)
        self.async_session_maker = async_sessionmaker(...)
```

### Connection Pooling

**Features**:
- Async engine for non-blocking operations
- Connection pooling for performance
- Statement timeout configuration
- Replica support for read scaling

---

## Repository Pattern

**Location**: `base_repository.py`

### Base Repository

```python
class BaseRepository:
    def __init__(self, session: AsyncSession):
        self.session = session
```

### Repository Implementations

**Location**: `repositories/`

- `tasks.py` - Task repository
- `workflows.py` - Workflow repository
- `workflow_runs.py` - Workflow run repository
- `artifacts.py` - Artifact repository
- `credentials.py` - Credential repository
- `browser_sessions.py` - Browser session repository
- `scripts.py` - Script repository
- `schedules.py` - Schedule repository
- `organizations.py` - Organization repository
- `observer.py` - Observer repository (workflow run blocks)
- `debug.py` - Debug session repository

---

## Key Patterns

### 1. Custom ID Generation

**Pattern**: Generate IDs before database insert

**Benefits**:
- No dependency on database sequence
- Globally unique across entities
- URL-safe for APIs
- Predictable format

### 2. Soft Delete Pattern

**Pattern**: Mark records as deleted instead of hard delete

**Benefits**:
- Data retention for audit
- Recovery capability
- Compliance with retention policies

### 3. Repository Pattern

**Pattern**: Separate data access logic from business logic

**Benefits**:
- Separation of concerns
- Testable data access layer
- Consistent data access patterns

### 4. Async Database Operations

**Pattern**: Use SQLAlchemy async for non-blocking operations

**Benefits**:
- Non-blocking I/O
- Better performance under load
- Scalability

### 5. Replica Support

**Pattern**: Separate read replica for scaling

**Benefits**:
- Read scaling
- Reduced load on primary
- Better performance

---

## Database Schema

### Indexes

**Composite Indexes**:
- `idx_tasks_org_created` - (organization_id, created_at)
- Custom indexes for query optimization

### Foreign Keys

- All tables reference organization_id
- Cascading deletes where appropriate
- Soft delete compatibility

---

## Testing Considerations

### Test Scenarios

1. **ID generation** - Verify uniqueness and format
2. **Soft delete** - Verify soft delete logic
3. **Repository operations** - Verify CRUD operations
4. **Async operations** - Verify non-blocking behavior
5. **Replica fallback** - Verify replica logic

---

## References

- **Models**: `forge/sdk/db/models.py` (1,348 lines)
- **AgentDB**: `forge/sdk/db/agent_db.py`
- **Base Repository**: `forge/sdk/db/base_repository.py`
- **ID Generation**: `forge/sdk/db/id.py`
- **Soft Delete**: `forge/sdk/db/_soft_delete.py`
- **Repositories**: `forge/sdk/db/repositories/`
