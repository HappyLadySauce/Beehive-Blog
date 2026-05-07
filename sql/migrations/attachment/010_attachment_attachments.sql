CREATE SCHEMA IF NOT EXISTS attachment;

-- attachment.attachments: unified table for s3 / oss / local attachments,
-- aligned with GORM soft-deletion model (deleted_at is gorm.DeletedAt).
-- attachment.attachments：统一登记 s3 / oss / local 三类附件，
-- 与 GORM 软删模型对齐（deleted_at 对应 gorm.DeletedAt）。
CREATE TABLE attachment.attachments (
  id              BIGSERIAL PRIMARY KEY,

  -- Business fields. / 业务字段。
  filename        VARCHAR(255) NOT NULL,
  original_name   VARCHAR(255),
  mime_type       VARCHAR(127),
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

  -- Business status; soft-deletion is expressed by deleted_at, not status.
  -- 业务状态；软删一律由 deleted_at 表达，不再用 status='deleted'。
  status          VARCHAR(16) NOT NULL DEFAULT 'active',

  -- GORM-standard timestamps. / GORM 标准时间字段。
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at      TIMESTAMPTZ NULL,

  CONSTRAINT chk_attachment_storage_type
    CHECK (storage_type IN ('s3', 'oss', 'local')),
  CONSTRAINT chk_attachment_status
    CHECK (status IN ('active', 'archived')),
  CONSTRAINT chk_attachment_remote_required
    CHECK (
      storage_type = 'local'
      OR (bucket IS NOT NULL AND object_key IS NOT NULL)
    ),
  CONSTRAINT chk_attachment_local_required
    CHECK (
      storage_type <> 'local'
      OR local_path IS NOT NULL
    )
);

-- Index by storage_type, restricted to live rows for cheaper scans.
-- 按 storage_type 查询的索引；仅覆盖活跃行以减小体积。
CREATE INDEX idx_attachment_attachments_storage_type
  ON attachment.attachments (storage_type)
  WHERE deleted_at IS NULL;

-- Lookup by remote object_key (s3 / oss); skip NULL and soft-deleted rows.
-- 通过远端 object_key 定位（s3 / oss）；跳过 NULL 与软删行。
CREATE INDEX idx_attachment_attachments_object_key
  ON attachment.attachments (object_key)
  WHERE deleted_at IS NULL AND object_key IS NOT NULL;

-- Listing by created_at on live rows (typical timeline queries).
-- 活跃行按 created_at 排序的列表查询。
CREATE INDEX idx_attachment_attachments_created_at
  ON attachment.attachments (created_at)
  WHERE deleted_at IS NULL;

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
  'Object key for s3/oss; URL is derived at read time. / 远端对象键，访问 URL 在读取时拼装或签名。';
COMMENT ON COLUMN attachment.attachments.local_path IS
  'Relative path under configured local root. / 配置的本地根目录下的相对路径。';
COMMENT ON COLUMN attachment.attachments.etag IS
  'Provider-returned entity tag, used for cache validation. / 提供方返回的实体标签，用于缓存校验。';
COMMENT ON COLUMN attachment.attachments.checksum IS
  'Content checksum, algorithm fixed in app layer (e.g. sha256). / 内容校验和，算法在应用层固定，例如 sha256。';
COMMENT ON COLUMN attachment.attachments.status IS
  'Business status: active | archived. Soft-deletion uses deleted_at. / 业务状态；软删由 deleted_at 表示。';
COMMENT ON COLUMN attachment.attachments.created_at IS
  'Row creation timestamp, maintained by GORM CreatedAt. / 行创建时间，由 GORM CreatedAt 维护。';
COMMENT ON COLUMN attachment.attachments.updated_at IS
  'Row last-update timestamp, maintained by GORM UpdatedAt. / 行最近更新时间，由 GORM UpdatedAt 维护。';
COMMENT ON COLUMN attachment.attachments.deleted_at IS
  'Soft-deletion timestamp aligned with gorm.DeletedAt. / 与 gorm.DeletedAt 对齐的软删时间戳。';
