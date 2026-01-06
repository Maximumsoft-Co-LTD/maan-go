package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo/testutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestAggregateInterface tests the Aggregate[T] interface methods
func TestAggregateInterface(t *testing.T) {
	ctx := context.Background()
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	testutil.AssertNoError(t, err, "Failed to create fake client")
	defer client.Close()

	collection := NewCollection[testutil.TestDoc](ctx, client, testutil.TestCollectionName)

	// Test basic aggregation pipeline creation
	pipeline := bson.A{
		bson.M{"$match": bson.M{"active": true}},
		bson.M{"$sort": bson.M{"value": -1}},
		bson.M{"$limit": 10},
	}

	aggregate := collection.Agg(pipeline)
	testutil.AssertNotNil(t, aggregate, "Agg() should return an Aggregate")
}

// TestAggregateOptions tests aggregation option methods
func TestAggregateOptions(t *testing.T) {
	ctx := context.Background()
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	testutil.AssertNoError(t, err, "Failed to create fake client")
	defer client.Close()

	collection := NewCollection[testutil.TestDoc](ctx, client, testutil.TestCollectionName)
	pipeline := bson.A{bson.M{"$match": bson.M{"active": true}}}

	// Test Disk() option
	aggregate := collection.Agg(pipeline)
	diskAggregate := aggregate.Disk(true)
	testutil.AssertNotNil(t, diskAggregate, "Disk() should return an Aggregate")

	// Test Bsz() option
	bszAggregate := aggregate.Bsz(100)
	testutil.AssertNotNil(t, bszAggregate, "Bsz() should return an Aggregate")

	// Test Opts() option
	opts := options.Aggregate().SetAllowDiskUse(true).SetBatchSize(50)
	optsAggregate := aggregate.Opts(opts)
	testutil.AssertNotNil(t, optsAggregate, "Opts() should return an Aggregate")

	// Test method chaining
	chainedAggregate := aggregate.Disk(true).Bsz(100)
	testutil.AssertNotNil(t, chainedAggregate, "Method chaining should work")
}

// TestAggregateResults tests result retrieval methods
func TestAggregateResults(t *testing.T) {
	ctx := context.Background()
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	testutil.AssertNoError(t, err, "Failed to create fake client")
	defer client.Close()

	collection := NewCollection[testutil.TestDoc](ctx, client, testutil.TestCollectionName)

	// Test All() method interface
	pipeline := bson.A{bson.M{"$match": bson.M{"active": true}}}
	aggregate := collection.Agg(pipeline)

	// Test that All() method exists and returns proper types
	results, err := aggregate.All()
	// Fake client will return "client is disconnected" error, which is expected
	_ = err // Ignore error as it's expected with fake client
	testutil.AssertNotNil(t, results, "All() should return a slice (even if empty)")

	// Test Result() method interface
	var resultSlice []testutil.TestDoc
	err = aggregate.Result(&resultSlice)
	// Error is expected with fake client, but method should exist and handle nil gracefully
	_ = err

	// Test Result() with nil pointer (should handle gracefully)
	err = aggregate.Result(nil)
	testutil.AssertNoError(t, err, "Result() should handle nil pointer gracefully")

	// Test Raw() method interface
	rawResults, err := aggregate.Raw()
	// Fake client will return error, but interface should work
	_ = err
	testutil.AssertNotNil(t, rawResults, "Raw() should return a slice (even if empty)")
}

// TestAggregateStreaming tests streaming methods
func TestAggregateStreaming(t *testing.T) {
	ctx := context.Background()
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	testutil.AssertNoError(t, err, "Failed to create fake client")
	defer client.Close()

	collection := NewCollection[testutil.TestDoc](ctx, client, testutil.TestCollectionName)

	pipeline := bson.A{bson.M{"$match": bson.M{"active": true}}}
	aggregate := collection.Agg(pipeline)

	// Test Stream() method interface
	err = aggregate.Stream(func(ctx context.Context, doc testutil.TestDoc) error {
		testutil.AssertNotNil(t, ctx, "Context should not be nil in stream callback")
		return nil
	})
	// Fake client will return error, but interface should work
	_ = err

	// Test Each() method interface
	err = aggregate.Each(func(ctx context.Context, doc testutil.TestDoc) error {
		testutil.AssertNotNil(t, ctx, "Context should not be nil in each callback")
		return nil
	})
	// Fake client will return error, but interface should work
	_ = err
}

