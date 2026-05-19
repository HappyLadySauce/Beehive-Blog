CREATE SCHEMA IF NOT EXISTS content;

-- content.contents: unified content table supporting multiple content types.
-- content.contents：统一内容表，支持多种内容类型（article / note / project / experience / reflection / portfolio）。
CREATE TABLE IF NOT EXISTS content.contents (
  id              BIGSERIAL PRIMARY KEY,

  -- Content type discriminator. / 内容类型区分。
  type            VARCHAR(32) NOT NULL,
  -- Human-readable title. / 人类可读标题。
  title           VARCHAR(512) NOT NULL,
  -- URL-safe slug, unique per type among live rows. / URL 友好标识符，在活跃行内按类型唯一。
  slug            VARCHAR(512) NOT NULL,
  -- Short excerpt / summary. / 短摘要。
  excerpt         TEXT NULL,
  -- Main body in Markdown. / 正文（Markdown）。
  body            TEXT NULL,

  -- Cover image attachment. / 封面图片附件。
  cover_attachment_id BIGINT NULL
    CONSTRAINT fk_content_contents_cover_attachment
    REFERENCES attachment.attachments (id)
    ON DELETE SET NULL,

  -- Author (identity.users FK). / 作者外键。
  author_id       BIGINT NOT NULL
    CONSTRAINT fk_content_contents_author
    REFERENCES identity.users (id)
    ON DELETE CASCADE,

  -- Publication lifecycle. / 发布生命周期。
  status          VARCHAR(16) NOT NULL DEFAULT 'draft',
  -- Human read access policy. / 人类读取访问策略。
  visibility      VARCHAR(16) NOT NULL DEFAULT 'public',
  -- AI training access orthogonal to human visibility. / AI 训练访问（与人类可见性正交）。
  ai_access       VARCHAR(16) NOT NULL DEFAULT 'allowed',

  -- When the content was first published. / 首次发布时间。
  published_at    TIMESTAMPTZ NULL,
  -- Computed word count of body. / 计算得到的字数。
  word_count      INT NOT NULL DEFAULT 0,
  -- Estimated reading time in minutes. / 预计阅读分钟数。
  reading_time_minutes INT NOT NULL DEFAULT 0,

  -- Type-specific structured data (JSONB). / 类型特定的结构化数据。
  metadata        JSONB NOT NULL DEFAULT '{}',

  -- View counter for public reads. / 公开阅读的浏览计数器。
  view_count      BIGINT NOT NULL DEFAULT 0,

  -- GORM-standard timestamps. / GORM 标准时间字段。
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at      TIMESTAMPTZ NULL,

  CONSTRAINT chk_content_contents_type
    CHECK (type IN ('article', 'note', 'project', 'experience', 'reflection', 'portfolio')),
  CONSTRAINT chk_content_contents_status
    CHECK (status IN ('draft', 'review', 'published', 'archived')),
  CONSTRAINT chk_content_contents_visibility
    CHECK (visibility IN ('public', 'member', 'private')),
  CONSTRAINT chk_content_contents_ai_access
    CHECK (ai_access IN ('allowed', 'denied'))
);

-- Slug uniqueness per content type among live rows.
-- 按内容类型在活跃行内保证 slug 唯一。
CREATE UNIQUE INDEX IF NOT EXISTS ux_content_contents_type_slug
  ON content.contents (type, slug)
  WHERE deleted_at IS NULL;

-- Listing active content by type and status.
-- 按类型和状态列出活跃内容。
CREATE INDEX IF NOT EXISTS idx_content_contents_type_status
  ON content.contents (type, status)
  WHERE deleted_at IS NULL;

-- Published content ordering (public listing).
-- 已发布内容的排序索引（公开列表）。
CREATE INDEX IF NOT EXISTS idx_content_contents_published_at
  ON content.contents (published_at DESC, id DESC)
  WHERE deleted_at IS NULL AND status = 'published';

-- Author-scoped queries.
-- 作者维度的查询索引。
CREATE INDEX IF NOT EXISTS idx_content_contents_author
  ON content.contents (author_id)
  WHERE deleted_at IS NULL;

