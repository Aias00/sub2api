package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterPromptRoutes registers public prompt catalog routes shared by
// Sub2API web frontend surfaces.
func RegisterPromptRoutes(v1 *gin.RouterGroup, h *handler.Handlers, adminAuth middleware.AdminAuthMiddleware) {
	if h == nil {
		return
	}

	if h.PromptCatalog != nil {
		prompts := v1.Group("/prompts")
		{
			prompts.GET("/cases", h.PromptCatalog.ListCases)
			prompts.GET("/cases/:id", h.PromptCatalog.GetCase)
		}
	}

	if adminAuth != nil {
		adminPrompts := v1.Group("/admin/prompts")
		adminPrompts.Use(gin.HandlerFunc(adminAuth))
		{
			if h.PromptCatalog != nil {
				adminPrompts.POST("/cases", h.PromptCatalog.UpsertCase)
			}
			if h.TwitterImport != nil {
				adminPrompts.POST("/import-twitter", h.TwitterImport.Import)
			}
		}
	}
}
