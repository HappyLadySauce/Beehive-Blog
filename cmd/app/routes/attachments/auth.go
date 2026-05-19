package attachments

import (
	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	pkgattachment "github.com/HappyLadySauce/Beehive-Blog/pkg/attachment"
)

func actorFromClaims(ctx *gin.Context) pkgattachment.Actor {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		return pkgattachment.Actor{}
	}
	return pkgattachment.Actor{UID: claims.UID, Role: claims.Role}
}
