CREATE SCHEMA IF NOT EXISTS identity;

-- identity.users: core user rows; avatar points at attachment.attachments (FK).
-- Run after 000_attachment_attachments.sql (basename sort: 000 before 001).
-- identity.users：核心用户；头像外键指向 attachment.attachments。
-- 须在 000_attachment_attachments.sql 之后执行（按文件名前缀 000 < 001）。
CREATE TABLE identity.users (
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

-- Join / preload avatar attachment for live users who set one.
-- 为设置了头像的活跃用户预加载附件行。
CREATE INDEX idx_identity_users_avatar_attachment_id
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
  'FK to attachment.attachments; resolve URL from that row. DB trigger clears this when the attachment row is soft-deleted (deleted_at set). / 外键指向附件表；URL 从该行解析。附件行软删时由库触发器自动清空本列。';
COMMENT ON COLUMN identity.users.role IS
  'Authorization role: member | admin. / 授权角色。';
COMMENT ON COLUMN identity.users.status IS
  'Account lifecycle: pending | active | disabled | locked. Soft account removal uses deleted_at. / 账户生命周期；销户软删用 deleted_at。';
COMMENT ON COLUMN identity.users.last_login_at IS
  'Last successful login time. / 上次成功登录时间。';
COMMENT ON COLUMN identity.users.created_at IS
  'Row creation timestamp, maintained by GORM CreatedAt. / 行创建时间，由 GORM CreatedAt 维护。';
COMMENT ON COLUMN identity.users.updated_at IS
  'Row last-update timestamp, maintained by GORM UpdatedAt. / 行最近更新时间，由 GORM UpdatedAt 维护。';
COMMENT ON COLUMN identity.users.deleted_at IS
  'Soft-deletion timestamp aligned with gorm.DeletedAt. / 与 gorm.DeletedAt 对齐的软删时间戳。';

-- When an attachment becomes soft-deleted, unlink it from any user avatar FK.
-- 附件行一旦软删，自动解除所有用户头像外键引用。
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
  'Clears identity.users.avatar_attachment_id when attachment.attachments is soft-deleted. / 附件软删时清空 identity.users.avatar_attachment_id。';

CREATE TRIGGER trg_attachment_attachments_clear_users_avatar_on_soft_delete
  AFTER UPDATE OF deleted_at ON attachment.attachments
  FOR EACH ROW
  WHEN (OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL)
  EXECUTE PROCEDURE attachment.fn_clear_identity_users_avatar_on_attachment_soft_delete();

COMMENT ON TRIGGER trg_attachment_attachments_clear_users_avatar_on_soft_delete ON attachment.attachments IS
  'Unlink user avatars when this attachment row is soft-deleted. / 本附件软删时解除用户头像引用。';
