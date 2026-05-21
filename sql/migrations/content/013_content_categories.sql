-- content.categories: category definitions for content taxonomy.
-- content.categories：用于内容分类的分类定义。
CREATE TABLE IF NOT EXISTS content.categories (
  id          BIGSERIAL PRIMARY KEY,

  name        VARCHAR(64) NOT NULL,
  -- URL-safe unique slug among live rows. / 活跃行内 URL 友好的唯一标识符。
  slug        VARCHAR(64) NOT NULL,
  description TEXT NULL,
  -- Optional parent category for simple hierarchy. / 可选父级分类，用于简单层级关系。
  parent_id   BIGINT NULL
    CONSTRAINT fk_content_categories_parent
    REFERENCES content.categories (id)
    ON DELETE SET NULL,
  -- Manual sort order for display ordering. / 手动排序，用于展示顺序。
  sort_order  INT NOT NULL DEFAULT 0,

  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at  TIMESTAMPTZ NULL
);

-- Slug uniqueness among live rows.
-- 活跃行内 slug 唯一。
CREATE UNIQUE INDEX IF NOT EXISTS ux_content_categories_slug
  ON content.categories (slug)
  WHERE deleted_at IS NULL;

-- content.content_categories: many-to-many junction between content and categories.
-- content.content_categories：内容与分类的多对多联结表。
CREATE TABLE IF NOT EXISTS content.content_categories (
  content_id  BIGINT NOT NULL
    CONSTRAINT fk_content_categories_content
    REFERENCES content.contents (id)
    ON DELETE CASCADE,
  category_id BIGINT NOT NULL
    CONSTRAINT fk_content_categories_category
    REFERENCES content.categories (id)
    ON DELETE CASCADE,

  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  PRIMARY KEY (content_id, category_id)
);

-- Reverse lookup: all content for a given category.
-- 反向查找：某个分类下的所有内容。
CREATE INDEX IF NOT EXISTS idx_content_categories_category_id
  ON content.content_categories (category_id, content_id);

COMMENT ON TABLE content.categories IS
  'Category definitions for content taxonomy. Each category has a unique slug among live rows. / 内容分类的分类定义，每个分类在活跃行内有唯一标识符。';

COMMENT ON COLUMN content.categories.name IS
  'Display name of the category. / 分类展示名称。';
COMMENT ON COLUMN content.categories.slug IS
  'URL-safe unique identifier among live rows. / 活跃行内 URL 友好的唯一标识符。';
COMMENT ON COLUMN content.categories.description IS
  'Optional description of the category. / 分类的可选描述。';
COMMENT ON COLUMN content.categories.parent_id IS
  'Optional parent category reference for simple hierarchy. / 可选父级分类引用，用于简单层级关系。';
COMMENT ON COLUMN content.categories.sort_order IS
  'Manual sort order for display ordering. / 手动排序，用于控制展示顺序。';
COMMENT ON COLUMN content.categories.created_at IS
  'Row creation timestamp, maintained by GORM CreatedAt. / 行创建时间，由 GORM CreatedAt 维护。';
COMMENT ON COLUMN content.categories.updated_at IS
  'Row last-update timestamp, maintained by GORM UpdatedAt. / 行最近更新时间，由 GORM UpdatedAt 维护。';
COMMENT ON COLUMN content.categories.deleted_at IS
  'Soft-deletion timestamp aligned with gorm.DeletedAt. / 与 gorm.DeletedAt 对齐的软删时间戳。';

COMMENT ON TABLE content.content_categories IS
  'Many-to-many junction between content.contents and content.categories. / 内容与分类的多对多联结表。';

COMMENT ON COLUMN content.content_categories.content_id IS
  'FK to content.contents. / 内容外键。';
COMMENT ON COLUMN content.content_categories.category_id IS
  'FK to content.categories. / 分类外键。';
COMMENT ON COLUMN content.content_categories.created_at IS
  'Junction row creation timestamp. / 联结行创建时间。';
