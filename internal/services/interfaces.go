package services

import (
	"context"
	"time"

	"kb-platform-gateway/internal/models"

	"go.temporal.io/api/workflowservice/v1"
)

//go:generate mockgen -destination=mocks/mock_interfaces.go -package=mocks github.com/kb-platform-gateway/internal/services S3ClientInterface,TemporalClientInterface,QdrantClientInterface,PythonCoreClientInterface

// S3ClientInterface defines the interface for S3 operations.
type S3ClientInterface interface {
	// GeneratePresignedUploadURL generates a presigned URL for uploading an object.
	GeneratePresignedUploadURL(ctx context.Context, key string, expires time.Duration) (string, error)

	// GeneratePresignedDownloadURL generates a presigned URL for downloading an object.
	GeneratePresignedDownloadURL(ctx context.Context, key string, expires time.Duration) (string, error)

	// DeleteObject deletes an object from S3.
	DeleteObject(ctx context.Context, key string) error
}

// TemporalClientInterface defines the interface for Temporal workflow operations.
type TemporalClientInterface interface {
	// Close closes the Temporal client connection.
	Close()

	// StartUploadWorkflow starts the document upload workflow.
	StartUploadWorkflow(ctx context.Context, documentID, s3Key string) (string, error)

	// SignalUploadComplete signals that the upload is complete.
	SignalUploadComplete(ctx context.Context, documentID string) error

	// StartIndexWorkflow starts the document indexing workflow.
	StartIndexWorkflow(ctx context.Context, documentID string) (string, error)

	// QueryWorkflowStatus queries the status of a workflow.
	QueryWorkflowStatus(ctx context.Context, workflowID string) (*workflowservice.DescribeWorkflowExecutionResponse, error)

	// CancelWorkflow cancels a workflow.
	CancelWorkflow(ctx context.Context, workflowID string) error

	// HealthCheck checks the health of the Temporal service.
	HealthCheck(ctx context.Context) error
}

// QdrantClientInterface defines the interface for Qdrant vector database operations.
type QdrantClientInterface interface {
	// Close closes the Qdrant client connection.
	Close() error

	// DeleteDocumentVectors deletes all vectors associated with a document.
	DeleteDocumentVectors(ctx context.Context, documentID string) error
}

// PythonCoreClientInterface defines the interface for Python Core service operations.
type PythonCoreClientInterface interface {
	// Query sends a query to the RAG system and returns a stream of events.
	Query(query string, conversationID string, topK int) (<-chan models.SSEEvent, error)

	// HealthCheck checks the health of the Python Core service.
	HealthCheck() (map[string]string, error)
}
