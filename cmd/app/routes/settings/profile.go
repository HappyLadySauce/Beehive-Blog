package settings

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	pkgsettings "github.com/HappyLadySauce/Beehive-Blog/pkg/settings"
	settingtypes "github.com/HappyLadySauce/Beehive-Blog/pkg/settings/types"
)

// GetProfileSettings handles GET /api/v1/settings/profile.
// GetProfileSettings 处理 GET /api/v1/settings/profile。
func (h *SettingsController) GetProfileSettings(ctx *gin.Context) {
	if h.provider == nil {
		common.Fail(ctx, common.NewInternal("settings provider is not configured", errors.New("nil settings provider")))
		return
	}
	common.Success(ctx, toResponse(h.provider.Current(), h.provider.CachedRevision()))
}

func (h *SettingsController) patchProfileSettings(ctx context.Context, req *v1.ProfilePatchJSON) (v1.SettingsResponse, error) {
	if h.provider == nil {
		return v1.SettingsResponse{}, common.NewInternal("settings provider is not configured", errors.New("nil settings provider"))
	}
	if h.store == nil {
		return v1.SettingsResponse{}, common.NewInternal("settings store is not configured", errors.New("nil settings store"))
	}
	patch := &settingtypes.SettingsPatchRequest{Profile: patchProfileFromV1(req)}
	out, rev, err := h.store.Patch(ctx, patch)
	if err != nil {
		if errors.Is(err, pkgsettings.ErrInvalidSettings) {
			return v1.SettingsResponse{}, common.NewBadRequest("invalid settings", err)
		}
		return v1.SettingsResponse{}, common.NewInternal("failed to patch settings", err)
	}
	h.provider.Replace(out, rev)
	return toResponse(out, rev), nil
}

// PatchProfileSettings handles PATCH /api/v1/settings/profile.
// PatchProfileSettings 处理 PATCH /api/v1/settings/profile。
func (h *SettingsController) PatchProfileSettings(ctx *gin.Context) {
	var req v1.ProfilePatchJSON
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	out, err := h.patchProfileSettings(ctx.Request.Context(), &req)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, out)
}
