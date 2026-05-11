package settings

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	pkgsettings "github.com/HappyLadySauce/Beehive-Blog/pkg/settings"
	settingtypes "github.com/HappyLadySauce/Beehive-Blog/pkg/settings/types"
)

// toResponse maps internal settings to a sanitized API response.
// toResponse 将内部设置映射为脱敏 API 响应。
func toResponse(s settingtypes.ApplicationSettings, revision int64) v1.SettingsResponse {
	e := s.Email
	pwdSet := strings.TrimSpace(e.Password) != ""
	return v1.SettingsResponse{
		Revision: revision,
		Email: v1.EmailSettingsPublic{
			Enabled:     e.Enabled,
			Host:        e.Host,
			Port:        e.Port,
			Username:    e.Username,
			PasswordSet: pwdSet,
			From:        e.From,
			FromName:    e.FromName,
			TLS:         e.TLS,
		},
	}
}

func patchFromV1(p *v1.EmailSMTPPatchJSON) *settingtypes.EmailSMTPPatch {
	if p == nil {
		return nil
	}
	return &settingtypes.EmailSMTPPatch{
		Enabled:  p.Enabled,
		Host:     p.Host,
		Port:     p.Port,
		Username: p.Username,
		Password: p.Password,
		From:     p.From,
		FromName: p.FromName,
		TLS:      p.TLS,
	}
}

// ServeGet handles GET /api/v1/settings.
// ServeGet 处理 GET /api/v1/settings。
//
// @Summary      Get application settings (admin)
// @Description  Returns sanitized settings including email SMTP flags without secrets. 中文：返回脱敏应用设置（含 SMTP 开关，不含密码明文）。
// @Tags         settings
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  common.BaseResponse{data=v1.SettingsResponse}
// @Failure      401  {object}  common.BaseResponse
// @Failure      403  {object}  common.BaseResponse
// @Router       /api/v1/settings [get]
func (h *SettingsController) ServeGet(ctx *gin.Context) {
	if h.svc.Settings == nil {
		common.Fail(ctx, common.NewInternal("settings provider is not configured", errors.New("nil settings provider")))
		return
	}
	s := h.svc.Settings.Current()
	rev := h.svc.Settings.CachedRevision()
	common.Success(ctx, toResponse(s, rev))
}

// ServePatch handles PATCH /api/v1/settings (partial merge).
// ServePatch 处理 PATCH /api/v1/settings（部分合并）。
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
// @Router       /api/v1/settings [patch]
func (h *SettingsController) ServePatch(ctx *gin.Context) {
	if h.svc.Settings == nil {
		common.Fail(ctx, common.NewInternal("settings provider is not configured", errors.New("nil settings provider")))
		return
	}
	dec := json.NewDecoder(ctx.Request.Body)
	dec.DisallowUnknownFields()
	var req v1.SettingsPatchRequestJSON
	if err := dec.Decode(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	if req.Email == nil {
		common.Fail(ctx, common.NewBadRequest("email field is required for patch", nil))
		return
	}

	patch := &settingtypes.SettingsPatchRequest{Email: patchFromV1(req.Email)}
	if err := h.svc.Settings.PatchAndRefresh(ctx.Request.Context(), patch); err != nil {
		if errors.Is(err, pkgsettings.ErrInvalidSettings) {
			common.Fail(ctx, common.NewBadRequest("invalid settings", err))
			return
		}
		common.Fail(ctx, common.NewInternal("failed to patch settings", err))
		return
	}
	out := toResponse(h.svc.Settings.Current(), h.svc.Settings.CachedRevision())
	common.Success(ctx, out)
}
