CREATE TABLE IF NOT EXISTS content.comments (
    id BIGSERIAL PRIMARY KEY,
    content_id BIGINT NOT NULL REFERENCES content.contents(id) ON DELETE CASCADE,
    nickname VARCHAR(80) NOT NULL,
    email_hash VARCHAR(64) NOT NULL,
    website VARCHAR(512),
    body TEXT NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'review',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_content_comments_content_status_created
    ON content.comments(content_id, status, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_content_comments_status_created
    ON content.comments(status, created_at DESC)
    WHERE deleted_at IS NULL;
