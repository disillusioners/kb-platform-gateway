package handlers

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"kb-platform-gateway/internal/config"
	"kb-platform-gateway/internal/models"
	"kb-platform-gateway/internal/repository"
	"kb-platform-gateway/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Handlers struct {
	CoreClient   *services.PythonCoreClient
	S3Client     *services.S3Client
	Temporal     *services.TemporalClient
	QdrantClient *services.QdrantClient
	Repository   repository.Repository
	Logger       zerolog.Logger
}

func NewHandlers(cfg *config.Config, repo repository.Repository, logger zerolog.Logger) (*Handlers, error) {
	s3Client, err := services.NewS3Client(&cfg.S3)
	if err != nil {
		return nil, err
	}

	temporalClient, err := services.NewTemporalClient(&cfg.Temporal)
	if err != nil {
		return nil, err
	}

	qdrantClient, err := services.NewQdrantClient(&cfg.Qdrant)
	if err != nil {
		return nil, err
	}

	return &Handlers{
		CoreClient:   services.NewPythonCoreClient(cfg.Services.PythonCoreHost, cfg.Services.PythonCorePort),
		S3Client:     s3Client,
		Temporal:     temporalClient,
		QdrantClient: qdrantClient,
		Repository:   repo,
		Logger:       logger,
	}, nil
}

func (h *Handlers) Close() {
	if h.Temporal != nil {
		h.Temporal.Close()
	}
	if h.QdrantClient != nil {
		h.QdrantClient.Close()
	}
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

	documentID := generateUUID()
	s3Key := "documents/" + documentID + "/" + file.Filename

	uploadURL, err := h.S3Client.GeneratePresignedUploadURL(c.Request.Context(), s3Key, 15*time.Minute)
	if err != nil {
		h.Logger.Error().Err(err).Msg("Failed to generate presigned URL")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to generate upload URL",
			},
		})
		return
	}

	doc := &models.Document{
		ID:        documentID,
		S3Key:     s3Key,
		Filename:  file.Filename,
		FileSize:  file.Size,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	if err := h.Repository.CreateDocument(c.Request.Context(), doc); err != nil {
		h.Logger.Error().Err(err).Msg("Failed to save document to database")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to save document",
			},
		})
		return
	}

	// Start two-phase upload workflow
	_, err = h.Temporal.StartUploadWorkflow(c.Request.Context(), documentID, s3Key)
	if err != nil {
		h.Logger.Error().Err(err).Msg("Failed to start upload workflow")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to start upload workflow",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.Document{
		ID:        doc.ID,
		UploadURL: uploadURL,
		S3Key:     doc.S3Key,
		Filename:  doc.Filename,
		FileSize:  doc.FileSize,
		Status:    doc.Status,
		CreatedAt: doc.CreatedAt,
	})
}

func (h *Handlers) ListDocuments(c *gin.Context) {
	limit := 50
	offset := 0
	statusFilter := c.Query("status")

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	documents, total, err := h.Repository.ListDocuments(c.Request.Context(), limit, offset, statusFilter)
	if err != nil {
		h.Logger.Error().Err(err).Msg("Failed to list documents")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to list documents",
			},
		})
		return
	}

	docList := make([]models.Document, len(documents))
	for i, doc := range documents {
		docList[i] = *doc
	}

	c.JSON(http.StatusOK, models.DocumentListResponse{
		Documents: docList,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	})
}

func (h *Handlers) GetDocument(c *gin.Context) {
	documentID := c.Param("id")

	doc, err := h.Repository.GetDocument(c.Request.Context(), documentID)
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

	doc, err := h.Repository.GetDocument(c.Request.Context(), documentID)
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

	if doc != nil && doc.S3Key != "" {
		if err := h.S3Client.DeleteObject(c.Request.Context(), doc.S3Key); err != nil {
			h.Logger.Error().Err(err).Str("s3_key", doc.S3Key).Msg("Failed to delete from S3")
		}
	}

	if err := h.QdrantClient.DeleteDocumentVectors(c.Request.Context(), documentID); err != nil {
		h.Logger.Error().Err(err).Str("document_id", documentID).Msg("Failed to delete vectors")
	}

	if err := h.Repository.DeleteDocument(c.Request.Context(), documentID); err != nil {
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

	// Signal upload completion to workflow
	if err := h.Temporal.SignalUploadComplete(c.Request.Context(), documentID); err != nil {
		h.Logger.Error().Err(err).Str("document_id", documentID).Msg("Failed to signal upload complete")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to signal upload complete",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.Document{
		ID:     documentID,
		Status: "indexing",
	})
}

func (h *Handlers) ListConversations(c *gin.Context) {
	limit := 50
	offset := 0

	userID := c.GetString("username")

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	conversations, total, err := h.Repository.ListConversations(c.Request.Context(), userID, limit, offset)
	if err != nil {
		h.Logger.Error().Err(err).Msg("Failed to list conversations")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to list conversations",
			},
		})
		return
	}

	convList := make([]models.Conversation, len(conversations))
	for i, conv := range conversations {
		convList[i] = *conv
	}

	c.JSON(http.StatusOK, models.ConversationListResponse{
		Conversations: convList,
		Total:         total,
		Limit:         limit,
		Offset:        offset,
	})
}

func (h *Handlers) CreateConversation(c *gin.Context) {
	now := time.Now()

	conv := &models.Conversation{
		ID:        generateUUID(),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.Repository.CreateConversation(c.Request.Context(), conv); err != nil {
		h.Logger.Error().Err(err).Msg("Failed to create conversation")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to create conversation",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, conv)
}

func (h *Handlers) GetConversationMessages(c *gin.Context) {
	conversationID := c.Param("id")
	limit := 50
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	messages, err := h.Repository.GetMessagesByConversationID(c.Request.Context(), conversationID, limit, offset)
	if err != nil {
		h.Logger.Error().Err(err).Str("conversation_id", conversationID).Msg("Failed to get messages")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get messages",
			},
		})
		return
	}

	msgList := make([]models.Message, len(messages))
	for i, msg := range messages {
		msgList[i] = *msg
	}

	c.JSON(http.StatusOK, models.MessageListResponse{
		Messages: msgList,
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
	return uuid.New().String()
}
