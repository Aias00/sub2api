package routes

import (
	"time"

	"github.com/Aias00/cloudbase/internal/handler"
	"github.com/Aias00/cloudbase/internal/middleware"
	servermiddleware "github.com/Aias00/cloudbase/internal/server/middleware"
	"github.com/Aias00/cloudbase/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RegisterImageWorkspaceRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	jwtAuth servermiddleware.JWTAuthMiddleware,
	settingService *service.SettingService,
	redisClient *redis.Client,
) {
	if h == nil || h.ImageWorkspace == nil {
		return
	}

	authenticated := v1.Group("/image-workspace")
	authenticated.Use(gin.HandlerFunc(jwtAuth))
	authenticated.Use(servermiddleware.BackendModeUserGuard(settingService))
	{
		// Per-user rate limit on task creation to stop scripted queue flooding.
		// Keyed by authenticated user id (set by jwtAuth) so users behind a
		// shared NAT/CDN don't share a bucket. FailOpen: a Redis blip must not
		// lock paying users out. Limit value is intentionally hardcoded for
		// now; promote to a setting if tuning becomes needed.
		createTaskLimit := middleware.NewRateLimiter(redisClient).LimitByUserID(
			"image-workspace-create",
			30,
			time.Minute,
			middleware.RateLimitOptions{FailureMode: middleware.RateLimitFailOpen},
			func(c *gin.Context) (int64, bool) {
				subject, ok := servermiddleware.GetAuthSubjectFromContext(c)
				return subject.UserID, ok && subject.UserID > 0
			},
		)
		authenticated.POST("/tasks", createTaskLimit, h.ImageWorkspace.CreateTask)
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
