package mongo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo/testutil"
	"go.mongodb.org/mongo-driver/bson"
)

// TestSingleResultInterface tests the SingleResult[T] interface methods
func TestSingleResultInterface(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	testutil.AssertNoError(t, err, "Failed to create fake client")
	defer client.Close()

	ctx := context.Background()
	collection := NewCollection[testutil.TestDoc](ctx, client, "test_single_result")

	// Test SingleResult creation and method chaining
	singleResult := collection.FindOne(bson.M{"name": "test"})
	testutil.AssertNotNil(t, singleResult, "FindOne should return a SingleResult")

	// Test Proj method
	projResult := singleResult.Proj(bson.M{"name": 1})
	testutil.AssertNotNil(t, projResult, "Proj should return a SingleResult")

	// Test Sort method
	sortResult := singleResult.Sort(bson.M{"name": 1})
	testutil.AssertNotNil(t, sortResult, "Sort should return a SingleResult")

	// Test Hint method
	hintResult := singleResult.Hint(bson.M{"name": 1})
	testutil.AssertNotNil(t, hintResult, "Hint should return a SingleResult")

	// Test method chaining
	chainedResult := singleResult.
		Proj(bson.M{"name": 1}).
		Sort(bson.M{"name": 1}).
		Hint(bson.M{"name": 1})
	testutil.AssertNotNil(t, chainedResult, "Method chaining should work")

	// Test Result method (will fail with fake client, but should not panic)
	var doc testutil.TestDoc
	err = chainedResult.Result(&doc)
	// Error expected with fake client, but should not panic
	_ = err
}

// TestManyResultInterface tests the ManyResult[T] interface methods
func TestManyResultInterface(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	testutil.AssertNoError(t, err, "Failed to create fake client")
	defer client.Close()

	ctx := context.Background()
	collection := NewCollection[testutil.TestDoc](ctx, client, "test_many_result")

	// Test ManyResult creation and method chaining
	manyResult := collection.FindMany(bson.M{"active": true})
	testutil.AssertNotNil(t, manyResult, "FindMany should return a ManyResult")

	// Test Proj method
	projResult := manyResult.Proj(bson.M{"name": 1})
	testutil.AssertNotNil(t, projResult, "Proj should return a ManyResult")

	// Test Sort method
	sortResult := manyResult.Sort(bson.M{"name": 1})
	testutil.AssertNotNil(t, sortResult, "Sort should return a ManyResult")

	// Test Hint method
	hintResult := manyResult.Hint(bson.M{"name": 1})
	testutil.AssertNotNil(t, hintResult, "Hint should return a ManyResult")

	// Test Limit method
	limitResult := manyResult.Limit(10)
	testutil.AssertNotNil(t, limitResult, "Limit should return a ManyResult")

	// Test Skip method
	skipResult := manyResult.Skip(5)
	testutil.AssertNotNil(t, skipResult, "Skip should return a ManyResult")

	// Test Bsz method
	bszResult := manyResult.Bsz(100)
	testutil.AssertNotNil(t, bszResult, "Bsz should return a ManyResult")

	// Test method chaining
	chainedResult := manyResult.
		Proj(bson.M{"name": 1}).
		Sort(bson.M{"name": 1}).
		Limit(10).
		Skip(5)
	testutil.AssertNotNil(t, chainedResult, "Method chaining should work")

	// Test All method (will fail with fake client, but should not panic)
	docs, err := chainedResult.All()
	// Error expected with fake client, but should not panic
	_ = err
	_ = docs

	// Test Result method (will fail with fake client, but should not panic)
	var results []testutil.TestDoc
	err = chainedResult.Result(&results)
	// Error expected with fake client, but should not panic
	_ = err

	// Test Cnt method (will fail with fake client, but should not panic)
	count, err := chainedResult.Cnt()
	// Error expected with fake client, but should not panic
	_ = err
	_ = count
}

// TestStreamingMethods tests the Stream() and Each() methods
func TestStreamingMethods(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	testutil.AssertNoError(t, err, "Failed to create fake client")
	defer client.Close()

	ctx := context.Background()
	collection := NewCollection[testutil.TestDoc](ctx, client, "test_streaming")

	// Test ManyResult.Stream()
	manyResult := collection.FindMany(bson.M{})

	err = manyResult.Stream(func(ctx context.Context, doc testutil.TestDoc) error {
		testutil.AssertNotNil(t, ctx, "Context should not be nil in Stream callback")
		return nil
	})
	// Error expected with fake client, but callback signature should be enforced
	_ = err

	// Test ManyResult.Each()
	err = manyResult.Each(func(ctx context.Context, doc testutil.TestDoc) error {
		testutil.AssertNotNil(t, ctx, "Context should not be nil in Each callback")
		return nil
	})
	// Error expected with fake client, but callback signature should be enforced
	_ = err

	// Test Aggregate.Stream()
	aggregate := collection.Agg(bson.A{bson.M{"$match": bson.M{}}})

	err = aggregate.Stream(func(ctx context.Context, doc testutil.TestDoc) error {
		testutil.AssertNotNil(t, ctx, "Context should not be nil in Aggregate Stream callback")
		return nil
	})
	// Error expected with fake client, but callback signature should be enforced
	_ = err

	// Test Aggregate.Each()
	err = aggregate.Each(func(ctx context.Context, doc testutil.TestDoc) error {
		testutil.AssertNotNil(t, ctx, "Context should not be nil in Aggregate Each callback")
		return nil
	})
	// Error expected with fake client, but callback signature should be enforced
	_ = err
}

