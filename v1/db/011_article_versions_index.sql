-- ============================================
-- 11. 文章版本历史索引
-- 加速 MAX(version) 查询与按 article_id 分页列表
-- ============================================

CREATE INDEX IF NOT EXISTS idx_article_versions_article_id_version
  ON article_versions(article_id, version DESC);
