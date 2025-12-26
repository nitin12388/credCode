package models

import (
	"testing"
	"time"
)

func TestNode_Structure(t *testing.T) {
	node := &Node{
		PhoneNumber: "7379037972",
		Name:        "John Doe",
	}

	if node.PhoneNumber != "7379037972" {
		t.Errorf("Expected phone number '7379037972', got '%s'", node.PhoneNumber)
	}

	if node.Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got '%s'", node.Name)
	}
}

func TestEdgeType_Constants(t *testing.T) {
	if EdgeTypeContact != "has_contact" {
		t.Errorf("Expected EdgeTypeContact to be 'has_contact', got '%s'", EdgeTypeContact)
	}

	if EdgeTypeCall != "call" {
		t.Errorf("Expected EdgeTypeCall to be 'call', got '%s'", EdgeTypeCall)
	}
}

func TestContactMetadata_EdgeType(t *testing.T) {
	meta := &ContactMetadata{
		Name:    "John",
		AddedAt: time.Now(),
	}

	if meta.EdgeType() != EdgeTypeContact {
		t.Errorf("Expected EdgeTypeContact, got %s", meta.EdgeType())
	}
}

func TestContactMetadata_ToProperties(t *testing.T) {
	now := time.Now()
	meta := &ContactMetadata{
		Name:    "John Doe",
		AddedAt: now,
	}

	props := meta.ToProperties()

	if props["name"] != "John Doe" {
		t.Errorf("Expected name 'John Doe', got '%v'", props["name"])
	}

	if addedAtStr, ok := props["added_at"].(string); !ok {
		t.Errorf("Expected added_at to be string, got %T", props["added_at"])
	} else {
		parsed, err := time.Parse(time.RFC3339, addedAtStr)
		if err != nil {
			t.Errorf("Failed to parse added_at: %v", err)
		}
		if !parsed.Equal(now.Truncate(time.Second)) {
			t.Errorf("Expected added_at to match, got %v", parsed)
		}
	}
}

func TestContactMetadata_FromProperties(t *testing.T) {
	now := time.Now()
	props := map[string]interface{}{
		"name":      "John Doe",
		"added_at":  now.Format(time.RFC3339),
	}

	meta := &ContactMetadata{}
	err := meta.FromProperties(props)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if meta.Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got '%s'", meta.Name)
	}

	if !meta.AddedAt.Equal(now.Truncate(time.Second)) {
		t.Errorf("Expected added_at to match")
	}
}

