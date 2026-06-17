# Junie CLI Lessons Learned for Ti CLI

> **Version**: 1.0.0
> **Purpose**: Learn from Junie CLI (JetBrains AI coding agent) and apply best practices to Ti CLI
> **Last Updated**: 2026-04-30

## Overview

Junie CLI is a mature, production-ready AI coding agent from JetBrains. This document analyzes its architecture, design patterns, and best practices to apply to Ti CLI development.

## 1. Architecture & Design Patterns

### Junie CLI Architecture

```
Junie CLI Architecture:
├── Shim Script (entry point)
│   ├── Version management
│   ├── Update handling
│   └── Binary execution
├── Binary Distribution
│   ├── Platform-specific builds
│   └── Version isolation
├── Update Mechanism
│   ├── JSONL version manifest
│   ├── SHA-256 verification
│   └── Atomic updates
└── Plugin Registry (ACP)
    ├── Agent discovery
    ├── Platform-specific distribution
    └── Dependency management
```

### Key Design Patterns

#### 1. Shim Pattern
**Junie**: Shell script shim as entry point
- Handles version selection
- Applies pending updates
- Executes appropriate binary
- Provides shim-specific commands

**Benefits**:
- Atomic updates without downtime
- Version rollback capability
- Clean separation of concerns

#### 2. Multi-Version Support
**Junie**: Multiple versions coexist
```
~/.local/share/junie/
├── versions/
│   ├── 888.12/
│   ├── 888.26/
│   └── 888.46/
└── current -> 888.46 (symlink)
```

**Benefits**:
- Easy version switching
- A/B testing
- Rollback capability

#### 3. Plugin Registry (ACP Protocol)
**Junie**: Agent Client Protocol for plugin management
```json
{
  "agents": [
    {
      "id": "amp-acp",
      "name": "Amp",
      "version": "0.7.0",
      "distribution": {
        "binary": {
          "linux-amd64": {
            "archive": "https://...",
            "cmd": "./amp-acp"
          }
        }
      }
    }
  ]
}
```

**Benefits**:
- Standardized plugin interface
- Platform-specific distribution
- Dependency management

## 2. Installation Mechanism

### Junie Installation Script

**Features**:
1. **Platform Detection**: Auto-detect OS and architecture
2. **Version Fetching**: Fetch latest version from update-info.jsonl
3. **Checksum Verification**: SHA-256 verification for security
4. **Shim Installation**: Install shell script shim
5. **PATH Management**: Auto-add to PATH in appropriate shell profile
6. **Multi-Shell Support**: zsh, bash, fish

### Installation Script Structure

```bash
#!/bin/bash
set -euo pipefail

# Configuration
CHANNEL="release"
UPDATE_INFO_URL="https://raw.githubusercontent.com/jetbrains-junie/junie/main/update-info.jsonl"
GITHUB_RELEASES="https://github.com/jetbrains-junie/junie/releases"
JUNIE_BIN="$HOME/.local/bin"
JUNIE_DATA="$HOME/.local/share/junie"

# Platform Detection
OS=$(uname -s)
ARCH=$(uname -m)
PLATFORM="${OS_NAME}-${ARCH_NAME}"

# Version Resolution
if [[ -n "${JUNIE_VERSION:-}" ]]; then
  VERSION="$JUNIE_VERSION"
else
  fetch_latest_version
fi

# Download & Install
mkdir -p "$JUNIE_BIN"
mkdir -p "$JUNIE_DATA/versions"
curl -fSL --progress-bar -o "$TMP_ZIP" "$DOWNLOAD_URL"
# Verify checksum
# Extract
# Install shim
# Add to PATH
```

### Lessons Learned

✅ **DO THIS**:
- Use shim pattern for entry point
- Support multiple versions
- Auto-detect platform
- Verify checksums
- Auto-manage PATH
- Support multiple shells

❌ **DON'T DO THIS**:
- Hardcode platform-specific paths
- Skip checksum verification
- Assume single version
- Manual PATH setup

