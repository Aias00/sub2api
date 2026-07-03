package routes

import (
	"github.com/Aias00/cloudbase/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterHotContentRoutes(v1 *gin.RouterGroup, h *handler.Handlers) {
	if h == nil || h.HotContent == nil {
		return
	}
	hot := v1.Group("/hot")
	{
		hot.GET("/sources", h.HotContent.ListSources)
		hot.GET("/items", h.HotContent.ListItems)
		hot.GET("/run-events", h.HotContent.ListRunEvents)
	}
}
