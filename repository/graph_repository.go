package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"credCode/models"

	"github.com/cayleygraph/cayley"
	_ "github.com/cayleygraph/cayley/graph/memstore"
	"github.com/cayleygraph/quad"
)

var (
	ErrNodeNotFound    = errors.New("node not found")
	ErrEdgeNotFound    = errors.New("edge not found")
	ErrNodeExists      = errors.New("node already exists")
	ErrInvalidEdgeType = errors.New("invalid edge type")
)

// CallFilters defines filters for querying call edges
type CallFilters struct {
	IsAnswered     *bool      // filter by answered/unanswered
	MaxDuration    *int       // maximum duration in seconds
	MinDuration    *int       // minimum duration in seconds
	TimeRangeStart *time.Time // start of time range
	TimeRangeEnd   *time.Time // end of time range
}

// GraphRepository defines the interface for graph operations
type GraphRepository interface {
	// Node operations
	AddNode(phoneNumber string) error
	GetNode(phoneNumber string) (*models.Node, error)
	NodeExists(phoneNumber string) bool
	GetAllNodes() ([]*models.Node, error)
	DeleteNode(phoneNumber string) error

	// Edge operations
	AddContactEdge(phone1, phone2 string) error
	AddCallEdge(from, to string, isAnswered bool, duration int, timestamp time.Time) (*models.Edge, error)
	GetEdge(edgeID string) (*models.Edge, error)
	DeleteEdge(edgeID string) error

	// Query operations
	GetUsersWithContact(phoneNumber string) ([]string, int)
	GetOutgoingEdges(phoneNumber string, edgeType models.EdgeType) []*models.Edge
	GetIncomingEdges(phoneNumber string, edgeType models.EdgeType) []*models.Edge
	GetCallsWithFilters(phoneNumber string, filters CallFilters, direction string) ([]*models.Edge, int)

	// Seed data operations
	LoadSeedData(filePath string) error
}

// CayleyGraphRepository implements GraphRepository using Cayley
type CayleyGraphRepository struct {
	store       *cayley.Handle
	edgeCounter int
	mu          sync.RWMutex
}

// NewCayleyGraphRepository creates a new Cayley-based graph repository
func NewCayleyGraphRepository() (*CayleyGraphRepository, error) {
	// Initialize Cayley in-memory store
	store, err := cayley.NewMemoryGraph()
	if err != nil {
		return nil, fmt.Errorf("failed to create memory graph: %w", err)
	}

	return &CayleyGraphRepository{
		store:       store,
		edgeCounter: 0,
	}, nil
}

// NewInMemoryGraphRepository creates a new in-memory graph repository (alias for backwards compatibility)
func NewInMemoryGraphRepository() *CayleyGraphRepository {
	repo, err := NewCayleyGraphRepository()
	if err != nil {
		panic(fmt.Sprintf("failed to create graph repository: %v", err))
	}
	return repo
}

// AddNode adds a new node to the graph
func (r *CayleyGraphRepository) AddNode(phoneNumber string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if node already exists
	if r.nodeExistsUnsafe(phoneNumber) {
		return ErrNodeExists
	}

	// Add node as a quad: phoneNumber -> type -> "node"
	quad := quad.Make(phoneNumber, "type", "node", nil)
	if err := r.store.AddQuad(quad); err != nil {
		return fmt.Errorf("failed to add node: %w", err)
	}

	return nil
}

// GetNode retrieves a node by phone number
func (r *CayleyGraphRepository) GetNode(phoneNumber string) (*models.Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.nodeExistsUnsafe(phoneNumber) {
		return nil, ErrNodeNotFound
	}

	return &models.Node{
		PhoneNumber: phoneNumber,
	}, nil
}

// NodeExists checks if a node exists
func (r *CayleyGraphRepository) NodeExists(phoneNumber string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.nodeExistsUnsafe(phoneNumber)
}

// nodeExistsUnsafe checks if a node exists (must be called with lock held)
func (r *CayleyGraphRepository) nodeExistsUnsafe(phoneNumber string) bool {
	p := cayley.StartPath(r.store, quad.String(phoneNumber)).Out(quad.String("type"))
	ctx := context.TODO()

	it, _ := p.BuildIterator().Optimize()
	defer it.Close()

	return it.Next(ctx)
}

