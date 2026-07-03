package routes

import (
	"github.com/Aias00/cloudbase/internal/handler"
	"github.com/gin-gonic/gin"
)

// RegisterWebRoutes registers browser-safe web session and checkout routes
// used by first-party frontend surfaces.
func RegisterWebRoutes(v1 *gin.RouterGroup, h *handler.Handlers) {
	if h == nil || h.Web == nil {
		return
	}
	registerWebSessionRoutes(v1.Group("/web"), h.Web)
}

func registerWebSessionRoutes(web *gin.RouterGroup, h *handler.WebHandler) {
	if web == nil || h == nil {
		return
	}
	auth := web.Group("/auth")
	{
		auth.GET("/oauth/:provider/start", h.OAuthStart)
		auth.POST("/login", h.Login)
		auth.POST("/register", h.Register)
		auth.POST("/oauth/session", h.OAuthSession)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/logout", h.Logout)
		auth.GET("/me", h.Me)
		auth.GET("/credits", h.Credits)
	}
	web.POST("/payments/checkout", h.Checkout)
	web.POST("/prompts/import-twitter", h.ImportTwitter)
}
