-- Seed the storage_drivers table with the three built-in drivers.
-- Uses ON CONFLICT DO NOTHING so the migration is safe to re-run.
-- 种子三个内置驱动到 storage_drivers 表。
-- 使用 ON CONFLICT DO NOTHING 确保重复执行安全。

INSERT INTO attachment.storage_drivers (name, display_name, description, config_schema, capabilities)
VALUES (
    'local',
    'Local Storage',
    'Server-local filesystem storage. Files are stored under a configurable root directory.',
    '{"type":"object","properties":{"root":{"type":"string","title":"Root Directory","description":"Absolute or relative path to the storage root."}},"required":["root"]}',
    '{"upload":true,"download":true,"delete":true,"presign":false,"health_check":true}'
)
ON CONFLICT DO NOTHING;

-- Seed one default local mount so uploads have a database-owned default target.
-- 创建默认 local 挂载项，使未指定 storage_mount_id 的上传由数据库配置决定。
INSERT INTO attachment.storage_mounts (
    driver_name,
    mount_path,
    name,
    config,
    order_index,
    is_default,
    disabled,
    status
)
VALUES (
    'local',
    '/local',
    'Local Storage',
    '{"root":"data/attachments"}',
    0,
    true,
    false,
    'unknown'
)
ON CONFLICT DO NOTHING;

-- Seed disabled remote placeholders so old s3/oss rows can be mapped during
-- destructive migration. Administrators must edit config before using them.
-- 创建禁用的远端占位 mount，便于破坏式迁移映射旧 s3/oss 行。
-- 管理员必须补齐配置后才能启用使用。
INSERT INTO attachment.storage_mounts (
    driver_name,
    mount_path,
    name,
    config,
    order_index,
    is_default,
    disabled,
    status,
    last_error
)
VALUES
(
    's3',
    '/s3',
    'S3 Storage',
    '{"bucket":"","upload_base_url":"","download_base_url":""}',
    10,
    false,
    true,
    'error',
    'remote storage config requires bucket, upload_base_url and download_base_url'
),
(
    'oss',
    '/oss',
    'OSS Storage',
    '{"bucket":"","upload_base_url":"","download_base_url":""}',
    20,
    false,
    true,
    'error',
    'remote storage config requires bucket, upload_base_url and download_base_url'
)
ON CONFLICT DO NOTHING;

INSERT INTO attachment.storage_drivers (name, display_name, description, config_schema, capabilities)
VALUES (
    's3',
    'Amazon S3 / Compatible',
    'S3-compatible object storage. Supports presigned upload and download URLs.',
    '{"type":"object","properties":{"bucket":{"type":"string","title":"Bucket"},"upload_base_url":{"type":"string","title":"Upload Base URL"},"download_base_url":{"type":"string","title":"Download Base URL"}},"required":["bucket","upload_base_url","download_base_url"]}',
    '{"upload":false,"download":false,"delete":false,"presign":true,"health_check":true}'
)
ON CONFLICT DO NOTHING;

INSERT INTO attachment.storage_drivers (name, display_name, description, config_schema, capabilities)
VALUES (
    'oss',
    'Alibaba Cloud OSS',
    'Alibaba Cloud Object Storage Service. Supports presigned upload and download URLs.',
    '{"type":"object","properties":{"bucket":{"type":"string","title":"Bucket"},"upload_base_url":{"type":"string","title":"Upload Base URL"},"download_base_url":{"type":"string","title":"Download Base URL"}},"required":["bucket","upload_base_url","download_base_url"]}',
    '{"upload":false,"download":false,"delete":false,"presign":true,"health_check":true}'
)
ON CONFLICT DO NOTHING;
