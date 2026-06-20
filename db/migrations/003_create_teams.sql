CREATE TABLE IF NOT EXISTS teams (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255) NOT NULL,
    owner_id   UUID NOT NULL REFERENCES users(id),
    plan       VARCHAR(20) NOT NULL DEFAULT 'team',
    sync_quota INT NOT NULL DEFAULT -1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
