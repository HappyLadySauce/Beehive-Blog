// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package gateway

import (
	"net/http"

	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/logic/gateway"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/svc"
	xhttp "github.com/zeromicro/x/http"
)

func HealthHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := gateway.NewHealthLogic(r.Context(), svcCtx)
		resp, err := l.Health()
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
