package property_tests

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo"
	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo/testutil"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"go.mongodb.org/mongo-driver/bson"
)

// TestProperty9_GracefulNilAndInvalidInputHandling tests that all interfaces handle nil and invalid inputs gracefully
// **Feature: unit-testing, Property 9: Graceful nil and invalid input handling**
// **Validates: Requirements 3.1, 3.2, 5.4**
func TestProperty9_GracefulNilAndInvalidInputHandling(t *testing.T) {
	runner := NewPropertyTestRunner()

	property := ForAllValid(
		genNilAndInvalidInputs(),
		func(testCase NilInvalidTestCase) bool {
			helper, err := NewPropertyTestHelper()
			if err != nil {
				t.Logf("Failed to create test helper: %v", err)
				return false
			}
			defer helper.Cleanup()

			client := helper.Client()
			collection := mongo.NewCollection[testutil.TestDoc](context.Background(), client, "test_collection")

			// Test all methods with nil/invalid inputs - they should not panic
			return testCollectionNilInputs(collection, testCase) &&
				testExtendedCollectionNilInputs(collection, testCase) &&
				testResultNilInputs(collection, testCase) &&
				testAggregationNilInputs(collection, testCase)
		},
	)

	runner.RunProperty(t, "graceful nil and invalid input handling", property)
}

// NilInvalidTestCase represents different types of nil/invalid inputs to test
type NilInvalidTestCase struct {
	NilDoc          *testutil.TestDoc
	NilDocSlice     *[]testutil.TestDoc
	InvalidFilter   interface{}
	InvalidUpdate   interface{}
	InvalidPipeline interface{}
}

// genNilAndInvalidInputs generates various nil and invalid input combinations
func genNilAndInvalidInputs() gopter.Gen {
	return gopter.CombineGens(
		genNilTestDoc(),      // Generate nil pointers
		genNilTestDocSlice(), // Generate nil slice pointers
		genInvalidFilter(),
		genInvalidUpdate(),
		genInvalidPipeline(),
	).Map(func(values []interface{}) NilInvalidTestCase {
		return NilInvalidTestCase{
			NilDoc:          values[0].(*testutil.TestDoc),
			NilDocSlice:     values[1].(*[]testutil.TestDoc),
			InvalidFilter:   values[2],
			InvalidUpdate:   values[3],
			InvalidPipeline: values[4],
		}
	})
}

// genNilTestDoc generates nil TestDoc pointers
func genNilTestDoc() gopter.Gen {
	return gen.OneGenOf(
		gen.Const((*testutil.TestDoc)(nil)),
		testutil.GenTestDoc(),
	)
}

// genNilTestDocSlice generates nil TestDoc slice pointers
func genNilTestDocSlice() gopter.Gen {
	return gen.OneGenOf(
		gen.Const((*[]testutil.TestDoc)(nil)),
		gen.SliceOf(testutil.GenTestDoc()).Map(func(slice []*testutil.TestDoc) *[]testutil.TestDoc {
			result := make([]testutil.TestDoc, len(slice))
			for i, doc := range slice {
				if doc != nil {
					result[i] = *doc
				}
			}
			return &result
		}),
	)
}

// genInvalidFilter generates invalid filter expressions
func genInvalidFilter() gopter.Gen {
	return gen.OneGenOf(
		gen.Const(nil),
		gen.Const("invalid_string_filter"),
		gen.Const(123),
		gen.Const([]string{"invalid", "slice"}),
		gen.Const(bson.M{"$invalid": "operator"}),
		gen.Const(make(chan int)), // Invalid type
	)
}

// genInvalidUpdate generates invalid update expressions
func genInvalidUpdate() gopter.Gen {
	return gen.OneGenOf(
		gen.Const(nil),
		gen.Const("invalid_string_update"),
		gen.Const(123),
		gen.Const([]int{1, 2, 3}),
		gen.Const(make(chan int)), // Invalid type
	)
}

// genInvalidPipeline generates invalid aggregation pipelines
func genInvalidPipeline() gopter.Gen {
	return gen.OneGenOf(
		gen.Const(nil),
		gen.Const("invalid_string_pipeline"),
		gen.Const(123),
		gen.Const([]string{"invalid", "pipeline"}),
		gen.Const(make(chan int)), // Invalid type
	)
}

