package settings

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	pkgsettings "github.com/HappyLadySauce/Beehive-Blog/pkg/settings"
	settingtypes "github.com/HappyLadySauce/Beehive-Blog/pkg/settings/types"
)

// GetAttachmentSettings handles GET /api/v1/settings/attachment.
// GetAttachmentSettings 处理 GET /api/v1/settings/attachment。
//
// @Summary      Get attachment settings (admin)
// @Description  Returns sanitized attachment storage and validation settings.
// @Tags         settings
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  common.BaseResponse{data=v1.SettingsResponse}
// @Failure      401  {object}  common.BaseResponse
// @Failure      403  {object}  common.BaseResponse
// @Router       /api/v1/settings/attachment [get]
func (h *SettingsController) GetAttachmentSettings(ctx *gin.Context) {
	if h.provider == nil {
		common.Fail(ctx, common.NewInternal("settings provider is not configured", errors.New("nil settings provider")))
		return
	}
	s := h.provider.Current()
	rev := h.provider.CachedRevision()
	common.Success(ctx, toResponse(s, rev))
}

func (h *SettingsController) patchAttachmentSettings(ctx context.Context, req *v1.AttachmentPatchJSON) (v1.SettingsResponse, error) {
	if h.provider == nil {
		return v1.SettingsResponse{}, common.NewInternal("settings provider is not configured", errors.New("nil settings provider"))
	}
	if h.store == nil {
		return v1.SettingsResponse{}, common.NewInternal("settings store is not configured", errors.New("nil settings store"))
	}

	patch := &settingtypes.SettingsPatchRequest{Attachment: patchAttachmentFromV1(req)}
	next, rev, err := h.store.Patch(ctx, patch)
	if err != nil {
		if errors.Is(err, pkgsettings.ErrInvalidSettings) {
			return v1.SettingsResponse{}, common.NewBadRequest("invalid settings", err)
		}
		return v1.SettingsResponse{}, common.NewInternal("failed to patch settings", err)
	}
	h.provider.Replace(next, rev)
	return toResponse(next, rev), nil
}

// PatchAttachmentSettings handles PATCH /api/v1/settings/attachment (partial merge).
// PatchAttachmentSettings 处理 PATCH /api/v1/settings/attachment（部分合并）。
//
// @Summary      Patch attachment settings (admin)
// @Description  Merges attachment subtree; omit fields to keep existing.
// @Tags         settings
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      v1.AttachmentPatchJSON  true  "Partial attachment settings"
// @Success      200   {object}  common.BaseResponse{data=v1.SettingsResponse}
// @Failure      400   {object}  common.BaseResponse
// @Failure      401   {object}  common.BaseResponse
// @Failure      403   {object}  common.BaseResponse
// @Failure      500   {object}  common.BaseResponse
// @Router       /api/v1/settings/attachment [patch]
func (h *SettingsController) PatchAttachmentSettings(ctx *gin.Context) {
	var req v1.AttachmentPatchJSON
	dec := json.NewDecoder(ctx.Request.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		common.Fail(ctx, common.NewBadRequest("invalid request body", errors.New("request body must contain a single JSON object")))
		return
	}

	out, err := h.patchAttachmentSettings(ctx.Request.Context(), &req)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, out)
}