// GetAllNodes retrieves all nodes
func (r *CayleyGraphRepository) GetAllNodes() ([]*models.Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make([]*models.Node, 0)
	ctx := context.TODO()

	// Find all subjects that have type "node"
	p := cayley.StartPath(r.store).Has(quad.String("type"), quad.String("node"))

	it, _ := p.BuildIterator().Optimize()
	defer it.Close()

	for it.Next(ctx) {
		token := it.Result()
		phoneNumber := quad.ToString(r.store.NameOf(token))
		nodes = append(nodes, &models.Node{
			PhoneNumber: phoneNumber,
		})
	}

	if err := it.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate nodes: %w", err)
	}

	return nodes, nil
}

// DeleteNode removes a node and all its edges
func (r *CayleyGraphRepository) DeleteNode(phoneNumber string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.nodeExistsUnsafe(phoneNumber) {
		return ErrNodeNotFound
	}

	ctx := context.TODO()

	// Delete all quads where this phone number is subject or object
	// This includes the node itself and all edges
	quadsToDelete := make([]*quad.Quad, 0)

	// Find quads where phoneNumber is subject
	it := r.store.QuadsAllIterator()
	defer it.Close()

	for it.Next(ctx) {
		q := r.store.Quad(it.Result())
		subject := quad.ToString(q.Subject)
		object := quad.ToString(q.Object)

		if subject == phoneNumber || object == phoneNumber {
			quadsToDelete = append(quadsToDelete, &q)
		}
	}

	// Delete all found quads
	for _, q := range quadsToDelete {
		if err := r.store.RemoveQuad(*q); err != nil {
			return fmt.Errorf("failed to remove quad: %w", err)
		}
	}

	return nil
}

// AddContactEdge adds a bidirectional contact edge between two phone numbers
func (r *CayleyGraphRepository) AddContactEdge(phone1, phone2 string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Ensure both nodes exist
	if !r.nodeExistsUnsafe(phone1) {
		r.store.AddQuad(quad.Make(phone1, "type", "node", nil))
	}
	if !r.nodeExistsUnsafe(phone2) {
		r.store.AddQuad(quad.Make(phone2, "type", "node", nil))
	}

	// Create bidirectional contact edges
	// phone1 -has_contact-> phone2
	r.store.AddQuad(quad.Make(phone1, "has_contact", phone2, nil))
	// phone2 -has_contact-> phone1
	r.store.AddQuad(quad.Make(phone2, "has_contact", phone1, nil))

	return nil
}

// AddCallEdge adds a directional call edge from one phone to another
func (r *CayleyGraphRepository) AddCallEdge(from, to string, isAnswered bool, duration int, timestamp time.Time) (*models.Edge, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Ensure both nodes exist
	if !r.nodeExistsUnsafe(from) {
		r.store.AddQuad(quad.Make(from, "type", "node", nil))
	}
	if !r.nodeExistsUnsafe(to) {
		r.store.AddQuad(quad.Make(to, "type", "node", nil))
	}

	// Generate unique call ID
	r.edgeCounter++
	callID := fmt.Sprintf("call_%d", r.edgeCounter)

	// Store call as multiple quads (RDF-style)
	r.store.AddQuad(quad.Make(callID, "type", "call", nil))
	r.store.AddQuad(quad.Make(callID, "from", from, nil))
	r.store.AddQuad(quad.Make(callID, "to", to, nil))
	r.store.AddQuad(quad.Make(callID, "is_answered", isAnswered, nil))
	r.store.AddQuad(quad.Make(callID, "duration", duration, nil))
	r.store.AddQuad(quad.Make(callID, "created_at", timestamp.Format(time.RFC3339), nil))

	// Create edge object to return
	callProps := &models.CallProperties{
		IsAnswered:        isAnswered,
		DurationInSeconds: duration,
	}

	edge := &models.Edge{
		ID:         callID,
		From:       from,
		To:         to,
		Type:       models.EdgeTypeCall,
		Properties: callProps.ToMap(),
		CreatedAt:  timestamp,
	}

	return edge, nil
}

