# Beehive-Blog 数据库文档

## 数据库信息

- **数据库类型**: PostgreSQL
- **数据库名称**: beehive_blog
- **字符集**: UTF-8

## 文件说明

| 文件 | 说明 |
|------|------|
| `init.sql` | 主初始化脚本，按顺序加载所有模块 |
| `001_users.sql` | 用户相关表（users, user_levels） |
| `002_articles.sql` | 文章相关表（articles, categories, tags, article_tags 等） |
| `003_comments.sql` | 评论相关表（comments, comment_likes） |
| `004_attachments.sql` | 附件相关表（attachments, storage_policies, attachment_groups） |
| `005_settings.sql` | 系统设置相关表（settings, links, themes, menus 等） |
| `006_notifications.sql` | 通知相关表（notifications, subscriptions, webhooks） |
| `007_seed.sql` | 初始数据（等级配置、默认设置、菜单等） |
| `008_triggers.sql` | 数据库触发器（自动更新时间戳、计数器） |
| `014_hexo_settings.sql` | Hexo 运行时设置默认行（`settings.group=hexo`；存量库可单独执行） |

## 已有环境升级

若仓库新增了 `014_hexo_settings.sql` 等增量脚本，在已存在的数据库上执行（幂等）：

```bash
psql -U postgres -d beehive_blog -f db/014_hexo_settings.sql
```

全量 `init.sql` 已包含该文件，**新库**无需单独执行。

## 快速开始

### 1. 创建数据库

```bash
# 创建数据库
createdb -U postgres beehive_blog

# 或使用 psql
psql -U postgres -c "CREATE DATABASE beehive_blog WITH ENCODING='UTF8';"
```

### 2. 执行初始化脚本

```bash
# 方式一：使用主初始化脚本
psql -U postgres -d beehive_blog -f init.sql

# 方式二：使用环境变量
export PGHOST=localhost
export PGPORT=5432
export PGUSER=postgres
export PGPASSWORD=your_password
export PGDATABASE=beehive_blog
psql -f init.sql
```

### 3. 验证安装

```sql
-- 查看所有表
\dt

-- 查看用户等级配置
SELECT * FROM user_levels;

-- 查看系统设置
SELECT * FROM settings;

-- 查看存储策略
SELECT * FROM storage_policies;
```

## 数据库表结构

### 核心表关系

```
users (用户)
  ├── articles (文章) [author_id]
  │     ├── article_tags (文章标签关联)
  │     ├── article_likes (点赞)
  │     ├── article_versions (版本历史)
  │     └── article_view_logs (浏览记录)
  ├── comments (评论) [user_id]
  ├── user_favorites (收藏)
  ├── notifications (通知)
  └── notification_settings (通知设置)

categories (分类)
  └── articles [category_id]

tags (标签)
  └── article_tags (文章标签关联)

attachments (附件)
  ├── storage_policies (存储策略)
  └── attachment_groups (分组)

settings (系统设置)
links (友情链接)
themes (主题)
menus (菜单)
  └── menu_items (菜单项)
pages (独立页面)

webhooks (Webhook配置)
  └── webhook_logs (调用日志)

subscriptions (邮件订阅)
```

## 触发器说明

### 自动更新时间戳

所有包含 `updated_at` 字段的表都会自动更新时间戳。

### 计数器自动更新

| 触发器 | 说明 |
|--------|------|
| `update_tag_article_count` | 标签文章数量自动更新 |
| `update_category_article_count` | 分类文章数量自动更新 |
| `update_user_comment_count` | 用户评论数量自动更新 |
| `update_article_comment_count` | 文章评论数量自动更新 |
| `update_article_like_count` | 文章点赞数量自动更新 |
| `update_comment_like_count` | 评论点赞数量自动更新 |

## 索引说明

### 主要索引

- **用户表**: email, role, status
- **文章表**: slug, status, author_id, category_id, published_at
- **评论表**: article_id, user_id, status
- **附件表**: type, policy_id, uploaded_by
- **通知表**: user_id, is_read

### 性能优化建议

1. 定期执行 `VACUUM ANALYZE` 清理和统计
2. 大表考虑分区（如 article_view_logs）
3. 根据查询模式添加复合索引

## 备份与恢复

### 备份

```bash
# 完整备份
pg_dump -U postgres beehive_blog > backup_$(date +%Y%m%d).sql

# 仅数据
pg_dump -U postgres --data-only beehive_blog > data_backup.sql

# 仅结构
pg_dump -U postgres --schema-only beehive_blog > schema_backup.sql
```

### 恢复

```bash
psql -U postgres -d beehive_blog -f backup_20260326.sql
```

## 常用查询

### 统计查询

```sql
-- 文章统计
SELECT status, COUNT(*) FROM articles GROUP BY status;

-- 用户活跃度
SELECT 
    u.username,
    u.level,
    u.comment_count,
    u.article_view_count
FROM users u
ORDER BY u.article_view_count DESC
LIMIT 10;

-- 热门标签
SELECT t.name, t.article_count
FROM tags t
ORDER BY t.article_count DESC
LIMIT 10;
```

### 清理操作

```sql
-- 清理90天前的浏览记录
DELETE FROM article_view_logs 
WHERE viewed_at < CURRENT_TIMESTAMP - INTERVAL '90 days';

-- 清理已删除的文章（物理删除）
DELETE FROM articles WHERE deleted_at IS NOT NULL;

-- 清理未验证的订阅
DELETE FROM subscriptions 
WHERE verified_at IS NULL AND created_at < CURRENT_TIMESTAMP - INTERVAL '7 days';
```