## 3. Version Management

### Junie Version Management

**Directory Structure**:
```
~/.local/share/junie/
├── versions/
│   ├── 888.12/
│   │   ├── junie/bin/junie
│   │   └── junie.app/ (macOS)
│   ├── 888.26/
│   └── 888.46/
├── current -> 888.46 (symlink)
└── updates/
    └── pending-update.json
```

**Version Resolution Priority**:
1. `--use-version=<version>` flag
2. `JUNIE_VERSION` environment variable
3. `current` symlink

**Shim Commands**:
- `--list-versions`: List installed versions
- `--switch-version=<version>`: Switch version
- `--shim-version`: Show shim version

### Update Mechanism

**Update Process**:
1. Check for pending update
2. Verify checksum
3. Extract to versions directory
4. Update symlink atomically
5. Cleanup old files

**Atomic Update**:
```bash
# Extract to new version directory
mkdir -p "$VERSIONS_DIR/$version"
unzip -q -o "$zip_path" -d "$VERSIONS_DIR/$version"

# Atomic symlink update
ln -sfn "$VERSIONS_DIR/$version" "$CURRENT_LINK"
```

### Lessons Learned

✅ **DO THIS**:
- Use symlinks for version switching
- Support atomic updates
- Allow version rollback
- Provide version management commands

❌ **DON'T DO THIS**:
- Overwrite existing binary
- Break existing installations during update
- Skip atomic operations

## 4. Update Mechanism

### Junie Update Mechanism

**Update Info Format (JSONL)**:
```jsonl
{"version":"888.12","platform":"linux-amd64","downloadUrl":"https://...","sha256":"...","size":176132980}
{"version":"888.26","platform":"linux-amd64","downloadUrl":"https://...","sha256":"...","size":176133069}
```

**Why JSONL?**
- Easy to parse (one JSON per line)
- Easy to append new versions
- Easy to find latest version (tail -1)
- Human-readable

**Update Process**:
1. Fetch update-info.jsonl
2. Find latest entry for platform
3. Parse version, download URL, SHA-256
4. Download zip file
5. Verify checksum
6. Extract to versions directory
7. Update symlink

### Pending Update Mechanism

**Pending Update File**:
```json
{
  "version": "888.46",
  "zipPath": "/path/to/junie-release-888.46-linux-amd64.zip",
  "sha256": "..."
}
```

**Apply on Next Run**:
```bash
if [[ -f "$PENDING_UPDATE" ]]; then
  apply_pending_update
fi
```

### Lessons Learned

✅ **DO THIS**:
- Use JSONL for version manifests
- Verify checksums
- Support pending updates
- Atomic updates via symlinks
- Platform-specific downloads

❌ **DON'T DO THIS**:
- Use complex JSON for version info
- Skip checksum verification
- Break installations during update
- Hardcode download URLs

## 5. Plugin Registry (ACP Protocol)

### Junie Plugin Registry

**Registry Format**:
```json
{
  "version": "1.0.0",
  "agents": [
    {
      "id": "amp-acp",
      "name": "Amp",
      "version": "0.7.0",
      "description": "ACP wrapper for Amp",
      "repository": "https://github.com/...",
      "authors": ["tao12345666333"],
      "license": "Apache-2.0",
      "icon": "https://cdn.agentclientprotocol.com/...",
      "distribution": {
        "binary": {
          "linux-amd64": {
            "archive": "https://github.com/.../amp-acp-linux-amd64.tar.gz",
            "cmd": "./amp-acp"
          },
          "windows-amd64": {
            "archive": "https://github.com/.../amp-acp-windows-amd64.zip",
            "cmd": "./amp-acp.exe"
          }
        }
      }
    }
  ]
}
```

**Key Features**:
- Standardized agent interface (ACP)
- Platform-specific distribution
- Metadata (description, authors, license)
- Icon support
- Multiple distribution types (binary, source)

