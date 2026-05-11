-- Add the owner FK after identity.users exists. Migration files are sorted by
-- basename first, so 014 runs after identity/013_identity_user_identities.sql.
-- 在 identity.users 创建完成后追加归属外键。迁移文件先按文件名排序，
-- 因此 014 会在 identity/013_identity_user_identities.sql 之后执行。
ALTER TABLE attachment.attachments
  ADD CONSTRAINT fk_attachment_attachments_owner_user
  FOREIGN KEY (owner_user_id)
  REFERENCES identity.users (id)
  ON DELETE RESTRICT;

COMMENT ON CONSTRAINT fk_attachment_attachments_owner_user ON attachment.attachments IS
  'Owner user FK for non-system attachments. Hard-deleting a user with attachments is restricted; account removal should use identity.users.deleted_at. / 非 system 附件的归属用户外键。拥有附件的用户不允许物理删除；账号移除应使用 identity.users.deleted_at 软删。';
