# Unit Testing Design Document

## Overview

This design document outlines the comprehensive unit testing strategy for the maan-go MongoDB library. The testing approach combines traditional unit tests for specific examples and edge cases with property-based testing for universal correctness properties. The design ensures thorough coverage of all public interfaces while maintaining fast execution and independence from external MongoDB instances.

The testing strategy leverages the existing fake client infrastructure and extends it with comprehensive test coverage for all components including collections, query builders, aggregation pipelines, transactions, and model defaults.

## Architecture

### Testing Framework Stack

- **Go Testing Framework**: Standard `testing` package for test execution and assertions
- **Property-Based Testing**: `github.com/leanovate/gopter` for property-based testing with random input generation
- **Test Utilities**: Custom test helpers for common patterns and data generation
- **Fake Client Infrastructure**: Existing `FakeClient` implementation for MongoDB-independent testing

### Test Organization Structure

```
internal/mongo/
├── *_test.go              # Unit tests alongside implementation
├── testutil/              # Shared test utilities and helpers
│   ├── generators.go      # Property test data generators
│   ├── assertions.go      # Custom assertion helpers
│   └── fixtures.go        # Test data fixtures
└── property_tests/        # Property-based test suites
    ├── collection_properties_test.go
    ├── query_properties_test.go
    └── aggregation_properties_test.go
```

### Test Isolation Strategy

- Each test uses independent fake client instances
- Context isolation ensures no cross-test contamination
- Property tests use fresh random seeds for reproducibility
- Test data generators create isolated document instances

## Components and Interfaces

### Core Testing Components

#### Test Data Generators
- **Document Generators**: Create random test documents with various field types
- **Filter Generators**: Generate valid MongoDB filter expressions
- **Pipeline Generators**: Create aggregation pipeline stages
- **Context Generators**: Generate contexts with various values and cancellation

#### Test Utilities
- **Assertion Helpers**: Custom assertions for MongoDB-specific operations
- **Mock Factories**: Factories for creating test clients and collections
- **Comparison Utilities**: Deep equality checks for documents and results
- **Error Matchers**: Pattern matching for expected error conditions

#### Property Test Infrastructure
- **Property Definitions**: Formal specifications of universal properties
- **Test Runners**: Execution framework for property-based tests
- **Shrinking Support**: Minimal failing examples when properties fail
- **Coverage Tracking**: Ensure property tests exercise all code paths

## Data Models

### Test Document Types

```go
// Basic test document for simple operations
type TestDoc struct {
    ID     primitive.ObjectID `bson:"_id"`
    Name   string             `bson:"name"`
    Value  int                `bson:"value"`
    Active bool               `bson:"active"`
}

// Document with default interfaces for testing model defaults
type DefaultTestDoc struct {
    ID        primitive.ObjectID `bson:"_id"`
    Name      string             `bson:"name"`
    CreatedAt time.Time          `bson:"created_at"`
    UpdatedAt time.Time          `bson:"updated_at"`
}

func (d *DefaultTestDoc) DefaultId() primitive.ObjectID { /* implementation */ }
func (d *DefaultTestDoc) DefaultCreatedAt() time.Time { /* implementation */ }
func (d *DefaultTestDoc) DefaultUpdatedAt() time.Time { /* implementation */ }

// Complex nested document for advanced testing
type ComplexTestDoc struct {
    ID       primitive.ObjectID `bson:"_id"`
    Metadata map[string]any     `bson:"metadata"`
    Tags     []string           `bson:"tags"`
    Nested   NestedDoc          `bson:"nested"`
}

type NestedDoc struct {
    Field1 string `bson:"field1"`
    Field2 int    `bson:"field2"`
}
```

### Test Configuration Models

