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

type SubmitReviewLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitReviewLogic {
	return &SubmitReviewLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SubmitReviewLogic) SubmitReview(in *pb.SubmitReviewRequest) (*pb.ReviewTask, error) {
	out, err := l.svcCtx.Store.SubmitReview(l.ctx, in)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return out, nil
}
