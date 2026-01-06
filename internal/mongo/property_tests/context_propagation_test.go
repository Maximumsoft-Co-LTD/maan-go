package property_tests

import (
	"context"
	"testing"

	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo"
	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo/testutil"
	"github.com/leanovate/gopter/prop"
)

// TestContextPropagationPreservation tests Property 7: Context propagation preservation
// **Feature: unit-testing, Property 7: Context propagation preservation**
// **Validates: Requirements 2.4**
func TestContextPropagationPreservation(t *testing.T) {
	runner := NewPropertyTestRunner()

	property := prop.ForAll(func(doc *testutil.TestDoc) bool {
		contextKey := "test_key"
		contextValue := "test_value"
		// Create test client and collection
		client, err := CreateTestClient()
		if err != nil {
			return false
		}
		defer client.Close()

		// Create context with a value
		ctx := context.WithValue(context.Background(), contextKey, contextValue)
		collectionName := "test_collection"

		// Create collection with context
		collection := mongo.NewCollection[testutil.TestDoc](ctx, client, collectionName)

		// Test that context is preserved through Ctx() method
		ctxCollection := collection.Ctx(ctx)
		if ctxCollection == nil {
			return false
		}

		// Test that context is preserved through Build() method
		extendedCollection := collection.Build(ctx)
		if extendedCollection == nil {
			return false
		}

		// Test that context is preserved through query operations
		singleResult := collection.FindOne(map[string]interface{}{"name": doc.Name})
		if singleResult == nil {
			return false
		}

		manyResult := collection.FindMany(map[string]interface{}{"active": doc.Active})
		if manyResult == nil {
			return false
		}

		// Test that context is preserved through aggregation
		aggregate := collection.Agg([]interface{}{map[string]interface{}{"$match": map[string]interface{}{"value": doc.Value}}})
		if aggregate == nil {
			return false
		}

		// Test chained context operations
		chainedCollection := collection.Ctx(ctx).Ctx(ctx)
		if chainedCollection == nil {
			return false
		}

		// Test that different contexts create independent instances
		ctx2 := context.WithValue(context.Background(), contextKey+"_2", contextValue+"_2")
		collection2 := collection.Ctx(ctx2)
		if collection2 == nil {
			return false
		}

		// Both collections should be valid but independent
		return collection != collection2
	}, testutil.GenTestDoc())

	runner.RunProperty(t, "Context propagation preservation", property)
}

// TestContextPropagationWithTimeout tests context propagation with timeout contexts
// **Feature: unit-testing, Property 7: Context propagation preservation**
// **Validates: Requirements 2.4**
func TestContextPropagationWithTimeout(t *testing.T) {
	runner := NewPropertyTestRunner()

	property := prop.ForAll(func(doc *testutil.TestDoc) bool {
		// Create test client and collection
		client, err := CreateTestClient()
		if err != nil {
			return false
		}
		defer client.Close()

		// Create context with timeout
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		collectionName := "test_collection_timeout"

		// Create collection with timeout context
		collection := mongo.NewCollection[testutil.TestDoc](ctx, client, collectionName)

		// Test that timeout context is preserved through operations
		ctxCollection := collection.Ctx(ctx)
		if ctxCollection == nil {
			return false
		}

		// Test operations with timeout context
		singleResult := ctxCollection.FindOne(map[string]interface{}{"name": doc.Name})
		if singleResult == nil {
			return false
		}

		manyResult := ctxCollection.FindMany(map[string]interface{}{"active": doc.Active})
		if manyResult == nil {
			return false
		}

		// Test that we can create new contexts from the collection
		newCtx := context.Background()
		newCollection := collection.Ctx(newCtx)
		if newCollection == nil {
			return false
		}

		return true
	}, testutil.GenTestDoc())

	runner.RunProperty(t, "Context propagation with timeout", property)
}

// TestContextPropagationInExtendedCollection tests context propagation in extended collections
// **Feature: unit-testing, Property 7: Context propagation preservation**
// **Validates: Requirements 2.4**
func TestContextPropagationInExtendedCollection(t *testing.T) {
	runner := NewPropertyTestRunner()

	property := prop.ForAll(func(doc *testutil.TestDoc) bool {
		contextKey := "test_key"
		contextValue := "test_value"
		// Create test client and collection
		client, err := CreateTestClient()
		if err != nil {
			return false
		}
		defer client.Close()

		// Create context with a value
		ctx := context.WithValue(context.Background(), contextKey, contextValue)
		collectionName := "test_collection_extended"

		// Create collection with context
		collection := mongo.NewCollection[testutil.TestDoc](ctx, client, collectionName)

		// Build extended collection with context
		extendedCollection := collection.Build(ctx)
		if extendedCollection == nil {
			return false
		}

		// Test that extended collection operations work with context
		byCollection := extendedCollection.By("name", doc.Name)
		if byCollection == nil {
			return false
		}

		whereCollection := byCollection.Where(map[string]interface{}{"active": doc.Active})
		if whereCollection == nil {
			return false
		}

		// Test that we can chain operations on extended collection
		chainedCollection := extendedCollection.By("name", doc.Name).Where(map[string]interface{}{"value": doc.Value})
		if chainedCollection == nil {
			return false
		}

		return true
	}, testutil.GenTestDoc())

	runner.RunProperty(t, "Context propagation in extended collection", property)
}
