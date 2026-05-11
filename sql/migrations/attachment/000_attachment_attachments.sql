CREATE SCHEMA IF NOT EXISTS attachment;

-- attachment.attachments: unified metadata table for s3 / oss / local attachments.
-- It separates ownership, upload completion, access policy, business lifecycle,
-- and GORM soft-deletion so handlers can enforce each concern explicitly.
-- attachment.attachments：统一登记 s3 / oss / local 三类附件元数据。
-- 表结构将归属、上传完成状态、访问策略、业务生命周期与 GORM 软删拆开，
-- 便于接口层分别执行权限、安全与清理策略。
CREATE TABLE attachment.attachments (
  id              BIGSERIAL PRIMARY KEY,

  -- Business owner. The FK is declared by a later migration because this
  -- migration must run before identity.users; application code must still
  -- validate the owner in request scope.
  -- 业务归属。此处不声明外键，因为本迁移必须先于 identity.users 执行；
  -- 外键由后续迁移追加；应用层仍需在请求上下文中校验归属。
  owner_user_id   BIGINT,

  -- Attachment purpose drives validation policy in the application layer.
  -- 附件用途用于驱动应用层校验策略。
  purpose         VARCHAR(32) NOT NULL DEFAULT 'content',

  -- Business fields. / 业务字段。
  filename        VARCHAR(255) NOT NULL,
  original_name   VARCHAR(255),
  mime_type       VARCHAR(127) NOT NULL,
  size            BIGINT NOT NULL CHECK (size >= 0),

  -- Storage backend selector. / 存储后端选择。
  storage_type    VARCHAR(16) NOT NULL DEFAULT 'local',

  -- Remote (s3 / oss). / 远端字段。
  bucket          VARCHAR(63),
  object_key      VARCHAR(1024),

  -- Local. / 本地字段。
  local_path      VARCHAR(1024),

  -- Optional integrity / cache fields. / 可选完整性与缓存字段。
  etag            VARCHAR(80),
  checksum        VARCHAR(128),

  -- Read access policy. public rows may be served without an owner check
  -- after application-level publication checks; private rows require auth.
  -- 读取访问策略。public 可在应用层发布校验后匿名读取；private 必须鉴权。
  access_scope    VARCHAR(16) NOT NULL DEFAULT 'private',

  -- Upload state is separate from business visibility. Direct-to-object-store
  -- flows create pending rows first; only ready rows can be downloaded or bound
  -- as avatars.
  -- 上传状态与业务可见性分离。对象存储直传会先创建 pending 行；
  -- 只有 ready 行可下载或绑定为头像。
  upload_status   VARCHAR(16) NOT NULL DEFAULT 'ready',

  -- Business visibility / lifecycle (orthogonal to soft-delete):
  --   active   — default; shown in normal listings.
  --   hidden   — not shown in public/default UI; file still retained; row still "live" (deleted_at NULL).
  --   archived — long-term / cold retention; off active workflows; still not deleted until deleted_at set.
  -- Soft-delete (deleted_at): logical removal; GORM omits row by default; may trigger cleanup of storage later.
  -- 业务可见性与生命周期（与软删正交）：
  --   active   — 默认；出现在常规列表。
  --   hidden   — 公共/默认列表不展示；文件仍保留；行仍为「存活」（deleted_at 为空）。
  --   archived — 长期归档/冷数据；退出活跃业务流；在未设置 deleted_at 前不算删除。
  -- 软删（deleted_at）：逻辑删除；GORM 默认查询会排除；可配合后续清理对象存储。
  status          VARCHAR(16) NOT NULL DEFAULT 'active',

  -- GORM-standard timestamps. / GORM 标准时间字段。
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at      TIMESTAMPTZ NULL,

  CONSTRAINT chk_attachment_purpose
    CHECK (purpose IN ('avatar', 'content', 'system', 'other')),
  CONSTRAINT chk_attachment_storage_type
    CHECK (storage_type IN ('s3', 'oss', 'local')),
  CONSTRAINT chk_attachment_access_scope
    CHECK (access_scope IN ('private', 'public')),
  CONSTRAINT chk_attachment_upload_status
    CHECK (upload_status IN ('pending', 'ready', 'failed')),
  CONSTRAINT chk_attachment_status
    CHECK (status IN ('active', 'archived', 'hidden')),
  CONSTRAINT chk_attachment_owner_required
    CHECK (owner_user_id IS NOT NULL OR purpose = 'system'),
  CONSTRAINT chk_attachment_avatar_mime_type
    CHECK (purpose <> 'avatar' OR mime_type LIKE 'image/%'),
  CONSTRAINT chk_attachment_public_requires_ready_upload
    CHECK (access_scope <> 'public' OR upload_status = 'ready'),
  CONSTRAINT chk_attachment_storage_location
    CHECK (
      (
        storage_type = 'local'
        AND local_path IS NOT NULL
        AND bucket IS NULL
        AND object_key IS NULL
      )
      OR
      (
        storage_type IN ('s3', 'oss')
        AND bucket IS NOT NULL
        AND object_key IS NOT NULL
        AND local_path IS NULL
      )
    )
);

