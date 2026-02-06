package mocks

import (
	"context"
	"time"

	"kb-platform-gateway/internal/models"

	"github.com/stretchr/testify/mock"
	"go.temporal.io/api/workflowservice/v1"
)

// MockPythonCoreClient is a mock implementation of PythonCoreClientInterface.
type MockPythonCoreClient struct {
	mock.Mock
}

func NewMockPythonCoreClient() *MockPythonCoreClient {
	return &MockPythonCoreClient{}
}

func (m *MockPythonCoreClient) Query(query string, conversationID string, topK int) (<-chan models.SSEEvent, error) {
	args := m.Called(query, conversationID, topK)
	return args.Get(0).(<-chan models.SSEEvent), args.Error(1)
}

func (m *MockPythonCoreClient) HealthCheck() (map[string]string, error) {
	args := m.Called()
	if len(args) > 0 {
		if err := args.Error(1); err != nil {
			return nil, err
		}
		if args.Get(0) != nil {
			return args.Get(0).(map[string]string), nil
		}
	}
	return nil, nil
}

// MockS3Client is a mock implementation of S3ClientInterface.
type MockS3Client struct {
	mock.Mock
}

func NewMockS3Client() *MockS3Client {
	return &MockS3Client{}
}

func (m *MockS3Client) GeneratePresignedUploadURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	args := m.Called(ctx, key, expires)
	return args.String(0), args.Error(1)
}

func (m *MockS3Client) GeneratePresignedDownloadURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	args := m.Called(ctx, key, expires)
	return args.String(0), args.Error(1)
}

func (m *MockS3Client) DeleteObject(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	if len(args) > 0 {
		if err := args.Error(0); err != nil {
			return err
		}
	}
	return nil
}

// MockTemporalClient is a mock implementation of TemporalClientInterface.
type MockTemporalClient struct {
	mock.Mock
}

func NewMockTemporalClient() *MockTemporalClient {
	return &MockTemporalClient{}
}

func (m *MockTemporalClient) Close() {
	m.Called()
}

func (m *MockTemporalClient) StartUploadWorkflow(ctx context.Context, documentID, s3Key string) (string, error) {
	args := m.Called(ctx, documentID, s3Key)
	if len(args) > 1 {
		if err := args.Error(1); err != nil {
			return "", err
		}
		if args.String(0) != "" {
			return args.String(0), nil
		}
	}
	return "", nil
}

func (m *MockTemporalClient) SignalUploadComplete(ctx context.Context, documentID string) error {
	args := m.Called(ctx, documentID)
	if len(args) > 0 {
		if err := args.Error(0); err != nil {
			return err
		}
	}
	return nil
}

func (m *MockTemporalClient) StartIndexWorkflow(ctx context.Context, documentID string) (string, error) {
	args := m.Called(ctx, documentID)
	return args.String(0), args.Error(1)
}

func (m *MockTemporalClient) QueryWorkflowStatus(ctx context.Context, workflowID string) (*workflowservice.DescribeWorkflowExecutionResponse, error) {
	args := m.Called(ctx, workflowID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*workflowservice.DescribeWorkflowExecutionResponse), args.Error(1)
}

func (m *MockTemporalClient) CancelWorkflow(ctx context.Context, workflowID string) error {
	args := m.Called(ctx, workflowID)
	return args.Error(0)
}

func (m *MockTemporalClient) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockQdrantClient is a mock implementation of QdrantClientInterface.
type MockQdrantClient struct {
	mock.Mock
}

func NewMockQdrantClient() *MockQdrantClient {
	return &MockQdrantClient{}
}

func (m *MockQdrantClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockQdrantClient) DeleteDocumentVectors(ctx context.Context, documentID string) error {
	args := m.Called(ctx, documentID)
	if len(args) > 0 {
		if err := args.Error(0); err != nil {
			return err
		}
	}
	return nil
}
