# Edge Metadata Framework

## Overview

A robust, extensible framework for managing edge metadata in the graph database. This framework allows you to add any edge type with any metadata structure while maintaining type safety and backward compatibility.

## Architecture

### Core Components

1. **EdgeMetadata Interface** - Base interface for all edge metadata types
2. **EdgeMetadataRegistry** - Manages registration and deserialization of metadata types
3. **Specific Metadata Types** - ContactMetadata, CallMetadata, and extensible for future types
4. **Node Model** - Enhanced with name field

## Node Model

```go
type Node struct {
    PhoneNumber string `json:"phone_number"`
    Name        string `json:"name"`  // NEW: Node name
}
```

### Node Operations

```go
// Add node with name
repo.AddNodeWithName("7379037972", "John Doe")

// Add node without name (backward compatible)
repo.AddNode("7379037972")

// Get node (includes name)
node, _ := repo.GetNode("7379037972")
fmt.Println(node.Name) // "John Doe"
```

## Edge Metadata Framework

### EdgeMetadata Interface

All edge metadata types must implement this interface:

```go
type EdgeMetadata interface {
    EdgeType() EdgeType                    // Returns the edge type
    ToProperties() map[string]interface{}  // Serializes to map
    FromProperties(props map[string]interface{}) error  // Deserializes from map
    Validate() error                       // Validates metadata
}
```

### ContactMetadata

Metadata for contact edges:

```go
type ContactMetadata struct {
    Name    string    // Contact name as saved by user
    AddedAt time.Time // When contact was added
}
```

**Usage:**
```go
contactMeta := &models.ContactMetadata{
    Name:    "Priya",
    AddedAt: time.Now().Add(-5 * 24 * time.Hour),
}

edge, err := repo.AddEdgeWithMetadata("7379037972", "9876543210", contactMeta)
```

### CallMetadata

Metadata for call edges:

```go
type CallMetadata struct {
    IsAnswered        bool      // Whether call was answered
    DurationInSeconds int       // Call duration
    Timestamp         time.Time // When call occurred
}
```

**Usage:**
```go
callMeta := &models.CallMetadata{
    IsAnswered:        true,
    DurationInSeconds: 120,
    Timestamp:         time.Now().Add(-2 * time.Hour),
}

edge, err := repo.AddEdgeWithMetadata("7379037972", "1234567890", callMeta)
```

## Edge Metadata Registry

The registry manages registration and deserialization of edge metadata types:

```go
registry := models.NewEdgeMetadataRegistry()

// Register a new edge type
registry.Register(EdgeTypeCustom, func() EdgeMetadata {
    return &CustomMetadata{}
})

// Deserialize metadata from properties
metadata, err := registry.Deserialize(EdgeTypeCall, properties)
```

### Default Registered Types

- `EdgeTypeContact` → `ContactMetadata`
- `EdgeTypeCall` → `CallMetadata`

## Generic Edge Operations

### AddEdgeWithMetadata

Generic method to add any edge with metadata:

```go
edge, err := repo.AddEdgeWithMetadata(from, to, metadata)
```

**Benefits:**
- Type-safe metadata
- Automatic validation
- Consistent storage format
- Easy to extend

### GetEdgeWithMetadata

Retrieve edge with its metadata:

```go
edge, metadata, err := repo.GetEdgeWithMetadata(edgeID)

// Type assertion to get specific metadata
if callMeta, ok := metadata.(*models.CallMetadata); ok {
    fmt.Println(callMeta.DurationInSeconds)
}
```

## Backward Compatibility

All existing methods continue to work:

```go
// Old methods still work
repo.AddContactEdge("7379037972", "9876543210")
repo.AddCallEdge("7379037972", "1234567890", true, 120, time.Now())

// Old methods now use default metadata
// ContactEdge: empty name, current time
// CallEdge: uses provided parameters
```

## Query Operations

All query operations work with metadata:

```go
// Get outgoing edges (includes metadata)
edges := repo.GetOutgoingEdges("7379037972", models.EdgeTypeContact)
for _, edge := range edges {
    if cm, ok := edge.Metadata.(*models.ContactMetadata); ok {
        fmt.Println(cm.Name)
    }
}

// Call filtering (uses metadata)
calls, count := repo.GetCallsWithFilters("7379037972", CallFilters{
    IsAnswered: &answered,
}, "outgoing")
```

## Adding New Edge Types

### Step 1: Create Metadata Struct

