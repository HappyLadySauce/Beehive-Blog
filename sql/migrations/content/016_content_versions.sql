-- content.content_versions: revision history snapshots for content.
-- content.content_versions：内容的版本/修订历史快照。
CREATE TABLE IF NOT EXISTS content.content_versions (
  id              BIGSERIAL PRIMARY KEY,

  -- FK to content.contents. Cascade-deletes with content. / 内容外键，级联删除。
  content_id      BIGINT NOT NULL
    CONSTRAINT fk_content_versions_content
    REFERENCES content.contents (id)
    ON DELETE CASCADE,

  -- Monotonic version number per content_id. / 每个内容单调递增的版本号。
  version_number  INT NOT NULL,

  -- Snapshot of content fields at this version. / 此版本的内容字段快照。
  title           VARCHAR(512) NOT NULL,
  body            TEXT NULL,
  excerpt         TEXT NULL,

  -- Optional human-readable summary of changes. / 可选的可读变更说明。
  change_summary  VARCHAR(512) NULL,

  -- Who created this version. / 版本创建者。
  created_by      BIGINT NOT NULL
    CONSTRAINT fk_content_versions_created_by
    REFERENCES identity.users (id),

  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT ux_content_versions_content_version
    UNIQUE (content_id, version_number)
);

-- Ordered listing for a single content item.
-- 单个内容的版本有序列表。
CREATE INDEX IF NOT EXISTS idx_content_versions_content_id
  ON content.content_versions (content_id, version_number DESC);

COMMENT ON TABLE content.content_versions IS
  'Revision history for content. Snapshots of title, body, excerpt at each version. / 内容的修订历史。每个版本的标题、正文和摘要的快照。';

COMMENT ON COLUMN content.content_versions.content_id IS
  'FK to content.contents. / 内容外键。';
COMMENT ON COLUMN content.content_versions.version_number IS
  'Monotonic version number; auto-incremented per content_id. / 每个内容单调递增的版本号。';
COMMENT ON COLUMN content.content_versions.title IS
  'Snapshot of title at this version. / 此版本的标题快照。';
COMMENT ON COLUMN content.content_versions.body IS
  'Snapshot of body (Markdown) at this version. / 此版本的正文（Markdown）快照。';
COMMENT ON COLUMN content.content_versions.excerpt IS
  'Snapshot of excerpt at this version. / 此版本的摘要快照。';
COMMENT ON COLUMN content.content_versions.change_summary IS
  'Optional human-readable summary of what changed. / 可选的可读变更说明。';
COMMENT ON COLUMN content.content_versions.created_by IS
  'FK to identity.users; the admin who created this version snapshot. / 创建此版本的管理员用户外键。';
COMMENT ON COLUMN content.content_versions.created_at IS
  'Version creation timestamp. / 版本创建时间。';
