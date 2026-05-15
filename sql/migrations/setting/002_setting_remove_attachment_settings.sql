-- Remove obsolete file-service settings from the singleton payload.
-- 文件服务已改由 storage_drivers/storage_mounts 管理，清理旧 settings.attachment。
UPDATE setting.application_settings
SET payload = payload - 'attachment',
    revision = revision + 1,
    updated_at = NOW()
WHERE id = 1
  AND payload ? 'attachment';
