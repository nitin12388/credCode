package repository

import (
	"context"

	"credCode/models"
)

// QueryRepository defines the interface for graph query operations
type QueryRepository interface {
	GetUsersWithContact(ctx context.Context, phoneNumber string) ([]string, int)
	GetOutgoingEdges(ctx context.Context, phoneNumber string, edgeType models.EdgeType) []*models.Edge
	GetIncomingEdges(ctx context.Context, phoneNumber string, edgeType models.EdgeType) []*models.Edge
	GetCallsWithFilters(ctx context.Context, phoneNumber string, filters CallFilters, direction string) ([]*models.Edge, int)
	IsDirectContact(ctx context.Context, userPhone, callerPhone string) bool
	GetSecondLevelContactCount(ctx context.Context, userPhone, callerPhone string) int
}
