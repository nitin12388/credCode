package repository

import (
	"context"
	"testing"
	"time"

	"credCode/models"
)

func TestNewCayleyGraphRepository(t *testing.T) {
	repo, err := NewCayleyGraphRepository()
	if err != nil {
		t.Fatalf("Unexpected error creating repository: %v", err)
	}

	if repo == nil {
		t.Fatal("Expected repository to be created, got nil")
	}

	if repo.store == nil {
		t.Error("Expected store to be initialized")
	}

	if repo.registry == nil {
		t.Error("Expected registry to be initialized")
	}
}

func TestCayleyGraphRepository_AddNode(t *testing.T) {
	repo := NewInMemoryGraphRepository()
	ctx := context.Background()

	// Test successful node creation
	err := repo.AddNode(ctx, "7379037972")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Test duplicate node
	err = repo.AddNode(ctx, "7379037972")
	if err != ErrNodeExists {
		t.Errorf("Expected ErrNodeExists, got %v", err)
	}
}

func TestCayleyGraphRepository_AddNodeWithName(t *testing.T) {
	repo := NewInMemoryGraphRepository()
	ctx := context.Background()

	// Test successful node creation with name
	err := repo.AddNodeWithName(ctx, "7379037972", "John Doe")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify node exists
	node, err := repo.GetNode(ctx, "7379037972")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if node.Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got '%s'", node.Name)
	}
}

func TestCayleyGraphRepository_GetNode(t *testing.T) {
	repo := NewInMemoryGraphRepository()
	ctx := context.Background()

	repo.AddNodeWithName(ctx, "7379037972", "John Doe")

	// Test successful retrieval
	node, err := repo.GetNode(ctx, "7379037972")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if node.PhoneNumber != "7379037972" {
		t.Errorf("Expected phone '7379037972', got '%s'", node.PhoneNumber)
	}

	// Test not found
	_, err = repo.GetNode(ctx, "9999999999")
	if err != ErrNodeNotFound {
		t.Errorf("Expected ErrNodeNotFound, got %v", err)
	}
}

func TestCayleyGraphRepository_NodeExists(t *testing.T) {
	repo := NewInMemoryGraphRepository()
	ctx := context.Background()

	repo.AddNode(ctx, "7379037972")

	// Test existing node
	if !repo.NodeExists(ctx, "7379037972") {
		t.Error("Expected node to exist")
	}

	// Test non-existing node
	if repo.NodeExists(ctx, "9999999999") {
		t.Error("Expected node to not exist")
	}
}

func TestCayleyGraphRepository_GetAllNodes(t *testing.T) {
	repo := NewInMemoryGraphRepository()
	ctx := context.Background()

	// Create multiple nodes
	nodes := []struct {
		phone string
		name  string
	}{
		{"7379037972", "John"},
		{"9876543210", "Jane"},
		{"1234567890", "Bob"},
	}

	for _, n := range nodes {
		repo.AddNodeWithName(ctx, n.phone, n.name)
	}

	allNodes, err := repo.GetAllNodes(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(allNodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(allNodes))
	}
}

func TestCayleyGraphRepository_DeleteNode(t *testing.T) {
	repo := NewInMemoryGraphRepository()
	ctx := context.Background()

	repo.AddNode(ctx, "7379037972")

	// Test successful deletion
	err := repo.DeleteNode(ctx, "7379037972")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify deletion
	if repo.NodeExists(ctx, "7379037972") {
		t.Error("Expected node to be deleted")
	}

	// Test delete non-existent node
	err = repo.DeleteNode(ctx, "9999999999")
	if err != ErrNodeNotFound {
		t.Errorf("Expected ErrNodeNotFound, got %v", err)
	}
}

func TestCayleyGraphRepository_AddEdgeWithMetadata_Contact(t *testing.T) {
	repo := NewInMemoryGraphRepository()
	ctx := context.Background()

	// Create nodes
	repo.AddNodeWithName(ctx, "7379037972", "John")
	repo.AddNodeWithName(ctx, "9876543210", "Jane")

	// Create contact metadata
	meta := &models.ContactMetadata{
		Name:    "Jane",
		AddedAt: time.Now(),
	}

	// Test successful edge creation
	edge, err := repo.AddEdgeWithMetadata(ctx, "7379037972", "9876543210", meta)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if edge.From != "7379037972" {
		t.Errorf("Expected From '7379037972', got '%s'", edge.From)
	}

	if edge.To != "9876543210" {
		t.Errorf("Expected To '9876543210', got '%s'", edge.To)
	}

	if edge.Type != models.EdgeTypeContact {
		t.Errorf("Expected EdgeTypeContact, got %s", edge.Type)
	}
}

