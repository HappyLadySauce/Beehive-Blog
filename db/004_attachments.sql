-- ============================================
-- 4. 附件相关表
-- ============================================

-- 存储策略表
CREATE TABLE IF NOT EXISTS storage_policies (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('local', 'aliyun-oss', 'aws-s3', 'minio')),
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    base_url VARCHAR(500),
    upload_path VARCHAR(255),
    config JSONB,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_storage_policies_type ON storage_policies(type);
CREATE INDEX idx_storage_policies_is_default ON storage_policies(is_default);

COMMENT ON TABLE storage_policies IS '存储策略表';

-- 附件分组表
CREATE TABLE IF NOT EXISTS attachment_groups (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    parent_id BIGINT REFERENCES attachment_groups(id) ON DELETE SET NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_attachment_groups_parent_id ON attachment_groups(parent_id);

COMMENT ON TABLE attachment_groups IS '附件分组表';

-- 附件表
CREATE TABLE IF NOT EXISTS attachments (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    original_name VARCHAR(255),
    path VARCHAR(500) NOT NULL,
    url VARCHAR(500) NOT NULL,
    thumb_url VARCHAR(500),
    type VARCHAR(20) NOT NULL CHECK (type IN ('image', 'document', 'video', 'audio', 'other')),
    mime_type VARCHAR(100),
    size BIGINT NOT NULL,
    width INT,
    height INT,
    policy_id BIGINT NOT NULL REFERENCES storage_policies(id) ON DELETE RESTRICT,
    group_id BIGINT REFERENCES attachment_groups(id) ON DELETE SET NULL,
    uploaded_by BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_attachments_type ON attachments(type);
CREATE INDEX idx_attachments_policy_id ON attachments(policy_id);
CREATE INDEX idx_attachments_group_id ON attachments(group_id);
CREATE INDEX idx_attachments_uploaded_by ON attachments(uploaded_by);
CREATE INDEX idx_attachments_created_at ON attachments(created_at);

COMMENT ON TABLE attachments IS '附件表';
