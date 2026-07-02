package routes

import (
	"github.com/Wei-Shaw/cloudbase/internal/handler"
	"github.com/Wei-Shaw/cloudbase/internal/server/middleware"
	"github.com/Wei-Shaw/cloudbase/internal/service"
	"github.com/gin-gonic/gin"
)

func RegisterWeChatExportRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	jwtAuth middleware.JWTAuthMiddleware,
	settingService *service.SettingService,
) {
	if h == nil || h.WeChatExport == nil {
		return
	}

	authenticated := v1.Group("/wechat")
	authenticated.Use(gin.HandlerFunc(jwtAuth))
	authenticated.Use(middleware.BackendModeUserGuard(settingService))
	{
		authenticated.GET("/session", h.WeChatExport.GetSession)
		authenticated.POST("/session/qrcode", h.WeChatExport.CreateQRCodeSession)
		authenticated.GET("/session/poll/:sessionID", h.WeChatExport.PollSession)
		authenticated.POST("/session/validate", h.WeChatExport.ValidateSession)
		authenticated.POST("/session/logout", h.WeChatExport.LogoutSession)

		authenticated.GET("/accounts/search", h.WeChatExport.SearchAccounts)
		authenticated.POST("/accounts/bind", h.WeChatExport.BindAccount)
		authenticated.POST("/accounts/:accountID/sync", h.WeChatExport.SyncAccount)

		authenticated.GET("/articles", h.WeChatExport.ListArticles)
		authenticated.POST("/articles/import-link", h.WeChatExport.ImportArticleLink)

		authenticated.POST("/tasks/quote", h.WeChatExport.QuoteTask)
		authenticated.POST("/tasks", h.WeChatExport.CreateTask)
		authenticated.GET("/tasks", h.WeChatExport.ListTasks)
		authenticated.GET("/worker/status", h.WeChatExport.GetWorkerStatus)
		authenticated.GET("/tasks/:taskID", h.WeChatExport.GetTask)
		authenticated.POST("/tasks/:taskID/cancel", h.WeChatExport.CancelTask)
		authenticated.POST("/tasks/:taskID/retry", h.WeChatExport.RetryTask)
		authenticated.GET("/tasks/:taskID/logs", h.WeChatExport.ListTaskLogs)
		authenticated.GET("/tasks/:taskID/artifacts", h.WeChatExport.ListArtifacts)
		authenticated.GET("/tasks/:taskID/artifacts.zip", h.WeChatExport.DownloadTaskArtifactsZip)
		authenticated.GET("/artifacts/:artifactID/download", h.WeChatExport.DownloadArtifact)
	}

	worker := v1.Group("/wechat/worker")
	{
		worker.GET("/health", h.WeChatExport.WorkerHealth)
		worker.GET("/runtime-config", h.WeChatExport.WorkerRuntimeConfig)
		worker.POST("/tasks/claim", h.WeChatExport.WorkerClaimTask)
		worker.POST("/articles/:articleID/enrich", h.WeChatExport.WorkerEnrichArticle)
		worker.POST("/articles/:articleID/engagement", h.WeChatExport.WorkerFetchArticleEngagement)
		worker.POST("/tasks/:taskID/logs", h.WeChatExport.WorkerAddTaskLog)
		worker.POST("/tasks/:taskID/complete", h.WeChatExport.WorkerCompleteTask)
		worker.POST("/tasks/:taskID/fail", h.WeChatExport.WorkerFailTask)
	}
}
