package gateway

import (
	"context"
	"time"

	contentrpc "github.com/HappyLadySauce/Beehive-Blog/services/content/content"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DeleteTagLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteTagLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTagLogic {
	return &DeleteTagLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *DeleteTagLogic) DeleteTag(req *types.TagPathRequest) (err error) {
	if req == nil || req.Id <= 0 {
		return status.Error(codes.InvalidArgument, "id is required")
	}
	userID, _ := parseAccessTokenUserIDFromContext(l.ctx)
	fields := map[string]any{"user_id": userID, "tag_id": req.Id}
	startedAt := time.Now()
	defer func() {
		auditRPC(l.ctx, "content.delete_tag", "content.DeleteTag", startedAt, err, fields)
	}()

	_, err = l.svcCtx.Content.DeleteTag(l.ctx, &contentrpc.DeleteTagRequest{Id: req.Id})
	return err
}