func TestContactMetadata_Validate(t *testing.T) {
	tests := []struct {
		name    string
		meta    *ContactMetadata
		wantErr bool
	}{
		{
			name: "valid metadata",
			meta: &ContactMetadata{
				Name:    "John",
				AddedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "metadata with empty name",
			meta: &ContactMetadata{
				Name:    "",
				AddedAt: time.Now(),
			},
			wantErr: false, // Name can be empty for backward compatibility
		},
		{
			name: "metadata with zero time",
			meta: &ContactMetadata{
				Name:    "John",
				AddedAt: time.Time{},
			},
			wantErr: false, // Will default to current time
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.meta.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCallMetadata_EdgeType(t *testing.T) {
	meta := &CallMetadata{
		IsAnswered:        true,
		DurationInSeconds: 120,
		Timestamp:         time.Now(),
	}

	if meta.EdgeType() != EdgeTypeCall {
		t.Errorf("Expected EdgeTypeCall, got %s", meta.EdgeType())
	}
}

func TestCallMetadata_ToProperties(t *testing.T) {
	now := time.Now()
	meta := &CallMetadata{
		IsAnswered:        true,
		DurationInSeconds: 120,
		Timestamp:         now,
	}

	props := meta.ToProperties()

	if props["is_answered"] != true {
		t.Errorf("Expected is_answered to be true, got %v", props["is_answered"])
	}

	if props["duration_in_seconds"] != 120 {
		t.Errorf("Expected duration_in_seconds to be 120, got %v", props["duration_in_seconds"])
	}
}

func TestCallMetadata_FromProperties(t *testing.T) {
	now := time.Now()
	props := map[string]interface{}{
		"is_answered":         true,
		"duration_in_seconds":  120,
		"timestamp":            now.Format(time.RFC3339),
	}

	meta := &CallMetadata{}
	err := meta.FromProperties(props)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if meta.IsAnswered != true {
		t.Errorf("Expected IsAnswered to be true, got %v", meta.IsAnswered)
	}

	if meta.DurationInSeconds != 120 {
		t.Errorf("Expected DurationInSeconds to be 120, got %d", meta.DurationInSeconds)
	}
}

func TestCallMetadata_Validate(t *testing.T) {
	tests := []struct {
		name    string
		meta    *CallMetadata
		wantErr bool
	}{
		{
			name: "valid metadata",
			meta: &CallMetadata{
				IsAnswered:        true,
				DurationInSeconds: 120,
				Timestamp:         time.Now(),
			},
			wantErr: false,
		},
		{
			name: "negative duration",
			meta: &CallMetadata{
				IsAnswered:        true,
				DurationInSeconds: -10,
				Timestamp:         time.Now(),
			},
			wantErr: true,
		},
		{
			name: "zero timestamp",
			meta: &CallMetadata{
				IsAnswered:        true,
				DurationInSeconds: 120,
				Timestamp:         time.Time{},
			},
			wantErr: false, // Will default to current time
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.meta.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEdge_GetProperties(t *testing.T) {
	meta := &ContactMetadata{
		Name:    "John",
		AddedAt: time.Now(),
	}

	edge := &Edge{
		ID:        "e1",
		From:      "7379037972",
		To:        "9876543210",
		Type:      EdgeTypeContact,
		Metadata:  meta,
		CreatedAt: time.Now(),
	}

	props := edge.GetProperties()

	if props == nil {
		t.Errorf("Expected properties, got nil")
	}

	if props["name"] != "John" {
		t.Errorf("Expected name 'John', got '%v'", props["name"])
	}
}

func TestEdge_SetMetadata(t *testing.T) {
	edge := &Edge{
		ID:   "e1",
		From: "7379037972",
		To:   "9876543210",
	}

	meta := &ContactMetadata{
		Name:    "John",
		AddedAt: time.Now(),
	}

	err := edge.SetMetadata(meta)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if edge.Metadata != meta {
		t.Errorf("Expected metadata to be set")
	}

	if edge.Type != EdgeTypeContact {
		t.Errorf("Expected type to be EdgeTypeContact, got %s", edge.Type)
	}
}

func TestEdgeMetadataRegistry_Register(t *testing.T) {
	registry := NewEdgeMetadataRegistry()

	// Test that default types are registered
	types := registry.GetRegisteredTypes()

	if len(types) < 2 {
		t.Errorf("Expected at least 2 registered types, got %d", len(types))
	}

	// Check that Contact and Call types are registered
	hasContact := false
	hasCall := false
	for _, t := range types {
		if t == EdgeTypeContact {
			hasContact = true
		}
		if t == EdgeTypeCall {
			hasCall = true
		}
	}

	if !hasContact {
		t.Errorf("Expected EdgeTypeContact to be registered")
	}

	if !hasCall {
		t.Errorf("Expected EdgeTypeCall to be registered")
	}
}

func TestEdgeMetadataRegistry_Deserialize(t *testing.T) {
	registry := NewEdgeMetadataRegistry()

	// Test deserializing ContactMetadata
	contactProps := map[string]interface{}{
		"name":      "John",
		"added_at":  time.Now().Format(time.RFC3339),
	}

	contactMeta, err := registry.Deserialize(EdgeTypeContact, contactProps)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if contactMeta.EdgeType() != EdgeTypeContact {
		t.Errorf("Expected EdgeTypeContact, got %s", contactMeta.EdgeType())
	}

	// Test deserializing CallMetadata
	callProps := map[string]interface{}{
		"is_answered":         true,
		"duration_in_seconds":  120,
		"timestamp":            time.Now().Format(time.RFC3339),
	}

	callMeta, err := registry.Deserialize(EdgeTypeCall, callProps)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if callMeta.EdgeType() != EdgeTypeCall {
		t.Errorf("Expected EdgeTypeCall, got %s", callMeta.EdgeType())
	}

	// Test unknown type
	_, err = registry.Deserialize(EdgeType("unknown"), contactProps)
	if err == nil {
		t.Errorf("Expected error for unknown edge type")
	}
}

func TestEdgeJSON_ToEdge(t *testing.T) {
	registry := NewEdgeMetadataRegistry()

	ej := &EdgeJSON{
		ID:   "e1",
		From: "7379037972",
		To:   "9876543210",
		Type: EdgeTypeContact,
		Properties: map[string]interface{}{
			"name":      "John",
			"added_at":  time.Now().Format(time.RFC3339),
		},
		CreatedAt: time.Now(),
	}

	edge, err := ej.ToEdge(registry)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if edge.ID != "e1" {
		t.Errorf("Expected ID 'e1', got '%s'", edge.ID)
	}

	if edge.Metadata == nil {
		t.Errorf("Expected metadata to be set")
	}
}

func TestEdgeJSON_FromEdge(t *testing.T) {
	meta := &ContactMetadata{
		Name:    "John",
		AddedAt: time.Now(),
	}

	edge := &Edge{
		ID:        "e1",
		From:      "7379037972",
		To:        "9876543210",
		Type:      EdgeTypeContact,
		Metadata:  meta,
		CreatedAt: time.Now(),
	}

	ej := &EdgeJSON{}
	ej.FromEdge(edge)

	if ej.ID != "e1" {
		t.Errorf("Expected ID 'e1', got '%s'", ej.ID)
	}

	if ej.Properties == nil {
		t.Errorf("Expected properties to be set")
	}
}

