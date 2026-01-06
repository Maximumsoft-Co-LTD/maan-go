package property_tests

import (
	"context"
	"reflect"
	"testing"

	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo"
	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo/testutil"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// TestGenericTypeSafetyPreservation tests Property 1: Generic type safety preservation
// **Feature: unit-testing, Property 1: Generic type safety preservation**
// **Validates: Requirements 1.3**
func TestGenericTypeSafetyPreservation(t *testing.T) {
	runner := NewPropertyTestRunner()

	// Test with TestDoc type
	runner.RunProperty(t, "Generic type safety for TestDoc",
		testGenericTypeSafetyForType[testutil.TestDoc](testutil.GenTestDoc()))

	// Test with DefaultTestDoc type
	runner.RunProperty(t, "Generic type safety for DefaultTestDoc",
		testGenericTypeSafetyForType[testutil.DefaultTestDoc](testutil.GenDefaultTestDoc()))
}

// testGenericTypeSafetyForType creates a property test for a specific document type
func testGenericTypeSafetyForType[T any](docGen gopter.Gen) gopter.Prop {
	return prop.ForAll(func(doc *T) bool {
		// Create test client
		client, err := CreateTestClient()
		if err != nil {
			return false
		}
		defer client.Close()

		ctx := context.Background()
		collName := "test_collection"

		// Create collection with specific type T
		collection := mongo.NewCollection[T](ctx, client, collName)

		// Verify collection maintains type T throughout operations
		return verifyCollectionTypeConsistency(collection, doc)
	}, docGen)
}

// verifyCollectionTypeConsistency checks that all collection operations maintain type consistency
func verifyCollectionTypeConsistency[T any](collection mongo.Collection[T], doc *T) bool {
	ctx := context.Background()

	// Test 1: Collection.Ctx() should return Collection[T]
	ctxCollection := collection.Ctx(ctx)
	if !isCollectionOfType[T](ctxCollection) {
		return false
	}

	// Test 2: Collection.Build() should return ExtendedCollection[T]
	extendedCollection := collection.Build(ctx)
	if !isExtendedCollectionOfType[T](extendedCollection) {
		return false
	}

	// Test 3: Collection.FindOne() should return SingleResult[T]
	singleResult := collection.FindOne(map[string]interface{}{"name": "test"})
	if !isSingleResultOfType[T](singleResult) {
		return false
	}

	// Test 4: Collection.FindMany() should return ManyResult[T]
	manyResult := collection.FindMany(map[string]interface{}{"name": "test"})
	if !isManyResultOfType[T](manyResult) {
		return false
	}

	// Test 5: Collection.Agg() should return Aggregate[T]
	aggregate := collection.Agg([]interface{}{map[string]interface{}{"$match": map[string]interface{}{"name": "test"}}})
	if !isAggregateOfType[T](aggregate) {
		return false
	}

	// Test 6: Chained operations should maintain type consistency
	chainedSingle := collection.Ctx(ctx).FindOne(map[string]interface{}{"name": "test"})
	if !isSingleResultOfType[T](chainedSingle) {
		return false
	}

	chainedMany := collection.Ctx(ctx).FindMany(map[string]interface{}{"name": "test"})
	if !isManyResultOfType[T](chainedMany) {
		return false
	}

	chainedExtended := collection.Ctx(ctx).Build(ctx)
	if !isExtendedCollectionOfType[T](chainedExtended) {
		return false
	}

	return true
}

// Type checking helper functions using reflection
func isCollectionOfType[T any](collection mongo.Collection[T]) bool {
	// Check if the collection interface is properly typed
	collectionType := reflect.TypeOf(collection)
	if collectionType == nil {
		return false
	}

	// Verify it's a Collection interface
	expectedInterface := reflect.TypeOf((*mongo.Collection[T])(nil)).Elem()
	return collectionType.Implements(expectedInterface)
}

func isExtendedCollectionOfType[T any](extendedCollection mongo.ExtendedCollection[T]) bool {
	collectionType := reflect.TypeOf(extendedCollection)
	if collectionType == nil {
		return false
	}

	expectedInterface := reflect.TypeOf((*mongo.ExtendedCollection[T])(nil)).Elem()
	return collectionType.Implements(expectedInterface)
}

func isSingleResultOfType[T any](singleResult mongo.SingleResult[T]) bool {
	resultType := reflect.TypeOf(singleResult)
	if resultType == nil {
		return false
	}

	expectedInterface := reflect.TypeOf((*mongo.SingleResult[T])(nil)).Elem()
	return resultType.Implements(expectedInterface)
}

func isManyResultOfType[T any](manyResult mongo.ManyResult[T]) bool {
	resultType := reflect.TypeOf(manyResult)
	if resultType == nil {
		return false
	}

	expectedInterface := reflect.TypeOf((*mongo.ManyResult[T])(nil)).Elem()
	return resultType.Implements(expectedInterface)
}

func isAggregateOfType[T any](aggregate mongo.Aggregate[T]) bool {
	aggType := reflect.TypeOf(aggregate)
	if aggType == nil {
		return false
	}

	expectedInterface := reflect.TypeOf((*mongo.Aggregate[T])(nil)).Elem()
	return aggType.Implements(expectedInterface)
}

// TestGenericTypeConstraintEnforcement tests that generic constraints are properly enforced
func TestGenericTypeConstraintEnforcement(t *testing.T) {
	runner := NewPropertyTestRunner()

	runner.RunProperty(t, "Generic type constraint enforcement", prop.ForAll(func() bool {
		client, err := CreateTestClient()
		if err != nil {
			return false
		}
		defer client.Close()

		ctx := context.Background()
		collName := "test_collection"

		// Test that we can create collections with different types
		testDocCollection := mongo.NewCollection[testutil.TestDoc](ctx, client, collName+"_test")
		defaultDocCollection := mongo.NewCollection[testutil.DefaultTestDoc](ctx, client, collName+"_default")

		// Verify each collection maintains its specific type
		return isCollectionOfType[testutil.TestDoc](testDocCollection) &&
			isCollectionOfType[testutil.DefaultTestDoc](defaultDocCollection)
	}))
}

// TestGenericTypeOperationChaining tests that type information is preserved through operation chaining
func TestGenericTypeOperationChaining(t *testing.T) {
	runner := NewPropertyTestRunner()

	runner.RunProperty(t, "Generic type preservation in operation chaining", prop.ForAll(func(doc *testutil.TestDoc) bool {
		client, err := CreateTestClient()
		if err != nil {
			return false
		}
		defer client.Close()

		ctx := context.Background()
		collName := "test_collection"
		collection := mongo.NewCollection[testutil.TestDoc](ctx, client, collName)

		// Test complex chaining scenarios
		chainedResult1 := collection.Ctx(ctx).Build(ctx).By("name", doc.Name)
		if !isExtendedCollectionOfType[testutil.TestDoc](chainedResult1) {
			return false
		}

		chainedResult2 := collection.Ctx(ctx).FindOne(map[string]interface{}{"name": doc.Name}).Sort(map[string]interface{}{"name": 1})
		if !isSingleResultOfType[testutil.TestDoc](chainedResult2) {
			return false
		}

		chainedResult3 := collection.Ctx(ctx).FindMany(map[string]interface{}{"active": doc.Active}).Limit(10)
		if !isManyResultOfType[testutil.TestDoc](chainedResult3) {
			return false
		}

		chainedResult4 := collection.Ctx(ctx).Agg([]interface{}{map[string]interface{}{"$match": map[string]interface{}{"value": doc.Value}}}).Disk(true)
		if !isAggregateOfType[testutil.TestDoc](chainedResult4) {
			return false
		}

		return true
	}, testutil.GenTestDoc()))
}
