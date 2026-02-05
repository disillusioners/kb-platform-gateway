package services

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"kb-platform-gateway/internal/models"
)

type PythonCoreClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewPythonCoreClient(host string, port int) *PythonCoreClient {
	return &PythonCoreClient{
		baseURL: fmt.Sprintf("http://%s:%d", host, port),
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *PythonCoreClient) Query(query string, conversationID string, topK int) (<-chan models.SSEEvent, error) {
	req := models.QueryRequest{
		Query:          query,
		ConversationID: conversationID,
		TopK:           topK,
	}

	jsonData, _ := json.Marshal(req)

	httpReq, _ := http.NewRequest("POST", c.baseURL+"/api/v1/query", bytes.NewBuffer(jsonData))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("query failed with status: %d", resp.StatusCode)
	}

	eventChan := make(chan models.SSEEvent, 100)

	go func() {
		defer resp.Body.Close()
		defer close(eventChan)

		reader := bufio.NewReader(resp.Body)
		var buffer bytes.Buffer

		for {
			line, err := reader.ReadBytes('\n')
			if err != nil && len(line) == 0 {
				if err.Error() != "EOF" {
					eventChan <- models.SSEEvent{
						Type:    "error",
						Code:    "STREAM_ERROR",
						Message: err.Error(),
					}
				}
				return
			}

			if len(line) > 0 {
				buffer.Write(line)

				if bytes.HasSuffix(line, []byte("\n\n")) {
					data := buffer.String()
					if len(data) > 6 && data[:6] == "data: " {
						jsonData := data[6:]
						var event models.SSEEvent
						if err := json.Unmarshal([]byte(jsonData), &event); err == nil {
							eventChan <- event
						}
					}
					buffer.Reset()
				}
			}
		}
	}()

	return eventChan, nil
}

func (c *PythonCoreClient) HealthCheck() (map[string]string, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/readyz")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	var readiness models.ReadinessResponse
	if err := json.NewDecoder(resp.Body).Decode(&readiness); err != nil {
		return nil, err
	}

	return readiness.Dependencies, nil
}
