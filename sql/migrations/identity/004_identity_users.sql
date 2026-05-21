CREATE SCHEMA IF NOT EXISTS identity;

-- identity.users: core user rows; avatar points at attachment.attachments (FK).
-- Run after attachment migrations so attachment.attachments exists for the avatar FK.
-- identity.users：核心用户；头像外键指向 attachment.attachments。
-- 须在 attachment 迁移之后执行，确保 attachment.attachments 存在。
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

-- Default bootstrap admin for fresh installs (password set in 005_identity_user_credentials.sql).
-- 全新安装时的默认管理员（密码哈希在 005_identity_user_credentials.sql 中写入）。
INSERT INTO identity.users (username, nickname, role, status, created_at, updated_at)
SELECT 'admin', 'Administrator', 'admin', 'active', NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM identity.users u WHERE u.username = 'admin' AND u.deleted_at IS NULL
);

-- =========================================================================
-- Add the owner FK from attachment.attachments to identity.users.
-- This was deferred until identity.users existed.
-- 追加 attachment.attachments → identity.users 的归属外键。
-- =========================================================================
ALTER TABLE attachment.attachments
  ADD CONSTRAINT IF NOT EXISTS fk_attachment_attachments_owner_user
  FOREIGN KEY (owner_user_id)
  REFERENCES identity.users (id)
  ON DELETE RESTRICT;

COMMENT ON CONSTRAINT fk_attachment_attachments_owner_user ON attachment.attachments IS
  'Owner user FK for non-system attachments. Hard-deleting a user with attachments is restricted; account removal should use identity.users.deleted_at. / 非 system 附件的归属用户外键。拥有附件的用户不允许物理删除；账号移除应使用 identity.users.deleted_at 软删。';

-- =========================================================================
-- When an attachment becomes soft-deleted, unlink it from any user avatar FK
-- so those users fall back to the application default avatar.
-- 附件行一旦软删，自动解除所有用户头像外键引用，
-- 使这些用户回退到应用层默认头像。
-- =========================================================================
CREATE OR REPLACE FUNCTION attachment.fn_clear_identity_users_avatar_on_attachment_soft_delete()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
  UPDATE identity.users
  SET avatar_attachment_id = NULL,
      updated_at = NOW()
  WHERE avatar_attachment_id = NEW.id;
  RETURN NEW;
END;
$$;

COMMENT ON FUNCTION attachment.fn_clear_identity_users_avatar_on_attachment_soft_delete() IS
  'When attachment.attachments.deleted_at changes from NULL to non-NULL, clears identity.users.avatar_attachment_id and refreshes identity.users.updated_at so affected users fall back to the application default avatar. / 当 attachment.attachments.deleted_at 从 NULL 变为非 NULL 时，清空 identity.users.avatar_attachment_id 并刷新 identity.users.updated_at，使受影响用户回退到应用层默认头像。';

DROP TRIGGER IF EXISTS trg_attachment_attachments_clear_users_avatar_on_soft_delete ON attachment.attachments;

CREATE TRIGGER trg_attachment_attachments_clear_users_avatar_on_soft_delete
  AFTER UPDATE OF deleted_at ON attachment.attachments
  FOR EACH ROW
  WHEN (OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL)
  EXECUTE PROCEDURE attachment.fn_clear_identity_users_avatar_on_attachment_soft_delete();

COMMENT ON TRIGGER trg_attachment_attachments_clear_users_avatar_on_soft_delete ON attachment.attachments IS
  'On soft-delete only (deleted_at NULL -> non-NULL), unlink user avatars from this attachment, refresh affected identity.users.updated_at values, and make those users fall back to the application default avatar. / 仅在软删时（deleted_at 从 NULL 变为非 NULL）解除用户头像对本附件的引用，刷新受影响 identity.users.updated_at，并使这些用户回退到应用层默认头像。';
