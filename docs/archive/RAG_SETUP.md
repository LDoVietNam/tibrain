# TiBrain RAG System Setup Guide

## Overview
This RAG system provides intelligent knowledge retrieval for TiBrain using local file indexing and vector similarity search.

## Installation

### 1. Install Python Dependencies
```bash
pip install flask flask-cors
pip install chromadb sentence-transformers watchdog numpy
```

### 2. Directory Structure
```
Z:/10_WORKPLACE/Ti/apps/tibrain/
├── rag_builder.py          # RAG index builder
├── rag_api_server.py       # HTTP API server
├── start_rag_system.py     # System startup script
└── rag_index/             # Generated index (auto-created)
    ├── chroma_db/         # ChromaDB storage
    ├── metadata.json      # Index metadata
    └── file_hashes.pkl   # File change tracking
```

## Usage

### Quick Start
```bash
cd Z:/10_WORKPLACE/Ti/apps/tibrain
python start_rag_system.py
```

### Advanced Options
```bash
# Build index only
python start_rag_system.py --build-only

# Custom host/port
python start_rag_system.py --host 0.0.0.0 --port 9090

# Enable file watching
python start_rag_system.py --watch

# Custom paths
python start_rag_system.py --tibrain-path /path/to/tibrain --output-path /path/to/index
```

### Manual Index Building
```bash
# Full rebuild
python rag_builder.py --tibrain-path Z:/10_WORKPLACE/Ti/apps/tibrain

# Incremental update (only changed files)
python rag_builder.py --tibrain-path Z:/10_WORKPLACE/Ti/apps/tibrain --incremental

# With file watching
python rag_builder.py --tibrain-path Z:/10_WORKPLACE/Ti/apps/tibrain --watch
```

### API Server Only
```bash
python rag_api_server.py --host 127.0.0.1 --port 8080
```

## API Endpoints

### Health Check
```bash
GET http://127.0.0.1:8080/health
```

### Status
```bash
GET http://127.0.0.1:8080/status
```

### Query (POST)
```bash
POST http://127.0.0.1:8080/query
Content-Type: application/json

{
  "question": "How does the retrieval router work?",
  "context": "TiBrain architecture",
  "top_k": 5
}
```

### Query (GET)
```bash
GET http://127.0.0.1:8080/query?q=retrieval%20router&top_k=5
```

### Rebuild Index
```bash
POST http://127.0.0.1:8080/rebuild
```

## Integration with Ti Agent

The Ti agent can use the RAG system through the enhanced interface:

```go
// In Ti agent code
import "Z:/10_WORKPLACE/Ti/apps/ticrew/Member/Ti/tibrain_interface_rag"

// Create RAG-enabled interface
tbi := tibrain.NewTiBrainRAGInterface()
tbi.Initialize()

// Query for knowledge
result, err := tbi.Query("How does the retrieval router work?", "")

// Get guidance
guidance, err := tbi.GetGuidance("architecture")

// Check status
status, err := tbi.GetStatus()
```

## Features

### 1. Intelligent File Processing
- Go source code structure extraction
- Documentation parsing
- Incremental updates (only changed files)
- File hash tracking for efficiency

### 2. Multiple Storage Backends
- ChromaDB (recommended for production)
- Fallback pickle storage (for development)
- Configurable embedding models

### 3. Real-time Updates
- File system watching
- Automatic rebuild on changes
- Cooldown period to prevent excessive rebuilds

### 4. HTTP API
- RESTful endpoints
- JSON request/response
- CORS support
- Health monitoring

## Configuration

### Environment Variables
```bash
RAG_INDEX_PATH=./rag_index
TIBRAIN_PATH=Z:/10_WORKPLACE/Ti/apps/tibrain
```

### Performance Tuning
- Chunk size: Adjust in `rag_builder.py`
- Embedding model: Change in `rag_builder.py`
- Top-k results: Configure via API
- File watching cooldown: Adjust in `rag_builder.py`

## Troubleshooting

### Dependencies Missing
```bash
pip install flask flask-cors chromadb sentence-transformers watchdog numpy
```

### Port Already in Use
```bash
# Use different port
python start_rag_system.py --port 8081
```

### Index Build Fails
```bash
# Check file permissions
# Ensure tibrain path is correct
# Check disk space
```

### API Server Not Responding
```bash
# Check health endpoint
curl http://127.0.0.1:8080/health

# Check logs
# Verify port is not blocked by firewall
```

## Performance

### Index Build Time
- Initial build: ~5-10 minutes (depending on file count)
- Incremental update: ~30-60 seconds (changed files only)

### Query Performance
- ChromaDB: ~50-200ms per query
- Fallback: ~100-500ms per query

### Memory Usage
- ChromaDB: ~500MB - 2GB (depending on index size)
- Fallback: ~200MB - 1GB

## Architecture

```
TiBrain Source Files
        ↓
   RAG Builder
        ↓
   File Processing
        ↓
   Chunking
        ↓
   Embedding Generation
        ↓
   Vector Storage
        ↓
   HTTP API
        ↓
   Ti Agent Integration
```

## Security Considerations

- API server runs on localhost by default
- No authentication by default (add if needed)
- File access restricted to tibrain directory
- Input validation on API endpoints

## Future Enhancements

- [ ] Authentication/Authorization
- [ ] Query caching
- [ ] Distributed indexing
- [ ] Advanced filtering
- [ ] Multi-language support
- [ ] Performance monitoring dashboard