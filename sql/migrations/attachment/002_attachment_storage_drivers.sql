-- attachment.storage_drivers: driver template registry.
-- Maps each supported storage driver to its display name, config schema,
-- capabilities, and status. The Go code registry is still authoritative for
-- whether a driver is truly available; this table serves admin UI rendering
-- and driver discovery.
-- attachment.storage_drivers：驱动模板注册表。
-- 记录每个受支持存储驱动的展示名、配置 schema、能力集合和状态。
-- Go 代码 registry 仍是驱动是否真正可用的权威来源；
-- 本表用于管理端 UI 渲染与驱动发现。
CREATE TABLE attachment.storage_drivers (
    id            BIGSERIAL PRIMARY KEY,
    name          VARCHAR(64)  NOT NULL,
    display_name  VARCHAR(128) NOT NULL,
    description   TEXT,
    config_schema JSONB        NOT NULL DEFAULT '{}',
    capabilities  JSONB        NOT NULL DEFAULT '{}',
    status        VARCHAR(16)  NOT NULL DEFAULT 'active',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ,

    CONSTRAINT chk_storage_driver_status CHECK (status IN ('active', 'disabled'))
);

-- One active driver per name.
-- 每个驱动名最多一条活跃行。
CREATE UNIQUE INDEX ux_storage_drivers_name
    ON attachment.storage_drivers (name)
    WHERE deleted_at IS NULL;

COMMENT ON COLUMN attachment.storage_drivers.config_schema IS
    'JSON Schema for rendering driver config forms in Studio. / 用于 Studio 渲染驱动配置表单的 JSON Schema。';
COMMENT ON COLUMN attachment.storage_drivers.capabilities IS
    'Capability flags e.g. upload, download, presign, delete, health-check. / 能力标记，如上传、下载、预签名、删除、健康检查。';
COMMENT ON COLUMN attachment.storage_drivers.status IS
    'active | disabled. Drivers can be disabled via admin UI without code changes. / 管理员可在 UI 中禁用驱动而无需修改代码。';
