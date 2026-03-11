# CLAUDE.md — maan-go MongoDB Library

## Project Overview

**maan-go** is a Go library for MongoDB operations with read/write separation. It provides a fluent, type-safe API using Go generics for CRUD, aggregation, transactions, and real-time change streams.

## Common Commands

```bash
# Build
go build ./...

# Unit tests
go test ./...

# Integration tests (requires MongoDB replica set)
MONGO_INTEGRATION_URI="mongodb://localhost:27017" go test ./internal/mongo -run ClientRoundTrip

# Race detection
go test -race ./...

# Format / vet
go fmt ./...
go vet ./...
```

## Package Structure

### Root
- `main.go` – Public API entry point; re-exports types and functions from `internal/mongo`
- `go.mod` / `go.sum` – Module definition and dependency lock

### `internal/mongo/`
- `api.go` – All interface definitions (`Client`, `Collection[T]`, `ExtendedCollection[T]`, `SingleResult[T]`, `ManyResult[T]`, `Aggregate[T]`, `TxSession`, `ChangeStream[T]`, etc.)
- `client.go` – MongoDB client with read/write separation (`WithWriteURI`, `WithReadURI`, `WithDatabase`)
- `coll.go` – `collection[T]` implements `Collection[T]`; CRUD operations (`Create`, `FindOne`, `Find`, `Upd`, `UpdMany`, `Del`, `DelMany`, `Agg`, `Watch`)
- `more-coll.go` – `extendedCollection[T]` implements `ExtendedCollection[T]`; dynamic query builder (`By`, `Where`, `First`, `Many`, `Count`, `Exists`)
- `sub-qry.go` – `singleResult[T]` and `manyResult[T]`; fluent result types
- `aggregate.go` – `aggregate[T]` implements `Aggregate[T]`; aggregation pipeline operations
- `transaction-session.go` – `txSession` implements `TxSession`; auto (`WithTx`) and manual (`StartTx`) transaction handling
- `change-stream.go` – `changeStream[T]` implements `ChangeStream[T]`; fluent builder for watching real-time events with operation filtering (`OnIst/OnUpd/OnDel/OnRep`), full document lookup, and resume token support. Requires MongoDB replica set.
- `model-defaults.go` – Auto-populates model fields (ID, `CreatedAt`, `UpdatedAt`) via interface methods
- `fake_client.go` – `FakeClient` for unit testing without a real MongoDB instance

## Key Design Patterns

**Interface-based design**: All public surface is interfaces defined in `api.go`; concrete types are unexported.

**Generics**: `Collection[T any]` pattern ensures compile-time type safety across all operations.

**Functional options**: Client is configured via `Option` functions (`WithWriteURI`, `WithDatabase`, etc.).

**Fluent API / builder pattern**: Methods return the same interface type for chaining (e.g., `coll.Ctx(ctx).FindOne(filter).Exec()`).

**ExtendedCollection bridge**: `coll.Build()` converts a `Collection[T]` to `ExtendedCollection[T]` for access to the dynamic query builder (`By`, `Where`, `First`, `Many`, `Count`, `Exists`).

**Transaction patterns**: Use `client.WithTx(ctx, fn)` for automatic commit/rollback, or `client.StartTx(ctx)` for manual session control.

**Model defaults**: Implement `DefaultId()`, `DefaultCreatedAt()`, `DefaultUpdatedAt()` on a document struct and the library will auto-populate those fields on create/update.

## Testing Conventions

- Unit tests co-located alongside implementation files (`*_test.go`)
- Integration tests in `client_integration_test.go` (require real MongoDB)
- All tests must pass with `-race` flag
- Use `FakeClient` / `FakeCollection` for unit tests — no real MongoDB required
- Shared test helpers are in `internal/mongo/testutil/`:
  - `assertions.go` – `AssertEqual`, `AssertNil`, `AssertNoError`, `AssertContains`, etc.
  - `fixtures.go` – `PropertyTestConfig` (100 iterations default), `FixtureTestDoc`, `FixtureTestDocs`, `TestScenario`
  - `generators.go` – gopter generators for property-based tests
- Property-based tests use `github.com/leanovate/gopter` with ≥100 iterations per property

## Architecture Notes

- **Read/Write Separation**: Pass `WithWriteURI` and optionally `WithReadURI`; read operations automatically route to the read replica
- **Context propagation**: Every operation accepts `context.Context`; use `.Ctx(ctx)` on collections to bind a context
- **Change Streams**: `collection.Watch(ctx, pipeline)` returns a `ChangeStream[T]`; requires a MongoDB replica set or Atlas cluster
- **Import alias**: MongoDB driver is imported as `mg "go.mongodb.org/mongo-driver/mongo"` throughout the internal package
