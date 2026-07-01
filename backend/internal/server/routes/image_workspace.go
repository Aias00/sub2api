package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func RegisterImageWorkspaceRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	jwtAuth middleware.JWTAuthMiddleware,
	settingService *service.SettingService,
) {
	if h == nil || h.ImageWorkspace == nil {
		return
	}

	authenticated := v1.Group("/image-workspace")
	authenticated.Use(gin.HandlerFunc(jwtAuth))
	authenticated.Use(middleware.BackendModeUserGuard(settingService))
	{
		authenticated.POST("/tasks", h.ImageWorkspace.CreateTask)
		authenticated.GET("/tasks", h.ImageWorkspace.ListTasks)
		authenticated.GET("/models", h.ImageWorkspace.ListModels)
		authenticated.GET("/tasks/:taskID", h.ImageWorkspace.GetTask)
		authenticated.POST("/tasks/:taskID/cancel", h.ImageWorkspace.CancelTask)
		authenticated.POST("/tasks/:taskID/retry", h.ImageWorkspace.RetryTask)
		authenticated.GET("/artifacts/:artifactID/download", h.ImageWorkspace.DownloadArtifact)
		authenticated.GET("/templates", h.ImageWorkspace.ListTemplates)
		authenticated.POST("/templates", h.ImageWorkspace.UpsertTemplate)
		authenticated.DELETE("/templates/:templateID", h.ImageWorkspace.DeleteTemplate)
		authenticated.GET("/usage-records", h.ImageWorkspace.ListUsageRecords)
	}

	worker := v1.Group("/image-workspace/worker")
	{
		worker.GET("/health", h.ImageWorkspace.WorkerHealth)
		worker.GET("/runtime-config", h.ImageWorkspace.WorkerRuntimeConfig)
		worker.GET("/status", h.ImageWorkspace.WorkerStatus)
		worker.POST("/tasks/claim", h.ImageWorkspace.WorkerClaimTask)
		worker.POST("/tasks/:taskID/complete", h.ImageWorkspace.WorkerCompleteTask)
		worker.POST("/tasks/:taskID/fail", h.ImageWorkspace.WorkerFailTask)
	}
}
