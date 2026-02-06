package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kb-platform-gateway/internal/api/handlers"
	"kb-platform-gateway/internal/api/routes"
	"kb-platform-gateway/internal/config"
	"kb-platform-gateway/internal/repository"
	"kb-platform-gateway/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	logger.Info().Msg("Starting KB Platform Gateway")

	// Set Gin mode
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()

	// Initialize repository
	repo, err := repository.NewPostgresRepository(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	// Initialize services
	pythonCoreClient := services.NewPythonCoreClient(cfg.Services.PythonCoreHost, cfg.Services.PythonCorePort)
	s3Client, err := services.NewS3Client(&cfg.S3)
	if err != nil {
		log.Fatalf("Failed to create S3 client: %v", err)
	}
	temporalClient, err := services.NewTemporalClient(&cfg.Temporal)
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	qdrantClient, err := services.NewQdrantClient(&cfg.Qdrant)
	if err != nil {
		log.Fatalf("Failed to create Qdrant client: %v", err)
	}

	// Setup middleware
	setupMiddleware(router, cfg, logger)

	// Initialize handlers with services
	h, err := handlers.NewHandlers(repo, pythonCoreClient, s3Client, temporalClient, qdrantClient, logger)
	if err != nil {
		log.Fatalf("Failed to create handlers: %v", err)
	}
	defer func() {
		if temporalClient != nil {
			temporalClient.Close()
		}
		if qdrantClient != nil {
			qdrantClient.Close()
		}
	}()

	// Setup routes
	routes.SetupRoutes(router, cfg, h, logger)

	// Create HTTP server
	srv := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:        router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start server in goroutine
	go func() {
		logger.Info().
			Str("host", cfg.Server.Host).
			Int("port", cfg.Server.Port).
			Msg("Server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Server shutting down...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("Server forced to shutdown")
	}

	logger.Info().Msg("Server exited")
}

func setupMiddleware(router *gin.Engine, cfg *config.Config, logger zerolog.Logger) {
	// Recovery middleware
	router.Use(gin.Recovery())

	// Logger middleware
	router.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Log after processing
		latency := time.Since(start)
		status := c.Writer.Status()

		logger.Info().
			Str("method", method).
			Str("path", path).
			Int("status", status).
			Dur("latency", latency).
			Str("client_ip", c.ClientIP()).
			Msg("Request processed")
	})

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})
}