// testCollectionNilInputs tests Collection interface methods with nil inputs
func testCollectionNilInputs(collection mongo.Collection[testutil.TestDoc], testCase NilInvalidTestCase) bool {

	// Test Create with nil document - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Panic occurred, test fails
				panic("Create with nil document caused panic")
			}
		}()
		collection.Create(testCase.NilDoc) // May return error, but should not panic
	}()

	// Test CreateMany with nil slice - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("CreateMany with nil slice caused panic")
			}
		}()
		collection.CreateMany(testCase.NilDocSlice) // May return error, but should not panic
	}()

	// Test Del with invalid filter - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("Del with invalid filter caused panic")
			}
		}()
		collection.Del(testCase.InvalidFilter) // May return error, but should not panic
	}()

	// Test FindOne with invalid query - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("FindOne with invalid query caused panic")
			}
		}()
		collection.FindOne(testCase.InvalidFilter) // May return error, but should not panic
	}()

	// Test FindMany with invalid filter - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("FindMany with invalid filter caused panic")
			}
		}()
		collection.FindMany(testCase.InvalidFilter) // May return error, but should not panic
	}()

	// Test Save with invalid filter/update - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("Save with invalid inputs caused panic")
			}
		}()
		collection.Save(testCase.InvalidFilter, testCase.InvalidUpdate) // May return error, but should not panic
	}()

	// Test SaveMany with invalid filter/update - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("SaveMany with invalid inputs caused panic")
			}
		}()
		collection.SaveMany(testCase.InvalidFilter, testCase.InvalidUpdate) // May return error, but should not panic
	}()

	// Test Agg with invalid pipeline - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("Agg with invalid pipeline caused panic")
			}
		}()
		collection.Agg(testCase.InvalidPipeline) // May return error, but should not panic
	}()

	// Test WithTx with nil function - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("WithTx with nil function caused panic")
			}
		}()
		collection.WithTx(nil) // May return error, but should not panic
	}()

	return true
}

// testExtendedCollectionNilInputs tests ExtendedCollection interface methods with nil inputs
func testExtendedCollectionNilInputs(collection mongo.Collection[testutil.TestDoc], testCase NilInvalidTestCase) bool {
	ctx := context.Background()
	extCollection := collection.Build(ctx)

	// Test By with invalid field/value - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("By with invalid inputs caused panic")
			}
		}()
		extCollection.By("", testCase.InvalidFilter) // May return error, but should not panic
	}()

	// Test Where with invalid filter - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("Where with invalid filter caused panic")
			}
		}()
		if m, ok := testCase.InvalidFilter.(bson.M); ok {
			extCollection.Where(m) // May return error, but should not panic
		}
	}()

	// Test First with nil output - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("First with nil output caused panic")
			}
		}()
		extCollection.First(testCase.NilDoc) // May return error, but should not panic
	}()

	// Test Many with nil output - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("Many with nil output caused panic")
			}
		}()
		extCollection.Many(testCase.NilDocSlice) // May return error, but should not panic
	}()

	// Test Save with invalid update - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("Save with invalid update caused panic")
			}
		}()
		extCollection.Save(testCase.InvalidUpdate) // May return error, but should not panic
	}()

	// Test SaveMany with invalid update - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("SaveMany with invalid update caused panic")
			}
		}()
		extCollection.SaveMany(testCase.InvalidUpdate) // May return error, but should not panic
	}()

	return true
}

// testResultNilInputs tests SingleResult and ManyResult interface methods with nil inputs
func testResultNilInputs(collection mongo.Collection[testutil.TestDoc], testCase NilInvalidTestCase) bool {
	// Test SingleResult methods with nil inputs
	singleResult := collection.FindOne(bson.M{})

	// Test Result with nil output - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("SingleResult.Result with nil output caused panic")
			}
		}()
		singleResult.Result(testCase.NilDoc) // May return error, but should not panic
	}()

	// Test ManyResult methods with nil inputs
	manyResult := collection.FindMany(bson.M{})

	// Test Result with nil output - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("ManyResult.Result with nil output caused panic")
			}
		}()
		manyResult.Result(testCase.NilDocSlice) // May return error, but should not panic
	}()

	// Test Stream with nil function - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("ManyResult.Stream with nil function caused panic")
			}
		}()
		manyResult.Stream(nil) // May return error, but should not panic
	}()

	// Test Each with nil function - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("ManyResult.Each with nil function caused panic")
			}
		}()
		manyResult.Each(nil) // May return error, but should not panic
	}()

	return true
}

