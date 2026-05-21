-- Converge public content types to article / note / project.
-- 将公开内容类型收敛为 article / note / project。

WITH live_after_mapping AS (
  SELECT
    id,
    slug,
    type AS old_type,
    CASE
      WHEN type IN ('experience', 'reflection') THEN 'note'
      WHEN type = 'portfolio' THEN 'project'
      ELSE type
    END AS target_type
  FROM content.contents
  WHERE deleted_at IS NULL
),
ranked_live_slugs AS (
  SELECT
    id,
    old_type,
    ROW_NUMBER() OVER (
      PARTITION BY target_type, slug
      ORDER BY CASE WHEN old_type IN ('article', 'note', 'project') THEN 0 ELSE 1 END, id
    ) AS slug_rank
  FROM live_after_mapping
)
UPDATE content.contents AS c
SET slug = LEFT(c.slug, 480) || '-' || c.id::TEXT
FROM ranked_live_slugs AS r
WHERE c.id = r.id
  AND r.slug_rank > 1;

UPDATE content.contents
SET type = CASE
  WHEN type IN ('experience', 'reflection') THEN 'note'
  WHEN type = 'portfolio' THEN 'project'
  ELSE type
END
WHERE type IN ('experience', 'reflection', 'portfolio');

ALTER TABLE content.contents
  DROP CONSTRAINT IF EXISTS chk_content_contents_type;

ALTER TABLE content.contents
  ADD CONSTRAINT chk_content_contents_type
  CHECK (type IN ('article', 'note', 'project'));

COMMENT ON TABLE content.contents IS
  'Unified content table. Carries article, note, and project content. Status, visibility and ai_access are independent axes. / 统一内容表。承载文章、笔记、项目内容。状态、可见性和 AI 访问是独立维度。';

COMMENT ON COLUMN content.contents.type IS
  'Content type discriminator: article | note | project. / 内容类型区分。';

COMMENT ON COLUMN content.contents.metadata IS
  'Type-specific structured data as JSONB. e.g. project: {repo_url, tech_stack, live_url}. / 类型特定的结构化数据，JSONB 格式。';