COMMENT ON TABLE attachment.attachments IS
  'Attachment metadata and storage registry. Ownership, upload_status, access_scope, status and deleted_at are intentionally separate so authorization, publication, lifecycle and soft-deletion do not overlap. / 附件元数据与存储登记表。owner、upload_status、access_scope、status、deleted_at 被刻意拆分，避免授权、发布、生命周期和软删语义混用。';

-- Listing live attachments by storage_type with stable newest-first pagination.
-- 活跃附件按 storage_type 过滤并按最新优先稳定分页。
CREATE INDEX idx_attachment_attachments_live_storage_type_created_at
  ON attachment.attachments (storage_type, created_at DESC, id DESC)
  WHERE deleted_at IS NULL;

-- Owner-scoped library listing with stable newest-first pagination.
-- 按归属用户查询附件库并按最新优先稳定分页。
CREATE INDEX idx_attachment_attachments_live_owner_created_at
  ON attachment.attachments (owner_user_id, created_at DESC, id DESC)
  WHERE deleted_at IS NULL AND owner_user_id IS NOT NULL;

-- Public ready assets for published content, ordered newest first.
-- 已完成上传且可公开访问的发布资源，按最新优先排序。
CREATE INDEX idx_attachment_attachments_ready_public_created_at
  ON attachment.attachments (created_at DESC, id DESC)
  WHERE deleted_at IS NULL
    AND upload_status = 'ready'
    AND access_scope = 'public'
    AND status = 'active';

-- Pending / failed uploads are uncommon but need cleanup and retry scans.
-- pending / failed 上传较少，但需要清理与重试扫描。
CREATE INDEX idx_attachment_attachments_upload_status
  ON attachment.attachments (upload_status, created_at)
  WHERE deleted_at IS NULL AND upload_status <> 'ready';

-- Audit / cleanup queries on soft-deleted rows.
-- 用于审计或清理软删行的索引。
CREATE INDEX idx_attachment_attachments_deleted_at
  ON attachment.attachments (deleted_at)
  WHERE deleted_at IS NOT NULL;

-- Per-bucket uniqueness for remote objects among live rows.
-- 远端对象在活跃行内按 bucket 维度去重。
CREATE UNIQUE INDEX ux_attachment_attachments_remote_object
  ON attachment.attachments (storage_type, bucket, object_key)
  WHERE deleted_at IS NULL
    AND storage_type IN ('s3', 'oss')
    AND object_key IS NOT NULL;

-- Per-path uniqueness for local objects among live rows.
-- 本地对象在活跃行内按 local_path 去重。
CREATE UNIQUE INDEX ux_attachment_attachments_local_path
  ON attachment.attachments (storage_type, local_path)
  WHERE deleted_at IS NULL
    AND storage_type = 'local'
    AND local_path IS NOT NULL;

COMMENT ON COLUMN attachment.attachments.owner_user_id IS
  'User id that owns this attachment. The FK is added after identity.users is created; handlers must still validate ownership before writes and reads. / 拥有该附件的用户 id。外键在 identity.users 创建后追加；接口层在读写前仍必须校验归属。';
COMMENT ON COLUMN attachment.attachments.purpose IS
  'Attachment purpose: avatar | content | system | other. Purpose selects validation policy; avatar rows must be image MIME types. / 附件用途：avatar | content | system | other。用途决定校验策略；avatar 必须是图片 MIME 类型。';
COMMENT ON COLUMN attachment.attachments.filename IS
  'Server-side safe filename used for storage. / 用于落盘 / 上传的安全文件名。';
