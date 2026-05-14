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

// GetEmailSettings handles GET /api/v1/settings/email.
// GetEmailSettings 处理 GET /api/v1/settings/email。
//
// @Summary      Get application settings (admin)
// @Description  Returns sanitized settings including email SMTP flags without secrets. 中文：返回脱敏应用设置（含 SMTP 开关，不含密码明文）。
// @Tags         settings
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  common.BaseResponse{data=v1.SettingsResponse}
// @Failure      401  {object}  common.BaseResponse
// @Failure      403  {object}  common.BaseResponse
// @Router       /api/v1/settings/email [get]
func (h *SettingsController) GetEmailSettings(ctx *gin.Context) {
	if h.provider == nil {
		common.Fail(ctx, common.NewInternal("settings provider is not configured", errors.New("nil settings provider")))
		return
	}
	s := h.provider.Current()
	rev := h.provider.CachedRevision()
	common.Success(ctx, toResponse(s, rev))
}

func (h *SettingsController) patchEmailSettings(ctx context.Context, req *v1.SettingsPatchRequestJSON) (v1.SettingsResponse, error) {
	if h.provider == nil {
		return v1.SettingsResponse{}, common.NewInternal("settings provider is not configured", errors.New("nil settings provider"))
	}
	if h.store == nil {
		return v1.SettingsResponse{}, common.NewInternal("settings store is not configured", errors.New("nil settings store"))
	}
	if req.Email == nil {
		return v1.SettingsResponse{}, common.NewBadRequest("email field is required for patch", nil)
	}

	patch := &settingtypes.SettingsPatchRequest{Email: patchFromV1(req.Email)}
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

// PatchEmailSettings handles PATCH /api/v1/settings/email (partial merge).
// PatchEmailSettings 处理 PATCH /api/v1/settings/email（部分合并）。
//
// @Summary      Patch application settings (admin)
// @Description  Merges email subtree; omit password to keep existing. Empty string clears password. 中文：合并 email 子树；省略 password 表示不改；传空字符串清空密码。
// @Tags         settings
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      v1.SettingsPatchRequestJSON  true  "Partial settings"
// @Success      200   {object}  common.BaseResponse{data=v1.SettingsResponse}
// @Failure      400   {object}  common.BaseResponse
// @Failure      401   {object}  common.BaseResponse
// @Failure      403   {object}  common.BaseResponse
// @Failure      500   {object}  common.BaseResponse
// @Router       /api/v1/settings/email [patch]
func (h *SettingsController) PatchEmailSettings(ctx *gin.Context) {
	var req v1.SettingsPatchRequestJSON
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

	out, err := h.patchEmailSettings(ctx.Request.Context(), &req)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, out)
}