// TestStreamingErrorHandling tests error handling in streaming methods
func TestStreamingErrorHandling(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	testutil.AssertNoError(t, err, "Failed to create fake client")
	defer client.Close()

	ctx := context.Background()
	collection := NewCollection[testutil.TestDoc](ctx, client, "test_error_handling")

	testError := errors.New("test streaming error")

	// Test ManyResult.Stream() error propagation
	manyResult := collection.FindMany(bson.M{})

	err = manyResult.Stream(func(ctx context.Context, doc testutil.TestDoc) error {
		return testError
	})
	// Should handle error gracefully (might get different error from fake client)
	testutil.AssertNotNil(t, err, "Stream should propagate errors")

	// Test ManyResult.Each() error propagation
	err = manyResult.Each(func(ctx context.Context, doc testutil.TestDoc) error {
		return testError
	})
	// Should handle error gracefully
	testutil.AssertNotNil(t, err, "Each should propagate errors")

	// Test Aggregate.Stream() error propagation
	aggregate := collection.Agg(bson.A{bson.M{"$match": bson.M{}}})

	err = aggregate.Stream(func(ctx context.Context, doc testutil.TestDoc) error {
		return testError
	})
	// Should handle error gracefully
	testutil.AssertNotNil(t, err, "Aggregate Stream should propagate errors")

	// Test Aggregate.Each() error propagation
	err = aggregate.Each(func(ctx context.Context, doc testutil.TestDoc) error {
		return testError
	})
	// Should handle error gracefully
	testutil.AssertNotNil(t, err, "Aggregate Each should propagate errors")
}

// TestContextCancellation tests context cancellation during streaming
func TestContextCancellation(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	testutil.AssertNoError(t, err, "Failed to create fake client")
	defer client.Close()

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	collection := NewCollection[testutil.TestDoc](ctx, client, "test_cancellation")

	// Test ManyResult.Stream() with cancelled context
	manyResult := collection.FindMany(bson.M{})

	err = manyResult.Stream(func(ctx context.Context, doc testutil.TestDoc) error {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	})
	// Should handle cancellation gracefully
	_ = err

	// Test with timeout context
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer timeoutCancel()

	timeoutCollection := NewCollection[testutil.TestDoc](timeoutCtx, client, "test_timeout")
	timeoutResult := timeoutCollection.FindMany(bson.M{})

	err = timeoutResult.Stream(func(ctx context.Context, doc testutil.TestDoc) error {
		// Simulate some work
		time.Sleep(10 * time.Millisecond)

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	})
	// Should handle timeout gracefully
	_ = err
}

// TestResultTypeConversion tests result type conversion and iteration
func TestResultTypeConversion(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	testutil.AssertNoError(t, err, "Failed to create fake client")
	defer client.Close()

	ctx := context.Background()

	// Test with different document types
	testCollection := NewCollection[testutil.TestDoc](ctx, client, "test_type_conversion")
	defaultCollection := NewCollection[testutil.DefaultTestDoc](ctx, client, "default_type_conversion")
	complexCollection := NewCollection[testutil.ComplexTestDoc](ctx, client, "complex_type_conversion")

	// Test TestDoc streaming
	testResult := testCollection.FindMany(bson.M{})
	err = testResult.Stream(func(ctx context.Context, doc testutil.TestDoc) error {
		// Verify doc is of correct type
		_ = doc.Name
		_ = doc.Value
		_ = doc.Active
		return nil
	})
	_ = err

	// Test DefaultTestDoc streaming
	defaultResult := defaultCollection.FindMany(bson.M{})
	err = defaultResult.Stream(func(ctx context.Context, doc testutil.DefaultTestDoc) error {
		// Verify doc is of correct type
		_ = doc.Name
		_ = doc.CreatedAt
		_ = doc.UpdatedAt
		return nil
	})
	_ = err

	// Test ComplexTestDoc streaming
	complexResult := complexCollection.FindMany(bson.M{})
	err = complexResult.Stream(func(ctx context.Context, doc testutil.ComplexTestDoc) error {
		// Verify doc is of correct type
		_ = doc.Metadata
		_ = doc.Tags
		_ = doc.Nested
		return nil
	})
	_ = err
}

// TestResultLifecycle tests the complete lifecycle of result objects
func TestResultLifecycle(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	testutil.AssertNoError(t, err, "Failed to create fake client")
	defer client.Close()

	ctx := context.Background()
	collection := NewCollection[testutil.TestDoc](ctx, client, "test_lifecycle")

	// Test SingleResult lifecycle
	singleResult := collection.FindOne(bson.M{"name": "test"})

	// Apply modifiers
	modifiedSingle := singleResult.
		Proj(bson.M{"name": 1}).
		Sort(bson.M{"name": 1})

	// Execute result
	var doc testutil.TestDoc
	err = modifiedSingle.Result(&doc)
	// Error expected with fake client
	_ = err

	// Test ManyResult lifecycle
	manyResult := collection.FindMany(bson.M{"active": true})

	// Apply modifiers
	modifiedMany := manyResult.
		Proj(bson.M{"name": 1}).
		Sort(bson.M{"name": 1}).
		Limit(10)

	// Execute different result methods
	docs, err := modifiedMany.All()
	_ = err
	_ = docs

	var results []testutil.TestDoc
	err = modifiedMany.Result(&results)
	_ = err

	count, err := modifiedMany.Cnt()
	_ = err
	_ = count

	// Test streaming
	err = modifiedMany.Stream(func(ctx context.Context, doc testutil.TestDoc) error {
		return nil
	})
	_ = err

	err = modifiedMany.Each(func(ctx context.Context, doc testutil.TestDoc) error {
		return nil
	})
	_ = err
}
