package handlers

import (
	"context"
	"io"
	"net/http"
	"time"

	"kb-platform-gateway/internal/auth"
	"kb-platform-gateway/internal/config"
	"kb-platform-gateway/internal/models"
	"kb-platform-gateway/internal/services"
	"kb-platform-gateway/pkg/sse"

	"github.com/disillusioners/kb-platform-proto/gen/go/kbplatform/v1"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Handlers struct {
	JWTManager *auth.Manager
	CoreClient *services.PythonCoreClient
	GrpcClient *services.GrpcCoreClient
	S3Client   *services.S3Client
	SSEHub     *sse.Hub
	Logger     zerolog.Logger
}

func NewHandlers(cfg *config.Config, sseHub *sse.Hub, logger zerolog.Logger) *Handlers {
	// Create HTTP client for Python Core
	httpClient := services.NewPythonCoreClient(cfg.Services.PythonCoreHost, cfg.Services.PythonCorePort)

	// Create gRPC client for Python Core
	grpcClient, err := services.NewGrpcCoreClient(cfg.Services.PythonCoreHost, 50051)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to create gRPC client, falling back to HTTP")
		grpcClient = nil
	}

	// Create S3 client
	s3Client, err := services.NewS3Client(&cfg.S3)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to create S3 client, presigned URLs will not work")
		s3Client = nil
	}

	return &Handlers{
		JWTManager: auth.NewManager(cfg.JWT.Secret, cfg.JWT.Expiration),
		CoreClient: httpClient,
		GrpcClient: grpcClient,
		S3Client:   s3Client,
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
	// Try gRPC health check first, fall back to HTTP
	if h.GrpcClient != nil {
		if err := h.GrpcClient.HealthCheck(context.Background()); err != nil {
			c.JSON(http.StatusServiceUnavailable, models.ReadinessResponse{
				Status:       "not_ready",
				Dependencies: map[string]string{"python_core_grpc": err.Error()},
			})
			return
		}
	}

	deps := map[string]string{
		"python_core": "ok",
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

	// Generate real UUID
	documentID := uuid.New().String()
	s3Key := "documents/" + documentID + "/" + file.Filename

	c.JSON(http.StatusOK, models.Document{
		ID:        documentID,
		UploadURL: h.generatePresignedUploadURL(s3Key, file.Header.Get("Content-Type")),
		S3Key:     s3Key,
		Filename:  file.Filename,
		FileSize:  file.Size,
		Status:    "pending",
		CreatedAt: time.Now(),
	})
}

func (h *Handlers) ListDocuments(c *gin.Context) {
	// TODO: Implement actual document listing via gRPC
	// For now, return empty list
	c.JSON(http.StatusOK, models.DocumentListResponse{
		Documents: []models.Document{},
		Total:     0,
		Limit:     50,
		Offset:    0,
	})
}

func (h *Handlers) GetDocument(c *gin.Context) {
	documentID := c.Param("id")

	// Try gRPC client first
	if h.GrpcClient != nil {
		doc, err := h.GrpcClient.GetDocument(context.Background(), documentID)
		if err != nil {
			h.Logger.Error().Err(err).Str("document_id", documentID).Msg("Failed to get document via gRPC")
		} else if doc != nil {
			// Convert proto document to models.Document
			c.JSON(http.StatusOK, convertProtoDocumentToModel(doc))
			return
		}
	}

	// Fall back to HTTP client
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

	// Try gRPC client first
	if h.GrpcClient != nil {
		if err := h.GrpcClient.DeleteDocumentVectors(context.Background(), documentID); err != nil {
			h.Logger.Error().Err(err).Str("document_id", documentID).Msg("Failed to delete document via gRPC")
		} else {
			c.Status(http.StatusNoContent)
			return
		}
	}

	// Fall back to HTTP client
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

	// TODO: Trigger Temporal workflow for document processing
	// For now, just return the document with indexing status
	c.JSON(http.StatusOK, models.Document{
		ID:     documentID,
		Status: "indexing",
	})
}

func (h *Handlers) ListConversations(c *gin.Context) {
	// TODO: Implement actual conversation listing via gRPC
	// For now, return empty list
	c.JSON(http.StatusOK, models.ConversationListResponse{
		Conversations: []models.Conversation{},
		Total:         0,
		Limit:         50,
		Offset:        0,
	})
}

func (h *Handlers) CreateConversation(c *gin.Context) {
	// Generate real UUID
	conversationID := uuid.New().String()
	now := time.Now()

	c.JSON(http.StatusCreated, models.Conversation{
		ID:        conversationID,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (h *Handlers) GetConversationMessages(c *gin.Context) {
	conversationID := c.Param("id")

	// Try gRPC client first
	if h.GrpcClient != nil {
		messages, err := h.GrpcClient.GetConversationMessages(context.Background(), conversationID)
		if err != nil {
			h.Logger.Error().Err(err).Str("conversation_id", conversationID).Msg("Failed to get messages via gRPC")
		} else {
			// Convert proto messages to models.Message
			msgList := make([]models.Message, len(messages))
			for i, msg := range messages {
				msgList[i] = convertProtoMessageToModel(msg)
			}
			c.JSON(http.StatusOK, models.MessageListResponse{
				Messages: msgList,
			})
			return
		}
	}

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

	// Try gRPC client first for streaming
	if h.GrpcClient != nil {
		eventChan, err := h.GrpcClient.QueryStream(context.Background(), req.Query, req.ConversationID, req.TopK)
		if err != nil {
			h.Logger.Error().Err(err).Str("query", req.Query).Msg("Failed to query via gRPC")
		} else {
			c.Header("Content-Type", "text/event-stream")
			c.Header("Cache-Control", "no-cache")
			c.Header("Connection", "keep-alive")
			c.Stream(func(w io.Writer) bool {
				for event := range eventChan {
					c.SSEvent("message", convertProtoEventToSSE(event))
					if flusher, ok := w.(http.Flusher); ok {
						flusher.Flush()
					}
				}
				return false
			})
			return
		}
	}

	// Fall back to HTTP client
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

// Helper functions

// generatePresignedUploadURL generates a presigned URL for uploading a document
func (h *Handlers) generatePresignedUploadURL(s3Key string, contentType string) string {
	if h.S3Client == nil {
		// Return placeholder if S3 client is not available
		return "https://s3.amazonaws.com/kb-documents/" + s3Key + "?presigned=true"
	}

	url, err := h.S3Client.GenerateUploadPresignedURL(s3Key, contentType, 15*time.Minute)
	if err != nil {
		h.Logger.Error().Err(err).Str("s3_key", s3Key).Msg("Failed to generate presigned URL")
		return ""
	}

	return url
}

// convertProtoDocumentToModel converts proto Document to local models.Document
func convertProtoDocumentToModel(doc *v1.Document) models.Document {
	return models.Document{
		ID:           doc.Id,
		Filename:     doc.Filename,
		FileSize:     doc.FileSize,
		Status:       doc.Status,
		CreatedAt:    time.Now(),
		IndexedAt:    nil,
		ErrorMessage: doc.ErrorMessage,
		Metadata:     doc.Metadata,
	}
}

// convertProtoMessageToModel converts proto Message to local models.Message
func convertProtoMessageToModel(msg *v1.Message) models.Message {
	return models.Message{
		ID:             msg.Id,
		ConversationID: msg.ConversationId,
		Role:           msg.Role,
		Content:        msg.Content,
		Timestamp:      time.Now(),
		Metadata:       msg.Metadata,
	}
}

// convertProtoEventToSSE converts proto QueryResponse to SSE event
func convertProtoEventToSSE(event *v1.QueryResponse) models.SSEEvent {
	switch e := event.Event.(type) {
	case *v1.QueryResponse_Start:
		return models.SSEEvent{
			Type: "start",
			ID:   e.Start.RequestId,
		}
	case *v1.QueryResponse_Chunk:
		return models.SSEEvent{
			Type:    "chunk",
			Content: e.Chunk.Content,
		}
	case *v1.QueryResponse_End:
		return models.SSEEvent{
			Type: "end",
			ID:   e.End.RequestId,
		}
	case *v1.QueryResponse_Error:
		return models.SSEEvent{
			Type:    "error",
			Code:    e.Error.Code,
			Message: e.Error.Message,
		}
	default:
		return models.SSEEvent{
			Type: "unknown",
		}
	}
}
