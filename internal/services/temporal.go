package services

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"kb-platform-gateway/internal/config"
)

type TemporalClient struct {
	client client.Client
	cfg    *config.TemporalConfig
}

func NewTemporalClient(cfg *config.TemporalConfig) (*TemporalClient, error) {
	c, err := client.Dial(client.Options{
		HostPort:  fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Namespace: cfg.Namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create temporal client: %w", err)
	}

	return &TemporalClient{
		client: c,
		cfg:    cfg,
	}, nil
}

func (tc *TemporalClient) Close() {
	tc.client.Close()
}

type UploadWorkflowInput struct {
	DocumentID string
	S3Key      string
}

type IndexWorkflowInput struct {
	DocumentID string
}

type QueryWorkflowInput struct {
	Query          string
	ConversationID string
	TopK           int
}

func (tc *TemporalClient) StartUploadWorkflow(ctx context.Context, documentID, s3Key string) (string, error) {
	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("upload-%s", documentID),
		TaskQueue: "upload-task-queue",
	}

	we, err := tc.client.ExecuteWorkflow(ctx, workflowOptions, "UploadWorkflow", UploadWorkflowInput{
		DocumentID: documentID,
		S3Key:      s3Key,
	})
	if err != nil {
		return "", fmt.Errorf("failed to start upload workflow: %w", err)
	}

	return we.GetID(), nil
}

func (tc *TemporalClient) SignalUploadComplete(ctx context.Context, documentID string) error {
	return tc.client.SignalWorkflow(ctx, fmt.Sprintf("upload-%s", documentID), "", "upload-complete", nil)
}

func (tc *TemporalClient) StartIndexWorkflow(ctx context.Context, documentID string) (string, error) {
	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("index-%s", documentID),
		TaskQueue: "index-task-queue",
	}

	we, err := tc.client.ExecuteWorkflow(ctx, workflowOptions, "IndexWorkflow", IndexWorkflowInput{
		DocumentID: documentID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to start index workflow: %w", err)
	}

	return we.GetID(), nil
}

func (tc *TemporalClient) QueryWorkflowStatus(ctx context.Context, workflowID string) (*workflowservice.DescribeWorkflowExecutionResponse, error) {
	return tc.client.DescribeWorkflowExecution(ctx, workflowID, "")
}

func (tc *TemporalClient) CancelWorkflow(ctx context.Context, workflowID string) error {
	return tc.client.CancelWorkflow(ctx, workflowID, "")
}

func (tc *TemporalClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := tc.client.WorkflowService().GetSystemInfo(ctx, &workflowservice.GetSystemInfoRequest{})
	return err
}
