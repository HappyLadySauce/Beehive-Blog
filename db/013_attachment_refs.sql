-- ============================================
-- 13. 附件父子与文章引用（兼容旧库增量迁移，可重复执行）
-- 全新 init 时 004 已含下列结构，本脚本多为 no-op
-- ============================================

ALTER TABLE attachments ADD COLUMN IF NOT EXISTS parent_id BIGINT;
ALTER TABLE attachments ADD COLUMN IF NOT EXISTS variant VARCHAR(32);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'attachments'::regclass
          AND conname = 'attachments_parent_id_fkey'
    ) THEN
        ALTER TABLE attachments
            ADD CONSTRAINT attachments_parent_id_fkey
            FOREIGN KEY (parent_id) REFERENCES attachments(id) ON DELETE CASCADE;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_attachments_parent_id ON attachments(parent_id);

CREATE TABLE IF NOT EXISTS article_attachments (
    article_id BIGINT NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    attachment_id BIGINT NOT NULL REFERENCES attachments(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (article_id, attachment_id)
);

CREATE INDEX IF NOT EXISTS idx_article_attachments_attachment_id ON article_attachments(attachment_id);

COMMENT ON TABLE article_attachments IS '文章与附件引用关系（正文 URL）';
