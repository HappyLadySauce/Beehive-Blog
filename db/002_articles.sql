-- ============================================
-- 2. 文章相关表
-- ============================================

-- 分类表
CREATE TABLE IF NOT EXISTS categories (
    id BIGSERIAL PRIMARY KEY COMMENT '分类ID',
    name VARCHAR(50) NOT NULL COMMENT '分类名称',
    slug VARCHAR(50) NOT NULL UNIQUE COMMENT 'URL别名',
    description VARCHAR(255) COMMENT '分类描述',
    parent_id BIGINT REFERENCES categories(id) ON DELETE SET NULL COMMENT '父分类ID(支持多级分类)',
    article_count BIGINT NOT NULL DEFAULT 0 COMMENT '文章数量',
    sort_order INT NOT NULL DEFAULT 0 COMMENT '排序权重',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间'
);

CREATE INDEX idx_categories_parent_id ON categories(parent_id);
CREATE INDEX idx_categories_slug ON categories(slug);

COMMENT ON TABLE categories IS '分类表';

-- 标签表
CREATE TABLE IF NOT EXISTS tags (
    id BIGSERIAL PRIMARY KEY COMMENT '标签ID',
    name VARCHAR(50) NOT NULL UNIQUE COMMENT '标签名称',
    slug VARCHAR(50) NOT NULL UNIQUE COMMENT 'URL别名',
    color VARCHAR(10) NOT NULL DEFAULT '#3B82F6' COMMENT '标签颜色(十六进制)',
    description VARCHAR(255) COMMENT '标签描述',
    article_count BIGINT NOT NULL DEFAULT 0 COMMENT '关联文章数量',
    sort_order INT NOT NULL DEFAULT 0 COMMENT '排序权重',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间'
);

CREATE INDEX idx_tags_slug ON tags(slug);

COMMENT ON TABLE tags IS '标签表';

-- 文章表
CREATE TABLE IF NOT EXISTS articles (
    id BIGSERIAL PRIMARY KEY COMMENT '文章ID',
    title VARCHAR(200) NOT NULL COMMENT '文章标题',
    slug VARCHAR(100) UNIQUE COMMENT 'URL别名',
    content TEXT NOT NULL COMMENT '文章内容(Markdown)',
    summary VARCHAR(500) COMMENT '文章摘要',
    cover_image VARCHAR(500) COMMENT '封面图片URL',
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived', 'private', 'scheduled')) COMMENT '状态: draft-草稿, published-已发布, archived-已归档, private-私密, scheduled-定时发布',
    password VARCHAR(100) COMMENT '访问密码(私密文章)',
    is_pinned BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否置顶',
    pin_order INT NOT NULL DEFAULT 0 COMMENT '置顶排序权重',
    view_count BIGINT NOT NULL DEFAULT 0 COMMENT '浏览次数',
    like_count BIGINT NOT NULL DEFAULT 0 COMMENT '点赞次数',
    comment_count BIGINT NOT NULL DEFAULT 0 COMMENT '评论次数',
    author_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE COMMENT '作者ID',
    category_id BIGINT REFERENCES categories(id) ON DELETE SET NULL COMMENT '分类ID',
    published_at TIMESTAMP COMMENT '发布时间',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP COMMENT '删除时间(软删除)'
);

CREATE INDEX idx_articles_slug ON articles(slug);
CREATE INDEX idx_articles_status ON articles(status);
CREATE INDEX idx_articles_author_id ON articles(author_id);
CREATE INDEX idx_articles_category_id ON articles(category_id);
CREATE INDEX idx_articles_published_at ON articles(published_at);
CREATE INDEX idx_articles_is_pinned ON articles(is_pinned, pin_order DESC);
CREATE INDEX idx_articles_deleted_at ON articles(deleted_at);

COMMENT ON TABLE articles IS '文章表';

-- 文章标签关联表
CREATE TABLE IF NOT EXISTS article_tags (
    article_id BIGINT NOT NULL REFERENCES articles(id) ON DELETE CASCADE COMMENT '文章ID',
    tag_id BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE COMMENT '标签ID',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (article_id, tag_id)
);

CREATE INDEX idx_article_tags_tag_id ON article_tags(tag_id);

COMMENT ON TABLE article_tags IS '文章标签关联表';

-- 文章版本历史表
CREATE TABLE IF NOT EXISTS article_versions (
    id BIGSERIAL PRIMARY KEY COMMENT '版本ID',
    article_id BIGINT NOT NULL REFERENCES articles(id) ON DELETE CASCADE COMMENT '文章ID',
    title VARCHAR(200) NOT NULL COMMENT '文章标题',
    content TEXT NOT NULL COMMENT '文章内容',
    version INT NOT NULL COMMENT '版本号',
    created_by BIGINT REFERENCES users(id) COMMENT '创建者ID',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'
);

CREATE INDEX idx_article_versions_article_id ON article_versions(article_id);

COMMENT ON TABLE article_versions IS '文章版本历史表';

-- 文章点赞表
CREATE TABLE IF NOT EXISTS article_likes (
    id BIGSERIAL PRIMARY KEY COMMENT '点赞ID',
    article_id BIGINT NOT NULL REFERENCES articles(id) ON DELETE CASCADE COMMENT '文章ID',
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE COMMENT '用户ID',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    UNIQUE (article_id, user_id)
);

CREATE INDEX idx_article_likes_user_id ON article_likes(user_id);

COMMENT ON TABLE article_likes IS '文章点赞表';

-- 用户收藏表
CREATE TABLE IF NOT EXISTS user_favorites (
    id BIGSERIAL PRIMARY KEY COMMENT '收藏ID',
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE COMMENT '用户ID',
    article_id BIGINT NOT NULL REFERENCES articles(id) ON DELETE CASCADE COMMENT '文章ID',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    UNIQUE (user_id, article_id)
);

CREATE INDEX idx_user_favorites_article_id ON user_favorites(article_id);

COMMENT ON TABLE user_favorites IS '用户收藏表';

-- 文章浏览记录表
CREATE TABLE IF NOT EXISTS article_view_logs (
    id BIGSERIAL PRIMARY KEY COMMENT '浏览记录ID',
    article_id BIGINT NOT NULL REFERENCES articles(id) ON DELETE CASCADE COMMENT '文章ID',
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL COMMENT '用户ID',
    ip VARCHAR(50) COMMENT 'IP地址',
    user_agent VARCHAR(500) COMMENT '用户代理',
    view_duration INT NOT NULL DEFAULT 0 COMMENT '浏览时长(秒)',
    is_valid BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否有效阅读(>=3分钟)',
    viewed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '浏览时间'
);

CREATE INDEX idx_article_view_logs_article_id ON article_view_logs(article_id);
CREATE INDEX idx_article_view_logs_user_id ON article_view_logs(user_id);
CREATE INDEX idx_article_view_logs_viewed_at ON article_view_logs(viewed_at);

COMMENT ON TABLE article_view_logs IS '文章浏览记录表';
