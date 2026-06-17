# Component Tree Integration Guide

> **Last Updated**: 2026-05-04
> **Purpose**: Integrate component tree analysis with file server dashboard and Notion sync

---

## 🎯 Overview

Component tree analysis tool is now integrated with:
- **File Server Dashboard** - Web API for on-demand generation and viewing
- **Notion Auto-Sync** - Automatic sync of component trees to Notion for AI context
- **Batch Generation** - Generate trees for multiple projects at once

---

## 📐 Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    INTEGRATION LAYERS                     │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌─────────────────────────────────────────────────┐   │
│  │  CORE TOOL                                       │   │
│  │  bin/generate-component-hierarchy.ts             │   │
│  │  - AST parsing                                   │   │
│  │  - Component analysis                            │   │
│  │  - Tree generation                              │   │
│  └─────────────────────────────────────────────────┘   │
│                         ↓                              │
│  ┌─────────────────────────────────────────────────┐   │
│  │  SKILL WRAPPER                                   │   │
│  │  .devin/skills/component-tree/SKILL.md           │   │
│  │  - Standard invocation patterns                 │   │
│  │  - Smart defaults                               │   │
│  │  - Best practices                               │   │
│  └─────────────────────────────────────────────────┘   │
│                         ↓                              │
│  ┌─────────────────────────────────────────────────┐   │
│  │  BATCH GENERATOR                                 │   │
│  │  generate-component-trees.js                     │   │
│  │  - Multi-project support                        │   │
│  │  - Output management                            │   │
│  │  - Index generation                             │   │
│  └─────────────────────────────────────────────────┘   │
│                         ↓                              │
│  ┌─────────────────────────────────────────────────┐   │
│  │  FILE SERVER API                                 │   │
│  │  GET /api/component-trees                        │   │
│  │  GET /api/component-tree/{name}                  │   │
│  │  GET /api/component-tree/generate                │   │
│  └─────────────────────────────────────────────────┘   │
│                         ↓                              │
│  ┌─────────────────────────────────────────────────┐   │
│  │  NOTION SYNC                                     │   │
│  │  notion-sync.js                                  │   │
│  │  notion-auto-sync.js                             │   │
│  │  - Auto-sync component trees                     │   │
│  │  - Watch for changes                            │   │
│  └─────────────────────────────────────────────────┘   │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

---

## 🚀 Usage

### 1. Generate Component Trees

**Manual Generation (Single Project):**
```bash
bun bin/generate-component-hierarchy.ts \
  --src "apps/mobile/src" \
  --entry "apps/mobile/src/app/_layout.tsx"
```

**Batch Generation (All Projects):**
```bash
cd Z:\10_WORKPLACE\Ti\Ti-learning-lab\03_Knowledge\Router\context-dashboard
npm run generate-trees
```

**Via File Server API:**
```bash
# Generate on-demand
curl "http://localhost:3000/api/component-tree/generate?project=mobile"

# With focus
curl "http://localhost:3000/api/component-tree/generate?project=mobile&focus=HomeScreen&scope=full"

# Layout-only
curl "http://localhost:3000/api/component-tree/generate?project=mobile&layoutOnly=true"
```

### 2. View Component Trees

**List Available Trees:**
```bash
curl http://localhost:3000/api/component-trees
```

**Get Specific Tree:**
```bash
curl http://localhost:3000/api/component-tree/mobile
```

**Via File Browser UI:**
- Navigate to `component-trees/` directory
- View `.txt` files directly in browser

### 3. Sync to Notion

**Manual Sync:**
```bash
npm run notion-sync
```

**Auto-Sync (Background):**
```bash
npm run notion-auto-sync
```

Component trees in `component-trees/` directory are automatically:
- Scanned by notion-sync.js
- Watched by notion-auto-sync.js
- Synced to Notion database

---

## 📁 File Structure

```
context-dashboard/
├── bin/
│   └── generate-component-hierarchy.ts    # Core tool
├── component-trees/                         # Generated trees
│   ├── INDEX.md                            # Index file
│   ├── mobile-component-tree.txt           # Mobile project tree
│   ├── web-component-tree.txt              # Web project tree
│   └── desktop-component-tree.txt          # Desktop project tree
├── generate-component-trees.js             # Batch generator
├── notion-sync.js                          # Updated to sync .txt files
├── notion-auto-sync.js                     # Updated to watch component-trees/
├── server.js                               # Updated with component tree APIs
└── package.json                            # Updated with generate-trees script
```

---

## 🔧 Configuration

### Project Configuration

Edit `generate-component-trees.js` to add/remove projects:

