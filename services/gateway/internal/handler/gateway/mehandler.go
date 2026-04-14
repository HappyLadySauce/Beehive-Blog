// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package gateway

import (
	"net/http"

	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/logic/gateway"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func MeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := gateway.NewMeLogic(r.Context(), svcCtx)
		resp, err := l.Me()
		if err != nil {
			writeError(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
