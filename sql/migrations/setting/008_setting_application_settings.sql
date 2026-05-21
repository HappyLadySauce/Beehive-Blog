-- setting.application_settings: singleton application configuration (JSONB + revision for hot reload).
-- setting.application_settings：单行应用配置（JSONB + revision 支持热加载）。
CREATE SCHEMA IF NOT EXISTS setting;

CREATE TABLE IF NOT EXISTS setting.application_settings (
  id SMALLINT PRIMARY KEY DEFAULT 1
    CONSTRAINT chk_setting_application_singleton CHECK (id = 1),

  revision BIGINT NOT NULL DEFAULT 1,

  payload JSONB NOT NULL DEFAULT '{}'::jsonb,

  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL,

  CONSTRAINT chk_setting_application_revision_positive CHECK (revision >= 1)
);

COMMENT ON TABLE setting.application_settings IS
  'Singleton row (id=1) holding JSON application settings; revision increments on each successful write for cache invalidation. / id=1 单行 JSON 应用配置；revision 在每次成功写入后递增用于缓存失效。';
COMMENT ON COLUMN setting.application_settings.revision IS
  'Monotonic version bumped on each persist; used for O(1) hot-reload probes. / 单调版本号，持久化后递增，用于 O(1) 热加载探测。';
COMMENT ON COLUMN setting.application_settings.payload IS
  'JSON document (e.g. email SMTP subtree). / JSON 文档（如 email SMTP 子树）。';

-- Seed singleton with default email subtree (SMTP disabled).
-- No attachment key — file service settings are managed via storage_drivers/storage_mounts.
-- 种子数据：默认 email 子树（SMTP 关闭）。不含 attachment key，文件服务改由 storage_drivers/storage_mounts 管理。
INSERT INTO setting.application_settings (id, revision, payload, created_at, updated_at)
VALUES (
  1,
  1,
  jsonb_build_object(
    'email', jsonb_build_object(
      'enabled', false,
      'host', '',
      'port', 587,
      'username', '',
      'password', '',
      'from', '',
      'from_name', '',
      'tls', 'starttls'
    )
  ),
  NOW(),
  NOW()
)
ON CONFLICT (id) DO NOTHING;

-- NOTIFY on application_settings changes for cross-process hot reload (LISTEN setting_revision).
-- 在 application_settings 变更时 NOTIFY，供跨进程热加载（LISTEN setting_revision）。
CREATE OR REPLACE FUNCTION setting.fn_notify_application_changed()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
  PERFORM pg_notify('setting_revision', NEW.revision::text);
  RETURN NEW;
END;
$$;

COMMENT ON FUNCTION setting.fn_notify_application_changed() IS
  'Emits pg_notify on setting_revision with new revision after INSERT/UPDATE. / 在 INSERT/UPDATE 后向 setting_revision 发送 pg_notify，负载为新 revision。';

DROP TRIGGER IF EXISTS trg_setting_application_notify ON setting.application_settings;

CREATE TRIGGER trg_setting_application_notify
  AFTER INSERT OR UPDATE ON setting.application_settings
  FOR EACH ROW
  EXECUTE PROCEDURE setting.fn_notify_application_changed();

COMMENT ON TRIGGER trg_setting_application_notify ON setting.application_settings IS
  'Pushes setting_revision NOTIFY so app instances can refresh in-memory settings without polling. / 推送 setting_revision NOTIFY，使应用实例无需轮询即可刷新内存设置。';
