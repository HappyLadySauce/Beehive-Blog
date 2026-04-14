package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/HappyLadySauce/Beehive-Blog/api/swagger/docs"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middlewares"
)

var (
	router *gin.Engine
	v1     *gin.RouterGroup
)

// Livez godoc
//
//	@Summary		健康检查
//	@Description	检查服务是否存活
//	@Tags			health
//	@Accept			json
//	@Produce		plain
//	@Success		200	{string}	string	"livez"
//	@Router			/livez [get]
func Livez(c *gin.Context) {
	c.String(200, "livez")
}

// Readyz godoc
//
//	@Summary		就绪检查
//	@Description	检查服务是否就绪
//	@Tags			health
//	@Accept			json
//	@Produce		plain
//	@Success		200	{string}	string	"readyz"
//	@Router			/readyz [get]
func Readyz(c *gin.Context) {
	c.String(200, "readyz")
}

func init() {
	router = gin.Default()
	// SetTrustedProxies sets the trusted proxies for the router.
	_ = router.SetTrustedProxies(nil)

	// Recovery middleware recovers from any panics that occur in the request cycle.
	router.Use(gin.Recovery())
	router.Use(middlewares.Cors())

	v1 = router.Group("/api/v1")

	// register health check endpoints
	router.GET("/livez", Livez)
	router.GET("/readyz", Readyz)

	// register swagger routes
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// V1 returns the router group for /api/v1 which for resources in control plane endpoints.
func V1() *gin.RouterGroup {
	return v1
}

// Router returns the main Gin engine instance.
func Router() *gin.Engine {
	return router
}

// EnableRateLimit 为 /api/v1 分组启用全局限流中间件。
func EnableRateLimit() {
	v1.Use(middlewares.RateLimit())
}

// SetupStatic registers a local-disk static file route for uploaded files.
// uploadDir is the filesystem directory; it is served under the /uploads URL prefix.
func SetupStatic(uploadDir string) {
	if uploadDir == "" {
		uploadDir = "uploads"
	}
	router.Static("/uploads", uploadDir)
}

// NewServer creates an http.Server with the given address using the Gin router.
// This allows for graceful shutdown.
func NewServer(addr string) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: router,
	}
}
