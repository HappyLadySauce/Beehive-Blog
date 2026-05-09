CREATE SCHEMA IF NOT EXISTS identity;

-- identity.users: core user rows; avatar points at attachment.attachments (FK).
-- Run after 000_attachment_attachments.sql (basename sort: 000 before 001).
-- identity.users：核心用户；头像外键指向 attachment.attachments。
-- 须在 000_attachment_attachments.sql 之后执行（按文件名前缀 000 < 001）。
-- IF NOT EXISTS allows -force re-apply after checksum changes without 42P07 on existing DBs.
-- 使用 IF NOT EXISTS 便于在 checksum 变更后用 -force 重跑，避免库中已有对象时报 42P07。
CREATE TABLE IF NOT EXISTS identity.users (
  id BIGSERIAL PRIMARY KEY,

  username VARCHAR(64) NOT NULL,
  email VARCHAR(320) NULL,
  nickname VARCHAR(128) NULL,
  phone VARCHAR(16) NULL,

  -- Avatar as registered attachment row (storage details live on attachments).
  -- 头像登记为附件表一行，具体存储信息在 attachments 上。
  avatar_attachment_id BIGINT NULL
    CONSTRAINT fk_identity_users_avatar_attachment
    REFERENCES attachment.attachments (id)
    ON DELETE SET NULL,

  role VARCHAR(16) NOT NULL DEFAULT 'member',
  status VARCHAR(16) NOT NULL DEFAULT 'active',

  last_login_at TIMESTAMPTZ NULL,

  -- GORM-standard timestamps. / GORM 标准时间字段。
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL,

  CONSTRAINT chk_identity_users_role CHECK (role IN ('member', 'admin')),
  CONSTRAINT chk_identity_users_status CHECK (status IN ('pending', 'active', 'disabled', 'locked'))
);

-- Unique login identifiers among live rows only (allows reuse after soft-delete).
-- 仅在未软删行上唯一，便于软删后重新注册同名 / 同邮。
CREATE UNIQUE INDEX IF NOT EXISTS ux_identity_users_username
  ON identity.users (username)
  WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS ux_identity_users_email
  ON identity.users (email)
  WHERE deleted_at IS NULL AND email IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_identity_users_role_status
  ON identity.users (role, status)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_identity_users_deleted_at
  ON identity.users (deleted_at)
  WHERE deleted_at IS NOT NULL;

-- Join / preload avatar attachment for live users who set one.
-- 为设置了头像的活跃用户预加载附件行。
CREATE INDEX IF NOT EXISTS idx_identity_users_avatar_attachment_id
  ON identity.users (avatar_attachment_id)
  WHERE deleted_at IS NULL AND avatar_attachment_id IS NOT NULL;

COMMENT ON COLUMN identity.users.username IS
  'Unique login name among live rows. / 活跃行内唯一的登录名。';
COMMENT ON COLUMN identity.users.email IS
  'Optional email; unique among live rows when set. / 可选邮箱；有值时在活跃行内唯一。';
COMMENT ON COLUMN identity.users.nickname IS
  'Display name. / 展示昵称。';
COMMENT ON COLUMN identity.users.phone IS
  'Optional phone number. / 可选手机号。';
COMMENT ON COLUMN identity.users.avatar_attachment_id IS
  'FK to attachment.attachments; resolve URL from that row. NULL means use the application default avatar rather than missing data; DB trigger clears this when the attachment row is soft-deleted (deleted_at set), causing fallback to the default avatar. / 外键指向附件表；URL 从该行解析。NULL 表示使用应用层默认头像，而不是数据缺失；附件行软删时由库触发器自动清空本列，从而回退到默认头像。';
COMMENT ON COLUMN identity.users.role IS
  'Authorization role: member | admin. / 授权角色。';
COMMENT ON COLUMN identity.users.status IS
  'Account lifecycle: pending | active | disabled | locked. Soft account removal uses deleted_at. / 账户生命周期；销户软删用 deleted_at。';
COMMENT ON COLUMN identity.users.last_login_at IS
  'Last successful login time. / 上次成功登录时间。';
COMMENT ON COLUMN identity.users.created_at IS
  'Row creation timestamp, maintained by GORM CreatedAt. / 行创建时间，由 GORM CreatedAt 维护。';
COMMENT ON COLUMN identity.users.updated_at IS
  'Row last-update timestamp, maintained by GORM UpdatedAt; refreshes when avatar reference changes, including DB-triggered fallback to default avatar after attachment soft-delete. / 行最近更新时间，由 GORM UpdatedAt 维护；头像引用变化时会刷新，包括附件软删后由数据库触发回退默认头像。';
COMMENT ON COLUMN identity.users.deleted_at IS
  'Soft-deletion timestamp aligned with gorm.DeletedAt. / 与 gorm.DeletedAt 对齐的软删时间戳。';

-- Default bootstrap admin for fresh installs (password set in 011_identity_user_credentials.sql).
-- 全新安装时的默认管理员（密码哈希在 011_identity_user_credentials.sql 中写入）。
INSERT INTO identity.users (username, nickname, role, status, created_at, updated_at)
SELECT 'admin', 'Administrator', 'admin', 'active', NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM identity.users u WHERE u.username = 'admin' AND u.deleted_at IS NULL
);
