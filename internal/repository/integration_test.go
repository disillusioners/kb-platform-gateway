package repository_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"kb-platform-gateway/internal/config"
	"kb-platform-gateway/internal/models"
	"kb-platform-gateway/internal/repository"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupIntegration loads config and connects to the DB, or skips the test.
func setupIntegration(t *testing.T) *repository.PostgresRepository {
	t.Helper()

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Try to locate .env file (walking up directories)
	findEnv := func() string {
		dir, _ := os.Getwd()
		for i := 0; i < 4; i++ { // limit search depth
			path := filepath.Join(dir, ".env")
			if _, err := os.Stat(path); err == nil {
				return path
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
		return ""
	}

	envPath := findEnv()
	if envPath != "" {
		_ = godotenv.Load(envPath)
	}

	// Only run if DB_HOST is explicitly set (or loaded from .env)
	if os.Getenv("DB_HOST") == "" {
		t.Skip("Skipping integration test: DB_HOST not set")
	}

	cfg, err := config.Load()
	require.NoError(t, err)

	repo, err := repository.NewPostgresRepository(&cfg.Database)
	if err != nil {
		t.Skipf("Skipping integration test: failed to connect to database: %v", err)
	}

	// Read and execute schema.sql
	// schema.sql is likely in the same directory as .env (project root)
	schemaPath := filepath.Join(filepath.Dir(envPath), "schema.sql")
	if envPath == "" {
		// Fallback: look relative to current package dir
		dir, _ := os.Getwd()
		schemaPath = filepath.Join(filepath.Dir(filepath.Dir(dir)), "schema.sql")
	}

	schemaContent, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Logf("Warning: Could not read schema.sql: %v", err)
	} else {
		if _, err := repo.DB().Exec(string(schemaContent)); err != nil {
			t.Fatalf("Failed to initialize database schema: %v", err)
		}
	}

	return repo
}

func TestPostgresRepository_Integration_CreateAndGetDocument(t *testing.T) {
	repo := setupIntegration(t)
	defer repo.Close()
	ctx := context.Background()

	docID := uuid.New().String()
	doc := &models.Document{
		ID:        docID,
		Filename:  "integration_test.pdf",
		FileSize:  12345,
		Status:    "pending",
		CreatedAt: time.Now().Truncate(time.Microsecond), // Postgres precision handling
		Metadata:  map[string]string{"type": "test", "priority": "high"},
	}

	// Cleanup first (just in case)
	defer repo.DeleteDocument(ctx, docID)

	// 1. Create
	err := repo.CreateDocument(ctx, doc)
	require.NoError(t, err, "Failed to create document")

	// 2. Get
	fetched, err := repo.GetDocument(ctx, docID)
	require.NoError(t, err, "Failed to get document")
	require.NotNil(t, fetched)

	assert.Equal(t, doc.ID, fetched.ID)
	assert.Equal(t, doc.Filename, fetched.Filename)
	assert.Equal(t, doc.Status, fetched.Status)
	// Check Metadata
	assert.Equal(t, "test", fetched.Metadata["type"])

	// 3. Update Status
	err = repo.UpdateDocumentStatus(ctx, docID, "indexing", "")
	require.NoError(t, err)

	fetched, err = repo.GetDocument(ctx, docID)
	require.NoError(t, err)
	assert.Equal(t, "indexing", fetched.Status)

	// 4. List (filter by status)
	list, total, err := repo.ListDocuments(ctx, 10, 0, "indexing")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 1)
	found := false
	for _, d := range list {
		if d.ID == docID {
			found = true
			break
		}
	}
	assert.True(t, found, "Created document should appear in list")
}

func TestPostgresRepository_Integration_ConversationsAndMessages(t *testing.T) {
	repo := setupIntegration(t)
	defer repo.Close()
	ctx := context.Background()

	convID := uuid.New().String()
	conv := &models.Conversation{
		ID:        convID,
		CreatedAt: time.Now().Truncate(time.Microsecond),
		UpdatedAt: time.Now().Truncate(time.Microsecond),
	}

	// 1. Create Conversation
	err := repo.CreateConversation(ctx, conv)
	require.NoError(t, err)

	// 2. Create Message
	msgID := uuid.New().String()
	msg := &models.Message{
		ID:             msgID,
		ConversationID: convID,
		Role:           "user",
		Content:        "Hello integration test",
		CreatedAt:      time.Now().Truncate(time.Microsecond),
	}
	err = repo.CreateMessage(ctx, msg)
	require.NoError(t, err)

	// 3. Get Messages
	msgs, err := repo.GetMessagesByConversationID(ctx, convID, 10, 0)
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	assert.Equal(t, msg.Content, msgs[0].Content)

	// Cleanup
	repo.DeleteMessage(ctx, msgID)
	// Usually we'd delete conversation too, but there's no DeleteConversation method in the interface?
	// Checking the interface... Repository interface wasn't shown fully, but let's assume no delete conversation for now or check PostgresRepository.
}
