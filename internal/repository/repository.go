package repository

import (
	"context"

	"kb-platform-gateway/internal/models"
)

type DocumentRepository interface {
	CreateDocument(ctx context.Context, doc *models.Document) error
	GetDocument(ctx context.Context, id string) (*models.Document, error)
	ListDocuments(ctx context.Context, limit, offset int, statusFilter string) ([]*models.Document, int, error)
	UpdateDocument(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteDocument(ctx context.Context, id string) error
	UpdateDocumentStatus(ctx context.Context, id, status string, errorMessage string) error
}

type ConversationRepository interface {
	CreateConversation(ctx context.Context, conv *models.Conversation) error
	GetConversation(ctx context.Context, id string) (*models.Conversation, error)
	ListConversations(ctx context.Context, userID string, limit, offset int) ([]*models.Conversation, int, error)
	UpdateMessageCount(ctx context.Context, id string, count int) error
}

type MessageRepository interface {
	CreateMessage(ctx context.Context, msg *models.Message) error
	GetMessagesByConversationID(ctx context.Context, conversationID string, limit, offset int) ([]*models.Message, error)
	DeleteMessage(ctx context.Context, id string) error
}

type Repository interface {
	DocumentRepository
	ConversationRepository
	MessageRepository
}
