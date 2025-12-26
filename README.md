# TrueCaller-like Application

A Go-based application implementing TrueCaller-like features with in-memory data storage and graph-based relationship tracking.

## Features

- **User Management**: CRUD operations for users with phone numbers
- **Contact Management**: Add, update, delete contacts for users
- **Graph-based Relationships**: Track contacts and call history using graph data structures
- **In-Memory Storage**: Fast, thread-safe in-memory repository implementation
- **Comprehensive Testing**: Full unit test coverage for repository layer

## Project Structure

```
credCode/
├── main.go                    # Main application entry point
├── graph_demo.go              # Graph operations demonstration
├── models/
│   ├── user.go               # User and Contact models
│   └── graph.go              # Graph node and edge models
├── repository/
│   ├── user_repository.go    # In-memory user repository implementation
│   ├── user_repository_test.go  # Unit tests for repository
│   ├── graph_repository.go   # Graph repository for relationships
│   ├── seed_data.json        # Seed data with 100 users
│   └── graph_seed_data.json  # Graph seed data with calls and contacts
├── HLD_ARCHITECTURE.md       # High-level design architecture
├── truecaller.md             # Feature requirements and specifications
└── CAYLEY_INTEGRATION.md     # Graph database integration docs
```

## Getting Started

### Prerequisites

- Go 1.25 or higher
- Git

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd credCode
```

2. Install dependencies:
```bash
go mod download
```

3. Run the application:
```bash
go run main.go
```

4. Run tests:
```bash
go test ./repository -v
```

## Seed Data

The project includes comprehensive seed data:
- **100 users** with varied contact relationships
- **107 graph edges** including:
  - Contact relationships (bidirectional)
  - Call history with various scenarios
  - Spam call detection data
  - Long calls, missed calls, and group calls

## Key Components

### User Repository
- Thread-safe in-memory storage
- Fast lookups by ID or phone number
- Full CRUD operations for users and contacts

### Graph Repository
- Relationship tracking between users
- Call history management
- Query capabilities for graph traversal

## Testing

Run all tests:
```bash
go test ./...
```

Run tests with coverage:
```bash
go test ./repository -cover
```

## Architecture

See `HLD_ARCHITECTURE.md` for detailed architecture documentation.

## License

This project is for educational purposes.

