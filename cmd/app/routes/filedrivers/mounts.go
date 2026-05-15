package filedrivers

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

var mountPathPattern = regexp.MustCompile(`^/[A-Za-z0-9][A-Za-z0-9._/-]*$`)

// ListMounts returns all storage mounts.
// ListMounts 返回所有存储挂载项。
func (h *FileDriversController) ListMounts(ctx *gin.Context) {
	rows, err := h.driverStore.ListMounts(ctx.Request.Context())
	if err != nil {
		writeFileDriverError(ctx, err)
		return
	}
	items := make([]v1.StorageMountResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toMountResponse(row))
	}
	common.Response(ctx, nil, v1.StorageMountListResponse{Items: items})
}

// GetMount returns a single storage mount.
// GetMount 返回单个存储挂载项。
func (h *FileDriversController) GetMount(ctx *gin.Context) {
	id, err := parseIDParam(ctx.Param("id"))
	if err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid mount id", err))
		return
	}
	row, err := h.driverStore.GetMountByID(ctx.Request.Context(), id)
	if err != nil {
		writeFileDriverError(ctx, err)
		return
	}
	common.Response(ctx, nil, toMountResponse(*row))
}

// CreateMount creates a new storage mount.
// CreateMount 创建新的存储挂载项。
func (h *FileDriversController) CreateMount(ctx *gin.Context) {
	var req v1.StorageMountCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid mount request", err))
		return
	}
	req.DriverName = strings.TrimSpace(req.DriverName)
	req.MountPath = strings.TrimRight(strings.TrimSpace(req.MountPath), "/")
	req.Name = strings.TrimSpace(req.Name)
	if err := h.validateMountConfig(ctx.Request.Context(), req.DriverName, req.MountPath, req.Name, req.Config); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid mount request", err))
		return
	}

	mount := &model.StorageMount{
		DriverName: req.DriverName,
		MountPath:  req.MountPath,
		Name:       req.Name,
		Remark:     req.Remark,
		Config:     req.Config,
		OrderIndex: req.OrderIndex,
		IsDefault:  req.IsDefault,
		Disabled:   false,
		Status:     "unknown",
		CreatedBy:  nil,
	}
	if err := h.db.WithContext(ctx.Request.Context()).Transaction(func(tx *gorm.DB) error {
		if mount.IsDefault {
			if err := tx.Model(&model.StorageMount{}).Where("deleted_at IS NULL").Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Create(mount).Error
	}); err != nil {
		writeFileDriverError(ctx, err)
		return
	}
	common.Response(ctx, nil, toMountResponse(*mount))
}

// PatchMount updates fields on an existing storage mount.
// PatchMount 更新已有存储挂载项的字段。
func (h *FileDriversController) PatchMount(ctx *gin.Context) {
	id, err := parseIDParam(ctx.Param("id"))
	if err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid mount id", err))
		return
	}
	var req v1.StorageMountPatchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid patch request", err))
		return
	}
	current, err := h.driverStore.GetMountByID(ctx.Request.Context(), id)
	if err != nil {
		writeFileDriverError(ctx, err)
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		updates["name"] = name
	}
	if req.Remark != nil {
		updates["remark"] = *req.Remark
	}
	if req.Config != nil {
		updates["config"] = *req.Config
	}
	if req.OrderIndex != nil {
		updates["order_index"] = *req.OrderIndex
	}
	if req.IsDefault != nil {
		updates["is_default"] = *req.IsDefault
	}
	if len(updates) == 0 {
		common.Response(ctx, nil, toMountResponse(*current))
		return
	}
	config := current.Config
	if req.Config != nil {
		config = *req.Config
	}
	name := current.Name
	if req.Name != nil {
		name = strings.TrimSpace(*req.Name)
	}
	if err := h.validateMountConfig(ctx.Request.Context(), current.DriverName, current.MountPath, name, config); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid patch request", err))
		return
	}

	if err := h.db.WithContext(ctx.Request.Context()).Transaction(func(tx *gorm.DB) error {
		if req.IsDefault != nil && *req.IsDefault {
			if err := tx.Model(&model.StorageMount{}).Where("deleted_at IS NULL AND id <> ?", id).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		res := tx.Model(&model.StorageMount{}).Where("id = ?", id).Updates(updates)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	}); err != nil {
		writeFileDriverError(ctx, err)
		return
	}
	row, err := h.driverStore.GetMountByID(ctx.Request.Context(), id)
	if err != nil {
		writeFileDriverError(ctx, err)
		return
	}
	common.Response(ctx, nil, toMountResponse(*row))
}

// EnableMount enables a disabled mount.
// EnableMount 启用已被禁用的挂载项。
func (h *FileDriversController) EnableMount(ctx *gin.Context) {
	h.setMountDisabled(ctx, false)
}

