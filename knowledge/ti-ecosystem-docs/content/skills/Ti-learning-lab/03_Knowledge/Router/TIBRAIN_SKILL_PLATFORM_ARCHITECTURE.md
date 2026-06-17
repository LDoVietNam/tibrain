# TiBrain Skill Platform Architecture

> **Inspired by**: claude-code-templates (aitmpl.com)
> **Goal**: Full-stack skill management platform for TiBrain

---

## 🎯 Overview

TiBrain Skill Platform là một platform hoàn chỉnh để:
- Browse, search, và filter skills từ multiple sources
- Install skills vào TiBrain server
- Track usage và analytics
- Manage collections (user favorites)
- CLI tool cho installation

---

## 📋 Architecture (Based on claude-code-templates)

### Directory Structure

```
tibrain-skill-platform/
├── api/                          # API backend (Node.js/Express)
│   ├── src/
│   │   ├── routes/              # API endpoints
│   │   │   ├── skills.ts        # GET /api/skills
│   │   │   ├── search.ts        # GET /api/skills/search
│   │   │   ├── install.ts       # POST /api/skills/install
│   │   │   ├── collections.ts    # GET/POST /api/collections
│   │   │   └── tracking.ts      # POST /api/track-download
│   │   ├── lib/
│   │   │   ├── tibrain.ts       # TiBrain API client
│   │   │   ├── cors.ts          # CORS headers
│   │   │   └── auth.ts          # Auth (Clerk)
│   │   └── index.ts
│   ├── package.json
│   └── README.md
├── dashboard/                    # UI (Astro + React + Tailwind)
│   ├── src/
│   │   ├── pages/
│   │   │   ├── index.astro      # Homepage
│   │   │   ├── skills/          # Skill listing
│   │   │   │   └── [slug].astro # Skill detail
│   │   │   ├── collections/     # User collections
│   │   │   └── api/             # Astro API routes
│   │   │       ├── skills.ts
│   │   │       ├── search.ts
│   │   │       ├── install.ts
│   │   │       └── tracking.ts
│   │   ├── components/          # React components
│   │   │   ├── SkillCard.tsx
│   │   │   ├── SearchBar.tsx
│   │   │   ├── FilterPanel.tsx
│   │   │   └── CollectionManager.tsx
│   │   └── lib/
│   │       ├── constants.ts     # Featured items, categories
│   │       └── api.ts           # API client
│   ├── public/
│   │   ├── skills.json          # Generated skill catalog
│   │   └── trending-data.json   # Download stats
│   ├── astro.config.mjs
│   ├── tailwind.config.mjs
│   └── package.json
├── cli-tool/                     # CLI tool for installation
│   ├── src/
│   │   ├── index.ts             # Main CLI entry
│   │   ├── commands/
│   │   │   ├── install.ts       # Install skill
│   │   │   ├── search.ts        # Search skills
│   │   │   └── list.ts          # List skills
│   │   └── lib/
│   │       ├── tibrain.ts       # TiBrain API client
│   │       └── catalog.ts       # Skill catalog
│   ├── package.json
│   └── README.md
├── database/                     # Database migrations
│   └── migrations/
│       ├── 001_initial.sql      # Initial schema
│       ├── 002_collections.sql  # User collections
│       └── 003_tracking.sql     # Usage tracking
├── scripts/                      # Utility scripts
│   ├── generate_skills_json.py  # Generate catalog
│   ├── sync_tibrain.py          # Sync skills to TiBrain
│   └── validate_skills.py       # Validate skill format
├── docs/                         # Documentation
│   ├── skills.json              # Generated catalog
│   └── README.md
├── package.json                  # Root package.json
├── README.md
└── CLAUDE.md                     # Guidance for Claude Code
```

---

## 🗄️ Database Schema

### Tables