COMMENT ON COLUMN attachment.attachments.original_name IS
  'Original filename uploaded by the user, for display only. / 用户上传时的原始文件名，仅用于展示。';
COMMENT ON COLUMN attachment.attachments.mime_type IS
  'IANA media type of the content. / 内容的 IANA 媒体类型。';
COMMENT ON COLUMN attachment.attachments.size IS
  'Content size in bytes; must be non-negative. / 内容字节数；必须非负。';
COMMENT ON COLUMN attachment.attachments.storage_type IS
  'Storage backend: s3 | oss | local. / 存储后端类型。';
COMMENT ON COLUMN attachment.attachments.bucket IS
  'Bucket name for s3/oss; NULL for local. / 远端桶名，本地为空。';
COMMENT ON COLUMN attachment.attachments.object_key IS
  'Object key for s3/oss; URL is derived at read time. Remote attachment lookup should use storage_type + bucket + object_key together, not object_key alone. / 远端对象键，访问 URL 在读取时拼装或签名。远端附件定位应联合使用 storage_type + bucket + object_key，而不是仅用 object_key。';
COMMENT ON COLUMN attachment.attachments.local_path IS
  'Relative path under configured local root. / 配置的本地根目录下的相对路径。';
COMMENT ON COLUMN attachment.attachments.etag IS
  'Provider-returned entity tag, used for cache validation. / 提供方返回的实体标签，用于缓存校验。';
COMMENT ON COLUMN attachment.attachments.checksum IS
  'Content checksum, algorithm fixed in app layer (e.g. sha256). / 内容校验和，算法在应用层固定，例如 sha256。';
COMMENT ON COLUMN attachment.attachments.access_scope IS
  'Read access policy: private | public. public rows can be served anonymously only after upload_status=ready and application publication checks pass. / 读取访问策略：private | public。public 行仅在 upload_status=ready 且应用层发布校验通过后才可匿名读取。';
COMMENT ON COLUMN attachment.attachments.upload_status IS
  'Upload completion state: pending | ready | failed. Only ready rows may be downloaded or bound as user avatars. / 上传完成状态：pending | ready | failed。只有 ready 行可下载或绑定为用户头像。';
COMMENT ON COLUMN attachment.attachments.status IS
  'Visibility/lifecycle: active | hidden | archived. hidden hides from default UI without deleting the row and hidden remains referenceable, including by user avatars; archived marks cold retention; only soft-deletion via deleted_at makes the attachment unusable as an avatar. / 可见性与生命周期：active | hidden | archived。hidden 为默认列表不可见但未软删，且 hidden 仍可被引用，包括被用户头像引用；archived 为归档；只有 deleted_at 软删才会使附件不可继续作为头像。';
COMMENT ON COLUMN attachment.attachments.created_at IS
  'Row creation timestamp, maintained by GORM CreatedAt. / 行创建时间，由 GORM CreatedAt 维护。';
COMMENT ON COLUMN attachment.attachments.updated_at IS
  'Row last-update timestamp, maintained by GORM UpdatedAt. / 行最近更新时间，由 GORM UpdatedAt 维护。';
COMMENT ON COLUMN attachment.attachments.deleted_at IS
  'Soft-deletion timestamp aligned with gorm.DeletedAt. / 与 gorm.DeletedAt 对齐的软删时间戳。';

-- When an attachment becomes soft-deleted, unlink it from any user avatar FK
-- so those users fall back to the application default avatar.
-- 附件行一旦软删，自动解除所有用户头像外键引用，
-- 使这些用户回退到应用层默认头像。
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

CREATE TRIGGER trg_attachment_attachments_clear_users_avatar_on_soft_delete
  AFTER UPDATE OF deleted_at ON attachment.attachments
  FOR EACH ROW
  WHEN (OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL)
  EXECUTE PROCEDURE attachment.fn_clear_identity_users_avatar_on_attachment_soft_delete();

COMMENT ON TRIGGER trg_attachment_attachments_clear_users_avatar_on_soft_delete ON attachment.attachments IS
  'On soft-delete only (deleted_at NULL -> non-NULL), unlink user avatars from this attachment, refresh affected identity.users.updated_at values, and make those users fall back to the application default avatar. / 仅在软删时（deleted_at 从 NULL 变为非 NULL）解除用户头像对本附件的引用，刷新受影响 identity.users.updated_at，并使这些用户回退到应用层默认头像。';
