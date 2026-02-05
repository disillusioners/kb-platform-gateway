package services

import (
	"context"
	"fmt"

	"kb-platform-gateway/internal/config"

	pb "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type QdrantClient struct {
	pointsClient pb.PointsClient
	collection   string
	conn         *grpc.ClientConn
}

func NewQdrantClient(cfg *config.QdrantConfig) (*QdrantClient, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to qdrant: %w", err)
	}

	return &QdrantClient{
		pointsClient: pb.NewPointsClient(conn),
		collection:   cfg.Collection,
		conn:         conn,
	}, nil
}

func (q *QdrantClient) Close() error {
	return q.conn.Close()
}

func (q *QdrantClient) DeleteDocumentVectors(ctx context.Context, documentID string) error {
	// Create filter for document_id
	filter := &pb.Filter{
		Must: []*pb.Condition{
			{
				Condition: &pb.Condition_Field{
					Field: &pb.FieldCondition{
						Key: "document_id",
						Match: &pb.Match{
							MatchValue: &pb.Match_Keyword{
								Keyword: documentID,
							},
						},
					},
				},
			},
		},
	}

	// Delete points matching the filter
	_, err := q.pointsClient.Delete(ctx, &pb.DeletePoints{
		CollectionName: q.collection,
		Points: &pb.PointsSelector{
			PointsSelectorOneOf: &pb.PointsSelector_Filter{
				Filter: filter,
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to delete vectors for document %s: %w", documentID, err)
	}

	return nil
}
