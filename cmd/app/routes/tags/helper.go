package tags

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// parseTagID extracts the :id path parameter as int64.
// parseTagID 将 :id 路径参数提取为 int64。
func parseTagID(ctx *gin.Context) (int64, bool) {
	raw := ctx.Param("id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id < 1 {
		common.Fail(ctx, common.NewBadRequest("invalid tag id", fmt.Errorf("parse: %w", err)))
		return 0, false
	}
	return id, true
}

// toTagItem converts a model.Tag to its admin API response item.
// toTagItem 将 model.Tag 转换为管理员 API 响应项。
func toTagItem(t model.Tag) v1.TagItem {
	return v1.TagItem{
		ID:          t.ID,
		Name:        t.Name,
		Slug:        t.Slug,
		Description: t.Description,
		Color:       t.Color,
		Status:      t.Status,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// toPublicTagItem converts a model.Tag to its public API response item (status omitted).
// toPublicTagItem 将 model.Tag 转换为公开 API 响应项（不含 status）。
func toPublicTagItem(t model.Tag) v1.TagItem {
	return v1.TagItem{
		ID:          t.ID,
		Name:        t.Name,
		Slug:        t.Slug,
		Description: t.Description,
		Color:       t.Color,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// mapTagCrudUniqueViolation maps a PostgreSQL unique-constraint violation to a public error.
// mapTagCrudUniqueViolation 将 PostgreSQL 唯一约束冲突映射为对外错误。
func mapTagCrudUniqueViolation(err error, resource string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		msg := "tag slug is already taken"
		if resource != "" {
			msg = resource + " slug is already taken"
		}
		return common.NewConflict(msg, err)
	}
	return common.NewInternal("failed to create tag", err)
}

// actor holds optional caller info for admin detection on public routes.
// actor 保存可选调用者信息，用于在公开路由上检测管理员。
type actor struct {
	uid  int64
	role string
}

func (a actor) isAdmin() bool {
	return a.role == "admin"
}

// actorFromContext extracts optional actor info from the Gin context.
// Returns zero-value actor if no valid claims are present (anonymous).
// actorFromContext 从 Gin 上下文提取可选调用者信息。若无有效 claims 则返回零值（匿名）。
func actorFromContext(ctx *gin.Context) actor {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		return actor{}
	}
	return actor{uid: claims.UID, role: claims.Role}
}