**skills**
```sql
CREATE TABLE skills (
  id VARCHAR(255) PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  category VARCHAR(100),
  source VARCHAR(100), -- 'awesome-omni', 'best-source'
  file_path TEXT,
  content TEXT,
  quality_score INT DEFAULT 80,
  security_score INT DEFAULT 80,
  downloads INT DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**collections**
```sql
CREATE TABLE collections (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**collection_skills**
```sql
CREATE TABLE collection_skills (
  collection_id UUID REFERENCES collections(id) ON DELETE CASCADE,
  skill_id VARCHAR(255) REFERENCES skills(id) ON DELETE CASCADE,
  added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (collection_id, skill_id)
);
```

**downloads**
```sql
CREATE TABLE downloads (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  skill_id VARCHAR(255) REFERENCES skills(id),
  user_id VARCHAR(255),
  ip_address VARCHAR(45),
  user_agent TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## 🔌 API Endpoints

### Skills

**GET /api/skills**
- List all skills with pagination
- Query params: `page`, `limit`, `category`, `source`

**GET /api/skills/search**
- Search skills by name/description
- Query params: `q`, `category`, `source`

**GET /api/skills/:id**
- Get skill details by ID

**POST /api/skills/install**
- Install skill to TiBrain
- Body: `{ skill_id, tibrain_url }`

### Collections

**GET /api/collections**
- List user collections
- Headers: `Authorization: Bearer <token>`

**POST /api/collections**
- Create collection
- Body: `{ name, description }`

**POST /api/collections/:id/skills**
- Add skill to collection
- Body: `{ skill_id }`

### Tracking

**POST /api/track-download**
- Track skill download
- Body: `{ skill_id, user_id }`

---

## 🎨 UI Components

### Homepage
- Featured skills carousel
- Trending skills
- Categories grid
- Search bar

### Skill Listing
- Filter panel (category, source, quality score)
- Skill cards (name, description, category, stats)
- Pagination

### Skill Detail
- Skill content
- Install button
- Add to collection
- Related skills

### Collections
- User's collections
- Create/edit/delete collections
- Add/remove skills

---

## 🔧 CLI Tool

### Commands

```bash
# Install skill
tibrain-skill install <skill-id>

# Search skills
tibrain-skill search <query>

# List skills
tibrain-skill list [--category <cat>] [--source <src>]

# Sync skills to TiBrain
tibrain-skill sync
```

---

## 📊 Data Flow

### Skill Catalog Generation

1. Scan awesome-omni-skills (skills/, skills_omni/)
2. Scan best_source/skills/
3. Parse skill files (SKILL.md, README.md)
4. Generate skills.json
5. Copy to dashboard/public/

### Installation Flow

1. User clicks "Install" on dashboard
2. API call: POST /api/skills/install
3. API calls TiBrain API to register tool
4. Track download in database
5. Return success/failure

### Sync Flow

1. Run sync script
2. Scan skill sources
3. Compare with TiBrain database
4. Register new tools
5. Update existing tools
6. Track stats

---

## 🚀 Deployment

### Vercel

Single Vercel project serves:
- Dashboard (www.tibrain-skills.com)
- API endpoints (api.tibrain-skills.com)

Environment Variables:
```
# Clerk Auth
PUBLIC_CLERK_PUBLISHABLE_KEY=xxx
CLERK_SECRET_KEY=xxx

# Database
DATABASE_URL=postgresql://...

# TiBrain
TIBRAIN_URL=http://localhost:1810
TIBRAIN_API_KEY=xxx
```

### CI/CD

Push to main → auto-deploy via GitHub Actions

---

## 📝 Next Steps

1. ✅ Architecture designed
2. ⏳ Create repo structure
3. ⏳ Import skills from sources
4. ⏳ Implement API backend
5. ⏳ Implement Dashboard UI
6. ⏳ Implement CLI tool
7. ⏳ Deploy to Vercel

---

**Status**: Architecture Complete
**Next**: Create repo structure
