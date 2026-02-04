package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/disillusioners/kb-platform-proto/gen/go/kbplatform/v1"
)

// GrpcCoreClient is a gRPC client for the Python Core service
type GrpcCoreClient struct {
	conn   *grpc.ClientConn
	client pb.KBPlatformServiceClient
}

// NewGrpcCoreClient creates a new gRPC client
func NewGrpcCoreClient(host string, port int) (*GrpcCoreClient, error) {
	addr := fmt.Sprintf("%s:%d", host, port)

	// Use insecure credentials for local development
	// In production, use secure credentials
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(10*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	return &GrpcCoreClient{
		conn:   conn,
		client: pb.NewKBPlatformServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection
func (c *GrpcCoreClient) Close() error {
	return c.conn.Close()
}

// QueryStream performs a streaming RAG query
func (c *GrpcCoreClient) QueryStream(ctx context.Context, query string, conversationID string, topK int) (<-chan *pb.QueryResponse, error) {
	req := &pb.QueryRequest{
		Query:          query,
		ConversationId: conversationID,
		TopK:           int32(topK),
	}

	stream, err := c.client.QueryStream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to start query stream: %w", err)
	}

	responseChan := make(chan *pb.QueryResponse, 100)

	go func() {
		defer close(responseChan)
		defer stream.CloseSend()

		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Printf("Error receiving from stream: %v", err)
				return
			}
			responseChan <- resp
		}
	}()

	return responseChan, nil
}

// GetDocument retrieves a document by ID
func (c *GrpcCoreClient) GetDocument(ctx context.Context, documentID string) (*pb.Document, error) {
	req := &pb.GetDocumentRequest{
		DocumentId: documentID,
	}

	resp, err := c.client.GetDocument(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return resp, nil
}

// DeleteDocumentVectors deletes document vectors from Qdrant
func (c *GrpcCoreClient) DeleteDocumentVectors(ctx context.Context, documentID string) error {
	req := &pb.DeleteDocumentVectorsRequest{
		DocumentId: documentID,
	}

	_, err := c.client.DeleteDocumentVectors(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete document vectors: %w", err)
	}

	return nil
}

// GetConversation retrieves a conversation by ID
func (c *GrpcCoreClient) GetConversation(ctx context.Context, conversationID string) (*pb.Conversation, error) {
	req := &pb.GetConversationRequest{
		ConversationId: conversationID,
	}

	resp, err := c.client.GetConversation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	return resp, nil
}

// GetConversationMessages retrieves messages for a conversation
func (c *GrpcCoreClient) GetConversationMessages(ctx context.Context, conversationID string) ([]*pb.Message, error) {
	req := &pb.GetConversationMessagesRequest{
		ConversationId: conversationID,
	}

	resp, err := c.client.GetConversationMessages(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation messages: %w", err)
	}

	return resp.Messages, nil
}

// SaveMessage saves a message to a conversation
func (c *GrpcCoreClient) SaveMessage(ctx context.Context, conversationID string, role string, content string, metadata map[string]string) (*pb.Message, error) {
	req := &pb.SaveMessageRequest{
		ConversationId: conversationID,
		Role:           role,
		Content:        content,
		Metadata:       metadata,
	}

	resp, err := c.client.SaveMessage(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	return resp, nil
}

// HealthCheck performs a health check on the Python Core service
func (c *GrpcCoreClient) HealthCheck(ctx context.Context) error {
	// Create a timeout context for health check
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Try to get document metadata with an empty ID to check connectivity
	// This will fail with a not found error if the service is running
	md := metadata.New(map[string]string{
		"health-check": "true",
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	_, err := c.client.GetDocument(ctx, &pb.GetDocumentRequest{DocumentId: "health-check"})
	if err != nil {
		// Not found is expected for a health check - means service is running
		if contains(err.Error(), "not found") || contains(err.Error(), "health-check") {
			return nil
		}
		return fmt.Errorf("health check failed: %w", err)
	}

	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
