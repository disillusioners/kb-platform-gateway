package handlers

import (
	"io"
	"net/http"
	"time"

	"kb-platform-gateway/internal/auth"
	"kb-platform-gateway/internal/config"
	"kb-platform-gateway/internal/models"
	"kb-platform-gateway/internal/services"
	"kb-platform-gateway/pkg/sse"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type Handlers struct {
	JWTManager *auth.Manager
	CoreClient *services.PythonCoreClient
	SSEHub     *sse.Hub
	Logger     zerolog.Logger
}

func NewHandlers(cfg *config.Config, sseHub *sse.Hub, logger zerolog.Logger) *Handlers {
	return &Handlers{
		JWTManager: auth.NewManager(cfg.JWT.Secret, cfg.JWT.Expiration),
		CoreClient: services.NewPythonCoreClient(cfg.Services.PythonCoreHost, cfg.Services.PythonCorePort),
		SSEHub:     sseHub,
		Logger:     logger,
	}
}

func (h *Handlers) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid request format",
			},
		})
		return
	}

	token, expiresAt, err := h.JWTManager.GenerateToken(req.Username)
	if err != nil {
		h.Logger.Error().Err(err).Msg("Failed to generate token")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to generate token",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	})
}

func (h *Handlers) Health(c *gin.Context) {
	c.JSON(http.StatusOK, models.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

func (h *Handlers) Ready(c *gin.Context) {
	deps, err := h.CoreClient.HealthCheck()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, models.ReadinessResponse{
			Status:       "not_ready",
			Dependencies: map[string]string{"python_core": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, models.ReadinessResponse{
		Status:       "ready",
		Dependencies: deps,
	})
}

func (h *Handlers) UploadDocument(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: "No file provided",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.Document{
		ID:        generateUUID(),
		UploadURL: "https://s3.amazonaws.com/bucket/presigned-url",
		S3Key:     "documents/" + generateUUID() + "/" + file.Filename,
		Filename:  file.Filename,
		FileSize:  file.Size,
		Status:    "pending",
		CreatedAt: time.Now(),
	})
}

func (h *Handlers) ListDocuments(c *gin.Context) {
	c.JSON(http.StatusOK, models.DocumentListResponse{
		Documents: []models.Document{},
		Total:     0,
		Limit:     50,
		Offset:    0,
	})
}

func (h *Handlers) GetDocument(c *gin.Context) {
	documentID := c.Param("id")

	doc, err := h.CoreClient.GetDocument(documentID)
	if err != nil {
		h.Logger.Error().Err(err).Str("document_id", documentID).Msg("Failed to get document")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get document",
			},
		})
		return
	}

	if doc == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "NOT_FOUND",
				Message: "Document not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, doc)
}

func (h *Handlers) DeleteDocument(c *gin.Context) {
	documentID := c.Param("id")

	if err := h.CoreClient.DeleteDocumentVectors(documentID); err != nil {
		h.Logger.Error().Err(err).Str("document_id", documentID).Msg("Failed to delete document")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to delete document",
			},
		})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handlers) CompleteUpload(c *gin.Context) {
	documentID := c.Param("id")

	c.JSON(http.StatusOK, models.Document{
		ID:     documentID,
		Status: "indexing",
	})
}

func (h *Handlers) ListConversations(c *gin.Context) {
	c.JSON(http.StatusOK, models.ConversationListResponse{
		Conversations: []models.Conversation{},
		Total:         0,
		Limit:         50,
		Offset:        0,
	})
}

func (h *Handlers) CreateConversation(c *gin.Context) {
	now := time.Now()
	c.JSON(http.StatusCreated, models.Conversation{
		ID:        generateUUID(),
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (h *Handlers) GetConversationMessages(c *gin.Context) {
	c.JSON(http.StatusOK, models.MessageListResponse{
		Messages: []models.Message{},
	})
}

func (h *Handlers) Query(c *gin.Context) {
	var req models.QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid request format",
			},
		})
		return
	}

	if req.TopK == 0 {
		req.TopK = 5
	}

	eventChan, err := h.CoreClient.Query(req.Query, req.ConversationID, req.TopK)
	if err != nil {
		h.Logger.Error().Err(err).Str("query", req.Query).Msg("Failed to query")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to query",
			},
		})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Stream(func(w io.Writer) bool {
		for event := range eventChan {
			c.SSEvent("message", event)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
		return false
	})
}

func generateUUID() string {
	return "550e8400-e29b-41d4-a716-446655440000"
}
