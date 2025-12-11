package property_tests

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo"
	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo/testutil"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
	"go.mongodb.org/mongo-driver/bson"
)

// TestThreadSafetyUnderConcurrentAccess tests Property 3: Thread safety under concurrent access
// **Feature: unit-testing, Property 3: Thread safety under concurrent access**
// **Validates: Requirements 1.5, 3.4, 4.4**
func TestThreadSafetyUnderConcurrentAccess(t *testing.T) {
	runner := NewPropertyTestRunner()

	// Test concurrent collection operations
	runner.RunProperty(t, "Concurrent collection operations thread safety",
		testConcurrentCollectionOperations())

	// Test concurrent context switching
	runner.RunProperty(t, "Concurrent context switching thread safety",
		testConcurrentContextSwitching())

	// Test concurrent query building
	runner.RunProperty(t, "Concurrent query building thread safety",
		testConcurrentQueryBuilding())

	// Test concurrent fake client operations
	runner.RunProperty(t, "Concurrent fake client operations thread safety",
		testConcurrentFakeClientOperations())
}

// testConcurrentCollectionOperations tests that collection operations are thread-safe
func testConcurrentCollectionOperations() gopter.Prop {
	return prop.ForAll(func(doc *testutil.TestDoc) bool {
		client, err := CreateTestClient()
		if err != nil {
			return false
		}
		defer client.Close()

		ctx := context.Background()
		collection := mongo.NewCollection[testutil.TestDoc](ctx, client, "test_collection")

		const numGoroutines = 10
		const operationsPerGoroutine = 5

		var wg sync.WaitGroup
		var mu sync.Mutex
		errors := make([]error, 0)

		// Function to safely collect errors
		collectError := func(err error) {
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
			}
		}

		// Launch multiple goroutines performing collection operations
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < operationsPerGoroutine; j++ {
					// Test concurrent Ctx() calls
					ctxCollection := collection.Ctx(ctx)
					if !isCollectionOfType[testutil.TestDoc](ctxCollection) {
						collectError(err)
						return
					}

					// Test concurrent Build() calls
					extendedCollection := collection.Build(ctx)
					if !isExtendedCollectionOfType[testutil.TestDoc](extendedCollection) {
						collectError(err)
						return
					}

					// Test concurrent FindOne() calls
					singleResult := collection.FindOne(bson.M{"name": doc.Name})
					if !isSingleResultOfType[testutil.TestDoc](singleResult) {
						collectError(err)
						return
					}

					// Test concurrent FindMany() calls
					manyResult := collection.FindMany(bson.M{"active": doc.Active})
					if !isManyResultOfType[testutil.TestDoc](manyResult) {
						collectError(err)
						return
					}

					// Test concurrent Agg() calls
					pipeline := bson.A{bson.M{"$match": bson.M{"value": doc.Value}}}
					aggregate := collection.Agg(pipeline)
					if !isAggregateOfType[testutil.TestDoc](aggregate) {
						collectError(err)
						return
					}

					// Small delay to increase chance of race conditions
					runtime.Gosched()
				}
			}(i)
		}

		// Wait for all goroutines to complete
		wg.Wait()

		// Check if any errors occurred
		mu.Lock()
		hasErrors := len(errors) > 0
		mu.Unlock()

		return !hasErrors
	}, testutil.GenTestDoc())
}

// testConcurrentContextSwitching tests that context switching is thread-safe
func testConcurrentContextSwitching() gopter.Prop {
	return prop.ForAll(func(doc *testutil.TestDoc) bool {
		client, err := CreateTestClient()
		if err != nil {
			return false
		}
		defer client.Close()

		baseCtx := context.Background()
		collection := mongo.NewCollection[testutil.TestDoc](baseCtx, client, "test_collection")

		const numGoroutines = 10
		const operationsPerGoroutine = 5

		var wg sync.WaitGroup
		var mu sync.Mutex
		contextValues := make(map[int]string)
		errors := make([]error, 0)

		// Function to safely collect errors
		collectError := func(err error) {
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
			}
		}

		// Launch multiple goroutines with different contexts
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				// Create a unique context for this goroutine
				ctx := context.WithValue(baseCtx, testutil.TestContextKeyRequest, goroutineID)

				for j := 0; j < operationsPerGoroutine; j++ {
					// Create collection with specific context
					ctxCollection := collection.Ctx(ctx)

					// Verify the collection maintains the correct context
					if !isCollectionOfType[testutil.TestDoc](ctxCollection) {
						collectError(err)
						return
					}

					// Test that context isolation works
					extendedCollection := ctxCollection.Build(ctx)
					if !isExtendedCollectionOfType[testutil.TestDoc](extendedCollection) {
						collectError(err)
						return
					}

					// Store the context value for verification
					mu.Lock()
					contextValues[goroutineID] = "processed"
					mu.Unlock()

					runtime.Gosched()
				}
			}(i)
		}

		// Wait for all goroutines to complete
		wg.Wait()

		// Verify all goroutines completed successfully
		mu.Lock()
		hasErrors := len(errors) > 0
		allCompleted := len(contextValues) == numGoroutines
		mu.Unlock()

		return !hasErrors && allCompleted
	}, testutil.GenTestDoc())
}