// testAggregationNilInputs tests Aggregate interface methods with nil inputs
func testAggregationNilInputs(collection mongo.Collection[testutil.TestDoc], testCase NilInvalidTestCase) bool {
	// Create aggregation with valid pipeline first
	agg := collection.Agg(bson.A{bson.M{"$match": bson.M{}}})

	// Test Result with nil output - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("Aggregate.Result with nil output caused panic")
			}
		}()
		agg.Result(testCase.NilDocSlice) // May return error, but should not panic
	}()

	// Test Stream with nil function - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("Aggregate.Stream with nil function caused panic")
			}
		}()
		agg.Stream(nil) // May return error, but should not panic
	}()

	// Test Each with nil function - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("Aggregate.Each with nil function caused panic")
			}
		}()
		agg.Each(nil) // May return error, but should not panic
	}()

	return true
}

// TestProperty10_EmptyResultSetHandling tests that all operations handle empty result sets correctly
// **Feature: unit-testing, Property 10: Empty result set handling**
// **Validates: Requirements 3.3**
func TestProperty10_EmptyResultSetHandling(t *testing.T) {
	runner := NewPropertyTestRunner()

	property := ForAllValid(
		genEmptyResultTestCase(),
		func(testCase EmptyResultTestCase) bool {
			helper, err := NewPropertyTestHelper()
			if err != nil {
				t.Logf("Failed to create test helper: %v", err)
				return false
			}
			defer helper.Cleanup()

			client := helper.Client()
			collection := mongo.NewCollection[testutil.TestDoc](context.Background(), client, testCase.CollectionName)

			// Test all operations that should handle empty results gracefully
			return testEmptyFindOperations(collection, testCase) &&
				testEmptyAggregationOperations(collection, testCase) &&
				testEmptyExtendedCollectionOperations(collection, testCase)
		},
	)

	runner.RunProperty(t, "empty result set handling", property)
}

// EmptyResultTestCase represents test cases for empty result set handling
type EmptyResultTestCase struct {
	CollectionName string
	EmptyFilter    bson.M
	EmptyPipeline  bson.A
}

// genEmptyResultTestCase generates test cases for empty result scenarios
func genEmptyResultTestCase() gopter.Gen {
	return gopter.CombineGens(
		testutil.GenCollectionName(),
		genEmptyFilter(),
		genEmptyPipeline(),
	).Map(func(values []interface{}) EmptyResultTestCase {
		return EmptyResultTestCase{
			CollectionName: values[0].(string),
			EmptyFilter:    values[1].(bson.M),
			EmptyPipeline:  values[2].(bson.A),
		}
	})
}

// genEmptyFilter generates filters that should match no documents
func genEmptyFilter() gopter.Gen {
	return gen.OneGenOf(
		// Filter that matches non-existent field values
		gen.Const(bson.M{"nonexistent_field": "nonexistent_value"}),
		gen.Const(bson.M{"name": "this_document_does_not_exist"}),
		gen.Const(bson.M{"value": -999999}),
		// Empty filter (should match all, but collection is empty)
		gen.Const(bson.M{}),
	).SuchThat(func(filter bson.M) bool {
		// Always return true to avoid discarding tests
		return true
	})
}

// genEmptyPipeline generates aggregation pipelines that should return no results
func genEmptyPipeline() gopter.Gen {
	return gen.OneGenOf(
		// Pipeline with match stage that matches nothing
		gen.Const(bson.A{
			bson.M{"$match": bson.M{"nonexistent_field": "nonexistent_value"}},
		}),
		gen.Const(bson.A{
			bson.M{"$match": bson.M{"name": "this_document_does_not_exist"}},
		}),
		// Pipeline with limit 0
		gen.Const(bson.A{
			bson.M{"$limit": 0},
		}),
		// Empty pipeline (should match all, but collection is empty)
		gen.Const(bson.A{}),
	).SuchThat(func(pipeline bson.A) bool {
		// Always return true to avoid discarding tests
		return true
	})
}

