-- 删除已废弃的 SMTP 设置键（规范键为 smtp.host 等；本脚本不写入新数据）
DELETE FROM settings
WHERE "group" = 'smtp'
  AND key IN (
    'smtp_host',
    'smtp_port',
    'smtp_encryption',
    'smtp_username',
    'smtp_password',
    'smtp_from_email',
    'smtp_from_name',
    'smtp_enabled',
    'smtp_user',
    'smtp_pass'
  );
