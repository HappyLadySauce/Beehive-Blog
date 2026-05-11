-- NOTIFY on application_settings changes for cross-process hot reload (LISTEN setting_revision).
-- 在 application_settings 变更时 NOTIFY，供多进程热加载（LISTEN setting_revision）。
-- Run after 000_setting_application.sql.
-- 在 000_setting_application.sql 之后执行。

CREATE OR REPLACE FUNCTION setting.fn_notify_application_changed()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
  PERFORM pg_notify('setting_revision', NEW.revision::text);
  RETURN NEW;
END;
$$;

COMMENT ON FUNCTION setting.fn_notify_application_changed() IS
  'Emits pg_notify on setting_revision with new revision after INSERT/UPDATE. / 在 INSERT/UPDATE 后向 setting_revision 发送 pg_notify，负载为新 revision。';

DROP TRIGGER IF EXISTS trg_setting_application_notify ON setting.application_settings;

CREATE TRIGGER trg_setting_application_notify
  AFTER INSERT OR UPDATE ON setting.application_settings
  FOR EACH ROW
  EXECUTE PROCEDURE setting.fn_notify_application_changed();

COMMENT ON TRIGGER trg_setting_application_notify ON setting.application_settings IS
  'Pushes setting_revision NOTIFY so app instances can refresh in-memory settings without polling. / 推送 setting_revision NOTIFY，使应用实例无需轮询即可刷新内存设置。';
