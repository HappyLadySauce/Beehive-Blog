-- Extend attachment.attachments with storage-mount, file-node, and
-- provider-metadata columns. Also relax the storage location CHECK constraint
-- so that rows with a valid storage_mount_id no longer need the old
-- local_path / bucket / object_key mutual-exclusivity rule.
-- 扩展 attachment.attachments，增加存储挂载引用、文件节点引用
-- 和提供方元数据列。同时放宽存储位置 CHECK 约束，
-- 使持有合法 storage_mount_id 的行不再受旧 mutual-exclusivity 规则限制。

-- 1) Add new columns.
-- 1) 添加新列。
ALTER TABLE attachment.attachments
    ADD COLUMN storage_mount_id BIGINT,
    ADD COLUMN file_node_id     BIGINT,
    ADD COLUMN storage_metadata JSONB NOT NULL DEFAULT '{}';

-- 2) Foreign keys.
-- 2) 外键。
ALTER TABLE attachment.attachments
    ADD CONSTRAINT fk_attachment_storage_mount
        FOREIGN KEY (storage_mount_id) REFERENCES attachment.storage_mounts (id),
    ADD CONSTRAINT fk_attachment_file_node
        FOREIGN KEY (file_node_id) REFERENCES attachment.file_nodes (id);

-- 3) Relax the storage-location constraint to permit mount-based storage.
--    When storage_mount_id IS NOT NULL the driver + config come from the
--    mount row, so local_path / bucket / object_key are optional.
-- 3) 放宽存储位置约束以允许 mount-based 存储。
--    当 storage_mount_id IS NOT NULL 时，驱动与配置从 mount 行获取，
--    local_path / bucket / object_key 为可选字段。
ALTER TABLE attachment.attachments
    DROP CONSTRAINT chk_attachment_storage_location;

ALTER TABLE attachment.attachments
    ADD CONSTRAINT chk_attachment_storage_location
        CHECK (
            storage_mount_id IS NOT NULL
            OR
            (
                (
                    storage_type = 'local'
                    AND local_path IS NOT NULL
                    AND bucket IS NULL
                    AND object_key IS NULL
                )
                OR
                (
                    storage_type IN ('s3', 'oss')
                    AND bucket IS NOT NULL
                    AND object_key IS NOT NULL
                    AND local_path IS NULL
                )
            )
        );

COMMENT ON COLUMN attachment.attachments.storage_mount_id IS
    'References attachment.storage_mounts(id). Resolves driver_name and config at upload/download time. / 关联 storage_mounts(id)，在上传/下载时解析驱动名与配置。';
COMMENT ON COLUMN attachment.attachments.file_node_id IS
    'Optional reference to attachment.file_nodes(id) for Studio file browsing. / 可选关联 file_nodes(id)，用于 Studio 文件浏览。';
COMMENT ON COLUMN attachment.attachments.storage_metadata IS
    'Provider-specific metadata such as version id, headers, etag details. / 提供方扩展元数据，如 version id、headers、etag 详情。';
