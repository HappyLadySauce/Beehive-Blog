CREATE SCHEMA IF NOT EXISTS identity;

-- identity.users: core user rows aligned with GORM soft-deletion (deleted_at).
-- identity.users：核心用户行，与 GORM 软删（deleted_at）对齐。
CREATE TABLE identity.users (
  id BIGSERIAL PRIMARY KEY,

  username VARCHAR(64) NOT NULL,
  email VARCHAR(320) NULL,
  nickname VARCHAR(128) NULL,
  phone VARCHAR(16) NULL,

  -- Avatar: stable key + storage backend; URL is derived at read time.
  -- 头像：稳定键 + 存储后端；访问 URL 在读取时拼装或签名。
  avatar_object_key TEXT NULL,
  avatar_storage VARCHAR(16) NULL,

  role VARCHAR(16) NOT NULL DEFAULT 'member',
  status VARCHAR(16) NOT NULL DEFAULT 'active',

  last_login_at TIMESTAMPTZ NULL,

  -- GORM-standard timestamps. / GORM 标准时间字段。
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL,

  CONSTRAINT chk_identity_users_role CHECK (role IN ('member', 'admin')),
  CONSTRAINT chk_identity_users_status CHECK (status IN ('pending', 'active', 'locked')),
  CONSTRAINT chk_identity_users_avatar_storage CHECK (
    avatar_storage IS NULL OR avatar_storage IN ('local', 'oss', 's3')
  ),
  CONSTRAINT chk_identity_users_avatar_pair CHECK (
    (avatar_object_key IS NULL AND avatar_storage IS NULL)
    OR (avatar_object_key IS NOT NULL AND avatar_storage IS NOT NULL)
  )
);

-- Unique login identifiers among live rows only (allows reuse after soft-delete).
-- 仅在未软删行上唯一，便于软删后重新注册同名 / 同邮。
CREATE UNIQUE INDEX ux_identity_users_username
  ON identity.users (username)
  WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX ux_identity_users_email
  ON identity.users (email)
  WHERE deleted_at IS NULL AND email IS NOT NULL;

CREATE INDEX idx_identity_users_role_status
  ON identity.users (role, status)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_identity_users_deleted_at
  ON identity.users (deleted_at)
  WHERE deleted_at IS NOT NULL;

COMMENT ON COLUMN identity.users.username IS
  'Unique login name among live rows. / 活跃行内唯一的登录名。';
COMMENT ON COLUMN identity.users.email IS
  'Optional email; unique among live rows when set. / 可选邮箱；有值时在活跃行内唯一。';
COMMENT ON COLUMN identity.users.nickname IS
  'Display name. / 展示昵称。';
COMMENT ON COLUMN identity.users.phone IS
  'Optional phone number. / 可选手机号。';
COMMENT ON COLUMN identity.users.avatar_object_key IS
  'Stable key: local relative path or remote object key (no scheme/host). / 稳定键：本地相对路径或远端对象键（不含协议与域名）。';
COMMENT ON COLUMN identity.users.avatar_storage IS
  'Avatar backend: local | oss | s3; pairs with avatar_object_key. / 头像存储后端，与 avatar_object_key 成对出现。';
COMMENT ON COLUMN identity.users.role IS
  'Authorization role: member | admin. / 授权角色。';
COMMENT ON COLUMN identity.users.status IS
  'Account lifecycle: pending | active | locked. Soft account removal uses deleted_at. / 账户生命周期；销户软删用 deleted_at。';
COMMENT ON COLUMN identity.users.last_login_at IS
  'Last successful login time. / 上次成功登录时间。';
COMMENT ON COLUMN identity.users.created_at IS
  'Row creation timestamp, maintained by GORM CreatedAt. / 行创建时间，由 GORM CreatedAt 维护。';
COMMENT ON COLUMN identity.users.updated_at IS
  'Row last-update timestamp, maintained by GORM UpdatedAt. / 行最近更新时间，由 GORM UpdatedAt 维护。';
COMMENT ON COLUMN identity.users.deleted_at IS
  'Soft-deletion timestamp aligned with gorm.DeletedAt. / 与 gorm.DeletedAt 对齐的软删时间戳。';
