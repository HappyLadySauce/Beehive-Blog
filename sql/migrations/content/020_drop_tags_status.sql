-- Drop tag lifecycle status; deletion uses deleted_at soft-delete only.
-- 移除标签生命周期 status；删除仅通过 deleted_at 软删。
ALTER TABLE content.tags DROP CONSTRAINT IF EXISTS chk_content_tags_status;
ALTER TABLE content.tags DROP COLUMN IF EXISTS status;
