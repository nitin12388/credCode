package repository

import (
	"context"

	"credCode/models"
)

// NodeRepository defines the interface for node operations
type NodeRepository interface {
	AddNode(ctx context.Context, phoneNumber string) error
	AddNodeWithName(ctx context.Context, phoneNumber, name string) error
	GetNode(ctx context.Context, phoneNumber string) (*models.Node, error)
	NodeExists(ctx context.Context, phoneNumber string) bool
	GetAllNodes(ctx context.Context) ([]*models.Node, error)
	DeleteNode(ctx context.Context, phoneNumber string) error
}
