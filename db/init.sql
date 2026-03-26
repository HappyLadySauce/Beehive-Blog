-- ============================================
-- Beehive-Blog 数据库主初始化脚本
-- 执行方式: psql -U postgres -d beehive_blog -f init.sql
-- ============================================

\echo '========================================'
\echo 'Beehive-Blog 数据库初始化开始'
\echo '========================================'

-- 加载各模块SQL
\i 001_users.sql
\i 002_articles.sql
\i 003_comments.sql
\i 004_attachments.sql
\i 005_settings.sql
\i 006_notifications.sql
\i 007_seed.sql
\i 008_triggers.sql

\echo '========================================'
\echo 'Beehive-Blog 数据库初始化完成'
\echo '========================================'