### Lessons Learned

✅ **DO THIS**:
- Use standardized plugin interface
- Support platform-specific distribution
- Include rich metadata
- Support multiple distribution types
- Use GitHub releases for distribution

❌ **DON'T DO THIS**:
- Hardcode plugin paths
- Skip platform detection
- Ignore metadata
- Use single distribution type

## 6. Configuration Management

### Junie Configuration

**Data Directory**:
```
~/.local/share/junie/
├── versions/
├── current (symlink)
├── updates/
└── config/
```

**Environment Variables**:
- `JUNIE_DATA`: Data directory location
- `JUNIE_VERSION`: Version to use
- `EJ_RUNNER_PWD`: Working directory

### Lessons Learned

✅ **DO THIS**:
- Use XDG base directory specification
- Support environment variables for configuration
- Separate data from binary
- Use symlinks for version management

❌ **DON'T DO THIS**:
- Hardcode paths
- Mix data with binary
- Ignore XDG specification

## 7. CLI UX Patterns

### Junie CLI UX

**Interactive Mode**:
```bash
junie
> Build a web dashboard...
```

**Direct Commands**:
```bash
junie review
junie fix bug
```

**GitHub Integration**:
```bash
/install-github-action
```

**Feedback Mechanism**:
```bash
/feedback
```

### Lessons Learned

✅ **DO THIS**:
- Support both interactive and direct commands
- Provide easy GitHub integration
- Include feedback mechanism
- Natural language interface

❌ **DON'T DO THIS**:
- Force interactive mode only
- Skip integration options
- Ignore user feedback

## 8. Applying to Ti CLI

### Recommended Architecture for Ti CLI

```
Ti CLI Architecture (Learned from Junie):
├── Shim Script (entry point)
│   ├── Version management
│   ├── Update handling
│   └── Binary execution
├── Binary Distribution
│   ├── Platform-specific builds
│   └── Version isolation
├── Update Mechanism
│   ├── JSONL version manifest
│   ├── SHA-256 verification
│   └── Atomic updates
└── Plugin System (MCP-based)
    ├── MCP Hub integration
    ├── Plugin registry
    └── Platform-specific distribution
```

### Implementation Plan

#### Phase 1: Shim Pattern
**Tasks**:
1. Create shell script shim for Ti CLI
2. Implement version selection logic
3. Add shim-specific commands (--list-versions, --switch-version)
4. Implement pending update mechanism

**Files**:
- `install.sh` / `install.ps1` (installation scripts)
- `ti-shim.sh` / `ti-shim.ps1` (shim scripts)

#### Phase 2: Version Management
**Tasks**:
1. Implement multi-version support
2. Create version directory structure
3. Implement version resolution priority
4. Add version management commands

**Directory Structure**:
```
~/.local/share/ti/
├── versions/
│   ├── 1.0.0/
│   ├── 1.1.0/
│   └── 1.2.0/
├── current -> 1.2.0 (symlink)
└── updates/
    └── pending-update.json
```

#### Phase 3: Update Mechanism
**Tasks**:
1. Create update-info.jsonl
2. Implement version fetching
3. Add SHA-256 verification
4. Implement atomic updates

**Update Info Format**:
```jsonl
{"version":"1.0.0","platform":"linux-amd64","downloadUrl":"https://...","sha256":"...","size":12345678}
{"version":"1.1.0","platform":"linux-amd64","downloadUrl":"https://...","sha256":"...","size":12345678}
```

#### Phase 4: Plugin Registry
**Tasks**:
1. Create MCP plugin registry
2. Implement plugin discovery
3. Add platform-specific distribution
4. Integrate with existing MCP Hub

**Registry Format**:
```json
{
  "version": "1.0.0",
  "plugins": [
    {
      "id": "mcp-github",
      "name": "GitHub MCP",
      "version": "1.0.0",
      "description": "GitHub operations via MCP",
      "distribution": {
        "binary": {
          "linux-amd64": {
            "archive": "https://.../mcp-github-linux-amd64.tar.gz",
            "cmd": "./mcp-github"
          }
        }
      }
    }
  ]
}
```

