# Skyvern Artifact Management

**Repository**: Skyvern  
**Location**: `skyvern/forge/sdk/artifact/`  
**Key Files**:
- `manager.py` - Artifact manager
- `models.py` - Artifact models
- `storage/base.py` - Base storage interface
- `storage/factory.py` - Storage factory
- `storage/local.py` - Local storage
- `storage/s3.py` - S3 storage
- `storage/azure.py` - Azure storage
- `signing.py` - Artifact signing

---

## Overview

Skyvern's Artifact Management system stores and manages workflow artifacts including screenshots, LLM prompts, request/response logs, and extracted data. It supports multiple storage backends (local, S3, Azure) with artifact signing for integrity verification.

---

## Artifact Types

**Location**: `forge/sdk/artifact/models.py`

```python
class ArtifactType(StrEnum):
    screenshot = "screenshot"
    llm_request = "llm_request"
    llm_response = "llm_response"
    action = "action"
    element_tree = "element_tree"
    extracted_data = "extracted_data"
    file = "file"
    other = "other"
```

---

## Artifact Manager

**Location**: `forge/sdk/artifact/manager.py`

**Purpose**: Manage artifact lifecycle

**Features**:
- Artifact creation
- Artifact retrieval
- Artifact deletion
- Bulk operations
- Artifact signing

### Bulk Artifact Creation

```python
class BulkArtifactCreationRequest(BaseModel):
    """Request for bulk artifact creation."""
```

---

## Storage Backends

### Base Storage

**Location**: `forge/sdk/artifact/storage/base.py`

**Purpose**: Abstract storage interface

**Methods**:
- `upload()` - Upload artifact
- `download()` - Download artifact
- `delete()` - Delete artifact
- `exists()` - Check existence
- `list()` - List artifacts

### Local Storage

**Location**: `forge/sdk/artifact/storage/local.py`

**Purpose**: Local filesystem storage

**Features**:
- File-based storage
- Directory management
- Path validation

### S3 Storage

**Location**: `forge/sdk/artifact/storage/s3.py`

**Purpose**: AWS S3 storage

**Features**:
- S3 integration
- Bucket management
- Multipart upload
- Presigned URLs

### Azure Storage

**Location**: `forge/sdk/artifact/storage/azure.py`

**Purpose**: Azure Blob Storage

**Features**:
- Azure integration
- Container management
- SAS tokens
- Blob operations

---

## Storage Factory

**Location**: `forge/sdk/artifact/storage/factory.py`

**Purpose**: Create storage instances

**Features**:
- Storage backend selection
- Configuration management
- Fallback logic

---

## Artifact Signing

**Location**: `forge/sdk/artifact/signing.py`

**Purpose**: Sign artifacts for integrity verification

**Features**:
- Artifact hashing
- Signature generation
- Signature verification
- Key management

---

## Key Patterns

### 1. Storage Abstraction

**Pattern**: Abstract storage interface for multiple backends

**Benefits**:
- Backend flexibility
- Easy testing
- Consistent API

### 2. Artifact Signing

**Pattern**: Sign artifacts for integrity verification

**Benefits**:
- Tamper detection
- Integrity verification
- Audit trail

### 3. Bulk Operations

**Pattern**: Batch artifact operations for efficiency

**Benefits**:
- Reduced overhead
- Faster operations
- Better performance

---

## Performance Optimizations

### 1. Bulk Operations

**Impact**: Reduced overhead

### 2. Multipart Upload

**Impact**: Faster large file uploads

### 3. Presigned URLs

**Impact**: Direct downloads, reduced server load

---

## Testing Considerations

### Test Scenarios

1. **Storage backends** - Verify all backends work
2. **Artifact signing** - Verify signature generation/verification
3. **Bulk operations** - Verify bulk create/delete
4. **Storage factory** - Verify backend selection
5. **Error handling** - Verify error recovery

### Test Commands

```bash
# Run artifact tests
python -m pytest tests/unit/test_artifacts.py -v

# Run storage tests
python -m pytest tests/unit/test_storage.py -v
```

---

## References

- **Artifact Manager**: `forge/sdk/artifact/manager.py`
- **Artifact Models**: `forge/sdk/artifact/models.py`
- **Storage Base**: `forge/sdk/artifact/storage/base.py`
- **Storage Factory**: `forge/sdk/artifact/storage/factory.py`
- **Artifact Signing**: `forge/sdk/artifact/signing.py`
