package filedrivers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/attachment/driver"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// ListDrivers returns all available storage driver templates.
// ListDrivers 返回所有可用的存储驱动模板。
func (h *FileDriversController) ListDrivers(ctx *gin.Context) {
	rows, err := h.driverStore.ListDrivers(ctx.Request.Context())
	if err != nil {
		writeFileDriverError(ctx, err)
		return
	}
	items := make([]v1.DriverResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDriverResponse(row))
	}
	common.Response(ctx, nil, v1.DriverListResponse{Items: items})
}

// GetDriver returns a single storage driver template by name.
// GetDriver 按名称返回单个存储驱动模板。
func (h *FileDriversController) GetDriver(ctx *gin.Context) {
	name := ctx.Param("name")
	row, err := h.driverStore.GetDriver(ctx.Request.Context(), name)
	if err != nil {
		writeFileDriverError(ctx, err)
		return
	}
	common.Response(ctx, nil, toDriverResponse(*row))
}

func toDriverResponse(row model.StorageDriver) v1.DriverResponse {
	return v1.DriverResponse{
		ID:           row.ID,
		Name:         row.Name,
		DisplayName:  row.DisplayName,
		Description:  row.Description,
		ConfigSchema: row.ConfigSchema,
		Capabilities: row.Capabilities,
		Status:       row.Status,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func toMountResponse(row model.StorageMount) v1.StorageMountResponse {
	return v1.StorageMountResponse{
		ID:            row.ID,
		DriverName:    row.DriverName,
		MountPath:     row.MountPath,
		Name:          row.Name,
		Remark:        row.Remark,
		Config:        row.Config,
		OrderIndex:    row.OrderIndex,
		IsDefault:     row.IsDefault,
		Disabled:      row.Disabled,
		Status:        row.Status,
		LastCheckedAt: row.LastCheckedAt,
		LastError:     row.LastError,
		CreatedBy:     row.CreatedBy,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

func writeFileDriverError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		common.Fail(ctx, common.NewNotFound("resource not found", err))
	case errors.Is(err, driver.ErrUnsupportedDriver):
		common.Fail(ctx, common.NewBadRequest("unsupported driver", err))
	default:
		common.Fail(ctx, common.NewInternal("internal error", err))
	}
}