#### Phase 5: Configuration Management
**Tasks**:
1. Use XDG base directory specification
2. Support environment variables
3. Separate data from binary
4. Implement configuration loading

**Environment Variables**:
- `TI_DATA`: Data directory location
- `TI_VERSION`: Version to use
- `TI_CONFIG`: Config file location

#### Phase 6: CLI UX
**Tasks**:
1. Support interactive mode
2. Add direct commands
3. Implement GitHub integration
4. Add feedback mechanism

**Commands**:
- `ti` (interactive mode)
- `ti mcp` (MCP operations)
- `ti beads` (BEADS operations)
- `ti brain` (TiBrain operations)
- `/install-github-action` (GitHub integration)

### Priority Matrix

| Priority | Feature | Complexity | Impact |
|----------|---------|------------|--------|
| **P0** | Shim Pattern | Medium | High |
| **P0** | Version Management | Medium | High |
| **P1** | Update Mechanism | High | High |
| **P1** | Configuration Management | Low | Medium |
| **P2** | Plugin Registry | High | Medium |
| **P2** | CLI UX Improvements | Medium | Medium |

### Quick Wins (Implement First)

1. **Shim Script** (1-2 days)
   - Create basic shim
   - Add version selection
   - Implement PATH management

2. **Version Management** (2-3 days)
   - Create version directory structure
   - Add version commands
   - Implement version switching

3. **Update Info JSONL** (1 day)
   - Create update-info.jsonl
   - Implement version fetching
   - Add checksum verification

### Long-term Goals

1. **Plugin Registry** (1-2 weeks)
   - Design registry format
   - Implement plugin discovery
   - Integrate with MCP Hub

2. **CLI UX Improvements** (1 week)
   - Interactive mode
   - Natural language interface
   - GitHub integration

## 9. Comparison: Junie vs Ti CLI

| Feature | Junie CLI | Ti CLI (Current) | Ti CLI (Target) |
|---------|-----------|------------------|-----------------|
| **Architecture** | Shim + Binary | Direct binary | Shim + Binary |
| **Version Management** | Multi-version | Single version | Multi-version |
| **Updates** | Atomic updates | Manual updates | Atomic updates |
| **Plugin System** | ACP Registry | MCP Hub | MCP Registry |
| **Configuration** | XDG + Env vars | JSON config | XDG + Env vars + JSON |
| **Installation** | Auto-install script | Manual build | Auto-install script |
| **Platform Support** | Linux, macOS, Windows | Windows (primary) | Linux, macOS, Windows |

## 10. Next Steps

### Immediate Actions

1. **Create shim script** for Ti CLI
2. **Implement version management** structure
3. **Create update-info.jsonl** format
4. **Update installation scripts**

### Documentation Updates

1. Update `MICROKERNEL_PLUGIN_PLAN.md` with shim pattern
2. Create `INSTALLATION.md` with new installation process
3. Update `AGENTS.md` with version management
4. Create `UPDATE_MECHANISM.md` for update process

### Testing

1. Test shim script on Linux, macOS, Windows
2. Test version management
3. Test atomic updates
4. Test platform detection

## Resources

- **Junie CLI**: https://github.com/jetbrains-junie/junie
- **Junie Documentation**: https://junie.jetbrains.com/docs
- **ACP Protocol**: https://agentclientprotocol.com
- **XDG Base Directory**: https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html

## Conclusion

Junie CLI provides excellent patterns for:
- **Shim-based architecture** for atomic updates
- **Multi-version support** for flexibility
- **JSONL format** for version manifests
- **ACP protocol** for plugin management
- **XDG specification** for configuration

Applying these patterns to Ti CLI will result in a more robust, maintainable, and user-friendly CLI.
