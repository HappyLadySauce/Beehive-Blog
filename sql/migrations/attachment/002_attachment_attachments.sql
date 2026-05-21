-- attachment.attachments: unified metadata table for storage attachments.
-- Separates ownership, upload completion, access policy, business lifecycle,
-- and GORM soft-deletion so handlers can enforce each concern explicitly.
-- Storage location is resolved via storage_mount_id → storage_mounts at
-- upload/download time; object_key is the sole object locator within a mount.
-- attachment.attachments：统一登记附件元数据。
-- 将归属、上传完成状态、访问策略、业务生命周期与 GORM 软删拆开，
-- 便于接口层分别执行权限、安全与清理策略。
-- 存储位置通过 storage_mount_id → storage_mounts 解析；object_key 是挂载项内的唯一对象定位符。
CREATE TABLE attachment.attachments (
  id              BIGSERIAL PRIMARY KEY,

  -- Business owner. FK is added in identity/004 after identity.users exists.
  -- 业务归属。外键在 identity/004 中追加（identity.users 创建之后）。
  owner_user_id   BIGINT,

  -- Attachment purpose drives validation policy in the application layer.
  -- 附件用途用于驱动应用层校验策略。
  purpose         VARCHAR(32) NOT NULL DEFAULT 'content',

  -- Business fields. / 业务字段。
  filename        VARCHAR(255) NOT NULL,
  original_name   VARCHAR(255),
  mime_type       VARCHAR(127) NOT NULL,
  size            BIGINT NOT NULL CHECK (size >= 0),

  -- Storage mount and object key resolve the driver + config at runtime.
  -- 通过 storage_mount_id + object_key 在运行时解析驱动与配置。
  storage_mount_id BIGINT NOT NULL
    CONSTRAINT fk_attachment_storage_mount
    REFERENCES attachment.storage_mounts (id),
  object_key      VARCHAR(1024) NOT NULL,

  -- Provider-specific metadata such as version id, headers, etag details.
  -- 提供方扩展元数据，如 version id、headers、etag 详情。
  storage_metadata JSONB NOT NULL DEFAULT '{}',

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
    CHECK (access_scope <> 'public' OR upload_status = 'ready')
);

COMMENT ON TABLE attachment.attachments IS
  'Attachment metadata and storage registry. Ownership, upload_status, access_scope, status and deleted_at are intentionally separate so authorization, publication, lifecycle and soft-deletion do not overlap. / 附件元数据与存储登记表。owner、upload_status、access_scope、status、deleted_at 被刻意拆分，避免授权、发布、生命周期和软删语义混用。';

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

-- Object keys are unique within an active mount.
-- 活跃行内，同一 mount 下 object_key 唯一。
CREATE UNIQUE INDEX ux_attachment_attachments_mount_object_key
  ON attachment.attachments (storage_mount_id, object_key)
  WHERE deleted_at IS NULL;

-- Listing live attachments by mount with stable newest-first pagination.
-- 活跃附件按 mount 过滤并按最新优先稳定分页。
CREATE INDEX idx_attachment_attachments_live_mount_created_at
  ON attachment.attachments (storage_mount_id, created_at DESC, id DESC)
  WHERE deleted_at IS NULL;

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
COMMENT ON COLUMN attachment.attachments.storage_mount_id IS
  'Required storage mount used to resolve driver and config. / 必填存储挂载项，用于解析驱动和配置。';
COMMENT ON COLUMN attachment.attachments.object_key IS
  'Object key inside the selected storage mount. / 所选存储挂载项内的对象键。';
COMMENT ON COLUMN attachment.attachments.storage_metadata IS
  'Provider-specific metadata such as version id, headers, etag details. / 提供方扩展元数据，如 version id、headers、etag 详情。';
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

-- Soft-delete trigger that clears identity.users.avatar_attachment_id is
-- created in identity/004 after identity.users exists.
-- 软删时清除 identity.users.avatar_attachment_id 的触发器在 identity/004 中创建。
