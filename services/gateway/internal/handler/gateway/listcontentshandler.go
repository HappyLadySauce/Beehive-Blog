// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package gateway

import (
	"net/http"

	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/logic/gateway"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ListContentsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ContentListRequest
		if err := httpx.Parse(r, &req); err != nil {
			writeError(r.Context(), w, err)
			return
		}

		l := gateway.NewListContentsLogic(r.Context(), svcCtx)
		resp, err := l.ListContents(&req)
		if err != nil {
			writeError(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
