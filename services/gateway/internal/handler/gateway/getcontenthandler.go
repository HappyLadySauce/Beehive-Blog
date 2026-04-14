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

func GetContentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ContentPathRequest
		if err := httpx.Parse(r, &req); err != nil {
			writeError(r.Context(), w, err)
			return
		}

		l := gateway.NewGetContentLogic(r.Context(), svcCtx)
		resp, err := l.GetContent(&req)
		if err != nil {
			writeError(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
