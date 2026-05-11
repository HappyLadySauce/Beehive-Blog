package settings

import (
	"testing"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// Ensure model table name is wired for GORM queries.
// 确保 GORM 使用正确的表名。
func TestApplicationSettingTableName(t *testing.T) {
	var m model.ApplicationSetting
	if m.TableName() != "setting.application_settings" {
		t.Fatalf("TableName = %q", m.TableName())
	}
}
