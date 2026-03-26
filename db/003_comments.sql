-- ============================================
-- 3. 评论相关表
-- ============================================

-- 评论表
CREATE TABLE IF NOT EXISTS comments (
    id BIGSERIAL PRIMARY KEY COMMENT '评论ID',
    content TEXT NOT NULL COMMENT '评论内容',
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'spam')) COMMENT '状态: pending-待审核, approved-已通过, rejected-已拒绝, spam-垃圾评论',
    article_id BIGINT NOT NULL REFERENCES articles(id) ON DELETE CASCADE COMMENT '文章ID',
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL COMMENT '用户ID(游客评论为NULL)',
    parent_id BIGINT REFERENCES comments(id) ON DELETE CASCADE COMMENT '父评论ID(支持嵌套回复)',
    author_name VARCHAR(50) COMMENT '游客评论者名称',
    author_email VARCHAR(100) COMMENT '游客评论者邮箱',
    author_ip VARCHAR(50) COMMENT '评论者IP',
    like_count BIGINT NOT NULL DEFAULT 0 COMMENT '点赞次数',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间'
);

CREATE INDEX idx_comments_article_id ON comments(article_id);
CREATE INDEX idx_comments_user_id ON comments(user_id);
CREATE INDEX idx_comments_parent_id ON comments(parent_id);
CREATE INDEX idx_comments_status ON comments(status);

COMMENT ON TABLE comments IS '评论表';

-- 评论点赞表
CREATE TABLE IF NOT EXISTS comment_likes (
    id BIGSERIAL PRIMARY KEY COMMENT '点赞ID',
    comment_id BIGINT NOT NULL REFERENCES comments(id) ON DELETE CASCADE COMMENT '评论ID',
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE COMMENT '用户ID',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    UNIQUE (comment_id, user_id)
);

CREATE INDEX idx_comment_likes_user_id ON comment_likes(user_id);

COMMENT ON TABLE comment_likes IS '评论点赞表';
