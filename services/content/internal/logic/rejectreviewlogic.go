package logic

import (
	"context"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/services/content/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/content/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RejectReviewLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRejectReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RejectReviewLogic {
	return &RejectReviewLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RejectReviewLogic) RejectReview(in *pb.RejectReviewRequest) (*pb.ReviewTask, error) {
	out, err := l.svcCtx.Store.RejectReview(l.ctx, in)
	if err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "not found") {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if strings.Contains(msg, "already claimed") || strings.Contains(msg, "not pending") {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return out, nil
}
