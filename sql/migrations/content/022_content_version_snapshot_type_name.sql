-- Add snapshot type and name to content versions.
-- 为内容版本增加快照类型与版本名称。
ALTER TABLE content.content_versions
  ADD COLUMN IF NOT EXISTS snapshot_type VARCHAR(16) NOT NULL DEFAULT 'manual',
  ADD COLUMN IF NOT EXISTS name VARCHAR(128) NOT NULL DEFAULT '';

-- Keep one auto snapshot per content item.
-- 每篇内容只保留一个自动快照。
CREATE UNIQUE INDEX IF NOT EXISTS ux_content_versions_single_auto
  ON content.content_versions (content_id, snapshot_type)
  WHERE snapshot_type = 'auto';

COMMENT ON COLUMN content.content_versions.snapshot_type IS
  'Snapshot type: manual or auto. / 快照类型：manual 或 auto。';
COMMENT ON COLUMN content.content_versions.name IS
  'Human-readable version name. / 可读版本名称。';
