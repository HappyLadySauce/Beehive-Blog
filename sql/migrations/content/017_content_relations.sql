-- content.content_relations: directed relationships between content rows.
-- content.content_relations：内容行之间的有向关系。
CREATE TABLE IF NOT EXISTS content.content_relations (
  id                  BIGSERIAL PRIMARY KEY,

  -- Source content (e.g. the article referencing a project). / 源内容。
  source_content_id   BIGINT NOT NULL
    CONSTRAINT fk_content_relations_source
    REFERENCES content.contents (id)
    ON DELETE CASCADE,

  -- Target content (e.g. the project being referenced). / 目标内容。
  target_content_id   BIGINT NOT NULL
    CONSTRAINT fk_content_relations_target
    REFERENCES content.contents (id)
    ON DELETE CASCADE,

  -- Relationship type. / 关系类型。
  relation_type       VARCHAR(32) NOT NULL,
  -- Optional human-readable label for the relation. / 关系的可选可读标签。
  label               VARCHAR(128) NULL,
  -- Ordering within the same source+type. / 在相同源+类型内的排序。
  sort_order          INT NOT NULL DEFAULT 0,

  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT chk_content_relations_type
    CHECK (relation_type IN ('references', 'part_of', 'derived_from', 'follows')),
  CONSTRAINT ux_content_relations_pair
    UNIQUE (source_content_id, target_content_id, relation_type),
  CONSTRAINT chk_content_relations_no_self
    CHECK (source_content_id <> target_content_id)
);

-- Listing all outgoing relations from a piece of content.
-- 列出某个内容的所有出向关系。
CREATE INDEX IF NOT EXISTS idx_content_relations_source
  ON content.content_relations (source_content_id, sort_order, id);

-- Listing all incoming relations to a piece of content.
-- 列出某个内容的所有入向关系。
CREATE INDEX IF NOT EXISTS idx_content_relations_target
  ON content.content_relations (target_content_id, sort_order, id);

COMMENT ON TABLE content.content_relations IS
  'Directed relationships between content rows, e.g. Project->Article, Experience->Project. / 内容行之间的有向关系。';

COMMENT ON COLUMN content.content_relations.source_content_id IS
  'FK to content.contents; the source of the relationship. / 关系源内容的外键。';
COMMENT ON COLUMN content.content_relations.target_content_id IS
  'FK to content.contents; the target of the relationship. / 关系目标内容的外键。';
COMMENT ON COLUMN content.content_relations.relation_type IS
  'Type: references | part_of | derived_from | follows. / 关系类型。';
COMMENT ON COLUMN content.content_relations.label IS
  'Optional human-readable label for display. / 用于展示的可选标签。';
COMMENT ON COLUMN content.content_relations.sort_order IS
  'Ordering among siblings with the same source+type. / 相同源+类型的兄弟关系之间的排序。';
COMMENT ON COLUMN content.content_relations.created_at IS
  'Row creation timestamp. / 行创建时间。';