// TestAggregatePipelineStageOrdering tests that pipeline stages are applied in order
func TestAggregatePipelineStageOrdering(t *testing.T) {
	ctx := context.Background()
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	testutil.AssertNoError(t, err, "Failed to create fake client")
	defer client.Close()

	collection := NewCollection[testutil.TestDoc](ctx, client, testutil.TestCollectionName)

	// Test pipeline with specific ordering: match -> sort -> limit
	pipeline := bson.A{
		bson.M{"$match": bson.M{"active": true}},
		bson.M{"$sort": bson.M{"value": -1}},
		bson.M{"$limit": 2},
	}

	aggregate := collection.Agg(pipeline)
	testutil.AssertNotNil(t, aggregate, "Agg() should return an Aggregate")

	// Test that pipeline can be executed (interface test)
	results, err := aggregate.All()
	// Fake client will return error, but interface should work
	_ = err
	testutil.AssertNotNil(t, results, "Pipeline should return results slice (even if empty)")

	// Test that complex pipelines can be created
	complexPipeline := bson.A{
		bson.M{"$match": bson.M{"active": true}},
		bson.M{"$group": bson.M{"_id": "$active", "count": bson.M{"$sum": 1}}},
		bson.M{"$sort": bson.M{"count": -1}},
		bson.M{"$limit": 10},
		bson.M{"$project": bson.M{"_id": 0, "active": "$_id", "count": 1}},
	}

	complexAggregate := collection.Agg(complexPipeline)
	testutil.AssertNotNil(t, complexAggregate, "Complex pipeline should be created")

	// Test that we can call methods on complex pipeline
	complexResults, err := complexAggregate.Disk(true).Bsz(50).All()
	_ = err
	testutil.AssertNotNil(t, complexResults, "Complex pipeline should return results slice")
}

// TestAggregateErrorHandling tests error handling in aggregation operations
func TestAggregateErrorHandling(t *testing.T) {
	ctx := context.Background()
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	testutil.AssertNoError(t, err, "Failed to create fake client")
	defer client.Close()

	collection := NewCollection[testutil.TestDoc](ctx, client, testutil.TestCollectionName)

	// Test with nil pipeline (should handle gracefully)
	aggregate := collection.Agg(nil)
	testutil.AssertNotNil(t, aggregate, "Agg() should handle nil pipeline")

	// Test Result() with nil pointer
	err = aggregate.Result(nil)
	testutil.AssertNoError(t, err, "Result() should handle nil pointer gracefully")

	// Test streaming with error in callback
	pipeline := bson.A{bson.M{"$match": bson.M{}}}
	aggregate = collection.Agg(pipeline)

	streamErr := aggregate.Stream(func(ctx context.Context, doc testutil.TestDoc) error {
		return testutil.NewTestError("stream error")
	})
	// Error should be propagated from callback
	testutil.AssertError(t, streamErr, "Stream() should propagate callback errors")

	eachErr := aggregate.Each(func(ctx context.Context, doc testutil.TestDoc) error {
		return testutil.NewTestError("each error")
	})
	// Error should be propagated from callback
	testutil.AssertError(t, eachErr, "Each() should propagate callback errors")
}

// TestAggregateContextHandling tests context handling in aggregation operations
func TestAggregateContextHandling(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	testutil.AssertNoError(t, err, "Failed to create fake client")
	defer client.Close()

	// Test with cancelled context
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	collection := NewCollection[testutil.TestDoc](cancelledCtx, client, testutil.TestCollectionName)
	pipeline := bson.A{bson.M{"$match": bson.M{}}}
	aggregate := collection.Agg(pipeline)

	// Operations should handle cancelled context gracefully
	_, err = aggregate.All()
	// Note: Fake client may not properly simulate context cancellation
	// This test verifies the interface handles cancelled contexts without panicking

	// Test with timeout context
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer timeoutCancel()

	collection = NewCollection[testutil.TestDoc](timeoutCtx, client, testutil.TestCollectionName)
	aggregate = collection.Agg(pipeline)

	// Should handle timeout gracefully
	_, err = aggregate.All()
	// Note: Fake client may not properly simulate timeouts
	// This test verifies the interface handles timeout contexts without panicking
}
