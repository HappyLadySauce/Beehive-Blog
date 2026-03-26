-- ============================================
-- 6. 通知相关表
-- ============================================

-- 站内通知表
CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY COMMENT '通知ID',
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE COMMENT '用户ID',
    type VARCHAR(20) NOT NULL CHECK (type IN ('system', 'comment', 'article', 'user', 'like')) COMMENT '通知类型: system-系统, comment-评论, article-文章, user-用户, like-点赞',
    title VARCHAR(200) NOT NULL COMMENT '通知标题',
    content TEXT COMMENT '通知内容',
    is_read BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否已读',
    source_id VARCHAR(50) COMMENT '关联资源ID',
    source_type VARCHAR(50) COMMENT '关联资源类型',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    read_at TIMESTAMP COMMENT '阅读时间'
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_is_read ON notifications(is_read);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);

COMMENT ON TABLE notifications IS '站内通知表';

-- 用户通知设置表
CREATE TABLE IF NOT EXISTS notification_settings (
    id BIGSERIAL PRIMARY KEY COMMENT '设置ID',
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE COMMENT '用户ID',
    email_on_comment BOOLEAN NOT NULL DEFAULT TRUE COMMENT '评论回复邮件通知',
    email_on_like BOOLEAN NOT NULL DEFAULT TRUE COMMENT '点赞邮件通知',
    email_on_follow BOOLEAN NOT NULL DEFAULT TRUE COMMENT '关注邮件通知',
    email_on_system BOOLEAN NOT NULL DEFAULT TRUE COMMENT '系统公告邮件通知',
    site_on_comment BOOLEAN NOT NULL DEFAULT TRUE COMMENT '评论回复站内通知',
    site_on_like BOOLEAN NOT NULL DEFAULT TRUE COMMENT '点赞站内通知',
    site_on_follow BOOLEAN NOT NULL DEFAULT TRUE COMMENT '关注站内通知',
    site_on_system BOOLEAN NOT NULL DEFAULT TRUE COMMENT '系统公告站内通知',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间'
);

COMMENT ON TABLE notification_settings IS '用户通知设置表';

-- 邮件订阅表
CREATE TABLE IF NOT EXISTS subscriptions (
    id BIGSERIAL PRIMARY KEY COMMENT '订阅ID',
    email VARCHAR(100) NOT NULL COMMENT '邮箱地址',
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL COMMENT '用户ID',
    type VARCHAR(20) NOT NULL CHECK (type IN ('all', 'category', 'tag', 'author')) COMMENT '订阅类型: all-全站, category-分类, tag-标签, author-作者',
    target_id VARCHAR(50) COMMENT '订阅目标ID(分类/标签/作者ID)',
    frequency VARCHAR(20) NOT NULL DEFAULT 'realtime' CHECK (frequency IN ('realtime', 'daily', 'weekly')) COMMENT '推送频率: realtime-实时, daily-每日, weekly-每周',
    is_active BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否激活',
    verify_token VARCHAR(100) COMMENT '验证令牌',
    verified_at TIMESTAMP COMMENT '验证时间',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE (email, type, target_id)
);

CREATE INDEX idx_subscriptions_email ON subscriptions(email);
CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_type ON subscriptions(type);

COMMENT ON TABLE subscriptions IS '邮件订阅表';

-- Webhook配置表
CREATE TABLE IF NOT EXISTS webhooks (
    id BIGSERIAL PRIMARY KEY COMMENT 'WebhookID',
    name VARCHAR(50) NOT NULL COMMENT 'Webhook名称',
    url VARCHAR(500) NOT NULL COMMENT 'WebhookURL',
    secret VARCHAR(255) COMMENT '签名密钥',
    events JSONB COMMENT '触发事件列表(JSON数组)',
    method VARCHAR(10) NOT NULL DEFAULT 'POST' COMMENT '请求方法',
    headers JSONB COMMENT '自定义请求头(JSON对象)',
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否启用',
    last_triggered_at TIMESTAMP COMMENT '最后触发时间',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间'
);

CREATE INDEX idx_webhooks_is_enabled ON webhooks(is_enabled);

COMMENT ON TABLE webhooks IS 'Webhook配置表';

-- Webhook调用日志表
CREATE TABLE IF NOT EXISTS webhook_logs (
    id BIGSERIAL PRIMARY KEY COMMENT '日志ID',
    webhook_id BIGINT NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE COMMENT 'WebhookID',
    event VARCHAR(50) NOT NULL COMMENT '触发事件',
    payload JSONB COMMENT '请求数据',
    response TEXT COMMENT '响应内容',
    status_code INT COMMENT 'HTTP状态码',
    is_success BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否成功',
    retry_count INT NOT NULL DEFAULT 0 COMMENT '重试次数',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'
);

CREATE INDEX idx_webhook_logs_webhook_id ON webhook_logs(webhook_id);
CREATE INDEX idx_webhook_logs_created_at ON webhook_logs(created_at);

COMMENT ON TABLE webhook_logs IS 'Webhook调用日志表';
