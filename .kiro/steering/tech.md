# Technology Stack

## Language & Runtime
- **Go 1.22.5** - Primary language with generics support required
- Uses Go modules for dependency management

## Dependencies
- **go.mongodb.org/mongo-driver v1.17.4** - Official MongoDB Go driver
- Standard library packages for context, time, errors handling

## Build System
- Standard Go toolchain (`go build`, `go test`, `go mod`)
- No additional build tools or frameworks required

## Common Commands

### Development
```bash
# Install dependencies
go mod tidy

# Build the library
go build ./...

# Run unit tests
go test ./...

# Run integration tests (requires MongoDB)
MONGO_INTEGRATION_URI="mongodb://localhost:27017" go test ./internal/mongo -run ClientRoundTrip

# Run tests with verbose output
go test -v ./...

# Check for race conditions
go test -race ./...
```

### Code Quality
```bash
# Format code
go fmt ./...

# Vet code for issues
go vet ./...

# Run linter (if golangci-lint is installed)
golangci-lint run
```

## Architecture Patterns
- **Interface-based design** - Core functionality exposed through interfaces
- **Generics** - Type-safe collections using `Collection[T]` pattern
- **Functional options** - Configuration via `Option` functions
- **Context propagation** - All operations accept `context.Context`
- **Fluent API** - Method chaining for query building
- **Separation of concerns** - Read/write clients handled separately