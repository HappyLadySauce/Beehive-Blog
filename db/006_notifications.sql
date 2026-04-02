-- ============================================
-- 6. 通知相关表
-- ============================================

-- 站内通知表
CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('system', 'comment', 'article', 'user', 'like')),
    title VARCHAR(200) NOT NULL,
    content TEXT,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    source_id VARCHAR(50),
    source_type VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    read_at TIMESTAMP
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_is_read ON notifications(is_read);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);

COMMENT ON TABLE notifications IS '站内通知表';

-- 用户通知设置表
CREATE TABLE IF NOT EXISTS notification_settings (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    email_on_comment BOOLEAN NOT NULL DEFAULT TRUE,
    email_on_like BOOLEAN NOT NULL DEFAULT TRUE,
    email_on_follow BOOLEAN NOT NULL DEFAULT TRUE,
    email_on_system BOOLEAN NOT NULL DEFAULT TRUE,
    site_on_comment BOOLEAN NOT NULL DEFAULT TRUE,
    site_on_like BOOLEAN NOT NULL DEFAULT TRUE,
    site_on_follow BOOLEAN NOT NULL DEFAULT TRUE,
    site_on_system BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE notification_settings IS '用户通知设置表';

-- 邮件订阅表
CREATE TABLE IF NOT EXISTS subscriptions (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(100) NOT NULL,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('all', 'category', 'tag', 'author')),
    target_id VARCHAR(50),
    frequency VARCHAR(20) NOT NULL DEFAULT 'realtime' CHECK (frequency IN ('realtime', 'daily', 'weekly')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    verify_token VARCHAR(100),
    verified_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (email, type, target_id)
);

CREATE INDEX idx_subscriptions_email ON subscriptions(email);
CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_type ON subscriptions(type);

COMMENT ON TABLE subscriptions IS '邮件订阅表';

-- Webhook配置表
CREATE TABLE IF NOT EXISTS webhooks (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    url VARCHAR(500) NOT NULL,
    secret VARCHAR(255),
    events JSONB,
    method VARCHAR(10) NOT NULL DEFAULT 'POST',
    headers JSONB,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_triggered_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_webhooks_is_enabled ON webhooks(is_enabled);

COMMENT ON TABLE webhooks IS 'Webhook配置表';

-- Webhook调用日志表
CREATE TABLE IF NOT EXISTS webhook_logs (
    id BIGSERIAL PRIMARY KEY,
    webhook_id BIGINT NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event VARCHAR(50) NOT NULL,
    payload JSONB,
    response TEXT,
    status_code INT,
    is_success BOOLEAN NOT NULL DEFAULT FALSE,
    retry_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_webhook_logs_webhook_id ON webhook_logs(webhook_id);
CREATE INDEX idx_webhook_logs_created_at ON webhook_logs(created_at);

COMMENT ON TABLE webhook_logs IS 'Webhook调用日志表';
