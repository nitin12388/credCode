package models

import (
	"errors"
	"fmt"
	"time"
)

// Node represents a phone number node in the graph
type Node struct {
	PhoneNumber string `json:"phone_number"`
	Name        string `json:"name"`
}

// EdgeType represents the type of relationship between nodes
type EdgeType string

const (
	EdgeTypeContact EdgeType = "has_contact"
	EdgeTypeCall    EdgeType = "call"
)

// EdgeMetadata is the interface that all edge metadata types must implement
// This allows for type-safe, extensible edge metadata
type EdgeMetadata interface {
	// EdgeType returns the edge type this metadata belongs to
	EdgeType() EdgeType
	
	// ToProperties converts metadata to a map for storage
	ToProperties() map[string]interface{}
	
	// Validate ensures the metadata is valid
	Validate() error
	
	// FromProperties deserializes metadata from a map
	FromProperties(props map[string]interface{}) error
}

// Edge represents a relationship between two nodes
type Edge struct {
	ID        string       `json:"id"`
	From      string       `json:"from"`        // phone number
	To        string       `json:"to"`          // phone number
	Type      EdgeType     `json:"type"`
	Metadata  EdgeMetadata `json:"-"`          // metadata object (not serialized directly)
	CreatedAt time.Time    `json:"created_at"`
}

// GetProperties returns the properties map from metadata
func (e *Edge) GetProperties() map[string]interface{} {
	if e.Metadata != nil {
		return e.Metadata.ToProperties()
	}
	return make(map[string]interface{})
}

// SetMetadata sets the metadata and updates the edge type
func (e *Edge) SetMetadata(metadata EdgeMetadata) error {
	if err := metadata.Validate(); err != nil {
		return err
	}
	e.Metadata = metadata
	e.Type = metadata.EdgeType()
	return nil
}

// ContactMetadata represents metadata for contact edges
type ContactMetadata struct {
	Name    string    `json:"name"`
	AddedAt time.Time `json:"added_at"`
}

// EdgeType implements EdgeMetadata interface
func (cm *ContactMetadata) EdgeType() EdgeType {
	return EdgeTypeContact
}

// ToProperties converts ContactMetadata to a map
func (cm *ContactMetadata) ToProperties() map[string]interface{} {
	return map[string]interface{}{
		"name":      cm.Name,
		"added_at":  cm.AddedAt.Format(time.RFC3339),
	}
}

// FromProperties deserializes ContactMetadata from a map
func (cm *ContactMetadata) FromProperties(props map[string]interface{}) error {
	if name, ok := props["name"].(string); ok {
		cm.Name = name
	}
	
	if addedAtStr, ok := props["added_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, addedAtStr); err == nil {
			cm.AddedAt = t
		} else {
			return fmt.Errorf("invalid added_at format: %w", err)
		}
	}
	
	return nil
}

// Validate ensures ContactMetadata is valid
func (cm *ContactMetadata) Validate() error {
	if cm.Name == "" {
		return errors.New("contact name cannot be empty")
	}
	if cm.AddedAt.IsZero() {
		cm.AddedAt = time.Now() // Default to current time if not set
	}
	return nil
}

// CallMetadata represents metadata for call edges
type CallMetadata struct {
	IsAnswered        bool      `json:"is_answered"`
	DurationInSeconds int       `json:"duration_in_seconds"`
	Timestamp         time.Time `json:"timestamp"`
}

// EdgeType implements EdgeMetadata interface
func (cm *CallMetadata) EdgeType() EdgeType {
	return EdgeTypeCall
}

// ToProperties converts CallMetadata to a map
func (cm *CallMetadata) ToProperties() map[string]interface{} {
	return map[string]interface{}{
		"is_answered":         cm.IsAnswered,
		"duration_in_seconds": cm.DurationInSeconds,
		"timestamp":           cm.Timestamp.Format(time.RFC3339),
	}
}

// FromProperties deserializes CallMetadata from a map
func (cm *CallMetadata) FromProperties(props map[string]interface{}) error {
	if isAnswered, ok := props["is_answered"].(bool); ok {
		cm.IsAnswered = isAnswered
	}
	
	if duration, ok := props["duration_in_seconds"].(int); ok {
		cm.DurationInSeconds = duration
	} else if duration, ok := props["duration_in_seconds"].(float64); ok {
		cm.DurationInSeconds = int(duration)
	}
	
	if timestampStr, ok := props["timestamp"].(string); ok {
		if t, err := time.Parse(time.RFC3339, timestampStr); err == nil {
			cm.Timestamp = t
		}
	} else if createdAt, ok := props["created_at"].(string); ok {
		// Backward compatibility with old format
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			cm.Timestamp = t
		}
	}
	
	return nil
}

