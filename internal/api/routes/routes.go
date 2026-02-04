package routes

import (
	"kb-platform-gateway/internal/api/handlers"
	"kb-platform-gateway/internal/api/middleware"
	"kb-platform-gateway/internal/config"
	"kb-platform-gateway/pkg/sse"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func SetupRoutes(router *gin.Engine, cfg *config.Config, sseHub *sse.Hub, logger zerolog.Logger) {
	h := handlers.NewHandlers(cfg, sseHub, logger)
	authMiddleware := middleware.AuthMiddleware(h.JWTManager)

	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", h.Login)
		}

		docs := api.Group("/documents")
		docs.Use(authMiddleware)
		{
			docs.POST("", h.UploadDocument)
			docs.GET("", h.ListDocuments)
			docs.GET("/:id", h.GetDocument)
			docs.DELETE("/:id", h.DeleteDocument)
			docs.POST("/:id/complete", h.CompleteUpload)
		}

		conversations := api.Group("/conversations")
		conversations.Use(authMiddleware)
		{
			conversations.GET("", h.ListConversations)
			conversations.POST("", h.CreateConversation)
			conversations.GET("/:id/messages", h.GetConversationMessages)
		}

		query := api.Group("/query")
		query.Use(authMiddleware)
		{
			query.POST("", h.Query)
		}
	}

	router.GET("/healthz", h.Health)
	router.GET("/readyz", h.Ready)
}
