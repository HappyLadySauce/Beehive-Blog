-- Enable the pg_trgm extension for trigram-based fuzzy text search indexes.
-- 启用 pg_trgm 扩展，用于基于 trigram 的模糊文本搜索索引。
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Trigram index on content titles to accelerate LOWER(title) LIKE '%...%' queries.
-- 内容标题的 trigram 索引，加速模糊搜索查询。
CREATE INDEX IF NOT EXISTS idx_content_contents_title_trgm
  ON content.contents USING GIN (title gin_trgm_ops);

-- Trigram index on content excerpts for search.
-- 内容摘要的 trigram 索引，用于搜索。
CREATE INDEX IF NOT EXISTS idx_content_contents_excerpt_trgm
  ON content.contents USING GIN (COALESCE(excerpt, '') gin_trgm_ops);

-- Trigram index on tag names for search.
-- 标签名称的 trigram 索引，用于搜索。
CREATE INDEX IF NOT EXISTS idx_content_tags_name_trgm
  ON content.tags USING GIN (name gin_trgm_ops);

-- Trigram index on tag slugs for search.
-- 标签 slug 的 trigram 索引，用于搜索。
CREATE INDEX IF NOT EXISTS idx_content_tags_slug_trgm
  ON content.tags USING GIN (slug gin_trgm_ops);
