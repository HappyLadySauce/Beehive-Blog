-- ============================================
-- 4. 附件相关表
-- ============================================

-- 存储策略表
CREATE TABLE IF NOT EXISTS storage_policies (
    id BIGSERIAL PRIMARY KEY COMMENT '策略ID',
    name VARCHAR(50) NOT NULL COMMENT '策略名称',
    type VARCHAR(20) NOT NULL CHECK (type IN ('local', 'aliyun-oss', 'aws-s3', 'minio')) COMMENT '存储类型: local-本地, aliyun-oss-阿里云OSS, aws-s3-AWS S3, minio-MinIO',
    is_default BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否为默认策略',
    base_url VARCHAR(500) COMMENT '访问域名',
    upload_path VARCHAR(255) COMMENT '上传路径前缀',
    config JSONB COMMENT '存储配置(JSON格式,包含密钥等)',
    sort_order INT NOT NULL DEFAULT 0 COMMENT '排序权重',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间'
);

CREATE INDEX idx_storage_policies_type ON storage_policies(type);
CREATE INDEX idx_storage_policies_is_default ON storage_policies(is_default);

COMMENT ON TABLE storage_policies IS '存储策略表';

-- 附件分组表
CREATE TABLE IF NOT EXISTS attachment_groups (
    id BIGSERIAL PRIMARY KEY COMMENT '分组ID',
    name VARCHAR(50) NOT NULL COMMENT '分组名称',
    parent_id BIGINT REFERENCES attachment_groups(id) ON DELETE SET NULL COMMENT '父分组ID(支持多级分组)',
    sort_order INT NOT NULL DEFAULT 0 COMMENT '排序权重',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间'
);

CREATE INDEX idx_attachment_groups_parent_id ON attachment_groups(parent_id);

COMMENT ON TABLE attachment_groups IS '附件分组表';

-- 附件表
CREATE TABLE IF NOT EXISTS attachments (
    id BIGSERIAL PRIMARY KEY COMMENT '附件ID',
    name VARCHAR(255) NOT NULL COMMENT '存储文件名',
    original_name VARCHAR(255) COMMENT '原始文件名',
    path VARCHAR(500) NOT NULL COMMENT '存储路径',
    url VARCHAR(500) NOT NULL COMMENT '访问URL',
    thumb_url VARCHAR(500) COMMENT '缩略图URL',
    type VARCHAR(20) NOT NULL CHECK (type IN ('image', 'document', 'video', 'audio', 'other')) COMMENT '附件类型: image-图片, document-文档, video-视频, audio-音频, other-其他',
    mime_type VARCHAR(100) COMMENT 'MIME类型',
    size BIGINT NOT NULL COMMENT '文件大小(字节)',
    width INT COMMENT '图片宽度',
    height INT COMMENT '图片高度',
    policy_id BIGINT NOT NULL REFERENCES storage_policies(id) ON DELETE RESTRICT COMMENT '存储策略ID',
    group_id BIGINT REFERENCES attachment_groups(id) ON DELETE SET NULL COMMENT '分组ID',
    uploaded_by BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT COMMENT '上传者ID',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间'
);

CREATE INDEX idx_attachments_type ON attachments(type);
CREATE INDEX idx_attachments_policy_id ON attachments(policy_id);
CREATE INDEX idx_attachments_group_id ON attachments(group_id);
CREATE INDEX idx_attachments_uploaded_by ON attachments(uploaded_by);
CREATE INDEX idx_attachments_created_at ON attachments(created_at);

COMMENT ON TABLE attachments IS '附件表';