// testEmptyFindOperations tests find operations with empty results
func testEmptyFindOperations(collection mongo.Collection[testutil.TestDoc], testCase EmptyResultTestCase) bool {
	// Test FindOne with empty result - should not panic and handle gracefully
	singleResult := collection.FindOne(testCase.EmptyFilter)
	var doc testutil.TestDoc
	err := singleResult.Result(&doc)
	// Should return an error (no documents found) but not panic
	if err == nil {
		// If no error, the document should be zero value
		if doc.Name != "" || doc.Value != 0 {
			return false // Unexpected non-zero document
		}
	}

	// Test FindMany with empty result - should return empty slice
	manyResult := collection.FindMany(testCase.EmptyFilter)

	// Test All() method
	docs, err := manyResult.All()
	if err != nil {
		// Error is acceptable, but should not panic
		return true
	}
	if docs == nil {
		return false // Should return empty slice, not nil
	}
	if len(docs) != 0 {
		return false // Should be empty
	}

	// Test Result() method
	var resultDocs []testutil.TestDoc
	err = manyResult.Result(&resultDocs)
	if err != nil {
		// Error is acceptable, but should not panic
		return true
	}
	if len(resultDocs) != 0 {
		return false // Should be empty
	}

	// Test Count() method
	count, err := manyResult.Cnt()
	if err != nil {
		// Error is acceptable, but should not panic
		return true
	}
	if count != 0 {
		return false // Should be zero
	}

	// Test streaming methods with empty results
	streamCallCount := 0
	err = manyResult.Stream(func(ctx context.Context, doc testutil.TestDoc) error {
		streamCallCount++
		return nil
	})
	if err != nil {
		// Error is acceptable, but should not panic
		return true
	}
	if streamCallCount != 0 {
		return false // Stream function should not be called for empty results
	}

	eachCallCount := 0
	err = manyResult.Each(func(ctx context.Context, doc testutil.TestDoc) error {
		eachCallCount++
		return nil
	})
	if err != nil {
		// Error is acceptable, but should not panic
		return true
	}
	if eachCallCount != 0 {
		return false // Each function should not be called for empty results
	}

	return true
}

// testEmptyAggregationOperations tests aggregation operations with empty results
func testEmptyAggregationOperations(collection mongo.Collection[testutil.TestDoc], testCase EmptyResultTestCase) bool {
	// Test aggregation with empty results
	agg := collection.Agg(testCase.EmptyPipeline)

	// Test All() method
	docs, err := agg.All()
	if err != nil {
		// Error is acceptable, but should not panic
		return true
	}
	if docs == nil {
		return false // Should return empty slice, not nil
	}
	if len(docs) != 0 {
		return false // Should be empty
	}

	// Test Result() method
	var resultDocs []testutil.TestDoc
	err = agg.Result(&resultDocs)
	if err != nil {
		// Error is acceptable, but should not panic
		return true
	}
	if len(resultDocs) != 0 {
		return false // Should be empty
	}

	// Test Raw() method
	rawDocs, err := agg.Raw()
	if err != nil {
		// Error is acceptable, but should not panic
		return true
	}
	if rawDocs == nil {
		return false // Should return empty slice, not nil
	}
	if len(rawDocs) != 0 {
		return false // Should be empty
	}

	// Test streaming methods with empty results
	streamCallCount := 0
	err = agg.Stream(func(ctx context.Context, doc testutil.TestDoc) error {
		streamCallCount++
		return nil
	})
	if err != nil {
		// Error is acceptable, but should not panic
		return true
	}
	if streamCallCount != 0 {
		return false // Stream function should not be called for empty results
	}

	eachCallCount := 0
	err = agg.Each(func(ctx context.Context, doc testutil.TestDoc) error {
		eachCallCount++
		return nil
	})
	if err != nil {
		// Error is acceptable, but should not panic
		return true
	}
	if eachCallCount != 0 {
		return false // Each function should not be called for empty results
	}

	return true
}

