package service

import (
	"context"
	"log"

	"credCode/models"
	"credCode/repository"
)

// GraphBuilder defines the interface for building graphs from user data
type GraphBuilder interface {
	BuildFromUsers(userRepo repository.UserRepository, graphRepo repository.GraphRepository) error
}

// graphBuilder implements GraphBuilder
type graphBuilder struct{}

// NewGraphBuilder creates a new graph builder service
func NewGraphBuilder() GraphBuilder {
	return &graphBuilder{}
}

// BuildFromUsers constructs the graph from user repository data
func (gb *graphBuilder) BuildFromUsers(userRepo repository.UserRepository, graphRepo repository.GraphRepository) error {
	ctx := context.Background()

	// Get all users
	users, err := userRepo.GetAllUsers(ctx)
	if err != nil {
		return err
	}

	// Add nodes to graph with names
	nodeCount := 0
	for _, user := range users {
		if err := graphRepo.AddNodeWithName(ctx, user.PhoneNumber, user.Name); err != nil {
			// Node might already exist, continue
			log.Printf("Note: Node %s already exists or error: %v", user.PhoneNumber, err)
		} else {
			nodeCount++
		}
	}
	log.Printf("  Added %d nodes to graph", nodeCount)

	// Add contact edges from user contacts
	contactCount := 0
	for _, user := range users {
		contacts, err := userRepo.GetUserContacts(ctx, user.ID)
		if err != nil {
			continue
		}

		for _, contact := range contacts {
			// Add contact edge with metadata
			contactMeta := &models.ContactMetadata{
				Name:    contact.Name,
				AddedAt: contact.AddedAt,
			}

			// Add bidirectional contact edge
			_, err := graphRepo.AddEdgeWithMetadata(ctx, user.PhoneNumber, contact.PhoneNumber, contactMeta)
			if err != nil {
				log.Printf("Error adding contact edge: %v", err)
				continue
			}
			contactCount++
		}
	}
	log.Printf("  Added %d contact edges to graph", contactCount)

	return nil
}
