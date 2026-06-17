-- 034_expand_memory_types.sql
-- Expand the memories table CHECK constraint to include `working` and `long_term`
-- memory types introduced by CognitiveMemoryManager.
--
-- SQLite does not support modifying CHECK constraints in place, so we:
--   1. Create the new table as memories_new
--   2. Copy data from memories
--   3. Drop the old table
--   4. Rename memories_new -> memories
--   5. Recreate indexes

CREATE TABLE IF NOT EXISTS memories_new (
  id TEXT PRIMARY KEY,
  api_key_id TEXT NOT NULL DEFAULT 'local',
  session_id TEXT,
  type TEXT NOT NULL CHECK(type IN ('factual', 'episodic', 'procedural', 'semantic', 'working', 'long_term')),
  key TEXT,
  content TEXT NOT NULL,
  metadata TEXT,
  importance REAL NOT NULL DEFAULT 0.5,
  access_count INTEGER NOT NULL DEFAULT 0,
  last_access TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  expires_at TEXT
);

INSERT INTO memories_new (id, api_key_id, session_id, type, key, content, metadata, importance, access_count, last_access, created_at, updated_at, expires_at)
SELECT
  id,
  COALESCE(api_key_id, 'local'),
  session_id,
  type,
  key,
  content,
  metadata,
  0.5,
  0,
  NULL,
  created_at,
  updated_at,
  expires_at
FROM memories;

DROP TABLE memories;

ALTER TABLE memories_new RENAME TO memories;

CREATE INDEX IF NOT EXISTS idx_memories_api_key ON memories(api_key_id);
CREATE INDEX IF NOT EXISTS idx_memories_session ON memories(session_id);
CREATE INDEX IF NOT EXISTS idx_memories_type ON memories(type);
CREATE INDEX IF NOT EXISTS idx_memories_expires ON memories(expires_at);
CREATE INDEX IF NOT EXISTS idx_memories_importance ON memories(importance);