// DisableMount disables a mount.
// DisableMount 禁用挂载项。
func (h *FileDriversController) DisableMount(ctx *gin.Context) {
	h.setMountDisabled(ctx, true)
}

func (h *FileDriversController) setMountDisabled(ctx *gin.Context, disabled bool) {
	id, err := parseIDParam(ctx.Param("id"))
	if err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid mount id", err))
		return
	}
	updates := map[string]interface{}{"disabled": disabled}
	if err := h.driverStore.UpdateMount(ctx.Request.Context(), id, updates); err != nil {
		writeFileDriverError(ctx, err)
		return
	}
	row, err := h.driverStore.GetMountByID(ctx.Request.Context(), id)
	if err != nil {
		writeFileDriverError(ctx, err)
		return
	}
	common.Response(ctx, nil, toMountResponse(*row))
}

// CheckMount runs a health check on the mount's driver.
// CheckMount 对挂载项对应的驱动执行健康检查。
func (h *FileDriversController) CheckMount(ctx *gin.Context) {
	id, err := parseIDParam(ctx.Param("id"))
	if err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid mount id", err))
		return
	}
	mount, err := h.driverStore.GetMountByID(ctx.Request.Context(), id)
	if err != nil {
		writeFileDriverError(ctx, err)
		return
	}

	backend, beErr := h.driverRegistry.CreateBackend(mount.DriverName, mount.Config)
	checked := time.Now().UTC().Format(time.RFC3339)

	updates := map[string]interface{}{
		"last_checked_at": time.Now().UTC(),
	}

	if beErr != nil {
		errMsg := beErr.Error()
		updates["status"] = "error"
		updates["last_error"] = errMsg
		if err := h.driverStore.UpdateMount(ctx.Request.Context(), id, updates); err != nil {
			writeFileDriverError(ctx, err)
			return
		}
		common.Response(ctx, nil, v1.StorageMountCheckResponse{
			Status:  "error",
			Error:   &errMsg,
			Checked: checked,
		})
		return
	}

	if hcErr := backend.HealthCheck(ctx.Request.Context()); hcErr != nil {
		errMsg := hcErr.Error()
		updates["status"] = "error"
		updates["last_error"] = errMsg
		if err := h.driverStore.UpdateMount(ctx.Request.Context(), id, updates); err != nil {
			writeFileDriverError(ctx, err)
			return
		}
		common.Response(ctx, nil, v1.StorageMountCheckResponse{
			Status:  "error",
			Error:   &errMsg,
			Checked: checked,
		})
		return
	}

	updates["status"] = "work"
	updates["last_error"] = nil
	if err := h.driverStore.UpdateMount(ctx.Request.Context(), id, updates); err != nil {
		writeFileDriverError(ctx, err)
		return
	}
	common.Response(ctx, nil, v1.StorageMountCheckResponse{
		Status:  "work",
		Checked: checked,
	})
}

func (h *FileDriversController) validateMountConfig(ctx context.Context, driverName, mountPath, name string, config json.RawMessage) error {
	if driverName == "" {
		return fmt.Errorf("driver_name is required")
	}
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if mountPath == "" || mountPath == "/" || !mountPathPattern.MatchString(mountPath) || strings.Contains(mountPath, "//") ||
		strings.Contains(mountPath, "/../") || strings.Contains(mountPath, "/./") || strings.HasSuffix(mountPath, "/..") || strings.HasSuffix(mountPath, "/.") {
		return fmt.Errorf("mount_path is invalid")
	}
	template, err := h.driverStore.GetDriver(ctx, driverName)
	if err != nil {
		return fmt.Errorf("driver template not found: %w", err)
	}
	if template.Status != "active" {
		return fmt.Errorf("driver template is disabled")
	}
	backend, err := h.driverRegistry.CreateBackend(driverName, config)
	if err != nil {
		return fmt.Errorf("driver config is invalid: %w", err)
	}
	if err := backend.HealthCheck(ctx); err != nil {
		return fmt.Errorf("driver config is invalid: %w", err)
	}
	return nil
}

// DeleteMount soft-deletes a storage mount.
// DeleteMount 软删除存储挂载项。
func (h *FileDriversController) DeleteMount(ctx *gin.Context) {
	id, err := parseIDParam(ctx.Param("id"))
	if err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid mount id", err))
		return
	}

	count, err := h.driverStore.CountAttachmentsOnMount(ctx.Request.Context(), id)
	if err != nil {
		writeFileDriverError(ctx, err)
		return
	}
	if count > 0 {
		common.Fail(ctx, common.NewConflict(
			fmt.Sprintf("mount has %d attachments; reassign or delete them first", count),
			fmt.Errorf("mount %d has %d attachments", id, count),
		))
		return
	}

	if err := h.driverStore.SoftDeleteMount(ctx.Request.Context(), id); err != nil {
		writeFileDriverError(ctx, err)
		return
	}
	common.Response(ctx, nil, nil)
}