-- Soft-delete scans.
-- 软删行扫描索引。
CREATE INDEX IF NOT EXISTS idx_content_contents_deleted_at
  ON content.contents (deleted_at)
  WHERE deleted_at IS NOT NULL;

-- JSONB GIN index for metadata queries.
-- JSONB GIN 索引，用于 metadata 查询。
CREATE INDEX IF NOT EXISTS idx_content_contents_metadata
  ON content.contents USING GIN (metadata);

COMMENT ON TABLE content.contents IS
  'Unified content table. Carries multiple content types (article, note, project, experience, reflection, portfolio) to avoid the "only articles" bottleneck. Status, visibility and ai_access are independent axes. / 统一内容表。携带多种内容类型（文章、笔记、项目、经历、反思、作品集），避免"仅有文章"瓶颈。状态、可见性和 AI 访问是独立维度。';

COMMENT ON COLUMN content.contents.type IS
  'Content type discriminator: article | note | project | experience | reflection | portfolio. / 内容类型区分。';
COMMENT ON COLUMN content.contents.title IS
  'Human-readable title, rendered in list views and detail pages. / 人类可读标题，在列表视图和详情页中渲染。';
COMMENT ON COLUMN content.contents.slug IS
  'URL-safe slug, unique per type among non-deleted rows. / URL 友好的标识符，在未删行内按类型唯一。';
COMMENT ON COLUMN content.contents.excerpt IS
  'Short excerpt or summary rendered in list views. / 列表视图中渲染的短摘要。';
COMMENT ON COLUMN content.contents.body IS
  'Main content body in Markdown format. / Markdown 格式的正文。';
COMMENT ON COLUMN content.contents.cover_attachment_id IS
  'FK to attachment.attachments for the cover image. NULL means no cover. / 封面图片的附件外键。NULL 表示无封面。';
COMMENT ON COLUMN content.contents.author_id IS
  'FK to identity.users; the content author. Cascade-deletes the content when the user is deleted. / 内容作者的用户外键；用户删除时级联删除内容。';
COMMENT ON COLUMN content.contents.status IS
  'Publication status: draft | review | published | archived. / 发布状态。';
COMMENT ON COLUMN content.contents.visibility IS
  'Human read access: public | member | private. / 人类读取访问策略。';
COMMENT ON COLUMN content.contents.ai_access IS
  'AI training access: allowed | denied. Separate from human visibility. / AI 训练访问，与人类可见性分离。';
COMMENT ON COLUMN content.contents.published_at IS
  'Timestamp when status first transitioned to published. NULL if never published. / 首次变为 published 状态的时间戳。从未发布则为 NULL。';
COMMENT ON COLUMN content.contents.word_count IS
  'Computed word count of body. Updated on save. / 正文的计算字数，保存时更新。';
COMMENT ON COLUMN content.contents.reading_time_minutes IS
  'Estimated reading time derived from word_count (word_count / 200, min 1). / 根据字数估算的阅读分钟数（字数÷200，最少 1 分钟）。';
COMMENT ON COLUMN content.contents.metadata IS
  'Type-specific structured data as JSONB. e.g. project: {repo_url, tech_stack, live_url}; experience: {company, role, start_date, end_date, location}; reflection: {mood, context}. / 类型特定的结构化数据，JSONB 格式。';
COMMENT ON COLUMN content.contents.view_count IS
  'View counter incremented on each public read. / 每次公开读取时递增的浏览计数器。';
COMMENT ON COLUMN content.contents.created_at IS
  'Row creation timestamp, maintained by GORM CreatedAt. / 行创建时间，由 GORM CreatedAt 维护。';
COMMENT ON COLUMN content.contents.updated_at IS
  'Row last-update timestamp, maintained by GORM UpdatedAt. / 行最近更新时间，由 GORM UpdatedAt 维护。';
COMMENT ON COLUMN content.contents.deleted_at IS
  'Soft-deletion timestamp aligned with gorm.DeletedAt. / 与 gorm.DeletedAt 对齐的软删时间戳。';
