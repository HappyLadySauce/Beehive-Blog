package attachments

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	pkgattachment "github.com/HappyLadySauce/Beehive-Blog/pkg/attachment"
)

func (h *AttachmentsController) optionalActor(ctx *gin.Context) (pkgattachment.Actor, bool, error) {
	header := strings.TrimSpace(ctx.GetHeader("Authorization"))
	if header == "" {
		return pkgattachment.Actor{}, false, nil
	}
	scheme, tokenString, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(tokenString) == "" {
		return pkgattachment.Actor{}, false, fmt.Errorf("authorization header must use Bearer scheme")
	}
	claims, err := h.svc.Token.ParseAccess(strings.TrimSpace(tokenString))
	if err != nil {
		return pkgattachment.Actor{}, false, err
	}
	return pkgattachment.Actor{UID: claims.UID, Role: claims.Role}, true, nil
}

func actorFromClaims(ctx *gin.Context) pkgattachment.Actor {
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		return pkgattachment.Actor{}
	}
	return pkgattachment.Actor{UID: claims.UID, Role: claims.Role}
}
