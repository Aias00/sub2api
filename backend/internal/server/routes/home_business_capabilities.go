package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterHomeBusinessCapabilityRoutes(v1 *gin.RouterGroup, h *handler.Handlers) {
	if h == nil || h.HomeBusiness == nil {
		return
	}
	v1.GET("/home/business-capabilities", h.HomeBusiness.GetStatuses)
}
