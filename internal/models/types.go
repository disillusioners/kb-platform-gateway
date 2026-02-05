package models

import "time"

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Document struct {
	ID           string            `json:"id"`
	UploadURL    string            `json:"upload_url,omitempty"`
	S3Key        string            `json:"s3_key,omitempty"`
	Filename     string            `json:"filename"`
	FileSize     int64             `json:"file_size"`
	Status       string            `json:"status"`
	ErrorMessage string            `json:"error_message,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	IndexedAt    *time.Time        `json:"indexed_at,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type DocumentListResponse struct {
	Documents []Document `json:"documents"`
	Total     int        `json:"total"`
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
}

type Conversation struct {
	ID           string    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	MessageCount int       `json:"message_count,omitempty"`
}

type ConversationListResponse struct {
	Conversations []Conversation `json:"conversations"`
	Total         int            `json:"total"`
	Limit         int            `json:"limit"`
	Offset        int            `json:"offset"`
}

type Message struct {
	ID             string            `json:"id"`
	ConversationID string            `json:"conversation_id,omitempty"`
	Role           string            `json:"role"`
	Content        string            `json:"content"`
	CreatedAt      time.Time         `json:"created_at"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type MessageListResponse struct {
	Messages []Message `json:"messages"`
}

type QueryRequest struct {
	Query          string `json:"query" binding:"required"`
	ConversationID string `json:"conversation_id,omitempty"`
	TopK           int    `json:"top_k,omitempty"`
}

type ConversationRequest struct {
}

type SaveMessageRequest struct {
	ConversationID string            `json:"conversation_id" binding:"required"`
	Role           string            `json:"role" binding:"required,oneof=user assistant"`
	Content        string            `json:"content" binding:"required"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

type ReadinessResponse struct {
	Status       string            `json:"status"`
	Dependencies map[string]string `json:"dependencies"`
}

type SSEEvent struct {
	Type    string `json:"type"`
	ID      string `json:"id,omitempty"`
	Content string `json:"content,omitempty"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}
