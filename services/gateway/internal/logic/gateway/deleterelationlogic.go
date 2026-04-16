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

type DeleteRelationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteRelationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteRelationLogic {
	return &DeleteRelationLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *DeleteRelationLogic) DeleteRelation(req *types.RelationPathRequest) (err error) {
	if req == nil || req.Id <= 0 {
		return status.Error(codes.InvalidArgument, "id is required")
	}
	userID, _ := parseAccessTokenUserIDFromContext(l.ctx)
	fields := map[string]any{"user_id": userID, "relation_id": req.Id}
	startedAt := time.Now()
	defer func() {
		auditRPC(l.ctx, "content.delete_relation", "content.DeleteRelation", startedAt, err, fields)
	}()

	_, err = l.svcCtx.Content.DeleteRelation(l.ctx, &contentrpc.DeleteRelationRequest{Id: req.Id})
	return err
}