// GetEdge retrieves an edge by ID
func (r *CayleyGraphRepository) GetEdge(edgeID string) (*models.Edge, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ctx := context.TODO()

	// Check if edge exists and get its type
	typePath := cayley.StartPath(r.store, quad.String(edgeID)).Out(quad.String("type"))
	typeIt, _ := typePath.BuildIterator().Optimize()
	defer typeIt.Close()

	if !typeIt.Next(ctx) {
		return nil, ErrEdgeNotFound
	}

	token := typeIt.Result()
	edgeType := quad.ToString(r.store.NameOf(token))

	if edgeType == "call" {
		return r.getCallEdgeUnsafe(edgeID)
	}

	return nil, ErrEdgeNotFound
}

// getCallEdgeUnsafe retrieves a call edge (must be called with lock held)
func (r *CayleyGraphRepository) getCallEdgeUnsafe(callID string) (*models.Edge, error) {
	ctx := context.TODO()

	edge := &models.Edge{
		ID:         callID,
		Type:       models.EdgeTypeCall,
		Properties: make(map[string]interface{}),
	}

	// Get "from" phone
	fromPath := cayley.StartPath(r.store, quad.String(callID)).Out(quad.String("from"))
	fromIt, _ := fromPath.BuildIterator().Optimize()
	if fromIt.Next(ctx) {
		token := fromIt.Result()
		edge.From = quad.ToString(r.store.NameOf(token))
	}
	fromIt.Close()

	// Get "to" phone
	toPath := cayley.StartPath(r.store, quad.String(callID)).Out(quad.String("to"))
	toIt, _ := toPath.BuildIterator().Optimize()
	if toIt.Next(ctx) {
		token := toIt.Result()
		edge.To = quad.ToString(r.store.NameOf(token))
	}
	toIt.Close()

	// Get "is_answered"
	ansPath := cayley.StartPath(r.store, quad.String(callID)).Out(quad.String("is_answered"))
	ansIt, _ := ansPath.BuildIterator().Optimize()
	if ansIt.Next(ctx) {
		token := ansIt.Result()
		val := r.store.NameOf(token)
		nativeVal := quad.NativeOf(val)
		// Handle bool conversion
		if boolVal, ok := nativeVal.(bool); ok {
			edge.Properties["is_answered"] = boolVal
		}
	}
	ansIt.Close()

	// Get "duration"
	durPath := cayley.StartPath(r.store, quad.String(callID)).Out(quad.String("duration"))
	durIt, _ := durPath.BuildIterator().Optimize()
	if durIt.Next(ctx) {
		token := durIt.Result()
		val := r.store.NameOf(token)
		nativeVal := quad.NativeOf(val)
		// Handle int conversion
		if intVal, ok := nativeVal.(int); ok {
			edge.Properties["duration_in_seconds"] = intVal
		} else if int64Val, ok := nativeVal.(int64); ok {
			edge.Properties["duration_in_seconds"] = int(int64Val)
		}
	}
	durIt.Close()

	// Get "created_at"
	timePath := cayley.StartPath(r.store, quad.String(callID)).Out(quad.String("created_at"))
	timeIt, _ := timePath.BuildIterator().Optimize()
	if timeIt.Next(ctx) {
		token := timeIt.Result()
		timeStr := quad.ToString(r.store.NameOf(token))
		if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
			edge.CreatedAt = t
		}
	}
	timeIt.Close()

	return edge, nil
}

// DeleteEdge removes an edge from the graph
func (r *CayleyGraphRepository) DeleteEdge(edgeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ctx := context.TODO()

	// Find and delete all quads related to this edge
	it := r.store.QuadsAllIterator()
	defer it.Close()

	quadsToDelete := make([]*quad.Quad, 0)

	for it.Next(ctx) {
		q := r.store.Quad(it.Result())
		subject := quad.ToString(q.Subject)

		if subject == edgeID {
			quadsToDelete = append(quadsToDelete, &q)
		}
	}

	if len(quadsToDelete) == 0 {
		return ErrEdgeNotFound
	}

	for _, q := range quadsToDelete {
		if err := r.store.RemoveQuad(*q); err != nil {
			return fmt.Errorf("failed to remove quad: %w", err)
		}
	}

	return nil
}

