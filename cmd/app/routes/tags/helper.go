package tags

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"

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

// toTagItem converts a model.Tag to its API response item.
// toTagItem 将 model.Tag 转换为 API 响应项。
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
