-- setting.application_settings: singleton application configuration (JSONB + revision for hot reload).
-- setting.application_settings：单行应用配置（JSONB + revision 支持热加载）。
-- Run after identity migrations (basename sort: 000_setting_* after 013_*).
-- 在 identity 迁移之后执行（按文件名 000_setting_* 在 013_* 之后）。
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
-- 种子数据：默认 email 子树（SMTP 关闭）。
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