// GetUsersWithContact returns all phone numbers that have the given phone number in their contacts
// Query 1: Give me count or all the users who have saved a phone number in their contact list
func (r *CayleyGraphRepository) GetUsersWithContact(phoneNumber string) ([]string, int) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ctx := context.TODO()
	users := make([]string, 0)

	// Find all phone numbers that have has_contact edge pointing to this phoneNumber
	// This is: ? -has_contact-> phoneNumber
	p := cayley.StartPath(r.store, quad.String(phoneNumber)).In(quad.String("has_contact"))

	it, _ := p.BuildIterator().Optimize()
	defer it.Close()

	for it.Next(ctx) {
		token := it.Result()
		user := quad.ToString(r.store.NameOf(token))
		users = append(users, user)
	}

	return users, len(users)
}

// GetOutgoingEdges returns all outgoing edges of a specific type from a phone number
func (r *CayleyGraphRepository) GetOutgoingEdges(phoneNumber string, edgeType models.EdgeType) []*models.Edge {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ctx := context.TODO()
	edges := make([]*models.Edge, 0)

	if edgeType == models.EdgeTypeContact {
		// Find all: phoneNumber -has_contact-> ?
		p := cayley.StartPath(r.store, quad.String(phoneNumber)).Out(quad.String("has_contact"))
		it, _ := p.BuildIterator().Optimize()
		defer it.Close()

		for it.Next(ctx) {
			token := it.Result()
			toPhone := quad.ToString(r.store.NameOf(token))
			edges = append(edges, &models.Edge{
				From:       phoneNumber,
				To:         toPhone,
				Type:       models.EdgeTypeContact,
				Properties: make(map[string]interface{}),
			})
		}
	} else if edgeType == models.EdgeTypeCall {
		// Find all calls where from = phoneNumber
		edges = r.getCallsByPhoneUnsafe(phoneNumber, "from")
	}

	return edges
}

// GetIncomingEdges returns all incoming edges of a specific type to a phone number
func (r *CayleyGraphRepository) GetIncomingEdges(phoneNumber string, edgeType models.EdgeType) []*models.Edge {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ctx := context.TODO()
	edges := make([]*models.Edge, 0)

	if edgeType == models.EdgeTypeContact {
		// Find all: ? -has_contact-> phoneNumber
		p := cayley.StartPath(r.store, quad.String(phoneNumber)).In(quad.String("has_contact"))
		it, _ := p.BuildIterator().Optimize()
		defer it.Close()

		for it.Next(ctx) {
			token := it.Result()
			fromPhone := quad.ToString(r.store.NameOf(token))
			edges = append(edges, &models.Edge{
				From:       fromPhone,
				To:         phoneNumber,
				Type:       models.EdgeTypeContact,
				Properties: make(map[string]interface{}),
			})
		}
	} else if edgeType == models.EdgeTypeCall {
		// Find all calls where to = phoneNumber
		edges = r.getCallsByPhoneUnsafe(phoneNumber, "to")
	}

	return edges
}

// getCallsByPhoneUnsafe retrieves call edges by phone number (must be called with lock held)
func (r *CayleyGraphRepository) getCallsByPhoneUnsafe(phoneNumber, direction string) []*models.Edge {
	ctx := context.TODO()
	edges := make([]*models.Edge, 0)

	// Find all call IDs where direction = phoneNumber
	p := cayley.StartPath(r.store, quad.String(phoneNumber)).In(quad.String(direction)).Has(quad.String("type"), quad.String("call"))

	it, _ := p.BuildIterator().Optimize()
	defer it.Close()

	for it.Next(ctx) {
		token := it.Result()
		callID := quad.ToString(r.store.NameOf(token))
		if edge, err := r.getCallEdgeUnsafe(callID); err == nil {
			edges = append(edges, edge)
		}
	}

	return edges
}