```javascript
const PROJECTS = [
  {
    name: 'mobile',
    src: 'apps/mobile/src',
    entry: 'apps/mobile/src/app/_layout.tsx',
    aliases: ['@=apps/mobile/src']
  },
  // Add more projects...
];
```

### File Server Configuration

Edit `server.js` to add more projects:

```javascript
const projects = {
  mobile: {
    src: 'apps/mobile/src',
    entry: 'apps/mobile/src/app/_layout.tsx',
    aliases: ['@=apps/mobile/src']
  },
  // Add more projects...
};
```

---

## 🎨 Use Cases

### 1. AI Context for Architecture

**Workflow:**
```bash
# Generate trees
npm run generate-trees

# Sync to Notion
npm run notion-sync

# AI now has component hierarchy context
```

**Benefit:** AI understands project structure when answering questions.

### 2. Team Documentation

**Workflow:**
```bash
# Generate trees for documentation
npm run generate-trees

# Commit to repo
git add component-trees/
git commit -m "docs: update component trees"

# View in file browser
# http://localhost:3000
```

**Benefit:** Onboarding docs for new developers.

### 3. Code Review

**Workflow:**
```bash
# Before changes
bun bin/generate-component-hierarchy.ts > before.txt

# Make changes...

# After changes
bun bin/generate-component-hierarchy.ts > after.txt

# Compare
diff before.txt after.txt
```

**Benefit:** Visualize structural changes in PRs.

### 4. Debug Navigation

**Workflow:**
```bash
# Focus on specific route
curl "http://localhost:3000/api/component-tree/generate?project=mobile&focus=ProfileScreen&scope=up"
```

**Benefit:** Understand navigation context and layout hierarchy.

---

## 🔍 API Reference

### GET /api/component-trees

**Description:** List all available component trees

**Response:**
```json
{
  "trees": [
    {
      "name": "mobile",
      "file": "mobile-component-tree.txt",
      "size": 12345,
      "modified": "2026-05-04T13:30:00.000Z"
    }
  ]
}
```

### GET /api/component-tree/{name}

**Description:** Get specific component tree

**Parameters:**
- `name` - Project name (e.g., "mobile", "web")

**Response:** Plain text ASCII tree

### GET /api/component-tree/generate

**Description:** Generate component tree on-demand

**Query Parameters:**
- `project` - Project name (default: "mobile")
- `focus` - Component name to focus on
- `scope` - Focus scope: "up" | "full" | "down" (default: "full")
- `layoutOnly` - Boolean flag for layout-only mode

**Example:**
```
GET /api/component-tree/generate?project=mobile&focus=HomeScreen&scope=full&layoutOnly=true
```

**Response:** Plain text ASCII tree

---

## 🛠️ Troubleshooting

### Issue: "Component tree not found"

**Solution:** Generate tree first:
```bash
npm run generate-trees
```

### Issue: "Unknown project" in API

**Solution:** Add project configuration to `server.js` and `generate-component-trees.js`

### Issue: Notion sync not including component trees

**Solution:** Ensure:
1. `component-trees/` directory exists
2. Files have `.txt` extension
3. `SYNC_EXTENSIONS` includes `.txt` in notion-sync.js

### Issue: Parse errors for complex TypeScript

**Solution:** Check:
1. TypeScript configuration
2. Babel plugins in generate-component-hierarchy.ts
3. File encoding (should be UTF-8)

---

## 📊 Best Practices

### 1. Version Control
- Commit component trees to track changes over time
- Use diff to understand structural changes
- Tag releases with component tree snapshots

### 2. Regular Updates
- Regenerate trees after major refactors
- Update before releases
- Sync to Notion for AI context

### 3. Multiple Views
- Generate different views for different purposes:
  - Full tree for architecture overview
  - Layout-only for styling review
  - Focused views for specific features

### 4. Team Collaboration
- Share component trees in PRs
- Use for code review discussions
- Include in onboarding documentation

---

## 🔮 Future Enhancements

- [ ] Interactive web UI for component tree visualization
- [ ] Diff viewer for before/after comparisons
- [ ] JSON output format for programmatic analysis
- [ ] Integration with CI/CD pipelines
- [ ] Automatic generation on file changes
- [ ] Component dependency graph visualization
- [ ] Performance metrics per component
- [ ] Search across component trees

---

## 📚 Related Documentation

- [Component Tree Skill](.devin/skills/component-tree/SKILL.md)
- [File Server Dashboard README](README.md)
- [Notion Integration Guide](NOTION_INTEGRATION.md)
- [Core Tool Documentation](bin/generate-component-hierarchy.ts)