func TestCayleyGraphRepository_AddEdgeWithMetadata_Call(t *testing.T) {
	repo := NewInMemoryGraphRepository()
	ctx := context.Background()

	// Create nodes
	repo.AddNode(ctx, "7379037972")
	repo.AddNode(ctx, "9876543210")

	// Create call metadata
	meta := &models.CallMetadata{
		IsAnswered:        true,
		DurationInSeconds: 120,
		Timestamp:         time.Now(),
	}

	// Test successful edge creation
	edge, err := repo.AddEdgeWithMetadata(ctx, "7379037972", "9876543210", meta)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if edge.Type != models.EdgeTypeCall {
		t.Errorf("Expected EdgeTypeCall, got %s", edge.Type)
	}

	// Verify metadata
	if callMeta, ok := edge.Metadata.(*models.CallMetadata); ok {
		if callMeta.DurationInSeconds != 120 {
			t.Errorf("Expected duration 120, got %d", callMeta.DurationInSeconds)
		}
	} else {
		t.Error("Expected CallMetadata type")
	}
}

func TestCayleyGraphRepository_GetEdge(t *testing.T) {
	repo := NewInMemoryGraphRepository()
	ctx := context.Background()

	repo.AddNode(ctx, "7379037972")
	repo.AddNode(ctx, "9876543210")

	meta := &models.CallMetadata{
		IsAnswered:        true,
		DurationInSeconds: 120,
		Timestamp:         time.Now(),
	}

	edge, _ := repo.AddEdgeWithMetadata(ctx, "7379037972", "9876543210", meta)

	// Test retrieval
	retrieved, err := repo.GetEdge(ctx, edge.ID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if retrieved.ID != edge.ID {
		t.Errorf("Expected ID '%s', got '%s'", edge.ID, retrieved.ID)
	}

	// Test not found
	_, err = repo.GetEdge(ctx, "nonexistent")
	if err != ErrEdgeNotFound {
		t.Errorf("Expected ErrEdgeNotFound, got %v", err)
	}
}

func TestCayleyGraphRepository_GetUsersWithContact(t *testing.T) {
	repo := NewInMemoryGraphRepository()
	ctx := context.Background()

	// Create nodes and contact edges
	phones := []string{"7379037972", "9876543210", "1234567890", "5555555555"}
	for _, phone := range phones {
		repo.AddNode(ctx, phone)
	}

	// Create contact relationships
	meta := &models.ContactMetadata{Name: "Contact", AddedAt: time.Now()}
	repo.AddEdgeWithMetadata(ctx, "9876543210", "7379037972", meta)
	repo.AddEdgeWithMetadata(ctx, "1234567890", "7379037972", meta)
	repo.AddEdgeWithMetadata(ctx, "5555555555", "7379037972", meta)

	// Test query
	users, count := repo.GetUsersWithContact(ctx, "7379037972")
	if count != 3 {
		t.Errorf("Expected 3 users, got %d", count)
	}

	if len(users) != 3 {
		t.Errorf("Expected 3 users in list, got %d", len(users))
	}
}

func TestCayleyGraphRepository_IsDirectContact(t *testing.T) {
	repo := NewInMemoryGraphRepository()
	ctx := context.Background()

	repo.AddNode(ctx, "7379037972")
	repo.AddNode(ctx, "9876543210")

	meta := &models.ContactMetadata{Name: "Contact", AddedAt: time.Now()}
	repo.AddEdgeWithMetadata(ctx, "7379037972", "9876543210", meta)

	// Test direct contact
	isDirect := repo.IsDirectContact(ctx, "7379037972", "9876543210")
	if !isDirect {
		t.Error("Expected direct contact to be true")
	}

	// Test not direct contact
	isDirect = repo.IsDirectContact(ctx, "7379037972", "9999999999")
	if isDirect {
		t.Error("Expected direct contact to be false")
	}
}

func TestCayleyGraphRepository_GetSecondLevelContactCount(t *testing.T) {
	repo := NewInMemoryGraphRepository()
	ctx := context.Background()

	// Create nodes
	phones := []string{"7379037972", "9876543210", "1234567890", "5555555555"}
	for _, phone := range phones {
		repo.AddNode(ctx, phone)
	}

	// Create level 1: 7379037972 -> 9876543210
	meta1 := &models.ContactMetadata{Name: "Contact1", AddedAt: time.Now()}
	repo.AddEdgeWithMetadata(ctx, "7379037972", "9876543210", meta1)

	// Create level 2: 9876543210 -> 1234567890 (so 1234567890 is level-2 from 7379037972)
	meta2 := &models.ContactMetadata{Name: "Contact2", AddedAt: time.Now()}
	repo.AddEdgeWithMetadata(ctx, "9876543210", "1234567890", meta2)

	// Test second level count
	count := repo.GetSecondLevelContactCount(ctx, "7379037972", "1234567890")
	if count != 1 {
		t.Errorf("Expected 1 level-2 contact, got %d", count)
	}

	// Test no level-2 contact
	count = repo.GetSecondLevelContactCount(ctx, "7379037972", "9999999999")
	if count != 0 {
		t.Errorf("Expected 0 level-2 contacts, got %d", count)
	}
}

func TestCayleyGraphRepository_GetCallsWithFilters(t *testing.T) {
	repo := NewInMemoryGraphRepository()
	ctx := context.Background()

	repo.AddNode(ctx, "7379037972")
	repo.AddNode(ctx, "9876543210")

	// Create call edges
	now := time.Now()
	call1 := &models.CallMetadata{
		IsAnswered:        true,
		DurationInSeconds: 120,
		Timestamp:         now.Add(-2 * time.Hour),
	}
	call2 := &models.CallMetadata{
		IsAnswered:        false,
		DurationInSeconds: 15,
		Timestamp:         now.Add(-1 * time.Hour),
	}

	repo.AddEdgeWithMetadata(ctx, "7379037972", "9876543210", call1)
	repo.AddEdgeWithMetadata(ctx, "7379037972", "9876543210", call2)

	// Test filter: answered calls only
	answered := true
	filters := CallFilters{IsAnswered: &answered}
	calls, count := repo.GetCallsWithFilters(ctx, "7379037972", filters, "outgoing")

	if count != 1 {
		t.Errorf("Expected 1 answered call, got %d", count)
	}

	if len(calls) != 1 {
		t.Errorf("Expected 1 call in list, got %d", len(calls))
	}

	// Test filter: short duration
	maxDuration := 20
	filters = CallFilters{MaxDuration: &maxDuration}
	calls, count = repo.GetCallsWithFilters(ctx, "7379037972", filters, "outgoing")

	if count != 1 {
		t.Errorf("Expected 1 short call, got %d", count)
	}
}

func TestCayleyGraphRepository_GetOutgoingEdges(t *testing.T) {
	repo := NewInMemoryGraphRepository()
	ctx := context.Background()

	repo.AddNode(ctx, "7379037972")
	repo.AddNode(ctx, "9876543210")

	meta := &models.ContactMetadata{Name: "Contact", AddedAt: time.Now()}
	repo.AddEdgeWithMetadata(ctx, "7379037972", "9876543210", meta)

	// Test outgoing contact edges
	edges := repo.GetOutgoingEdges(ctx, "7379037972", models.EdgeTypeContact)
	if len(edges) != 1 {
		t.Errorf("Expected 1 outgoing edge, got %d", len(edges))
	}

	if edges[0].To != "9876543210" {
		t.Errorf("Expected To '9876543210', got '%s'", edges[0].To)
	}
}

func TestCayleyGraphRepository_GetIncomingEdges(t *testing.T) {
	repo := NewInMemoryGraphRepository()
	ctx := context.Background()

	repo.AddNode(ctx, "7379037972")
	repo.AddNode(ctx, "9876543210")

	meta := &models.ContactMetadata{Name: "Contact", AddedAt: time.Now()}
	repo.AddEdgeWithMetadata(ctx, "7379037972", "9876543210", meta)

	// Test incoming contact edges
	edges := repo.GetIncomingEdges(ctx, "9876543210", models.EdgeTypeContact)
	if len(edges) != 1 {
		t.Errorf("Expected 1 incoming edge, got %d", len(edges))
	}

	if edges[0].From != "7379037972" {
		t.Errorf("Expected From '7379037972', got '%s'", edges[0].From)
	}
}