```go
type MessageMetadata struct {
    Content   string
    SentAt    time.Time
    IsRead    bool
}

func (mm *MessageMetadata) EdgeType() EdgeType {
    return EdgeTypeMessage
}

func (mm *MessageMetadata) ToProperties() map[string]interface{} {
    return map[string]interface{}{
        "content": mm.Content,
        "sent_at": mm.SentAt.Format(time.RFC3339),
        "is_read": mm.IsRead,
    }
}

func (mm *MessageMetadata) FromProperties(props map[string]interface{}) error {
    // Deserialization logic
    return nil
}

func (mm *MessageMetadata) Validate() error {
    // Validation logic
    return nil
}
```

### Step 2: Register Edge Type

```go
// In repository initialization
repo.registry.Register(EdgeTypeMessage, func() EdgeMetadata {
    return &MessageMetadata{}
})
```

### Step 3: Use It

```go
messageMeta := &MessageMetadata{
    Content: "Hello!",
    SentAt:  time.Now(),
    IsRead:  false,
}

edge, err := repo.AddEdgeWithMetadata("7379037972", "9876543210", messageMeta)
```

## Storage Format

### Contact Edges

```
from -> has_contact -> to
contactKey -> name -> "Contact Name"
contactKey -> added_at -> "2024-01-15T10:00:00Z"
```

### Call Edges

```
callID -> type -> "call"
callID -> from -> "7379037972"
callID -> to -> "9876543210"
callID -> is_answered -> true
callID -> duration_in_seconds -> 120
callID -> timestamp -> "2024-01-20T10:30:00Z"
```

## Benefits

### 1. **Type Safety**
- Compile-time checking
- No runtime property name errors
- IDE autocomplete support

### 2. **Extensibility**
- Easy to add new edge types
- No changes to core framework
- Registry-based registration

### 3. **Validation**
- Built-in validation per metadata type
- Consistent error handling
- Data integrity

### 4. **Backward Compatibility**
- Existing code continues to work
- Gradual migration path
- No breaking changes

### 5. **Maintainability**
- Clear separation of concerns
- Self-documenting code
- Easy to test

## Example: Complete Workflow

```go
// 1. Create repository
repo := repository.NewInMemoryGraphRepository()

// 2. Add nodes with names
repo.AddNodeWithName("7379037972", "John Doe")
repo.AddNodeWithName("9876543210", "Priya Kumar")

// 3. Add contact edge with metadata
contactMeta := &models.ContactMetadata{
    Name:    "Priya",
    AddedAt: time.Now(),
}
edge, _ := repo.AddEdgeWithMetadata("7379037972", "9876543210", contactMeta)

// 4. Add call edge with metadata
callMeta := &models.CallMetadata{
    IsAnswered:        true,
    DurationInSeconds: 120,
    Timestamp:         time.Now(),
}
callEdge, _ := repo.AddEdgeWithMetadata("7379037972", "9876543210", callMeta)

// 5. Query with metadata
edges := repo.GetOutgoingEdges("7379037972", models.EdgeTypeContact)
for _, e := range edges {
    if cm, ok := e.Metadata.(*models.ContactMetadata); ok {
        fmt.Printf("Contact: %s (added: %s)\n", cm.Name, cm.AddedAt)
    }
}

// 6. Retrieve with metadata
retrievedEdge, metadata, _ := repo.GetEdgeWithMetadata(callEdge.ID)
if cm, ok := metadata.(*models.CallMetadata); ok {
    fmt.Printf("Call duration: %ds\n", cm.DurationInSeconds)
}
```

## Migration Guide

### From Old Format

**Before:**
```go
edge.Properties["is_answered"] = true
edge.Properties["duration_in_seconds"] = 120
```

**After:**
```go
callMeta := &models.CallMetadata{
    IsAnswered:        true,
    DurationInSeconds: 120,
    Timestamp:         time.Now(),
}
edge, _ := repo.AddEdgeWithMetadata(from, to, callMeta)
```

### Benefits of Migration

1. Type safety
2. Better IDE support
3. Automatic validation
4. Consistent API
5. Easier testing

## Testing

The framework includes comprehensive test coverage:

- ✅ Node operations with names
- ✅ Contact edges with metadata
- ✅ Call edges with metadata
- ✅ Edge retrieval with metadata
- ✅ Query operations
- ✅ Call filtering
- ✅ Backward compatibility

## Future Enhancements

1. **Edge Versioning** - Support multiple metadata versions
2. **Metadata Inheritance** - Base metadata with extensions
3. **Custom Validators** - Per-field validation rules
4. **Metadata Migration** - Automatic schema migration
5. **Query Optimization** - Metadata-aware query planning

## Conclusion

The Edge Metadata Framework provides:
- ✅ Robust, type-safe edge metadata
- ✅ Easy extensibility for new edge types
- ✅ Full backward compatibility
- ✅ Production-ready implementation
- ✅ Comprehensive documentation

This framework makes it trivial to add new edge types with any metadata structure while maintaining code quality and type safety.