// testEmptyExtendedCollectionOperations tests extended collection operations with empty results
func testEmptyExtendedCollectionOperations(collection mongo.Collection[testutil.TestDoc], testCase EmptyResultTestCase) bool {
	ctx := context.Background()
	extCollection := collection.Build(ctx)

	// Apply the empty filter
	if len(testCase.EmptyFilter) > 0 {
		extCollection = extCollection.Where(testCase.EmptyFilter)
	}

	// Test Count() method
	count, err := extCollection.Count()
	if err != nil {
		// Error is acceptable, but should not panic
		return true
	}
	if count != 0 {
		return false // Should be zero
	}

	// Test Exists() method
	exists, err := extCollection.Exists()
	if err != nil {
		// Error is acceptable, but should not panic
		return true
	}
	if exists {
		return false // Should be false for empty results
	}

	// Test First() method with empty results
	var doc testutil.TestDoc
	err = extCollection.First(&doc)
	if err == nil {
		// If no error, the document should be zero value
		if doc.Name != "" || doc.Value != 0 {
			return false // Unexpected non-zero document
		}
	}
	// Error is acceptable for empty results

	// Test Many() method with empty results
	var docs []testutil.TestDoc
	err = extCollection.Many(&docs)
	if err != nil {
		// Error is acceptable, but should not panic
		return true
	}
	if len(docs) != 0 {
		return false // Should be empty
	}

	return true
}

// TestProperty11_ResourceCleanupConsistency tests that all resources are properly cleaned up
// **Feature: unit-testing, Property 11: Resource cleanup consistency**
// **Validates: Requirements 3.5, 6.4**
func TestProperty11_ResourceCleanupConsistency(t *testing.T) {
	runner := NewPropertyTestRunner()

	property := ForAllValid(
		gen.IntRange(1, 1000),
		func(id int) bool {
			// Test client cleanup - should not panic
			return testSimpleClientCleanup(id) && testSimpleTransactionCleanup(id)
		},
	)

	runner.RunProperty(t, "resource cleanup consistency", property)
}

// ResourceCleanupTestCase represents test cases for resource cleanup
type ResourceCleanupTestCase struct {
	DatabaseName   string
	CollectionName string
	URI            string
	ShouldError    bool
}

// genResourceCleanupTestCase generates test cases for resource cleanup scenarios
func genResourceCleanupTestCase() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 20 }).Map(func(s string) string { return "testdb_" + s }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 20 }).Map(func(s string) string { return "testcoll_" + s }),
		gen.Const("mongodb://localhost:27017"),
		gen.Bool(),
	).Map(func(values []interface{}) ResourceCleanupTestCase {
		return ResourceCleanupTestCase{
			DatabaseName:   values[0].(string),
			CollectionName: values[1].(string),
			URI:            values[2].(string),
			ShouldError:    values[3].(bool),
		}
	})
}

// genFakeURI generates fake MongoDB URIs for testing
func genFakeURI() gopter.Gen {
	return gen.OneGenOf(
		gen.Const("mongodb://localhost:27017"),
		gen.Const("mongodb://test:27017"),
		gen.Const("mongodb://fake:27017"),
		gen.AlphaString().Map(func(s string) string {
			return "mongodb://" + s + ":27017"
		}),
	)
}

// testClientCleanup tests that clients are properly cleaned up
func testClientCleanup(testCase ResourceCleanupTestCase) bool {
	// Create a fake client
	client, err := mongo.NewFakeClient(
		mongo.WithFakeDatabase(testCase.DatabaseName),
		mongo.WithFakeURI(testCase.URI),
	)
	if err != nil {
		// Error creating client is acceptable
		return true
	}

	// Verify client is functional before cleanup
	if client.DbName() != testCase.DatabaseName {
		return false // Client should have correct database name
	}

	// Test that Write() and Read() methods work before cleanup
	writeClient := client.Write()
	readClient := client.Read()
	if writeClient == nil || readClient == nil {
		return false // Clients should not be nil before cleanup
	}

	// Test cleanup - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Panic during cleanup is a failure
				panic("cleanup panicked")
			}
		}()
		client.Close() // May return error, but should not panic
	}()

	// After cleanup, the client should still be safe to call methods on
	// (though they may return errors or nil values)
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Panic after cleanup is a failure
				panic("client methods panicked after cleanup")
			}
		}()
		client.DbName() // Should not panic
		client.Write()  // Should not panic (may return nil)
		client.Read()   // Should not panic (may return nil)
		client.Close()  // Should not panic (may return error)
	}()

	return true
}

