package repository

import (
	"context"

	"credCode/models"
)

// EdgeRepository defines the interface for edge operations
type EdgeRepository interface {
	AddEdgeWithMetadata(ctx context.Context, from, to string, metadata models.EdgeMetadata) (*models.Edge, error)
	GetEdge(ctx context.Context, edgeID string) (*models.Edge, error)
	GetEdgeWithMetadata(ctx context.Context, edgeID string) (*models.Edge, models.EdgeMetadata, error)
	DeleteEdge(ctx context.Context, edgeID string) error
}
