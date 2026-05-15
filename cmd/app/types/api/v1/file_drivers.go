package v1

import (
	"encoding/json"
	"time"
)

// DriverResponse is the API representation of a storage driver template.
// DriverResponse 是存储驱动模板的 API 表示。
type DriverResponse struct {
	ID           int64           `json:"id"`
	Name         string          `json:"name"`
	DisplayName  string          `json:"display_name"`
	Description  *string         `json:"description,omitempty"`
	ConfigSchema json.RawMessage `json:"config_schema"`
	Capabilities json.RawMessage `json:"capabilities"`
	Status       string          `json:"status"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// DriverListResponse is the API response for listing storage drivers.
// DriverListResponse 是列出存储驱动的 API 响应。
type DriverListResponse struct {
	Items []DriverResponse `json:"items"`
}

// StorageMountResponse is the API representation of a storage mount.
// StorageMountResponse 是存储挂载项的 API 表示。
type StorageMountResponse struct {
	ID            int64           `json:"id"`
	DriverName    string          `json:"driver_name"`
	MountPath     string          `json:"mount_path"`
	Name          string          `json:"name"`
	Remark        *string         `json:"remark,omitempty"`
	Config        json.RawMessage `json:"config"`
	OrderIndex    int             `json:"order_index"`
	IsDefault     bool            `json:"is_default"`
	Disabled      bool            `json:"disabled"`
	Status        string          `json:"status"`
	LastCheckedAt *time.Time      `json:"last_checked_at,omitempty"`
	LastError     *string         `json:"last_error,omitempty"`
	CreatedBy     *int64          `json:"created_by,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// StorageMountListResponse is the API response for listing storage mounts.
// StorageMountListResponse 是列出存储挂载项的 API 响应。
type StorageMountListResponse struct {
	Items []StorageMountResponse `json:"items"`
}

// StorageMountCreateRequest is the request body for creating a storage mount.
// StorageMountCreateRequest 是创建存储挂载项的请求体。
type StorageMountCreateRequest struct {
	DriverName string          `json:"driver_name" binding:"required"`
	MountPath  string          `json:"mount_path" binding:"required,max=512"`
	Name       string          `json:"name" binding:"required,max=128"`
	Remark     *string         `json:"remark,omitempty"`
	Config     json.RawMessage `json:"config" binding:"required"`
	OrderIndex int             `json:"order_index"`
	IsDefault  bool            `json:"is_default"`
}

// StorageMountPatchRequest is the request body for patching a storage mount.
// StorageMountPatchRequest 是更新存储挂载项的请求体。
type StorageMountPatchRequest struct {
	Name       *string          `json:"name,omitempty"`
	Remark     *string          `json:"remark,omitempty"`
	Config     *json.RawMessage `json:"config,omitempty"`
	OrderIndex *int             `json:"order_index,omitempty"`
	IsDefault  *bool            `json:"is_default,omitempty"`
}

// StorageMountCheckResponse is the response for a health check.
// StorageMountCheckResponse 是健康检查的响应。
type StorageMountCheckResponse struct {
	Status  string  `json:"status"`
	Error   *string `json:"error,omitempty"`
	Checked string  `json:"checked"`
}