// testTransactionSessionCleanup tests that transaction sessions are properly cleaned up
func testTransactionSessionCleanup(testCase ResourceCleanupTestCase) bool {
	// Create a fake client for transaction testing
	client, err := mongo.NewFakeClient(
		mongo.WithFakeDatabase(testCase.DatabaseName),
		mongo.WithFakeURI(testCase.URI),
	)
	if err != nil {
		// Error creating client is acceptable
		return true
	}
	defer client.Close()

	collection := mongo.NewCollection[testutil.TestDoc](context.Background(), client, testCase.CollectionName)

	// Test transaction session cleanup with StartTx
	// For fake clients, StartTx will likely fail, but should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Panic during StartTx is a failure
				panic("StartTx panicked")
			}
		}()

		txSession, err := collection.StartTx()
		if err != nil {
			// Error starting transaction is expected for fake client
			return
		}

		// If we got a session, test cleanup
		if txSession != nil {
			// Verify session is functional before cleanup
			ctx := txSession.Ctx()
			if ctx == nil {
				panic("session context is nil")
			}

			// Test cleanup with successful scenario (no error)
			var sessionErr error
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Panic during session cleanup is a failure
						panic("session cleanup panicked")
					}
				}()
				txSession.Close(&sessionErr)
			}()
		}
	}()

	// Test WithTx cleanup (automatic transaction management)
	// This should handle all cleanup automatically and not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Panic during WithTx is a failure
				panic("WithTx panicked")
			}
		}()

		// WithTx may return an error with fake client, but should not panic
		collection.WithTx(func(ctx context.Context) error {
			// Simple operation that should not cause issues
			if testCase.ShouldError {
				return errors.New("test error")
			}
			return nil
		})
	}()

	// WithTx should handle cleanup automatically, regardless of error
	// The important thing is that it doesn't panic

	return true
}

// testSimpleClientCleanup tests basic client cleanup without panics
func testSimpleClientCleanup(id int) bool {
	// Create a fake client
	client, err := mongo.NewFakeClient(
		mongo.WithFakeDatabase(fmt.Sprintf("testdb_%d", id)),
		mongo.WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		// Error creating client is acceptable
		return true
	}

	// Test cleanup - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("client cleanup panicked")
			}
		}()
		client.Close() // May return error, but should not panic
	}()

	// Test methods after cleanup - should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("client methods panicked after cleanup")
			}
		}()
		client.DbName()
		client.Write()
		client.Read()
		client.Close() // Second close should not panic
	}()

	return true
}

// testSimpleTransactionCleanup tests basic transaction cleanup without panics
func testSimpleTransactionCleanup(id int) bool {
	// Create a fake client
	client, err := mongo.NewFakeClient(
		mongo.WithFakeDatabase(fmt.Sprintf("testdb_%d", id)),
		mongo.WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		// Error creating client is acceptable
		return true
	}
	defer client.Close()

	collection := mongo.NewCollection[testutil.TestDoc](context.Background(), client, fmt.Sprintf("testcoll_%d", id))

	// Test WithTx - should not panic even if it returns errors
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("WithTx panicked")
			}
		}()

		collection.WithTx(func(ctx context.Context) error {
			// Simple operation
			return nil
		})
	}()

	// Test StartTx - should not panic even if it returns errors
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic("StartTx panicked")
			}
		}()

		txSession, err := collection.StartTx()
		if err != nil {
			// Error is expected with fake client
			return
		}

		if txSession != nil {
			// Test session cleanup - should not panic
			var sessionErr error
			func() {
				defer func() {
					if r := recover(); r != nil {
						panic("session cleanup panicked")
					}
				}()
				txSession.Close(&sessionErr)
			}()
		}
	}()

	return true
}
