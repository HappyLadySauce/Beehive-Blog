-- content.tags: tag definitions for content categorization.
-- content.tags：用于内容分类的标签定义。
CREATE TABLE IF NOT EXISTS content.tags (
  id          BIGSERIAL PRIMARY KEY,

  name        VARCHAR(64) NOT NULL,
  -- URL-safe unique slug among live rows. / 活跃行内 URL 友好的唯一标识符。
  slug        VARCHAR(64) NOT NULL,
  description TEXT NULL,
  -- Hex color code including #, e.g. #FF5733. / 十六进制颜色码，包含 #。
  color       VARCHAR(7) NULL,
  -- Lifecycle status. / 生命周期状态。
  status      VARCHAR(16) NOT NULL DEFAULT 'active',

  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at  TIMESTAMPTZ NULL,

  CONSTRAINT chk_content_tags_status
    CHECK (status IN ('active', 'archived')),
  CONSTRAINT chk_content_tags_color_format
    CHECK (color IS NULL OR color ~ '^#[0-9A-Fa-f]{6}$')
);

-- Slug uniqueness among live rows.
-- 活跃行内 slug 唯一。
CREATE UNIQUE INDEX IF NOT EXISTS ux_content_tags_slug
  ON content.tags (slug)
  WHERE deleted_at IS NULL;

-- content.content_tags: many-to-many junction between content and tags.
-- content.content_tags：内容与标签的多对多联结表。
CREATE TABLE IF NOT EXISTS content.content_tags (
  content_id  BIGINT NOT NULL
    CONSTRAINT fk_content_tags_content
    REFERENCES content.contents (id)
    ON DELETE CASCADE,
  tag_id      BIGINT NOT NULL
    CONSTRAINT fk_content_tags_tag
    REFERENCES content.tags (id)
    ON DELETE CASCADE,

  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  PRIMARY KEY (content_id, tag_id)
);

-- Reverse lookup: all content for a given tag.
-- 反向查找：某个标签下的所有内容。
CREATE INDEX IF NOT EXISTS idx_content_tags_tag_id
  ON content.content_tags (tag_id, content_id);

COMMENT ON TABLE content.tags IS
  'Tag definitions for content categorization. Each tag has a unique slug among live rows. / 内容分类的标签定义，每个标签在活跃行内有唯一标识符。';

COMMENT ON COLUMN content.tags.name IS
  'Display name of the tag. / 标签展示名称。';
COMMENT ON COLUMN content.tags.slug IS
  'URL-safe unique identifier among live rows. / 活跃行内 URL 友好的唯一标识符。';
COMMENT ON COLUMN content.tags.description IS
  'Optional description of the tag. / 标签的可选描述。';
COMMENT ON COLUMN content.tags.color IS
  'Optional hex color code including the # prefix, e.g. #FF5733. / 可选十六进制颜色码，包含 # 前缀。';
COMMENT ON COLUMN content.tags.status IS
  'Lifecycle: active | archived. / 生命周期状态。';
COMMENT ON COLUMN content.tags.created_at IS
  'Row creation timestamp, maintained by GORM CreatedAt. / 行创建时间，由 GORM CreatedAt 维护。';
COMMENT ON COLUMN content.tags.updated_at IS
  'Row last-update timestamp, maintained by GORM UpdatedAt. / 行最近更新时间，由 GORM UpdatedAt 维护。';
COMMENT ON COLUMN content.tags.deleted_at IS
  'Soft-deletion timestamp aligned with gorm.DeletedAt. / 与 gorm.DeletedAt 对齐的软删时间戳。';

COMMENT ON TABLE content.content_tags IS
  'Many-to-many junction between content.contents and content.tags. / 内容与标签的多对多联结表。';

COMMENT ON COLUMN content.content_tags.content_id IS
  'FK to content.contents. / 内容外键。';
COMMENT ON COLUMN content.content_tags.tag_id IS
  'FK to content.tags. / 标签外键。';
COMMENT ON COLUMN content.content_tags.created_at IS
  'Junction row creation timestamp. / 联结行创建时间。';
