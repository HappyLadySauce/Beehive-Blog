package httpx

import (
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
)

// HandleJSON returns a gin.HandlerFunc that binds JSON into *Req, invokes fn, and writes common.Success or common.Fail.
// The factory shape allocates the wrapper closure once at route-registration time, not per request.
// HandleJSON 返回 gin.HandlerFunc：将 JSON 绑定到 *Req，调用 fn，再写入 common.Success 或 common.Fail。
// 工厂式签名仅在路由注册时分配一次包装闭包，而非每次请求都分配。
func HandleJSON[Req any, Resp any](fn func(*gin.Context, *Req) (Resp, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.ShouldBindJSON(&req); err != nil {
			klog.ErrorS(err, "Failed to bind JSON request body")
			common.Fail(ctx, common.NewBadRequest("invalid request body", err))
			return
		}
		resp, err := fn(ctx, &req)
		if err != nil {
			common.Fail(ctx, err)
			return
		}
		common.Success(ctx, resp)
	}
}