// Validate ensures CallMetadata is valid
func (cm *CallMetadata) Validate() error {
	if cm.DurationInSeconds < 0 {
		return errors.New("call duration cannot be negative")
	}
	if cm.Timestamp.IsZero() {
		cm.Timestamp = time.Now() // Default to current time if not set
	}
	return nil
}

// EdgeMetadataRegistry manages registration and deserialization of edge metadata types
type EdgeMetadataRegistry struct {
	factories map[EdgeType]func() EdgeMetadata
}

// NewEdgeMetadataRegistry creates a new registry with default edge types registered
func NewEdgeMetadataRegistry() *EdgeMetadataRegistry {
	registry := &EdgeMetadataRegistry{
		factories: make(map[EdgeType]func() EdgeMetadata),
	}
	
	// Register default edge types
	registry.Register(EdgeTypeContact, func() EdgeMetadata { return &ContactMetadata{} })
	registry.Register(EdgeTypeCall, func() EdgeMetadata { return &CallMetadata{} })
	
	return registry
}

// Register registers a factory function for creating edge metadata of a specific type
func (r *EdgeMetadataRegistry) Register(edgeType EdgeType, factory func() EdgeMetadata) {
	r.factories[edgeType] = factory
}

// Deserialize creates and populates edge metadata from properties
func (r *EdgeMetadataRegistry) Deserialize(edgeType EdgeType, props map[string]interface{}) (EdgeMetadata, error) {
	factory, exists := r.factories[edgeType]
	if !exists {
		return nil, fmt.Errorf("unknown edge type: %s", edgeType)
	}
	
	metadata := factory()
	if err := metadata.FromProperties(props); err != nil {
		return nil, fmt.Errorf("failed to deserialize metadata: %w", err)
	}
	
	return metadata, nil
}

// GetRegisteredTypes returns all registered edge types
func (r *EdgeMetadataRegistry) GetRegisteredTypes() []EdgeType {
	types := make([]EdgeType, 0, len(r.factories))
	for edgeType := range r.factories {
		types = append(types, edgeType)
	}
	return types
}

// EdgeJSON is used for JSON serialization/deserialization
// It includes properties as a map for storage
type EdgeJSON struct {
	ID         string                 `json:"id"`
	From       string                 `json:"from"`
	To         string                 `json:"to"`
	Type       EdgeType               `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	CreatedAt  time.Time              `json:"created_at"`
}

// ToEdge converts EdgeJSON to Edge using the registry
func (ej *EdgeJSON) ToEdge(registry *EdgeMetadataRegistry) (*Edge, error) {
	edge := &Edge{
		ID:        ej.ID,
		From:      ej.From,
		To:        ej.To,
		Type:      ej.Type,
		CreatedAt: ej.CreatedAt,
	}
	
	if ej.Properties != nil && len(ej.Properties) > 0 {
		metadata, err := registry.Deserialize(ej.Type, ej.Properties)
		if err != nil {
			return nil, err
		}
		edge.Metadata = metadata
	}
	
	return edge, nil
}

// FromEdge converts Edge to EdgeJSON
func (ej *EdgeJSON) FromEdge(edge *Edge) {
	ej.ID = edge.ID
	ej.From = edge.From
	ej.To = edge.To
	ej.Type = edge.Type
	ej.CreatedAt = edge.CreatedAt
	ej.Properties = edge.GetProperties()
}

// Backward compatibility helpers

// CallProperties is kept for backward compatibility
type CallProperties struct {
	IsAnswered        bool `json:"is_answered"`
	DurationInSeconds int  `json:"duration_in_seconds"`
}

// ToMap converts CallProperties to map for Edge.Properties
func (cp *CallProperties) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"is_answered":         cp.IsAnswered,
		"duration_in_seconds": cp.DurationInSeconds,
	}
}

// ParseCallProperties extracts call properties from an edge (backward compatibility)
func ParseCallProperties(properties map[string]interface{}) *CallProperties {
	cp := &CallProperties{}
	
	if isAnswered, ok := properties["is_answered"].(bool); ok {
		cp.IsAnswered = isAnswered
	}
	
	if duration, ok := properties["duration_in_seconds"].(int); ok {
		cp.DurationInSeconds = duration
	} else if duration, ok := properties["duration_in_seconds"].(float64); ok {
		cp.DurationInSeconds = int(duration)
	}
	
	return cp
}
