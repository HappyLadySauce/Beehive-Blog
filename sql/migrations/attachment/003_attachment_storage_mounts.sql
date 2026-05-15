-- attachment.storage_mounts: administrator-created storage instances.
-- Each row represents a configured instance of a storage driver (e.g. a
-- specific S3 bucket or local directory). Attachments reference mounts via
-- storage_mount_id to resolve the driver and config at upload/download time.
-- attachment.storage_mounts：管理员创建的存储实例。
-- 每一行代表一个存储驱动的配置实例（如某个 S3 bucket 或本地目录）。
-- 附件通过 storage_mount_id 引用挂载项，在上传/下载时解析驱动与配置。
CREATE TABLE attachment.storage_mounts (
    id              BIGSERIAL    PRIMARY KEY,
    driver_name     VARCHAR(64)  NOT NULL,
    mount_path      VARCHAR(512) NOT NULL,
    name            VARCHAR(128) NOT NULL,
    remark          TEXT,
    config          JSONB        NOT NULL DEFAULT '{}',
    order_index     INT          NOT NULL DEFAULT 0,
    is_default      BOOLEAN      NOT NULL DEFAULT false,
    disabled        BOOLEAN      NOT NULL DEFAULT false,
    status          VARCHAR(16)  NOT NULL DEFAULT 'unknown',
    last_checked_at TIMESTAMPTZ,
    last_error      TEXT,
    created_by      BIGINT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_storage_mount_status CHECK (status IN ('unknown', 'work', 'error')),
    CONSTRAINT chk_storage_mount_path
        CHECK (
            mount_path ~ '^/[A-Za-z0-9][A-Za-z0-9._/-]*$'
            AND mount_path <> '/'
            AND mount_path NOT LIKE '%//%'
            AND mount_path NOT LIKE '%/../%'
            AND mount_path NOT LIKE '%/./%'
            AND mount_path NOT LIKE '%/..'
            AND mount_path NOT LIKE '%/.'
            AND mount_path !~ '/$'
        )
);

-- mount_path is unique among active mounts.
-- 活跃挂载项中 mount_path 唯一。
CREATE UNIQUE INDEX ux_storage_mounts_mount_path
    ON attachment.storage_mounts (mount_path)
    WHERE deleted_at IS NULL;

-- Lookup mounts by driver for admin filtering.
-- 按驱动名查找挂载项。
CREATE INDEX idx_storage_mounts_driver_name
    ON attachment.storage_mounts (driver_name)
    WHERE deleted_at IS NULL;

-- At most one enabled default mount at a time.
-- 同时最多一个启用的默认挂载项。
CREATE UNIQUE INDEX ux_storage_mounts_default
    ON attachment.storage_mounts (is_default)
    WHERE is_default = true AND disabled = false AND deleted_at IS NULL;

-- Sortable listing for admin UI.
-- 管理端可排序列表。
CREATE INDEX idx_storage_mounts_order
    ON attachment.storage_mounts (order_index, id)
    WHERE deleted_at IS NULL;

-- Filter by status for health-check dashboard.
-- 按状态筛选用于健康检查面板。
CREATE INDEX idx_storage_mounts_status
    ON attachment.storage_mounts (status)
    WHERE deleted_at IS NULL;

COMMENT ON COLUMN attachment.storage_mounts.config IS
    'Driver-specific instance config e.g. {"root":"..."} for local, {"bucket":"..."} for s3. / 驱动实例专属配置。';
COMMENT ON COLUMN attachment.storage_mounts.is_default IS
    'Whether this mount is the default for uploads that do not specify storage_mount_id. / 未指定 storage_mount_id 时是否使用此挂载项作为默认存储。';
COMMENT ON COLUMN attachment.storage_mounts.disabled IS
    'When true new uploads are rejected; existing ready files remain readable. / 为 true 时拒绝新上传，已有 ready 文件默认仍可读取。';
COMMENT ON COLUMN attachment.storage_mounts.status IS
    'unknown | work | error. Updated by health-check endpoint or background task. / 由健康检查端点或后台任务更新。';
COMMENT ON COLUMN attachment.storage_mounts.created_by IS
    'Admin user id who created this mount. / 创建此挂载项的管理员用户 ID。';
