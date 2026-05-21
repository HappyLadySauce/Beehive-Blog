package categories

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// mapFirstError maps gorm.ErrRecordNotFound to 404 and other errors to 500.
// mapFirstError 将 gorm.ErrRecordNotFound 映射为 404，其余错误映射为 500。
func mapFirstError(err error, notFoundMsg, internalMsg string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return common.NewNotFound(notFoundMsg, err)
	}
	return common.NewInternal(internalMsg, err)
}

// parseCategoryID extracts the :id path parameter as int64.
// parseCategoryID 将 :id 路径参数提取为 int64。
func parseCategoryID(ctx *gin.Context) (int64, bool) {
	raw := ctx.Param("id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id < 1 {
		common.Fail(ctx, common.NewBadRequest("invalid category id", fmt.Errorf("parse: %w", err)))
		return 0, false
	}
	return id, true
}

// toCategoryItem converts a model.Category to its API response item.
// toCategoryItem 将 model.Category 转换为 API 响应项。
func toCategoryItem(c model.Category) v1.CategoryItem {
	return v1.CategoryItem{
		ID:          c.ID,
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		ParentID:    c.ParentID,
		SortOrder:   c.SortOrder,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

// mapCategoryCreateUniqueViolation maps a unique-constraint violation on category create.
// mapCategoryCreateUniqueViolation 映射分类创建时的唯一约束冲突。
func mapCategoryCreateUniqueViolation(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return common.NewConflict("category slug is already taken", err)
	}
	return common.NewInternal("failed to create category", err)
}

// mapCategoryUpdateUniqueViolation maps a unique-constraint violation on category update.
// mapCategoryUpdateUniqueViolation 映射分类更新时的唯一约束冲突。
func mapCategoryUpdateUniqueViolation(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return common.NewConflict("category slug is already taken", err)
	}
	return common.NewInternal("failed to update category", err)
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
