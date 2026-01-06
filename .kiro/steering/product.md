# Product Overview

**maan-go** is a Go library for MongoDB operations with read/write separation support. It provides a fluent API for CRUD operations, aggregation, and transactions with strongly typed collections.

## Key Features

- **Read/Write Separation**: Supports separate MongoDB URIs for read and write operations
- **Type Safety**: Strongly typed collections using Go generics (`Collection[T]`)
- **Fluent API**: Chainable methods for queries, aggregation, and operations
- **Transaction Support**: Both automatic (`WithTx`) and manual (`StartTx`) transaction handling
- **Model Defaults**: Automatic population of default values (ID, timestamps) via interface methods
- **Testing Support**: Includes fake client for unit testing without MongoDB dependency

## Target Use Cases

- Applications requiring MongoDB read/write separation for performance
- Projects needing type-safe database operations
- Systems with complex aggregation pipelines
- Applications requiring transaction support
- Repository pattern implementations

## API Surface

- `Client` - MongoDB connection management with read/write separation
- `Collection[T]` - Primary interface for CRUD and aggregation operations
- `ExtendedCollection[T]` - Query builder with chainable methods
- `SingleResult[T]` / `ManyResult[T]` - Fluent query results
- `Aggregate[T]` - Aggregation pipeline operations
- `TxSession` - Transaction session management