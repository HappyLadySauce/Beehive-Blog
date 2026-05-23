package comments

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// create stores a reader comment in review status.
// create 以待审核状态保存读者评论。
func (c *CommentsController) create(ctx context.Context, contentID int64, req *v1.CreateCommentRequest) (*v1.CreateCommentResponse, error) {
	if err := c.db.WithContext(ctx).Where("id = ? AND status = ? AND visibility = ?", contentID, "published", "public").
		First(&model.Content{}).Error; err != nil {
		return nil, mapCommentFirstError(err)
	}

	nickname, err := cleanRequiredText(req.Nickname, "nickname")
	if err != nil {
		return nil, err
	}
	body, err := cleanRequiredText(req.Body, "body")
	if err != nil {
		return nil, err
	}
	website, err := cleanOptionalWebsite(req.Website)
	if err != nil {
		return nil, err
	}

	now := nowUTC()
	row := model.Comment{
		ContentID: contentID,
		Nickname:  nickname,
		EmailHash: normalizeEmailHash(req.Email),
		Website:   website,
		Body:      body,
		Status:    "review",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := c.db.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, common.NewInternal("failed to create comment", fmt.Errorf("insert: %w", err))
	}
	return &v1.CreateCommentResponse{ID: row.ID, Status: row.Status}, nil
}

// Create handles POST /api/v1/contents/:id/comments.
// Create 处理 POST /api/v1/contents/:id/comments。
func (c *CommentsController) Create(ctx *gin.Context) {
	contentID, ok := parseContentID(ctx)
	if !ok {
		return
	}
	var req v1.CreateCommentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeCommentError(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	resp, err := c.create(ctx.Request.Context(), contentID, &req)
	if err != nil {
		writeCommentError(ctx, err)
		return
	}
	common.Success(ctx, resp)
}
