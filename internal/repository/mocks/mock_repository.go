package mocks

import (
	"context"

	"kb-platform-gateway/internal/models"
	"kb-platform-gateway/internal/repository"

	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of the Repository interface.
type MockRepository struct {
	mock.Mock
}

// NewMockRepository creates a new MockRepository instance.
func NewMockRepository() *MockRepository {
	return &MockRepository{}
}

// CreateDocument mocks the CreateDocument method.
func (m *MockRepository) CreateDocument(ctx context.Context, doc *models.Document) error {
	args := m.Called(ctx, doc)
	return args.Error(0)
}

// GetDocument mocks the GetDocument method.
func (m *MockRepository) GetDocument(ctx context.Context, id string) (*models.Document, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Document), args.Error(1)
}

// ListDocuments mocks the ListDocuments method.
func (m *MockRepository) ListDocuments(ctx context.Context, limit, offset int, statusFilter string) ([]*models.Document, int, error) {
	args := m.Called(ctx, limit, offset, statusFilter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.Document), args.Int(1), args.Error(2)
}

// UpdateDocument mocks the UpdateDocument method.
func (m *MockRepository) UpdateDocument(ctx context.Context, id string, updates map[string]interface{}) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

// DeleteDocument mocks the DeleteDocument method.
func (m *MockRepository) DeleteDocument(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// UpdateDocumentStatus mocks the UpdateDocumentStatus method.
func (m *MockRepository) UpdateDocumentStatus(ctx context.Context, id, status string, errorMessage string) error {
	args := m.Called(ctx, id, status, errorMessage)
	return args.Error(0)
}

// CreateConversation mocks the CreateConversation method.
func (m *MockRepository) CreateConversation(ctx context.Context, conv *models.Conversation) error {
	args := m.Called(ctx, conv)
	return args.Error(0)
}

// GetConversation mocks the GetConversation method.
func (m *MockRepository) GetConversation(ctx context.Context, id string) (*models.Conversation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Conversation), args.Error(1)
}

// ListConversations mocks the ListConversations method.
func (m *MockRepository) ListConversations(ctx context.Context, userID string, limit, offset int) ([]*models.Conversation, int, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.Conversation), args.Int(1), args.Error(2)
}

// UpdateMessageCount mocks the UpdateMessageCount method.
func (m *MockRepository) UpdateMessageCount(ctx context.Context, id string, count int) error {
	args := m.Called(ctx, id, count)
	return args.Error(0)
}

// CreateMessage mocks the CreateMessage method.
func (m *MockRepository) CreateMessage(ctx context.Context, msg *models.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

// GetMessagesByConversationID mocks the GetMessagesByConversationID method.
func (m *MockRepository) GetMessagesByConversationID(ctx context.Context, conversationID string, limit, offset int) ([]*models.Message, error) {
	args := m.Called(ctx, conversationID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Message), args.Error(1)
}

// DeleteMessage mocks the DeleteMessage method.
func (m *MockRepository) DeleteMessage(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Ensure MockRepository implements Repository interface
var _ repository.Repository = (*MockRepository)(nil)
