package graph

// schemaSQL creates the claim graph tables if they don't already exist.
// Session markdown remains the archive; claim rows and claim_vectors are the
// retrieval units used by search and tripwire checks.
const schemaSQL = `
CREATE TABLE IF NOT EXISTS sessions (
  id            TEXT PRIMARY KEY,
  project_id    TEXT     NOT NULL DEFAULT '',
  domain        TEXT     NOT NULL DEFAULT '',
  topic         TEXT     NOT NULL DEFAULT '',
  title         TEXT     NOT NULL DEFAULT '',
  prompt        TEXT     NOT NULL DEFAULT '',
  created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  full_markdown TEXT     NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_sessions_project ON sessions(project_id);
CREATE INDEX IF NOT EXISTS idx_sessions_domain_topic ON sessions(domain, topic);

CREATE TABLE IF NOT EXISTS claims (
  id                         TEXT PRIMARY KEY,
  session_id                 TEXT     NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
  text                       TEXT     NOT NULL,
  type                       TEXT     NOT NULL DEFAULT 'decision',
  status                     TEXT     NOT NULL DEFAULT 'active',
  importance                 INTEGER  NOT NULL DEFAULT 3,
  source_span                TEXT     NOT NULL DEFAULT '',
  scope_json                 TEXT     NOT NULL DEFAULT '{}',
  aliases_json               TEXT     NOT NULL DEFAULT '[]',
  retrieval_triggers_json    TEXT     NOT NULL DEFAULT '[]',
  blast_radius_json          TEXT     NOT NULL DEFAULT '{}',
  interrupt_conditions_json  TEXT     NOT NULL DEFAULT '[]'
);
CREATE INDEX IF NOT EXISTS idx_claims_session ON claims(session_id);
CREATE INDEX IF NOT EXISTS idx_claims_status ON claims(status);
CREATE INDEX IF NOT EXISTS idx_claims_type ON claims(type);

CREATE TABLE IF NOT EXISTS claim_vectors (
  id        TEXT PRIMARY KEY,
  claim_id  TEXT NOT NULL REFERENCES claims(id) ON DELETE CASCADE,
  kind      TEXT NOT NULL CHECK(kind IN ('claim','retrieval','interrupt')),
  provider  TEXT NOT NULL DEFAULT '',
  dims      INTEGER NOT NULL DEFAULT 0,
  text      TEXT NOT NULL,
  embedding BLOB NOT NULL,
  UNIQUE(claim_id, kind)
);
CREATE INDEX IF NOT EXISTS idx_claim_vectors_claim ON claim_vectors(claim_id);
CREATE INDEX IF NOT EXISTS idx_claim_vectors_kind ON claim_vectors(kind);
CREATE INDEX IF NOT EXISTS idx_claim_vectors_provider_dims ON claim_vectors(provider, dims);

CREATE TABLE IF NOT EXISTS embedding_config (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS edges (
  id            TEXT PRIMARY KEY,
  from_claim_id TEXT NOT NULL REFERENCES claims(id) ON DELETE CASCADE,
  to_claim_id   TEXT NOT NULL REFERENCES claims(id) ON DELETE CASCADE,
  type          TEXT NOT NULL CHECK(type IN (
    'CONSTRAINS','IMPLEMENTS','CAUSED_BY','SUPERSEDES','CONFLICTS_WITH','RELATED_CANDIDATE'
  )),
  confidence    REAL NOT NULL DEFAULT 0,
  evidence      TEXT NOT NULL DEFAULT '',
  source        TEXT NOT NULL DEFAULT '',
  status        TEXT NOT NULL DEFAULT 'active',
  created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(from_claim_id, to_claim_id, type)
);
CREATE INDEX IF NOT EXISTS idx_edges_from ON edges(from_claim_id);
CREATE INDEX IF NOT EXISTS idx_edges_to ON edges(to_claim_id);
CREATE INDEX IF NOT EXISTS idx_edges_type_status ON edges(type, status);

CREATE TABLE IF NOT EXISTS feedback (
  id          TEXT PRIMARY KEY,
  target_type TEXT NOT NULL,
  target_id   TEXT NOT NULL,
  action      TEXT NOT NULL,
  details     TEXT NOT NULL DEFAULT '',
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_feedback_target ON feedback(target_type, target_id);

CREATE TABLE IF NOT EXISTS interrupt_events (
  id              TEXT PRIMARY KEY,
  event_type      TEXT NOT NULL,
  project_id      TEXT NOT NULL DEFAULT '',
  event_json      TEXT NOT NULL,
  candidates_json TEXT NOT NULL DEFAULT '[]',
  user_action     TEXT NOT NULL DEFAULT '',
  created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_interrupt_events_project ON interrupt_events(project_id, event_type, created_at);

-- Legacy Section 2 tables are retained so existing graph.db files keep opening.
CREATE TABLE IF NOT EXISTS context_nodes (
  id           TEXT PRIMARY KEY,
  domain       TEXT    NOT NULL DEFAULT '',
  topic        TEXT    NOT NULL DEFAULT '',
  title        TEXT    NOT NULL DEFAULT '',
  prompt       TEXT    NOT NULL DEFAULT '',
  date         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  summary_text TEXT    NOT NULL DEFAULT '',
  embedding    BLOB
);

CREATE TABLE IF NOT EXISTS context_edges (
  from_id    TEXT REFERENCES context_nodes(id),
  to_id      TEXT REFERENCES context_nodes(id),
  edge_type  TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (from_id, to_id, edge_type)
);
`