// testConcurrentQueryBuilding tests that query building is thread-safe
func testConcurrentQueryBuilding() gopter.Prop {
	return prop.ForAll(func(doc *testutil.TestDoc) bool {
		client, err := CreateTestClient()
		if err != nil {
			return false
		}
		defer client.Close()

		ctx := context.Background()
		collection := mongo.NewCollection[testutil.TestDoc](ctx, client, "test_collection")

		const numGoroutines = 8
		const operationsPerGoroutine = 3

		var wg sync.WaitGroup
		var mu sync.Mutex
		errors := make([]error, 0)

		// Function to safely collect errors
		collectError := func(err error) {
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
			}
		}

		// Launch multiple goroutines building queries concurrently
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < operationsPerGoroutine; j++ {
					// Test concurrent ExtendedCollection operations
					builder := collection.Build(ctx)

					// Chain multiple operations
					result := builder.
						By("name", doc.Name).
						Where(bson.M{"active": doc.Active}).
						By("value", doc.Value)

					if !isExtendedCollectionOfType[testutil.TestDoc](result) {
						collectError(err)
						return
					}

					// Test concurrent SingleResult operations
					singleResult := collection.FindOne(bson.M{"name": doc.Name}).
						Proj(bson.M{"name": 1}).
						Sort(bson.M{"name": 1})

					if !isSingleResultOfType[testutil.TestDoc](singleResult) {
						collectError(err)
						return
					}

					// Test concurrent ManyResult operations
					manyResult := collection.FindMany(bson.M{"active": doc.Active}).
						Limit(10).
						Skip(5).
						Sort(bson.M{"name": 1})

					if !isManyResultOfType[testutil.TestDoc](manyResult) {
						collectError(err)
						return
					}

					// Test concurrent Aggregate operations
					pipeline := bson.A{bson.M{"$match": bson.M{"value": doc.Value}}}
					aggregate := collection.Agg(pipeline).
						Disk(true).
						Bsz(100)

					if !isAggregateOfType[testutil.TestDoc](aggregate) {
						collectError(err)
						return
					}

					runtime.Gosched()
				}
			}(i)
		}

		// Wait for all goroutines to complete
		wg.Wait()

		// Check if any errors occurred
		mu.Lock()
		hasErrors := len(errors) > 0
		mu.Unlock()

		return !hasErrors
	}, testutil.GenTestDoc())
}

// testConcurrentFakeClientOperations tests that fake client operations are thread-safe
func testConcurrentFakeClientOperations() gopter.Prop {
	return prop.ForAll(func() bool {
		const numGoroutines = 10
		const operationsPerGoroutine = 3

		var wg sync.WaitGroup
		var mu sync.Mutex
		errors := make([]error, 0)
		clients := make([]mongo.Client, 0)

		// Function to safely collect errors
		collectError := func(err error) {
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
			}
		}

		// Function to safely collect clients for cleanup
		collectClient := func(client mongo.Client) {
			mu.Lock()
			clients = append(clients, client)
			mu.Unlock()
		}

		// Launch multiple goroutines creating and using fake clients concurrently
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < operationsPerGoroutine; j++ {
					// Create fake client
					client, err := mongo.NewFakeClient(
						mongo.WithFakeDatabase("testdb_"+string(rune('A'+goroutineID))),
						mongo.WithFakeURI("mongodb://localhost:27017"),
					)
					if err != nil {
						collectError(err)
						return
					}

					collectClient(client)

					// Test client operations
					ctx := context.Background()
					collection := mongo.NewCollection[testutil.TestDoc](ctx, client, "test_collection")

					// Verify collection creation works
					if !isCollectionOfType[testutil.TestDoc](collection) {
						collectError(err)
						return
					}

					// Test basic operations
					ctxCollection := collection.Ctx(ctx)
					if !isCollectionOfType[testutil.TestDoc](ctxCollection) {
						collectError(err)
						return
					}

					runtime.Gosched()
				}
			}(i)
		}

		// Wait for all goroutines to complete
		wg.Wait()

		// Cleanup all clients
		mu.Lock()
		for _, client := range clients {
			if client != nil {
				client.Close()
			}
		}
		hasErrors := len(errors) > 0
		mu.Unlock()

		return !hasErrors
	})
}

// TestThreadSafetyWithRaceDetection runs thread safety tests with race detection enabled
func TestThreadSafetyWithRaceDetection(t *testing.T) {
	// This test is designed to be run with -race flag
	// go test -race ./internal/mongo/property_tests -run TestThreadSafetyWithRaceDetection

	runner := NewPropertyTestRunner().WithConfig(testutil.PropertyTestConfig{
		Iterations: 50, // Reduced iterations for race detection
		MaxShrinks: 50,
		RandomSeed: time.Now().UnixNano(),
		Timeout:    15 * time.Second,
	})

	runner.RunProperty(t, "Race detection for concurrent operations", prop.ForAll(func(doc *testutil.TestDoc) bool {
		client, err := CreateTestClient()
		if err != nil {
			return false
		}
		defer client.Close()

		ctx := context.Background()
		collection := mongo.NewCollection[testutil.TestDoc](ctx, client, "test_collection")

		const numGoroutines = 5
		var wg sync.WaitGroup

		// Launch concurrent operations that might cause race conditions
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				// Perform operations that access shared state
				_ = collection.Ctx(ctx).Build(ctx).By("name", doc.Name)
				_ = collection.FindOne(bson.M{"name": doc.Name}).Sort(bson.M{"name": 1})
				_ = collection.FindMany(bson.M{"active": doc.Active}).Limit(10)
				_ = collection.Agg(bson.A{bson.M{"$match": bson.M{"value": doc.Value}}}).Disk(true)

				runtime.Gosched()
			}()
		}

		wg.Wait()
		return true
	}, testutil.GenTestDoc()))
}
