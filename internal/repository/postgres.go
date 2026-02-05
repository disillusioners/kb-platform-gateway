package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"kb-platform-gateway/internal/config"
	"kb-platform-gateway/internal/models"

	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(cfg *config.DatabaseConfig) (*PostgresRepository, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

type DocumentRow struct {
	ID           string
	Filename     string
	FileSize     int64
	Status       string
	ErrorMessage *string
	S3Key        *string
	CreatedAt    time.Time
	IndexedAt    *time.Time
}

func (r *PostgresRepository) CreateDocument(ctx context.Context, doc *models.Document) error {
	query := `
		INSERT INTO documents (id, filename, file_size, status, s3_key, error_message, created_at, indexed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		doc.ID, doc.Filename, doc.FileSize, doc.Status,
		nullString(doc.S3Key), nullString(doc.ErrorMessage),
		doc.CreatedAt, nullTime(doc.IndexedAt),
	)

	return err
}

func (r *PostgresRepository) GetDocument(ctx context.Context, id string) (*models.Document, error) {
	query := `
		SELECT id, filename, file_size, status, s3_key, error_message, created_at, indexed_at
		FROM documents
		WHERE id = $1
	`

	var row DocumentRow
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&row.ID, &row.Filename, &row.FileSize, &row.Status,
		&row.S3Key, &row.ErrorMessage, &row.CreatedAt, &row.IndexedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return rowToDocument(&row), nil
}

func (r *PostgresRepository) ListDocuments(ctx context.Context, limit, offset int, statusFilter string) ([]*models.Document, int, error) {
	query := `
		SELECT id, filename, file_size, status, s3_key, error_message, created_at, indexed_at
		FROM documents
	`

	var args []interface{}
	var whereClauses []string

	if statusFilter != "" {
		args = append(args, statusFilter)
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", len(args)))
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + whereClauses[0]
	}

	query += " ORDER BY created_at DESC LIMIT $" + fmt.Sprintf("%d", len(args)+1) + " OFFSET $" + fmt.Sprintf("%d", len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var documents []*models.Document
	for rows.Next() {
		var row DocumentRow
		if err := rows.Scan(
			&row.ID, &row.Filename, &row.FileSize, &row.Status,
			&row.S3Key, &row.ErrorMessage, &row.CreatedAt, &row.IndexedAt,
		); err != nil {
			return nil, 0, err
		}
		documents = append(documents, rowToDocument(&row))
	}

	countQuery := "SELECT COUNT(*) FROM documents"
	if len(whereClauses) > 0 {
		countQuery += " WHERE " + whereClauses[0]
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args[:len(args)-2]...).Scan(&total); err != nil {
		return nil, 0, err
	}

	return documents, total, nil
}

func (r *PostgresRepository) UpdateDocument(ctx context.Context, id string, updates map[string]interface{}) error {
	setClauses := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argNum := 1

	for key, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", key, argNum))
		args = append(args, value)
		argNum++
	}
	args = append(args, id)

	query := fmt.Sprintf("UPDATE documents SET %s WHERE id = $%d", fmt.Sprintf("%s", setClauses), argNum)

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *PostgresRepository) DeleteDocument(ctx context.Context, id string) error {
	query := "DELETE FROM documents WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *PostgresRepository) UpdateDocumentStatus(ctx context.Context, id, status string, errorMessage string) error {
	query := `
		UPDATE documents
		SET status = $1, error_message = $2, indexed_at = $3
		WHERE id = $4
	`

	var indexedAt *time.Time
	if status == "complete" || status == "failed" {
		now := time.Now()
		indexedAt = &now
	}

	_, err := r.db.ExecContext(ctx, query, status, nullString(errorMessage), nullTime(indexedAt), id)
	return err
}

type ConversationRow struct {
	ID           sql.NullString
	CreatedAt    time.Time
	UpdatedAt    time.Time
	MessageCount sql.NullInt64
}

func (r *PostgresRepository) CreateConversation(ctx context.Context, conv *models.Conversation) error {
	query := `
		INSERT INTO conversations (id, created_at, updated_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.ExecContext(ctx, query, conv.ID, conv.CreatedAt, conv.UpdatedAt)
	return err
}

func (r *PostgresRepository) GetConversation(ctx context.Context, id string) (*models.Conversation, error) {
	query := `
		SELECT id, created_at, updated_at, message_count
		FROM conversations
		WHERE id = $1
	`

	var row ConversationRow
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&row.ID, &row.CreatedAt, &row.UpdatedAt, &row.MessageCount,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	conv := &models.Conversation{
		ID:        row.ID.String,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
	if row.MessageCount.Valid {
		conv.MessageCount = int(row.MessageCount.Int64)
	}

	return conv, nil
}

func (r *PostgresRepository) ListConversations(ctx context.Context, userID string, limit, offset int) ([]*models.Conversation, int, error) {
	query := `
		SELECT id, created_at, updated_at, message_count
		FROM conversations
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var conversations []*models.Conversation
	for rows.Next() {
		var row ConversationRow
		if err := rows.Scan(&row.ID, &row.CreatedAt, &row.UpdatedAt, &row.MessageCount); err != nil {
			return nil, 0, err
		}

		conv := &models.Conversation{
			ID:        row.ID.String,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		}
		if row.MessageCount.Valid {
			conv.MessageCount = int(row.MessageCount.Int64)
		}
		conversations = append(conversations, conv)
	}

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM conversations").Scan(&total); err != nil {
		return nil, 0, err
	}

	return conversations, total, nil
}

func (r *PostgresRepository) UpdateMessageCount(ctx context.Context, id string, count int) error {
	query := `
		UPDATE conversations
		SET message_count = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, count, time.Now(), id)
	return err
}

func (r *PostgresRepository) CreateMessage(ctx context.Context, msg *models.Message) error {
	query := `
		INSERT INTO messages (id, conversation_id, role, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query, msg.ID, msg.ConversationID, msg.Role, msg.Content, msg.CreatedAt)

	return err
}

func (r *PostgresRepository) GetMessagesByConversationID(ctx context.Context, conversationID string, limit, offset int) ([]*models.Message, error) {
	query := `
		SELECT id, conversation_id, role, content, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, conversationID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		var msg models.Message
		if err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}

	return messages, nil
}

func (r *PostgresRepository) DeleteMessage(ctx context.Context, id string) error {
	query := "DELETE FROM messages WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func rowToDocument(row *DocumentRow) *models.Document {
	doc := &models.Document{
		ID:        row.ID,
		Filename:  row.Filename,
		FileSize:  row.FileSize,
		Status:    row.Status,
		CreatedAt: row.CreatedAt,
	}

	if row.S3Key != nil {
		doc.S3Key = *row.S3Key
	}
	if row.ErrorMessage != nil {
		doc.ErrorMessage = *row.ErrorMessage
	}
	if row.IndexedAt != nil {
		doc.IndexedAt = row.IndexedAt
	}

	return doc
}

func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nullTime(t *time.Time) *time.Time {
	return t
}
