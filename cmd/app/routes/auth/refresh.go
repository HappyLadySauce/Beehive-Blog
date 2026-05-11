package auth

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
	authsession "github.com/HappyLadySauce/Beehive-Blog/pkg/auth/session"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// Refresh handles POST /api/v1/auth/refresh.
// Refresh 处理 POST /api/v1/auth/refresh。
//
// @Summary      Refresh access token
// @Description  Rotates the refresh session and returns a new access + refresh pair. Reuse of an old refresh token may revoke the session family. 中文：轮换 refresh 会话并返回新的令牌对；重复使用旧 refresh 可能导致会话族被吊销。
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      v1.RefreshRequest  true  "Refresh token"
// @Success      200   {object}  common.BaseResponse{data=v1.RefreshResponse}  "New token pair"
// @Failure      401   {object}  common.BaseResponse                          "Invalid, expired, or reused refresh token"
// @Failure      403   {object}  common.BaseResponse                          "User status disallows login"
// @Failure      500   {object}  common.BaseResponse                          "Internal error"
// @Router       /api/v1/auth/refresh [post]
func (a *AuthController) Refresh(ctx *gin.Context) {
	var req v1.RefreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	resp, err := a.refresh(ctx, &req)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}

// refresh rotates a valid server-side refresh session and returns a new token pair.
// refresh 轮换有效的服务端 refresh 会话并返回新的令牌对。
func (a *AuthController) refresh(ctx *gin.Context, req *v1.RefreshRequest) (*v1.RefreshResponse, error) {
	claims, err := a.svc.Token.ParseRefresh(req.RefreshToken)
	if err != nil {
		return nil, common.NewUnauthorized("invalid or expired refresh token", nil)
	}
	if claims.SID <= 0 || claims.ID == "" {
		return nil, common.NewUnauthorized("invalid or expired refresh token", nil)
	}

	var pair jwt.TokenPair
	var publicErr error
	err = a.svc.DB.Transaction(func(tx *gorm.DB) error {
		var current model.UserSession
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ?", claims.SID, claims.UID).
			First(&current).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return common.NewUnauthorized("invalid or expired refresh token", nil)
			}
			return common.NewInternal("failed to refresh token", fmt.Errorf("query session: %w", err))
		}

		if current.RotatedAt != nil {
			if err := authsession.RevokeSession(tx, current.ID, current.UserID, "refresh_reuse"); err != nil {
				return common.NewInternal("failed to refresh token", err)
			}
			publicErr = common.NewUnauthorized("invalid or expired refresh token", nil)
			return nil
		}
		if current.RevokedAt != nil || time.Now().After(current.ExpiresAt) {
			return common.NewUnauthorized("invalid or expired refresh token", nil)
		}
		if current.RefreshJTI != claims.ID || current.RefreshTokenHash != authsession.HashRefreshToken(req.RefreshToken) {
			if err := authsession.RevokeSession(tx, current.ID, current.UserID, "refresh_mismatch"); err != nil {
				return common.NewInternal("failed to refresh token", err)
			}
			publicErr = common.NewUnauthorized("invalid or expired refresh token", nil)
			return nil
		}

		// Refresh tokens must be re-authorized against current user state before minting access.
		// 刷新令牌签发 access 前必须基于当前用户状态重新授权。
		var user model.User
		if err := tx.First(&user, claims.UID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return common.NewUnauthorized("invalid or expired refresh token", nil)
			}
			return common.NewInternal("failed to refresh token", fmt.Errorf("query user: %w", err))
		}
		if err := assertUserMayLogin(&user); err != nil {
			return err
		}

		nextPair, _, err := authsession.Rotate(tx, a.svc.Token, &current, &user, authsession.ClientMeta{
			IP:        ctx.ClientIP(),
			UserAgent: ctx.Request.UserAgent(),
		})
		if err != nil {
			return common.NewInternal("failed to issue access token", err)
		}
		pair = nextPair
		return nil
	})
	if err != nil {
		return nil, err
	}
	if publicErr != nil {
		return nil, publicErr
	}

	return &v1.RefreshResponse{
		Token: v1.AuthToken{
			AccessToken:  pair.Access.Token,
			TokenType:    pair.TokenType,
			ExpiresIn:    pair.Access.ExpiresIn,
			RefreshToken: pair.Refresh.Token,
		},
	}, nil
}
