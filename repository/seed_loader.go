package repository

import "context"

// SeedDataLoader defines the interface for loading seed data
type SeedDataLoader interface {
	LoadSeedData(ctx context.Context, filePath string) error
}
