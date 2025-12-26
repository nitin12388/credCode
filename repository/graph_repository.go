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
// It embeds smaller interfaces for backward compatibility while allowing clients
// to depend only on the interfaces they need (Interface Segregation Principle)
type GraphRepository interface {
	// Embed smaller interfaces
	NodeRepository
	EdgeRepository
	QueryRepository
	SeedDataLoader
}

// CayleyGraphRepository implements GraphRepository using Cayley
type CayleyGraphRepository struct {
	store       *cayley.Handle
	registry    *models.EdgeMetadataRegistry
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
		registry:    models.NewEdgeMetadataRegistry(),
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

// AddNode adds a new node to the graph (backward compatible - without name)
func (r *CayleyGraphRepository) AddNode(ctx context.Context, phoneNumber string) error {
	return r.AddNodeWithName(ctx, phoneNumber, "")
}

// AddNodeWithName adds a new node with name to the graph
func (r *CayleyGraphRepository) AddNodeWithName(ctx context.Context, phoneNumber, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if node already exists
	if r.nodeExistsUnsafe(phoneNumber) {
		return ErrNodeExists
	}

	// Add node as a quad: phoneNumber -> type -> "node"
	r.store.AddQuad(quad.Make(phoneNumber, "type", "node", nil))

	// Add name if provided
	if name != "" {
		r.store.AddQuad(quad.Make(phoneNumber, "name", name, nil))
	}

	return nil
}

// GetNode retrieves a node by phone number
func (r *CayleyGraphRepository) GetNode(ctx context.Context, phoneNumber string) (*models.Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.nodeExistsUnsafe(phoneNumber) {
		return nil, ErrNodeNotFound
	}

	node := &models.Node{
		PhoneNumber: phoneNumber,
	}

	// Get name if it exists
	namePath := cayley.StartPath(r.store, quad.String(phoneNumber)).Out(quad.String("name"))
	nameIt, _ := namePath.BuildIterator().Optimize()
	if nameIt.Next(ctx) {
		token := nameIt.Result()
		node.Name = quad.ToString(r.store.NameOf(token))
	}
	nameIt.Close()

	return node, nil
}

// NodeExists checks if a node exists
func (r *CayleyGraphRepository) NodeExists(ctx context.Context, phoneNumber string) bool {
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
func (r *CayleyGraphRepository) GetAllNodes(ctx context.Context) ([]*models.Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make([]*models.Node, 0)

	// Find all subjects that have type "node"
	p := cayley.StartPath(r.store).Has(quad.String("type"), quad.String("node"))

	it, _ := p.BuildIterator().Optimize()
	defer it.Close()

	for it.Next(ctx) {
		token := it.Result()
		phoneNumber := quad.ToString(r.store.NameOf(token))

		node := &models.Node{
			PhoneNumber: phoneNumber,
		}

		// Get name if it exists
		namePath := cayley.StartPath(r.store, quad.String(phoneNumber)).Out(quad.String("name"))
		nameIt, _ := namePath.BuildIterator().Optimize()
		if nameIt.Next(ctx) {
			nameToken := nameIt.Result()
			node.Name = quad.ToString(r.store.NameOf(nameToken))
		}
		nameIt.Close()

		nodes = append(nodes, node)
	}

	if err := it.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate nodes: %w", err)
	}

	return nodes, nil
}

// DeleteNode removes a node and all its edges
func (r *CayleyGraphRepository) DeleteNode(ctx context.Context, phoneNumber string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.nodeExistsUnsafe(phoneNumber) {
		return ErrNodeNotFound
	}

	// Delete all quads where this phone number is subject or object
	quadsToDelete := make([]*quad.Quad, 0)

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

// AddContactEdge adds a bidirectional contact edge between two phone numbers (backward compatible)
func (r *CayleyGraphRepository) AddContactEdge(phone1, phone2 string) error {
	// Use default metadata
	metadata := &models.ContactMetadata{
		Name:    "", // Empty name for backward compatibility
		AddedAt: time.Now(),
	}

	// Add edge in both directions
	ctx := context.Background()
	_, err := r.AddEdgeWithMetadata(ctx, phone1, phone2, metadata)
	if err != nil {
		return err
	}

	return err
}

// AddContactEdgeWithMetadata adds a contact edge with full metadata
func (r *CayleyGraphRepository) AddContactEdgeWithMetadata(phone1, phone2 string, name string, addedAt time.Time) error {
	metadata := &models.ContactMetadata{
		Name:    name,
		AddedAt: addedAt,
	}

	// Add edge in both directions
	ctx := context.Background()
	_, err := r.AddEdgeWithMetadata(ctx, phone1, phone2, metadata)
	if err != nil {
		return err
	}

	return err
}

// AddCallEdge adds a directional call edge (backward compatible)
func (r *CayleyGraphRepository) AddCallEdge(from, to string, isAnswered bool, duration int, timestamp time.Time) (*models.Edge, error) {
	metadata := &models.CallMetadata{
		IsAnswered:        isAnswered,
		DurationInSeconds: duration,
		Timestamp:         timestamp,
	}

	ctx := context.Background()
	return r.AddEdgeWithMetadata(ctx, from, to, metadata)
}

// AddEdgeWithMetadata is the generic method to add any edge with metadata
func (r *CayleyGraphRepository) AddEdgeWithMetadata(ctx context.Context, from, to string, metadata models.EdgeMetadata) (*models.Edge, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Validate metadata
	if err := metadata.Validate(); err != nil {
		return nil, fmt.Errorf("invalid metadata: %w", err)
	}

	// Ensure both nodes exist
	if !r.nodeExistsUnsafe(from) {
		r.store.AddQuad(quad.Make(from, "type", "node", nil))
	}
	if !r.nodeExistsUnsafe(to) {
		r.store.AddQuad(quad.Make(to, "type", "node", nil))
	}

	properties := metadata.ToProperties()

	// Generate unique edge ID
	r.edgeCounter++
	edgeID := fmt.Sprintf("%s_%d", metadata.EdgeType(), r.edgeCounter)

	// Store edge based on type
	if metadata.EdgeType() == models.EdgeTypeContact {
		// Contact edges are stored directly: from -> has_contact -> to
		r.store.AddQuad(quad.Make(from, "has_contact", to, nil))

		// Store metadata as properties on the edge
		if name, ok := properties["name"].(string); ok && name != "" {
			// Store contact name: from -> contact_name_to -> name
			contactKey := fmt.Sprintf("%s_contact_%s", from, to)
			r.store.AddQuad(quad.Make(contactKey, "name", name, nil))
			r.store.AddQuad(quad.Make(contactKey, "from", from, nil))
			r.store.AddQuad(quad.Make(contactKey, "to", to, nil))
		}
		if addedAt, ok := properties["added_at"].(string); ok {
			contactKey := fmt.Sprintf("%s_contact_%s", from, to)
			r.store.AddQuad(quad.Make(contactKey, "added_at", addedAt, nil))
		}
	} else if metadata.EdgeType() == models.EdgeTypeCall {
		// Call edges are stored with an ID: call_id -> type -> "call"
		r.store.AddQuad(quad.Make(edgeID, "type", "call", nil))
		r.store.AddQuad(quad.Make(edgeID, "from", from, nil))
		r.store.AddQuad(quad.Make(edgeID, "to", to, nil))

		// Store all metadata properties
		for key, value := range properties {
			r.store.AddQuad(quad.Make(edgeID, key, value, nil))
		}
	}

	// Create edge object
	edge := &models.Edge{
		ID:        edgeID,
		From:      from,
		To:        to,
		Type:      metadata.EdgeType(),
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}

	// Set CreatedAt from metadata if available
	if callMeta, ok := metadata.(*models.CallMetadata); ok {
		edge.CreatedAt = callMeta.Timestamp
	} else if contactMeta, ok := metadata.(*models.ContactMetadata); ok {
		edge.CreatedAt = contactMeta.AddedAt
	}

	return edge, nil
}

// GetEdge retrieves an edge by ID (backward compatible - returns edge with properties map)
func (r *CayleyGraphRepository) GetEdge(ctx context.Context, edgeID string) (*models.Edge, error) {
	edge, _, err := r.GetEdgeWithMetadata(ctx, edgeID)
	return edge, err
}

// GetEdgeWithMetadata retrieves an edge with its metadata
func (r *CayleyGraphRepository) GetEdgeWithMetadata(ctx context.Context, edgeID string) (*models.Edge, models.EdgeMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check if edge exists and get its type
	typePath := cayley.StartPath(r.store, quad.String(edgeID)).Out(quad.String("type"))
	typeIt, _ := typePath.BuildIterator().Optimize()
	defer typeIt.Close()

	if !typeIt.Next(ctx) {
		// Try to find contact edge by checking if it's a contact metadata key
		return r.getContactEdgeByKey(edgeID)
	}

	token := typeIt.Result()
	edgeTypeStr := quad.ToString(r.store.NameOf(token))

	if edgeTypeStr == "call" {
		return r.getCallEdgeWithMetadataUnsafe(edgeID)
	}

	return nil, nil, ErrEdgeNotFound
}

// getContactEdgeByKey retrieves a contact edge by its metadata key
func (r *CayleyGraphRepository) getContactEdgeByKey(key string) (*models.Edge, models.EdgeMetadata, error) {
	ctx := context.TODO()

	// Check if this is a contact metadata key
	fromPath := cayley.StartPath(r.store, quad.String(key)).Out(quad.String("from"))
	fromIt, _ := fromPath.BuildIterator().Optimize()
	if !fromIt.Next(ctx) {
		fromIt.Close()
		return nil, nil, ErrEdgeNotFound
	}

	fromToken := fromIt.Result()
	from := quad.ToString(r.store.NameOf(fromToken))
	fromIt.Close()

	toPath := cayley.StartPath(r.store, quad.String(key)).Out(quad.String("to"))
	toIt, _ := toPath.BuildIterator().Optimize()
	if !toIt.Next(ctx) {
		toIt.Close()
		return nil, nil, ErrEdgeNotFound
	}

	toToken := toIt.Result()
	to := quad.ToString(r.store.NameOf(toToken))
	toIt.Close()

	// Get name
	name := ""
	namePath := cayley.StartPath(r.store, quad.String(key)).Out(quad.String("name"))
	nameIt, _ := namePath.BuildIterator().Optimize()
	if nameIt.Next(ctx) {
		nameToken := nameIt.Result()
		name = quad.ToString(r.store.NameOf(nameToken))
	}
	nameIt.Close()

	// Get added_at
	addedAt := time.Now()
	addedAtPath := cayley.StartPath(r.store, quad.String(key)).Out(quad.String("added_at"))
	addedAtIt, _ := addedAtPath.BuildIterator().Optimize()
	if addedAtIt.Next(ctx) {
		addedAtToken := addedAtIt.Result()
		addedAtStr := quad.ToString(r.store.NameOf(addedAtToken))
		if t, err := time.Parse(time.RFC3339, addedAtStr); err == nil {
			addedAt = t
		}
	}
	addedAtIt.Close()

	metadata := &models.ContactMetadata{
		Name:    name,
		AddedAt: addedAt,
	}

	edge := &models.Edge{
		ID:        key,
		From:      from,
		To:        to,
		Type:      models.EdgeTypeContact,
		Metadata:  metadata,
		CreatedAt: addedAt,
	}

	return edge, metadata, nil
}

// getCallEdgeWithMetadataUnsafe retrieves a call edge with metadata (must be called with lock held)
func (r *CayleyGraphRepository) getCallEdgeWithMetadataUnsafe(callID string) (*models.Edge, models.EdgeMetadata, error) {
	ctx := context.TODO()

	edge := &models.Edge{
		ID:   callID,
		Type: models.EdgeTypeCall,
	}

	properties := make(map[string]interface{})

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

	// Get all properties
	propertyPredicates := []string{"is_answered", "duration_in_seconds", "timestamp", "created_at"}
	for _, pred := range propertyPredicates {
		propPath := cayley.StartPath(r.store, quad.String(callID)).Out(quad.String(pred))
		propIt, _ := propPath.BuildIterator().Optimize()
		if propIt.Next(ctx) {
			token := propIt.Result()
			val := r.store.NameOf(token)
			nativeVal := quad.NativeOf(val)

			// Map property names
			if pred == "duration_in_seconds" {
				if intVal, ok := nativeVal.(int); ok {
					properties["duration_in_seconds"] = intVal
				} else if int64Val, ok := nativeVal.(int64); ok {
					properties["duration_in_seconds"] = int(int64Val)
				}
			} else if pred == "is_answered" {
				if boolVal, ok := nativeVal.(bool); ok {
					properties["is_answered"] = boolVal
				}
			} else if pred == "timestamp" || pred == "created_at" {
				if timeStr, ok := nativeVal.(string); ok {
					properties["timestamp"] = timeStr
					if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
						edge.CreatedAt = t
					}
				}
			}
		}
		propIt.Close()
	}

	// Deserialize metadata
	metadata, err := r.registry.Deserialize(models.EdgeTypeCall, properties)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to deserialize call metadata: %w", err)
	}

	edge.Metadata = metadata
	return edge, metadata, nil
}

