package services_test

import (
	"context"
	"testing"
	"time"

	"kb-platform-gateway/internal/services/mocks"

	"github.com/stretchr/testify/assert"
)

func TestS3Client(t *testing.T) {
	t.Run("GeneratePresignedUploadURL_Success", func(t *testing.T) {
		mockS3Client := mocks.NewMockS3Client()
		ctx := context.Background()
		mockS3Client.On("GeneratePresignedUploadURL", ctx, "documents/test.pdf", 15*time.Minute).Return("https://s3.example.com/upload?signature=abc", nil)

		url, err := mockS3Client.GeneratePresignedUploadURL(ctx, "documents/test.pdf", 15*time.Minute)

		assert.NoError(t, err)
		assert.Equal(t, "https://s3.example.com/upload?signature=abc", url)
		mockS3Client.AssertExpectations(t)
	})

	t.Run("GeneratePresignedDownloadURL_Success", func(t *testing.T) {
		mockS3Client := mocks.NewMockS3Client()
		ctx := context.Background()
		mockS3Client.On("GeneratePresignedDownloadURL", ctx, "documents/test.pdf", 15*time.Minute).Return("https://s3.example.com/download?signature=xyz", nil)

		url, err := mockS3Client.GeneratePresignedDownloadURL(ctx, "documents/test.pdf", 15*time.Minute)

		assert.NoError(t, err)
		assert.Equal(t, "https://s3.example.com/download?signature=xyz", url)
		mockS3Client.AssertExpectations(t)
	})

	t.Run("DeleteObject_Success", func(t *testing.T) {
		mockS3Client := mocks.NewMockS3Client()
		ctx := context.Background()
		mockS3Client.On("DeleteObject", ctx, "documents/test.pdf").Return(nil)

		err := mockS3Client.DeleteObject(ctx, "documents/test.pdf")

		assert.NoError(t, err)
		mockS3Client.AssertExpectations(t)
	})

	t.Run("DeleteObject_Error", func(t *testing.T) {
		mockS3Client := mocks.NewMockS3Client()
		ctx := context.Background()
		mockS3Client.On("DeleteObject", ctx, "documents/test.pdf").Return(assert.AnError)

		err := mockS3Client.DeleteObject(ctx, "documents/test.pdf")

		assert.Error(t, err)
		mockS3Client.AssertExpectations(t)
	})
}

func TestPythonCoreClient(t *testing.T) {
	t.Run("HealthCheck_Success", func(t *testing.T) {
		mockClient := mocks.NewMockPythonCoreClient()
		mockClient.On("HealthCheck").Return(map[string]string{"python_core": "ok"}, nil)

		deps, err := mockClient.HealthCheck()

		assert.NoError(t, err)
		assert.Equal(t, "ok", deps["python_core"])
		mockClient.AssertExpectations(t)
	})

	t.Run("HealthCheck_Error", func(t *testing.T) {
		mockClient := mocks.NewMockPythonCoreClient()
		mockClient.On("HealthCheck").Return(nil, assert.AnError)

		deps, err := mockClient.HealthCheck()

		assert.Error(t, err)
		assert.Nil(t, deps)
		mockClient.AssertExpectations(t)
	})
}

func TestTemporalClient(t *testing.T) {
	t.Run("StartUploadWorkflow_Success", func(t *testing.T) {
		mockClient := mocks.NewMockTemporalClient()
		ctx := context.Background()
		mockClient.On("StartUploadWorkflow", ctx, "doc-123", "s3://bucket/doc-123/test.pdf").Return("workflow-id-123", nil)

		workflowID, err := mockClient.StartUploadWorkflow(ctx, "doc-123", "s3://bucket/doc-123/test.pdf")

		assert.NoError(t, err)
		assert.Equal(t, "workflow-id-123", workflowID)
		mockClient.AssertExpectations(t)
	})

	t.Run("StartUploadWorkflow_Error", func(t *testing.T) {
		mockClient := mocks.NewMockTemporalClient()
		ctx := context.Background()
		mockClient.On("StartUploadWorkflow", ctx, "doc-123", "s3://bucket/doc-123/test.pdf").Return("", assert.AnError)

		workflowID, err := mockClient.StartUploadWorkflow(ctx, "doc-123", "s3://bucket/doc-123/test.pdf")

		assert.Error(t, err)
		assert.Empty(t, workflowID)
		mockClient.AssertExpectations(t)
	})

	t.Run("SignalUploadComplete_Success", func(t *testing.T) {
		mockClient := mocks.NewMockTemporalClient()
		ctx := context.Background()
		mockClient.On("SignalUploadComplete", ctx, "doc-123").Return(nil)

		err := mockClient.SignalUploadComplete(ctx, "doc-123")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("SignalUploadComplete_Error", func(t *testing.T) {
		mockClient := mocks.NewMockTemporalClient()
		ctx := context.Background()
		mockClient.On("SignalUploadComplete", ctx, "doc-123").Return(assert.AnError)

		err := mockClient.SignalUploadComplete(ctx, "doc-123")

		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("StartIndexWorkflow_Success", func(t *testing.T) {
		mockClient := mocks.NewMockTemporalClient()
		ctx := context.Background()
		mockClient.On("StartIndexWorkflow", ctx, "doc-123").Return("index-workflow-123", nil)

		workflowID, err := mockClient.StartIndexWorkflow(ctx, "doc-123")

		assert.NoError(t, err)
		assert.Equal(t, "index-workflow-123", workflowID)
		mockClient.AssertExpectations(t)
	})

	t.Run("HealthCheck_Success", func(t *testing.T) {
		mockClient := mocks.NewMockTemporalClient()
		ctx := context.Background()
		mockClient.On("HealthCheck", ctx).Return(nil)

		err := mockClient.HealthCheck(ctx)

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})
}

func TestQdrantClient(t *testing.T) {
	t.Run("DeleteDocumentVectors_Success", func(t *testing.T) {
		mockClient := mocks.NewMockQdrantClient()
		ctx := context.Background()
		mockClient.On("DeleteDocumentVectors", ctx, "doc-123").Return(nil)

		err := mockClient.DeleteDocumentVectors(ctx, "doc-123")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("DeleteDocumentVectors_Error", func(t *testing.T) {
		mockClient := mocks.NewMockQdrantClient()
		ctx := context.Background()
		mockClient.On("DeleteDocumentVectors", ctx, "doc-123").Return(assert.AnError)

		err := mockClient.DeleteDocumentVectors(ctx, "doc-123")

		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("Close_Success", func(t *testing.T) {
		mockClient := mocks.NewMockQdrantClient()
		mockClient.On("Close").Return(nil)

		err := mockClient.Close()

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})
}
