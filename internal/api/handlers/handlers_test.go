package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"kb-platform-gateway/internal/api/handlers"
	"kb-platform-gateway/internal/models"
	"kb-platform-gateway/internal/services/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestHealthHandler(t *testing.T) {
	t.Run("Health_Success", func(t *testing.T) {
		mockCoreClient := mocks.NewMockPythonCoreClient()
		mockS3Client := mocks.NewMockS3Client()
		mockTemporalClient := mocks.NewMockTemporalClient()
		mockQdrantClient := mocks.NewMockQdrantClient()

		h := &handlers.Handlers{
			CoreClient:   mockCoreClient,
			S3Client:     mockS3Client,
			Temporal:     mockTemporalClient,
			QdrantClient: mockQdrantClient,
		}

		router := setupTestRouter()
		router.GET("/healthz", h.Health)

		req, _ := http.NewRequest("GET", "/healthz", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		var response models.HealthResponse
		err := json.Unmarshal(resp.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "healthy", response.Status)
	})
}

func TestReadyHandler(t *testing.T) {
	t.Run("Ready_Success", func(t *testing.T) {
		mockCoreClient := mocks.NewMockPythonCoreClient()
		mockCoreClient.On("HealthCheck").Return(map[string]string{"python_core": "ok"}, nil)

		mockS3Client := mocks.NewMockS3Client()
		mockTemporalClient := mocks.NewMockTemporalClient()
		mockQdrantClient := mocks.NewMockQdrantClient()

		h := &handlers.Handlers{
			CoreClient:   mockCoreClient,
			S3Client:     mockS3Client,
			Temporal:     mockTemporalClient,
			QdrantClient: mockQdrantClient,
		}

		router := setupTestRouter()
		router.GET("/readyz", h.Ready)

		req, _ := http.NewRequest("GET", "/readyz", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		var response models.ReadinessResponse
		err := json.Unmarshal(resp.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "ready", response.Status)
		mockCoreClient.AssertExpectations(t)
	})

	t.Run("Ready_PythonCoreUnavailable", func(t *testing.T) {
		mockCoreClient := mocks.NewMockPythonCoreClient()
		mockCoreClient.On("HealthCheck").Return(nil, assert.AnError)

		mockS3Client := mocks.NewMockS3Client()
		mockTemporalClient := mocks.NewMockTemporalClient()
		mockQdrantClient := mocks.NewMockQdrantClient()

		h := &handlers.Handlers{
			CoreClient:   mockCoreClient,
			S3Client:     mockS3Client,
			Temporal:     mockTemporalClient,
			QdrantClient: mockQdrantClient,
		}

		router := setupTestRouter()
		router.GET("/readyz", h.Ready)

		req, _ := http.NewRequest("GET", "/readyz", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusServiceUnavailable, resp.Code)

		var response models.ReadinessResponse
		err := json.Unmarshal(resp.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "not_ready", response.Status)
		mockCoreClient.AssertExpectations(t)
	})
}

func TestUploadDocumentHandler_NoFile(t *testing.T) {
	t.Run("UploadDocument_NoFile_Returns400", func(t *testing.T) {
		mockCoreClient := mocks.NewMockPythonCoreClient()
		mockS3Client := mocks.NewMockS3Client()
		mockTemporalClient := mocks.NewMockTemporalClient()
		mockQdrantClient := mocks.NewMockQdrantClient()

		h := &handlers.Handlers{
			CoreClient:   mockCoreClient,
			S3Client:     mockS3Client,
			Temporal:     mockTemporalClient,
			QdrantClient: mockQdrantClient,
		}

		router := setupTestRouter()
		router.POST("/documents", h.UploadDocument)

		// Empty body (no file)
		req, _ := http.NewRequest("POST", "/documents", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})
}

func TestCompleteUploadHandler_TemporalError(t *testing.T) {
	t.Run("CompleteUpload_TemporalError_Returns500", func(t *testing.T) {
		mockCoreClient := mocks.NewMockPythonCoreClient()
		mockS3Client := mocks.NewMockS3Client()
		mockTemporalClient := mocks.NewMockTemporalClient()
		mockTemporalClient.On("SignalUploadComplete", mock.Anything, "test-doc-1").Return(assert.AnError)

		mockQdrantClient := mocks.NewMockQdrantClient()

		h := &handlers.Handlers{
			CoreClient:   mockCoreClient,
			S3Client:     mockS3Client,
			Temporal:     mockTemporalClient,
			QdrantClient: mockQdrantClient,
		}

		router := setupTestRouter()
		router.POST("/documents/:id/complete", h.CompleteUpload)

		req, _ := http.NewRequest("POST", "/documents/test-doc-1/complete", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		mockTemporalClient.AssertExpectations(t)
	})
}

func TestQueryHandler_ValidationError(t *testing.T) {
	t.Run("Query_InvalidJSON_Returns400", func(t *testing.T) {
		mockCoreClient := mocks.NewMockPythonCoreClient()
		mockS3Client := mocks.NewMockS3Client()
		mockTemporalClient := mocks.NewMockTemporalClient()
		mockQdrantClient := mocks.NewMockQdrantClient()

		h := &handlers.Handlers{
			CoreClient:   mockCoreClient,
			S3Client:     mockS3Client,
			Temporal:     mockTemporalClient,
			QdrantClient: mockQdrantClient,
		}

		router := setupTestRouter()
		router.POST("/query", h.Query)

		// Invalid JSON
		body := []byte(`{"invalid": "data"}`)
		req, _ := http.NewRequest("POST", "/query", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})
}
