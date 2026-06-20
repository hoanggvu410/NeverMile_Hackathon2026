CREATE TABLE IF NOT EXISTS contexts (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    local_id              VARCHAR(50) UNIQUE NOT NULL,
    user_id               UUID NOT NULL REFERENCES users(id),
    team_id               UUID REFERENCES teams(id),
    prompt                TEXT NOT NULL,
    reasoning             TEXT,
    decisions             TEXT,
    rejected_alternatives TEXT,
    trade_offs            TEXT,
    files                 JSONB NOT NULL DEFAULT '[]',
    commits               JSONB NOT NULL DEFAULT '[]',
    domain                VARCHAR(200),
    topic                 VARCHAR(200),
    agent                 VARCHAR(50),
    model                 VARCHAR(100),
    is_published          BOOLEAN NOT NULL DEFAULT FALSE,
    repo_name             VARCHAR(255),
    context_ts            TIMESTAMPTZ NOT NULL,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS ix_contexts_user_id      ON contexts(user_id);
CREATE INDEX IF NOT EXISTS ix_contexts_team_id      ON contexts(team_id);
CREATE INDEX IF NOT EXISTS ix_contexts_is_published ON contexts(is_published);
CREATE INDEX IF NOT EXISTS ix_contexts_domain       ON contexts(domain);
