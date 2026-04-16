package gateway

import (
	"context"
	"strings"
	"time"

	contentrpc "github.com/HappyLadySauce/Beehive-Blog/services/content/content"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateRelationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateRelationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRelationLogic {
	return &CreateRelationLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *CreateRelationLogic) CreateRelation(req *types.CreateRelationRequest) (resp *types.Relation, err error) {
	if req == nil || req.ContentId <= 0 || req.TargetContentId <= 0 || strings.TrimSpace(req.RelationType) == "" {
		return nil, status.Error(codes.InvalidArgument, "contentId,targetContentId,relationType are required")
	}
	userID, _ := parseAccessTokenUserIDFromContext(l.ctx)
	fields := map[string]any{"user_id": userID, "source_content_id": req.ContentId, "target_content_id": req.TargetContentId}
	startedAt := time.Now()
	defer func() {
		auditRPC(l.ctx, "content.create_relation", "content.CreateRelation", startedAt, err, fields)
	}()

	out, err := l.svcCtx.Content.CreateRelation(l.ctx, &contentrpc.CreateRelationRequest{
		SourceContentId: req.ContentId,
		TargetContentId: req.TargetContentId,
		RelationType:    req.RelationType,
		Weight:          req.Weight,
		Note:            req.Note,
	})
	if err != nil {
		return nil, err
	}
	return toRelation(out), nil
}