// GetCallsWithFilters returns call edges with applied filters
// Query 2: How many calls a phone number is making with filters
// direction: "outgoing", "incoming", or "both"
func (r *CayleyGraphRepository) GetCallsWithFilters(phoneNumber string, filters CallFilters, direction string) ([]*models.Edge, int) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var candidateEdges []*models.Edge

	// Gather candidate edges based on direction
	switch direction {
	case "outgoing":
		candidateEdges = r.getCallsByPhoneUnsafe(phoneNumber, "from")
	case "incoming":
		candidateEdges = r.getCallsByPhoneUnsafe(phoneNumber, "to")
	case "both":
		candidateEdges = append(
			r.getCallsByPhoneUnsafe(phoneNumber, "from"),
			r.getCallsByPhoneUnsafe(phoneNumber, "to")...,
		)
	default:
		candidateEdges = r.getCallsByPhoneUnsafe(phoneNumber, "from")
	}

	// Apply filters
	filteredEdges := make([]*models.Edge, 0)
	for _, edge := range candidateEdges {
		if r.matchesCallFilters(edge, filters) {
			filteredEdges = append(filteredEdges, edge)
		}
	}

	return filteredEdges, len(filteredEdges)
}

// matchesCallFilters checks if an edge matches the given call filters
func (r *CayleyGraphRepository) matchesCallFilters(edge *models.Edge, filters CallFilters) bool {
	if edge.Type != models.EdgeTypeCall {
		return false
	}

	callProps := models.ParseCallProperties(edge.Properties)

	// Filter by IsAnswered
	if filters.IsAnswered != nil && callProps.IsAnswered != *filters.IsAnswered {
		return false
	}

	// Filter by MaxDuration
	if filters.MaxDuration != nil && callProps.DurationInSeconds > *filters.MaxDuration {
		return false
	}

	// Filter by MinDuration
	if filters.MinDuration != nil && callProps.DurationInSeconds < *filters.MinDuration {
		return false
	}

	// Filter by TimeRangeStart
	if filters.TimeRangeStart != nil && edge.CreatedAt.Before(*filters.TimeRangeStart) {
		return false
	}

	// Filter by TimeRangeEnd
	if filters.TimeRangeEnd != nil && edge.CreatedAt.After(*filters.TimeRangeEnd) {
		return false
	}

	return true
}

// LoadSeedData loads graph seed data from a JSON file
func (r *CayleyGraphRepository) LoadSeedData(filePath string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Read the JSON file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Parse the JSON
	var seedData struct {
		Nodes []*models.Node `json:"nodes"`
		Edges []*models.Edge `json:"edges"`
	}

	if err := json.Unmarshal(data, &seedData); err != nil {
		return err
	}

	// Load nodes
	for _, node := range seedData.Nodes {
		if !r.nodeExistsUnsafe(node.PhoneNumber) {
			r.store.AddQuad(quad.Make(node.PhoneNumber, "type", "node", nil))
		}
	}

	// Load edges
	for _, edge := range seedData.Edges {
		// Ensure nodes exist
		if !r.nodeExistsUnsafe(edge.From) {
			r.store.AddQuad(quad.Make(edge.From, "type", "node", nil))
		}
		if !r.nodeExistsUnsafe(edge.To) {
			r.store.AddQuad(quad.Make(edge.To, "type", "node", nil))
		}

		if edge.Type == models.EdgeTypeContact {
			// Add contact edge (unidirectional as per seed data)
			r.store.AddQuad(quad.Make(edge.From, "has_contact", edge.To, nil))
		} else if edge.Type == models.EdgeTypeCall {
			// Add call edge with all properties
			r.store.AddQuad(quad.Make(edge.ID, "type", "call", nil))
			r.store.AddQuad(quad.Make(edge.ID, "from", edge.From, nil))
			r.store.AddQuad(quad.Make(edge.ID, "to", edge.To, nil))

			if isAnswered, ok := edge.Properties["is_answered"].(bool); ok {
				r.store.AddQuad(quad.Make(edge.ID, "is_answered", isAnswered, nil))
			}
			if duration, ok := edge.Properties["duration_in_seconds"].(float64); ok {
				r.store.AddQuad(quad.Make(edge.ID, "duration", int(duration), nil))
			} else if duration, ok := edge.Properties["duration_in_seconds"].(int); ok {
				r.store.AddQuad(quad.Make(edge.ID, "duration", duration, nil))
			}
			if !edge.CreatedAt.IsZero() {
				r.store.AddQuad(quad.Make(edge.ID, "created_at", edge.CreatedAt.Format(time.RFC3339), nil))
			}
		}
	}

	return nil
}
