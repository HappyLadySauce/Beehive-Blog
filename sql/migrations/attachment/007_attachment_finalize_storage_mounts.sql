-- Finalize the destructive storage-driver refactor.
-- This migration preserves existing attachment rows where possible, then removes
-- legacy storage_type / bucket / local_path columns from the final schema.
-- 完成破坏式存储驱动重构。
-- 先尽量保留已有附件数据，再从最终 schema 中删除旧 storage_type / bucket / local_path 字段。

-- Keep object_key as the single object locator. Old local rows used local_path.
-- object_key 成为唯一对象定位字段。旧本地行原先使用 local_path。
UPDATE attachment.attachments
SET object_key = local_path
WHERE storage_type = 'local'
  AND object_key IS NULL
  AND local_path IS NOT NULL;

-- Map old storage_type rows to the first active mount for the same driver.
-- 将旧 storage_type 行映射到同驱动的第一个活跃 mount。
UPDATE attachment.attachments a
SET storage_mount_id = (
    SELECT id FROM attachment.storage_mounts
    WHERE driver_name = a.storage_type
      AND deleted_at IS NULL
    ORDER BY is_default DESC, id ASC
    LIMIT 1
)
WHERE a.storage_mount_id IS NULL;

-- Fail the migration instead of silently keeping rows that the new code cannot resolve.
-- 无法被新代码解析的行直接阻止迁移，避免静默保留坏数据。
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM attachment.attachments
        WHERE storage_mount_id IS NULL OR object_key IS NULL OR object_key = ''
    ) THEN
        RAISE EXCEPTION 'attachment storage migration failed: storage_mount_id and object_key are required';
    END IF;
END $$;

DROP INDEX IF EXISTS attachment.idx_attachment_attachments_live_storage_type_created_at;
DROP INDEX IF EXISTS attachment.ux_attachment_attachments_remote_object;
DROP INDEX IF EXISTS attachment.ux_attachment_attachments_local_path;

ALTER TABLE attachment.attachments
    DROP CONSTRAINT IF EXISTS chk_attachment_storage_location,
    DROP CONSTRAINT IF EXISTS chk_attachment_storage_type;

ALTER TABLE attachment.attachments
    ALTER COLUMN storage_mount_id SET NOT NULL,
    ALTER COLUMN object_key SET NOT NULL,
    DROP COLUMN storage_type,
    DROP COLUMN bucket,
    DROP COLUMN local_path;

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

COMMENT ON COLUMN attachment.attachments.storage_mount_id IS
    'Required storage mount used to resolve driver and config. / 必填存储挂载项，用于解析驱动和配置。';
COMMENT ON COLUMN attachment.attachments.object_key IS
    'Object key inside the selected storage mount. / 所选存储挂载项内的对象键。';
