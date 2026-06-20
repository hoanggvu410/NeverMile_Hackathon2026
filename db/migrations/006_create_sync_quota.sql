CREATE TABLE IF NOT EXISTS sync_quota_usage (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    month      DATE NOT NULL,
    sync_count INT NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, month)
);