// DeleteEdge removes an edge from the graph
func (r *CayleyGraphRepository) DeleteEdge(ctx context.Context, edgeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

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
func (r *CayleyGraphRepository) GetUsersWithContact(ctx context.Context, phoneNumber string) ([]string, int) {
	r.mu.RLock()
	defer r.mu.RUnlock()
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
func (r *CayleyGraphRepository) GetOutgoingEdges(ctx context.Context, phoneNumber string, edgeType models.EdgeType) []*models.Edge {
	r.mu.RLock()
	defer r.mu.RUnlock()

	edges := make([]*models.Edge, 0)

	if edgeType == models.EdgeTypeContact {
		// Find all: phoneNumber -has_contact-> ?
		p := cayley.StartPath(r.store, quad.String(phoneNumber)).Out(quad.String("has_contact"))
		it, _ := p.BuildIterator().Optimize()
		defer it.Close()

		for it.Next(ctx) {
			token := it.Result()
			toPhone := quad.ToString(r.store.NameOf(token))

			// Try to get metadata
			contactKey := fmt.Sprintf("%s_contact_%s", phoneNumber, toPhone)
			edge := &models.Edge{
				From: phoneNumber,
				To:   toPhone,
				Type: models.EdgeTypeContact,
			}

			// Get metadata if available
			namePath := cayley.StartPath(r.store, quad.String(contactKey)).Out(quad.String("name"))
			nameIt, _ := namePath.BuildIterator().Optimize()
			if nameIt.Next(ctx) {
				nameToken := nameIt.Result()
				name := quad.ToString(r.store.NameOf(nameToken))
				addedAt := time.Now()

				addedAtPath := cayley.StartPath(r.store, quad.String(contactKey)).Out(quad.String("added_at"))
				addedAtIt, _ := addedAtPath.BuildIterator().Optimize()
				if addedAtIt.Next(ctx) {
					addedAtToken := addedAtIt.Result()
					addedAtStr := quad.ToString(r.store.NameOf(addedAtToken))
					if t, err := time.Parse(time.RFC3339, addedAtStr); err == nil {
						addedAt = t
					}
				}
				addedAtIt.Close()

				metadata := &models.ContactMetadata{
					Name:    name,
					AddedAt: addedAt,
				}
				edge.Metadata = metadata
				edge.CreatedAt = addedAt
			}
			nameIt.Close()

			edges = append(edges, edge)
		}
	} else if edgeType == models.EdgeTypeCall {
		// Find all calls where from = phoneNumber
		edges = r.getCallsByPhoneUnsafe(phoneNumber, "from")
	}

	return edges
}

// GetIncomingEdges returns all incoming edges of a specific type to a phone number
func (r *CayleyGraphRepository) GetIncomingEdges(ctx context.Context, phoneNumber string, edgeType models.EdgeType) []*models.Edge {
	r.mu.RLock()
	defer r.mu.RUnlock()

	edges := make([]*models.Edge, 0)

	if edgeType == models.EdgeTypeContact {
		// Find all: ? -has_contact-> phoneNumber
		p := cayley.StartPath(r.store, quad.String(phoneNumber)).In(quad.String("has_contact"))
		it, _ := p.BuildIterator().Optimize()
		defer it.Close()

		for it.Next(ctx) {
			token := it.Result()
			fromPhone := quad.ToString(r.store.NameOf(token))

			edge := &models.Edge{
				From: fromPhone,
				To:   phoneNumber,
				Type: models.EdgeTypeContact,
			}

			// Try to get metadata
			contactKey := fmt.Sprintf("%s_contact_%s", fromPhone, phoneNumber)
			namePath := cayley.StartPath(r.store, quad.String(contactKey)).Out(quad.String("name"))
			nameIt, _ := namePath.BuildIterator().Optimize()
			if nameIt.Next(ctx) {
				nameToken := nameIt.Result()
				name := quad.ToString(r.store.NameOf(nameToken))
				addedAt := time.Now()

				addedAtPath := cayley.StartPath(r.store, quad.String(contactKey)).Out(quad.String("added_at"))
				addedAtIt, _ := addedAtPath.BuildIterator().Optimize()
				if addedAtIt.Next(ctx) {
					addedAtToken := addedAtIt.Result()
					addedAtStr := quad.ToString(r.store.NameOf(addedAtToken))
					if t, err := time.Parse(time.RFC3339, addedAtStr); err == nil {
						addedAt = t
					}
				}
				addedAtIt.Close()

				metadata := &models.ContactMetadata{
					Name:    name,
					AddedAt: addedAt,
				}
				edge.Metadata = metadata
				edge.CreatedAt = addedAt
			}
			nameIt.Close()

			edges = append(edges, edge)
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
		if edge, _, err := r.getCallEdgeWithMetadataUnsafe(callID); err == nil {
			edges = append(edges, edge)
		}
	}

	return edges
}

// GetCallsWithFilters returns call edges with applied filters
// Query 2: How many calls a phone number is making with filters
// direction: "outgoing", "incoming", or "both"
func (r *CayleyGraphRepository) GetCallsWithFilters(ctx context.Context, phoneNumber string, filters CallFilters, direction string) ([]*models.Edge, int) {
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

	// Use metadata if available, otherwise fall back to properties
	var callMeta *models.CallMetadata
	if edge.Metadata != nil {
		if cm, ok := edge.Metadata.(*models.CallMetadata); ok {
			callMeta = cm
		}
	}

	// Fallback to parsing properties for backward compatibility
	if callMeta == nil {
		props := edge.GetProperties()
		if meta, err := r.registry.Deserialize(models.EdgeTypeCall, props); err == nil {
			if cm, ok := meta.(*models.CallMetadata); ok {
				callMeta = cm
			}
		}
	}

	if callMeta == nil {
		return false
	}

	// Filter by IsAnswered
	if filters.IsAnswered != nil && callMeta.IsAnswered != *filters.IsAnswered {
		return false
	}

	// Filter by MaxDuration
	if filters.MaxDuration != nil && callMeta.DurationInSeconds > *filters.MaxDuration {
		return false
	}

	// Filter by MinDuration
	if filters.MinDuration != nil && callMeta.DurationInSeconds < *filters.MinDuration {
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

// IsDirectContact checks if callerPhone is in userPhone's direct contacts (level 1)
// Uses Cayley path query: userPhone -has_contact-> callerPhone
func (r *CayleyGraphRepository) IsDirectContact(ctx context.Context, userPhone, callerPhone string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Query: userPhone -has_contact-> callerPhone
	p := cayley.StartPath(r.store, quad.String(userPhone)).
		Out(quad.String("has_contact")).
		Is(quad.String(callerPhone))

	it, _ := p.BuildIterator().Optimize()
	defer it.Close()

	return it.Next(ctx)
}

// GetSecondLevelContactCount counts how many of userPhone's contacts have callerPhone in their contacts
// Uses Cayley path query: userPhone -has_contact-> ? -has_contact-> callerPhone
// Returns the count of unique intermediate contacts (level 2 matches)
func (r *CayleyGraphRepository) GetSecondLevelContactCount(ctx context.Context, userPhone, callerPhone string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Query: Find all paths: userPhone -has_contact-> intermediate -has_contact-> callerPhone
	// We traverse: userPhone -> Out("has_contact") -> Out("has_contact") -> Is(callerPhone)
	// This finds all intermediate contacts that have callerPhone
	p := cayley.StartPath(r.store, quad.String(userPhone)).
		Out(quad.String("has_contact")).
		Out(quad.String("has_contact")).
		Is(quad.String(callerPhone))

	it, _ := p.BuildIterator().Optimize()
	defer it.Close()

	// Count unique intermediate contacts
	// We need to get the intermediate node (one hop from userPhone)
	// So we need to track: userPhone -> contact -> callerPhone
	// The intermediate contact is what we want to count

	// Alternative approach: Get all contacts of userPhone, then check each
	// But we can use a more efficient query by getting the path and extracting intermediate nodes

	// Actually, to count unique intermediates, we need to:
	// 1. Get all contacts of userPhone
	// 2. For each, check if they have callerPhone
	// But we can optimize with Cayley by using a different approach

	// Let's use: Get all contacts of userPhone, then check which ones have callerPhone
	contactsPath := cayley.StartPath(r.store, quad.String(userPhone)).Out(quad.String("has_contact"))
	contactsIt, _ := contactsPath.BuildIterator().Optimize()
	defer contactsIt.Close()

	count := 0
	checkedContacts := make(map[string]bool)

	for contactsIt.Next(ctx) {
		contactToken := contactsIt.Result()
		contactPhone := quad.ToString(r.store.NameOf(contactToken))

		// Skip if already checked
		if checkedContacts[contactPhone] {
			continue
		}
		checkedContacts[contactPhone] = true

		// Check if this contact has callerPhone: contactPhone -has_contact-> callerPhone
		hasCallerPath := cayley.StartPath(r.store, quad.String(contactPhone)).
			Out(quad.String("has_contact")).
			Is(quad.String(callerPhone))

		hasCallerIt, _ := hasCallerPath.BuildIterator().Optimize()
		if hasCallerIt.Next(ctx) {
			count++
		}
		hasCallerIt.Close()
	}

	return count
}

// LoadSeedData loads graph seed data from a JSON file
func (r *CayleyGraphRepository) LoadSeedData(ctx context.Context, filePath string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Read the JSON file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Parse the JSON using EdgeJSON format
	var seedData struct {
		Nodes []*models.Node     `json:"nodes"`
		Edges []*models.EdgeJSON `json:"edges"`
	}

	if err := json.Unmarshal(data, &seedData); err != nil {
		return err
	}

	// Load nodes
	for _, node := range seedData.Nodes {
		if !r.nodeExistsUnsafe(node.PhoneNumber) {
			r.store.AddQuad(quad.Make(node.PhoneNumber, "type", "node", nil))
			if node.Name != "" {
				r.store.AddQuad(quad.Make(node.PhoneNumber, "name", node.Name, nil))
			}
		}
	}

	// Load edges
	for _, edgeJSON := range seedData.Edges {
		// Ensure nodes exist
		if !r.nodeExistsUnsafe(edgeJSON.From) {
			r.store.AddQuad(quad.Make(edgeJSON.From, "type", "node", nil))
		}
		if !r.nodeExistsUnsafe(edgeJSON.To) {
			r.store.AddQuad(quad.Make(edgeJSON.To, "type", "node", nil))
		}

		// Deserialize edge using registry
		edge, err := edgeJSON.ToEdge(r.registry)
		if err != nil {
			return fmt.Errorf("failed to deserialize edge %s: %w", edgeJSON.ID, err)
		}

		// Store edge based on type
		if edge.Type == models.EdgeTypeContact {
			r.store.AddQuad(quad.Make(edge.From, "has_contact", edge.To, nil))

			// Store contact metadata if available
			if contactMeta, ok := edge.Metadata.(*models.ContactMetadata); ok {
				contactKey := fmt.Sprintf("%s_contact_%s", edge.From, edge.To)
				if contactMeta.Name != "" {
					r.store.AddQuad(quad.Make(contactKey, "name", contactMeta.Name, nil))
					r.store.AddQuad(quad.Make(contactKey, "from", edge.From, nil))
					r.store.AddQuad(quad.Make(contactKey, "to", edge.To, nil))
				}
				if !contactMeta.AddedAt.IsZero() {
					r.store.AddQuad(quad.Make(contactKey, "added_at", contactMeta.AddedAt.Format(time.RFC3339), nil))
				}
			}
		} else if edge.Type == models.EdgeTypeCall {
			// Store call edge
			r.store.AddQuad(quad.Make(edge.ID, "type", "call", nil))
			r.store.AddQuad(quad.Make(edge.ID, "from", edge.From, nil))
			r.store.AddQuad(quad.Make(edge.ID, "to", edge.To, nil))

			// Store all metadata properties
			props := edge.GetProperties()
			for key, value := range props {
				r.store.AddQuad(quad.Make(edge.ID, key, value, nil))
			}
		}
	}

	return nil
}
