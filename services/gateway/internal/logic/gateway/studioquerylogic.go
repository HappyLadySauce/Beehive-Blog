// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package gateway

import (
	"context"

	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type StudioQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewStudioQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StudioQueryLogic {
	return &StudioQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *StudioQueryLogic) StudioQuery(req *types.SearchRequest) (resp *types.SearchResponse, err error) {
	// Reuse the existing query pipeline so owner/public scope stays consistent.
	return NewQueryLogic(l.ctx, l.svcCtx).Query(req)
}
