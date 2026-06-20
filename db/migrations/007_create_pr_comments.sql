CREATE TABLE IF NOT EXISTS pr_comments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    context_id  UUID NOT NULL REFERENCES contexts(id),
    repo        VARCHAR(255) NOT NULL,
    pr_number   INT NOT NULL,
    comment_url TEXT,
    posted_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
