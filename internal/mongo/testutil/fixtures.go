package testutil

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PropertyTestConfig holds configuration for property-based tests
type PropertyTestConfig struct {
	Iterations int
	MaxShrinks int
	RandomSeed int64
	Timeout    time.Duration
}

// DefaultPropertyTestConfig returns default configuration for property tests
func DefaultPropertyTestConfig() PropertyTestConfig {
	return PropertyTestConfig{
		Iterations: 100,
		MaxShrinks: 100,
		RandomSeed: time.Now().UnixNano(),
		Timeout:    30 * time.Second,
	}
}

// TestScenario defines a test scenario with setup and cleanup
type TestScenario struct {
	Name        string
	Description string
	Setup       func() (interface{}, error)
	Cleanup     func(interface{}) error
}

// FixtureTestDoc creates a fixed test document for unit tests
func FixtureTestDoc() *TestDoc {
	return &TestDoc{
		ID:     primitive.NewObjectID(),
		Name:   "test_document",
		Value:  42,
		Active: true,
	}
}

// FixtureDefaultTestDoc creates a fixed default test document
func FixtureDefaultTestDoc() *DefaultTestDoc {
	now := time.Now().UTC()
	return &DefaultTestDoc{
		ID:        primitive.NewObjectID(),
		Name:      "default_test_document",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// FixtureComplexTestDoc creates a fixed complex test document
func FixtureComplexTestDoc() *ComplexTestDoc {
	return &ComplexTestDoc{
		ID: primitive.NewObjectID(),
		Metadata: map[string]any{
			"version": "1.0",
			"author":  "test",
			"tags":    []string{"test", "fixture"},
		},
		Tags: []string{"integration", "unit", "property"},
		Nested: NestedDoc{
			Field1: "nested_value",
			Field2: 100,
		},
	}
}

// FixtureEmptyTestDoc creates an empty test document
func FixtureEmptyTestDoc() *TestDoc {
	return &TestDoc{}
}

// FixtureTestDocs creates a slice of test documents for bulk operations
func FixtureTestDocs(count int) []*TestDoc {
	docs := make([]*TestDoc, count)
	for i := 0; i < count; i++ {
		docs[i] = &TestDoc{
			ID:     primitive.NewObjectID(),
			Name:   "test_document_" + string(rune('A'+i)),
			Value:  i * 10,
			Active: i%2 == 0,
		}
	}
	return docs
}

// FixtureContextWithValue creates a context with a test value
func FixtureContextWithValue(key, value string) context.Context {
	return context.WithValue(context.Background(), key, value)
}

// FixtureContextWithTimeout creates a context with a timeout
func FixtureContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// FixtureCancelledContext creates a pre-cancelled context
func FixtureCancelledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

// Common test collection names
const (
	TestCollectionName        = "test_collection"
	TestCollectionNameDefault = "test_collection_default"
	TestCollectionNameComplex = "test_collection_complex"
)

// Common test database names
const (
	TestDatabaseName = "test_database"
)

// Common context keys for testing
type ContextKey string

const (
	TestContextKeyRequest = ContextKey("request_id")
	TestContextKeyUser    = ContextKey("user_id")
	TestContextKeySession = ContextKey("session_id")
)

// TestError represents a test error for error handling tests
type TestError struct {
	message string
}

func (e *TestError) Error() string {
	return e.message
}

// NewTestError creates a new test error
func NewTestError(message string) error {
	return &TestError{message: message}
}
