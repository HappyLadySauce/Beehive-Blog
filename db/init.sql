-- ============================================
-- Beehive-Blog 数据库主初始化脚本
-- 执行方式: psql -U postgres -d beehive_blog -f init.sql
-- ============================================

\echo '========================================'
\echo 'Beehive-Blog 数据库初始化开始'
\echo '========================================'

-- 加载各模块SQL（\ir 按当前脚本目录解析）
\ir 001_users.sql
\ir 002_articles.sql
\ir 003_comments.sql
\ir 004_attachments.sql
\ir 005_settings.sql
\ir 006_notifications.sql
\ir 007_seed.sql
\ir 008_triggers.sql
\ir 010_comment_counts_approved_only.sql
\ir 011_article_versions_index.sql
\ir 012_article_versions_autosave.sql
\ir 013_attachment_refs.sql
\ir 014_hexo_settings.sql

\echo '========================================'
\echo 'Beehive-Blog 数据库初始化完成'
\echo '========================================'
