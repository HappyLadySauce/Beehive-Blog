-- ============================================
-- 可选：已有数据库收敛为一级分类（移除 parent_id）
-- 在新库上请直接执行 db/002_articles.sql；本脚本仅供从旧结构升级时手动执行。
-- ============================================

ALTER TABLE categories DROP CONSTRAINT IF EXISTS categories_parent_id_fkey;
DROP INDEX IF EXISTS idx_categories_parent_id;
ALTER TABLE categories DROP COLUMN IF EXISTS parent_id;