```go
// Configuration for property test execution
type PropertyTestConfig struct {
    Iterations    int
    MaxShrinks    int
    RandomSeed    int64
    Timeout       time.Duration
}

// Test scenario definitions
type TestScenario struct {
    Name        string
    Description string
    Setup       func() (Client, error)
    Cleanup     func(Client) error
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property Reflection

After analyzing all acceptance criteria, several properties can be consolidated to eliminate redundancy:

- Properties 3.1, 3.2, and 5.4 all deal with nil/invalid input handling and can be combined into a comprehensive input validation property
- Properties 1.5, 3.4, 4.4, and 6.5 all address concurrent access and can be unified into a thread safety property
- Properties 4.2 and 4.5 both verify fake client compatibility and can be merged
- Properties 6.1 and 6.3 both test transaction error handling and can be combined

### Core Properties

Property 1: Generic type safety preservation
*For any* valid Go type T, creating a Collection[T] and performing operations should maintain type safety throughout the operation chain
**Validates: Requirements 1.3**

Property 2: Fluent API method chaining consistency
*For any* collection operation chain, each method should return the correct interface type allowing further chaining
**Validates: Requirements 1.4**

Property 3: Thread safety under concurrent access
*For any* collection instance, concurrent access from multiple goroutines should not cause race conditions or data corruption
**Validates: Requirements 1.5, 3.4, 4.4**

Property 4: CRUD round-trip consistency
*For any* valid document, creating it in a collection and then reading it back should return an equivalent document
**Validates: Requirements 2.1**

Property 5: Query filter composition correctness
*For any* set of valid filters, composing them using query builders should produce logically correct combined filters
**Validates: Requirements 2.2**

Property 6: Aggregation pipeline stage ordering
*For any* aggregation pipeline, stages should be applied in the specified order maintaining data flow semantics
**Validates: Requirements 2.3, 8.1**

Property 7: Context propagation preservation
*For any* context with values, passing it through operation chains should preserve all context values
**Validates: Requirements 2.4**

Property 8: Collection instance isolation
*For any* collection, calling Ctx() with different contexts should create independent instances without affecting the original
**Validates: Requirements 2.5**

Property 9: Graceful nil and invalid input handling
*For any* collection method, providing nil or invalid inputs should handle gracefully without panics and return appropriate errors
**Validates: Requirements 3.1, 3.2, 5.4**

Property 10: Empty result set handling
*For any* query that matches no documents, the operation should return empty results without errors
**Validates: Requirements 3.3**

Property 11: Resource cleanup consistency
*For any* client or session, calling cleanup methods should properly release all associated resources
**Validates: Requirements 3.5, 6.4**

Property 12: Fake client configuration correctness
*For any* fake client configuration options, the resulting client should behave according to the specified configuration
**Validates: Requirements 4.1**

Property 13: Fake client API compatibility
*For any* operation available on real clients, fake clients should provide the same interface and behavior patterns
**Validates: Requirements 4.2, 4.5**

Property 14: Fake client lifecycle management
*For any* fake client, initialization and cleanup should follow the same patterns as real clients
**Validates: Requirements 4.3**

Property 15: Automatic ID generation consistency
*For any* document implementing DefaultId interface, creating the document should generate an ID when none exists
**Validates: Requirements 5.1**

Property 16: Automatic timestamp population
*For any* document implementing timestamp interfaces, creating or updating should set appropriate timestamps
**Validates: Requirements 5.2, 5.3**

Property 17: Default interface method invocation
*For any* document with default interfaces, the appropriate interface methods should be called during operations
**Validates: Requirements 5.5**

Property 18: Transaction commit and rollback semantics
*For any* transaction operation, successful completion should commit while errors should trigger rollback
**Validates: Requirements 6.1, 6.3**

Property 19: Manual transaction context propagation
*For any* manually started transaction, the session context should be properly propagated through all operations
**Validates: Requirements 6.2**

Property 20: Result streaming type correctness
*For any* streaming operation, callback functions should receive documents of the correct type
**Validates: Requirements 7.1**

Property 21: Streaming error handling
*For any* streaming operation with errors, proper error propagation should occur to callers
**Validates: Requirements 7.2, 7.5**

Property 22: Context cancellation during streaming
*For any* streaming operation with cancelled context, the operation should terminate gracefully
**Validates: Requirements 7.3**

Property 23: Aggregation option application
*For any* aggregation with options, the specified options should be properly applied to the operation
**Validates: Requirements 8.2**

Property 24: Aggregation result type handling
*For any* aggregation operation, both typed and raw results should be handled correctly
**Validates: Requirements 8.3**

Property 25: Aggregation streaming type conversion
*For any* aggregation streaming operation, documents should be properly converted to the target type
**Validates: Requirements 8.4**

Property 26: Aggregation error condition handling
*For any* aggregation operation with errors, appropriate error handling should occur
**Validates: Requirements 8.5**

<function_calls>
<invoke name="prework">
<parameter name="featureName">unit-testing

## Error Handling

### Error Classification Strategy

#### Input Validation Errors
- **Nil Pointer Errors**: Graceful handling when nil documents or pointers are provided
- **Type Mismatch Errors**: Clear error messages when incorrect types are used
- **Invalid Filter Errors**: Appropriate error propagation for malformed MongoDB filters

#### Resource Management Errors
- **Connection Errors**: Proper error handling for client connection failures
- **Session Errors**: Clear error messages for transaction session issues
- **Cleanup Errors**: Aggregated error reporting when multiple cleanup operations fail

#### Operation Errors
- **Query Errors**: Meaningful error messages for failed queries
- **Aggregation Errors**: Detailed error information for pipeline failures
- **Transaction Errors**: Clear distinction between commit and rollback errors

### Error Testing Strategy

- **Property-Based Error Testing**: Generate invalid inputs to verify error handling
- **Error Propagation Testing**: Ensure errors bubble up correctly through operation chains
- **Error Message Testing**: Verify error messages are informative and actionable
- **Recovery Testing**: Test system behavior after error conditions

## Testing Strategy

### Dual Testing Approach

The testing strategy employs both unit testing and property-based testing to provide comprehensive coverage:

#### Unit Testing Approach
- **Specific Examples**: Test concrete scenarios with known inputs and expected outputs
- **Edge Cases**: Test boundary conditions, empty inputs, and error scenarios
- **Integration Points**: Test interactions between components
- **Regression Tests**: Prevent previously fixed bugs from reoccurring

Unit tests cover:
- Specific configuration scenarios
- Known edge cases and error conditions
- Interface compliance verification
- Concrete examples of expected behavior

#### Property-Based Testing Approach
- **Universal Properties**: Test properties that should hold across all valid inputs
- **Random Input Generation**: Use generators to create diverse test scenarios
- **Shrinking**: Automatically find minimal failing examples when properties fail
- **High Iteration Count**: Run each property test with 100+ iterations for thorough coverage

Property-based tests cover:
- Type safety across different generic parameters
- Correctness of operations with random valid inputs
- Invariants that should hold regardless of specific data
- Behavioral consistency across operation chains

### Testing Framework Configuration

#### Property-Based Testing Setup
- **Library**: `github.com/leanovate/gopter` for Go property-based testing
- **Iterations**: Minimum 100 iterations per property test
- **Shrinking**: Enabled to find minimal failing examples
- **Seed Management**: Reproducible random seeds for debugging

#### Test Execution Requirements
- **Coverage Target**: Minimum 90% code coverage across all public interfaces
- **Performance Target**: All tests complete within 30 seconds
- **Isolation**: No external dependencies (MongoDB, network, filesystem)
- **Concurrency**: Tests must pass with race detection enabled

#### Test Organization
- **Co-location**: Unit tests alongside implementation files using `*_test.go` suffix
- **Property Tests**: Separate directory structure for property-based test suites
- **Utilities**: Shared test utilities in dedicated `testutil` package
- **Generators**: Reusable data generators for property tests

### Test Data Generation Strategy

#### Document Generators
- **Basic Documents**: Simple structs with primitive fields
- **Complex Documents**: Nested structures with arrays and maps
- **Default Interface Documents**: Structs implementing default value interfaces
- **Edge Case Documents**: Empty, nil, and boundary value scenarios

#### Filter Generators
- **Simple Filters**: Basic equality and comparison operations
- **Complex Filters**: Nested logical operations ($and, $or, $not)
- **Invalid Filters**: Malformed expressions for error testing
- **Edge Filters**: Empty filters and boundary conditions

#### Context Generators
- **Value Contexts**: Contexts with various key-value pairs
- **Cancelled Contexts**: Pre-cancelled contexts for cancellation testing
- **Timeout Contexts**: Contexts with various timeout durations
- **Background Contexts**: Standard background contexts

### Coverage and Quality Metrics

#### Code Coverage Requirements
- **Public Interface Coverage**: 90% minimum coverage of all exported functions and methods
- **Branch Coverage**: Ensure all conditional branches are tested
- **Error Path Coverage**: Test all error handling paths
- **Generic Type Coverage**: Test with multiple type parameters

#### Quality Assurance
- **Race Detection**: All tests must pass with `-race` flag
- **Memory Leak Detection**: Verify proper resource cleanup
- **Performance Benchmarks**: Ensure test execution remains fast
- **Documentation Coverage**: All public APIs have corresponding tests

### Continuous Integration Integration

#### Test Execution Pipeline
- **Unit Tests**: Run standard unit tests with coverage reporting
- **Property Tests**: Execute property-based tests with extended iteration counts
- **Race Testing**: Run all tests with race detection enabled
- **Coverage Analysis**: Generate and validate coverage reports

#### Quality Gates
- **Coverage Threshold**: Fail builds below 90% coverage
- **Performance Threshold**: Fail builds with tests taking over 30 seconds
- **Race Condition Detection**: Fail builds with race conditions
- **Property Test Failures**: Fail builds when any property test fails