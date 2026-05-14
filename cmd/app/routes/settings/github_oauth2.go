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

// GetGithubOAuth2Settings handles GET /api/v1/settings/github-oauth2.
// GetGithubOAuth2Settings 处理 GET /api/v1/settings/github-oauth2。
//
// @Summary      Get GitHub OAuth2 settings (admin)
// @Description  Returns sanitized GitHub OAuth2 settings without client secret.
// @Tags         settings
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  common.BaseResponse{data=v1.SettingsResponse}
// @Failure      401  {object}  common.BaseResponse
// @Failure      403  {object}  common.BaseResponse
// @Router       /api/v1/settings/github-oauth2 [get]
func (h *SettingsController) GetGithubOAuth2Settings(ctx *gin.Context) {
	if h.provider == nil {
		common.Fail(ctx, common.NewInternal("settings provider is not configured", errors.New("nil settings provider")))
		return
	}
	s := h.provider.Current()
	rev := h.provider.CachedRevision()
	common.Success(ctx, toResponse(s, rev))
}

func (h *SettingsController) patchGithubOAuth2Settings(ctx context.Context, req *v1.GithubOAuth2PatchJSON) (v1.SettingsResponse, error) {
	if h.provider == nil || h.store == nil {
		return v1.SettingsResponse{}, common.NewInternal("settings provider is not configured", errors.New("nil settings provider"))
	}

	patch := &settingtypes.SettingsPatchRequest{GithubOAuth2: patchGithubFromV1(req)}
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

// PatchGithubOAuth2Settings handles PATCH /api/v1/settings/github-oauth2 (partial merge).
// PatchGithubOAuth2Settings 处理 PATCH /api/v1/settings/github-oauth2（部分合并）。
//
// @Summary      Patch GitHub OAuth2 settings (admin)
// @Description  Merges github_oauth2 subtree; omit fields to keep existing. Pass empty string to clear.
// @Tags         settings
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      v1.GithubOAuth2PatchJSON  true  "Partial GitHub OAuth2 settings"
// @Success      200   {object}  common.BaseResponse{data=v1.SettingsResponse}
// @Failure      400   {object}  common.BaseResponse
// @Failure      401   {object}  common.BaseResponse
// @Failure      403   {object}  common.BaseResponse
// @Failure      500   {object}  common.BaseResponse
// @Router       /api/v1/settings/github-oauth2 [patch]
func (h *SettingsController) PatchGithubOAuth2Settings(ctx *gin.Context) {
	var req v1.GithubOAuth2PatchJSON
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

	out, err := h.patchGithubOAuth2Settings(ctx.Request.Context(), &req)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, out)
}
