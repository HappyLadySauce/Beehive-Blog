package gateway

import (
	"net/http"

	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/logic/gateway"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func DeleteSearchIndexHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SearchIndexPathRequest
		if err := httpx.Parse(r, &req); err != nil {
			writeError(r.Context(), w, err)
			return
		}
		l := gateway.NewDeleteSearchIndexLogic(r.Context(), svcCtx)
		if err := l.DeleteSearchIndex(&req); err != nil {
			writeError(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{"ok": true})
	}
}
