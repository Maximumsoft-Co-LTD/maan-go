# Project Structure

## Root Level
- `main.go` - Public API entry point, re-exports internal types and functions
- `go.mod` / `go.sum` - Go module definition and dependency lock
- `README.md` - Documentation in Thai language with usage examples

## Internal Package (`internal/mongo/`)
Contains the core implementation that is not directly importable by external packages:

- `api.go` - Interface definitions for all public types
- `client.go` - MongoDB client implementation with read/write separation
- `coll.go` - Collection implementation with CRUD operations
- `more-coll.go` - Extended collection with query builder functionality
- `sub-qry.go` - Single and many result query implementations
- `aggregate.go` - Aggregation pipeline implementation
- `transaction-session.go` - Transaction session management
- `model-defaults.go` - Default value population for models
- `fake_client.go` - Mock client for testing without MongoDB

## Testing Files
- `*_test.go` - Unit tests alongside implementation files
- `client_integration_test.go` - Integration tests requiring real MongoDB

## Code Organization Principles

### Package Structure
- **Public API** in root (`main.go`) - Clean interface for library users
- **Implementation** in `internal/mongo` - Hidden from external imports
- **Single responsibility** - Each file handles one major concern

### Naming Conventions
- **Interfaces** - Descriptive names (`Client`, `Collection[T]`, `SingleResult[T]`)
- **Implementations** - Lowercase struct names (`client`, `collection[T]`)
- **Methods** - Short, clear names (`Create`, `FindOne`, `Agg`)
- **Files** - Kebab-case for multi-word concepts (`model-defaults.go`)

### Type Organization
- **Generic types** use `[T any]` constraint for document types
- **Options pattern** for configuration (`WithWriteURI`, `WithDatabase`)
- **Context propagation** - All operations accept or return context-aware types
- **Fluent interfaces** - Methods return same type for chaining

### Import Patterns
- MongoDB driver imported as `mg "go.mongodb.org/mongo-driver/mongo"`
- BSON types used directly from `go.mongodb.org/mongo-driver/bson`
- Internal package imported in main.go for re-export