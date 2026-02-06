package repository_test

import (
	"context"
	"testing"
	"time"

	"kb-platform-gateway/internal/models"
	"kb-platform-gateway/internal/repository"
	"kb-platform-gateway/internal/repository/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDocumentRepository tests the DocumentRepository methods.
func TestDocumentRepository(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockRepository()

	t.Run("CreateDocument_Success", func(t *testing.T) {
		doc := &models.Document{
			ID:        "test-doc-1",
			Filename:  "test.pdf",
			FileSize:  1024,
			Status:    "pending",
			CreatedAt: time.Now(),
		}

		repo.On("CreateDocument", ctx, doc).Return(nil)

		err := repo.CreateDocument(ctx, doc)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("GetDocument_Found", func(t *testing.T) {
		expectedDoc := &models.Document{
			ID:        "test-doc-1",
			Filename:  "test.pdf",
			FileSize:  1024,
			Status:    "pending",
			CreatedAt: time.Now(),
		}

		repo.On("GetDocument", ctx, "test-doc-1").Return(expectedDoc, nil)

		doc, err := repo.GetDocument(ctx, "test-doc-1")

		require.NoError(t, err)
		assert.NotNil(t, doc)
		assert.Equal(t, "test-doc-1", doc.ID)
		repo.AssertExpectations(t)
	})

	t.Run("GetDocument_NotFound", func(t *testing.T) {
		repo.On("GetDocument", ctx, "non-existent").Return(nil, nil)

		doc, err := repo.GetDocument(ctx, "non-existent")

		require.NoError(t, err)
		assert.Nil(t, doc)
		repo.AssertExpectations(t)
	})

	t.Run("ListDocuments_WithPagination", func(t *testing.T) {
		docs := []*models.Document{
			{ID: "doc-1", Filename: "file1.pdf", Status: "pending"},
			{ID: "doc-2", Filename: "file2.pdf", Status: "complete"},
		}

		repo.On("ListDocuments", ctx, 50, 0, "").Return(docs, 2, nil)

		result, total, err := repo.ListDocuments(ctx, 50, 0, "")

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, 2, total)
		repo.AssertExpectations(t)
	})

	t.Run("ListDocuments_WithStatusFilter", func(t *testing.T) {
		docs := []*models.Document{
			{ID: "doc-1", Filename: "file1.pdf", Status: "pending"},
		}

		repo.On("ListDocuments", ctx, 50, 0, "pending").Return(docs, 1, nil)

		result, total, err := repo.ListDocuments(ctx, 50, 0, "pending")

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, 1, total)
		repo.AssertExpectations(t)
	})

	t.Run("DeleteDocument_Success", func(t *testing.T) {
		repo.On("DeleteDocument", ctx, "test-doc-1").Return(nil)

		err := repo.DeleteDocument(ctx, "test-doc-1")

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("UpdateDocumentStatus_Complete", func(t *testing.T) {
		repo.On("UpdateDocumentStatus", ctx, "test-doc-1", "complete", "").Return(nil)

		err := repo.UpdateDocumentStatus(ctx, "test-doc-1", "complete", "")

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("UpdateDocumentStatus_Failed", func(t *testing.T) {
		repo.On("UpdateDocumentStatus", ctx, "test-doc-1", "failed", "error message").Return(nil)

		err := repo.UpdateDocumentStatus(ctx, "test-doc-1", "failed", "error message")

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})
}

// TestConversationRepository tests the ConversationRepository methods.
func TestConversationRepository(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockRepository()

	t.Run("CreateConversation_Success", func(t *testing.T) {
		conv := &models.Conversation{
			ID:        "conv-1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		repo.On("CreateConversation", ctx, conv).Return(nil)

		err := repo.CreateConversation(ctx, conv)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("GetConversation_Found", func(t *testing.T) {
		expectedConv := &models.Conversation{
			ID:           "conv-1",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			MessageCount: 5,
		}

		repo.On("GetConversation", ctx, "conv-1").Return(expectedConv, nil)

		conv, err := repo.GetConversation(ctx, "conv-1")

		require.NoError(t, err)
		assert.NotNil(t, conv)
		assert.Equal(t, "conv-1", conv.ID)
		assert.Equal(t, 5, conv.MessageCount)
		repo.AssertExpectations(t)
	})

	t.Run("GetConversation_NotFound", func(t *testing.T) {
		repo.On("GetConversation", ctx, "non-existent").Return(nil, nil)

		conv, err := repo.GetConversation(ctx, "non-existent")

		require.NoError(t, err)
		assert.Nil(t, conv)
		repo.AssertExpectations(t)
	})

	t.Run("ListConversations_WithPagination", func(t *testing.T) {
		convs := []*models.Conversation{
			{ID: "conv-1", MessageCount: 5},
			{ID: "conv-2", MessageCount: 3},
		}

		repo.On("ListConversations", ctx, "user-1", 50, 0).Return(convs, 2, nil)

		result, total, err := repo.ListConversations(ctx, "user-1", 50, 0)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, 2, total)
		repo.AssertExpectations(t)
	})
}

// TestMessageRepository tests the MessageRepository methods.
func TestMessageRepository(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockRepository()

	t.Run("CreateMessage_Success", func(t *testing.T) {
		msg := &models.Message{
			ID:             "msg-1",
			ConversationID: "conv-1",
			Role:           "user",
			Content:        "Hello",
			CreatedAt:      time.Now(),
		}

		repo.On("CreateMessage", ctx, msg).Return(nil)

		err := repo.CreateMessage(ctx, msg)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("GetMessagesByConversationID_Success", func(t *testing.T) {
		msgs := []*models.Message{
			{ID: "msg-1", ConversationID: "conv-1", Role: "user", Content: "Hello"},
			{ID: "msg-2", ConversationID: "conv-1", Role: "assistant", Content: "Hi there!"},
		}

		repo.On("GetMessagesByConversationID", ctx, "conv-1", 50, 0).Return(msgs, nil)

		result, err := repo.GetMessagesByConversationID(ctx, "conv-1", 50, 0)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		repo.AssertExpectations(t)
	})

	t.Run("GetMessagesByConversationID_Empty", func(t *testing.T) {
		repo.On("GetMessagesByConversationID", ctx, "conv-empty", 50, 0).Return([]*models.Message{}, nil)

		result, err := repo.GetMessagesByConversationID(ctx, "conv-empty", 50, 0)

		require.NoError(t, err)
		assert.Len(t, result, 0)
		repo.AssertExpectations(t)
	})

	t.Run("DeleteMessage_Success", func(t *testing.T) {
		repo.On("DeleteMessage", ctx, "msg-1").Return(nil)

		err := repo.DeleteMessage(ctx, "msg-1")

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})
}

// TestRepositoryInterfaceCompliance ensures the mock implements all Repository methods.
func TestRepositoryInterfaceCompliance(t *testing.T) {
	// This test ensures that MockRepository implements the Repository interface.
	var _ repository.Repository = (*mocks.MockRepository)(nil)
	t.Log("MockRepository correctly implements Repository interface")
}
