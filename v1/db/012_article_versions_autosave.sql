-- 文章版本：自动保存单槽（每篇文章至多一条 is_autosave 记录，可反复覆盖）
ALTER TABLE article_versions
  ADD COLUMN IF NOT EXISTS is_autosave BOOLEAN NOT NULL DEFAULT FALSE;

CREATE UNIQUE INDEX IF NOT EXISTS uq_article_versions_article_autosave
  ON article_versions (article_id)
  WHERE is_autosave = TRUE;

COMMENT ON COLUMN article_versions.is_autosave IS 'true 表示该行为自动保存快照（每文最多一条，覆盖更新）';
