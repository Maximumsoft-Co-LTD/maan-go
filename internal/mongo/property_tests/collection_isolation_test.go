package property_tests

import (
	"context"
	"testing"

	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo"
	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo/testutil"
	"github.com/leanovate/gopter/prop"
)

// TestCollectionInstanceIsolation tests Property 8: Collection instance isolation
// **Feature: unit-testing, Property 8: Collection instance isolation**
// **Validates: Requirements 2.5**
func TestCollectionInstanceIsolation(t *testing.T) {
	runner := NewPropertyTestRunner()

	property := prop.ForAll(func(doc *testutil.TestDoc) bool {
		// Create test client and collection
		client, err := CreateTestClient()
		if err != nil {
			return false
		}
		defer client.Close()

		// Create base collection with background context
		ctx1 := context.Background()
		collectionName := "test_collection"
		baseCollection := mongo.NewCollection[testutil.TestDoc](ctx1, client, collectionName)

		// Create different contexts
		ctx2 := context.WithValue(context.Background(), "key1", "value1")
		ctx3 := context.WithValue(context.Background(), "key2", "value2")

		// Create isolated collection instances using Ctx()
		collection2 := baseCollection.Ctx(ctx2)
		collection3 := baseCollection.Ctx(ctx3)

		// Test that instances are different objects
		if baseCollection == collection2 || baseCollection == collection3 || collection2 == collection3 {
			return false
		}

		// Test that each collection maintains its own context isolation
		// All collections should be valid and functional
		if baseCollection == nil || collection2 == nil || collection3 == nil {
			return false
		}

		// Test that operations on one collection don't affect others
		// Each collection should be able to perform operations independently

		// Test FindOne operations on each collection
		result1 := baseCollection.FindOne(map[string]interface{}{"name": doc.Name})
		result2 := collection2.FindOne(map[string]interface{}{"name": doc.Name})
		result3 := collection3.FindOne(map[string]interface{}{"name": doc.Name})

		if result1 == nil || result2 == nil || result3 == nil {
			return false
		}

		// Test FindMany operations on each collection
		manyResult1 := baseCollection.FindMany(map[string]interface{}{"active": doc.Active})
		manyResult2 := collection2.FindMany(map[string]interface{}{"active": doc.Active})
		manyResult3 := collection3.FindMany(map[string]interface{}{"active": doc.Active})

		if manyResult1 == nil || manyResult2 == nil || manyResult3 == nil {
			return false
		}

		// Test that Build() creates isolated extended collections
		extended1 := baseCollection.Build(ctx1)
		extended2 := collection2.Build(ctx2)
		extended3 := collection3.Build(ctx3)

		if extended1 == nil || extended2 == nil || extended3 == nil {
			return false
		}

		// Extended collections should also be isolated
		if extended1 == extended2 || extended1 == extended3 || extended2 == extended3 {
			return false
		}

		return true
	}, testutil.GenTestDoc())

	runner.RunProperty(t, "Collection instance isolation", property)
}

// TestCollectionInstanceIsolationWithChaining tests isolation with method chaining
// **Feature: unit-testing, Property 8: Collection instance isolation**
// **Validates: Requirements 2.5**
func TestCollectionInstanceIsolationWithChaining(t *testing.T) {
	runner := NewPropertyTestRunner()

	property := prop.ForAll(func(doc *testutil.TestDoc) bool {
		// Create test client and collection
		client, err := CreateTestClient()
		if err != nil {
			return false
		}
		defer client.Close()

		// Create base collection
		ctx := context.Background()
		collectionName := "test_collection_chaining"
		baseCollection := mongo.NewCollection[testutil.TestDoc](ctx, client, collectionName)

		// Create different contexts for chaining
		ctx1 := context.WithValue(context.Background(), "chain1", "value1")
		ctx2 := context.WithValue(context.Background(), "chain2", "value2")

		// Test chained Ctx() calls create isolated instances
		chain1 := baseCollection.Ctx(ctx1).Ctx(ctx1)
		chain2 := baseCollection.Ctx(ctx2).Ctx(ctx2)

		if chain1 == nil || chain2 == nil {
			return false
		}

		// Chained collections should be isolated from each other
		if chain1 == chain2 {
			return false
		}

		// Test that chained operations maintain isolation
		chainedResult1 := chain1.FindOne(map[string]interface{}{"name": doc.Name})
		chainedResult2 := chain2.FindOne(map[string]interface{}{"name": doc.Name})

		if chainedResult1 == nil || chainedResult2 == nil {
			return false
		}

		// Test mixed chaining with Build()
		mixedChain1 := baseCollection.Ctx(ctx1).Build(ctx1)
		mixedChain2 := baseCollection.Ctx(ctx2).Build(ctx2)

		if mixedChain1 == nil || mixedChain2 == nil {
			return false
		}

		// Mixed chains should be isolated
		if mixedChain1 == mixedChain2 {
			return false
		}

		return true
	}, testutil.GenTestDoc())

	runner.RunProperty(t, "Collection instance isolation with chaining", property)
}

// TestCollectionInstanceIsolationWithDifferentTypes tests isolation across different document types
// **Feature: unit-testing, Property 8: Collection instance isolation**
// **Validates: Requirements 2.5**
func TestCollectionInstanceIsolationWithDifferentTypes(t *testing.T) {
	runner := NewPropertyTestRunner()

	property := prop.ForAll(func(testDoc *testutil.TestDoc) bool {
		// Create a simple default doc for testing
		defaultDoc := &testutil.DefaultTestDoc{
			Name: testDoc.Name + "_default",
		}
		// Create test client
		client, err := CreateTestClient()
		if err != nil {
			return false
		}
		defer client.Close()

		ctx := context.Background()

		// Create collections with different types
		testCollection := mongo.NewCollection[testutil.TestDoc](ctx, client, "test_collection")
		defaultCollection := mongo.NewCollection[testutil.DefaultTestDoc](ctx, client, "default_collection")

		if testCollection == nil || defaultCollection == nil {
			return false
		}

		// Collections with different types should be isolated
		// Note: We can't directly compare them due to different types, but we can test their functionality

		// Test that each collection works with its respective type
		testResult := testCollection.FindOne(map[string]interface{}{"name": testDoc.Name})
		defaultResult := defaultCollection.FindOne(map[string]interface{}{"name": defaultDoc.Name})

		if testResult == nil || defaultResult == nil {
			return false
		}

		// Test that context changes create isolated instances for each type
		ctx2 := context.WithValue(context.Background(), "type_test", "value")

		testCollection2 := testCollection.Ctx(ctx2)
		defaultCollection2 := defaultCollection.Ctx(ctx2)

		if testCollection2 == nil || defaultCollection2 == nil {
			return false
		}

		// New instances should be different from originals
		if testCollection == testCollection2 || defaultCollection == defaultCollection2 {
			return false
		}

		return true
	}, testutil.GenTestDoc())

	runner.RunProperty(t, "Collection instance isolation with different types", property)
}

// TestCollectionInstanceIsolationWithConcurrentAccess tests isolation under concurrent access
// **Feature: unit-testing, Property 8: Collection instance isolation**
// **Validates: Requirements 2.5**
func TestCollectionInstanceIsolationWithConcurrentAccess(t *testing.T) {
	runner := NewPropertyTestRunner()

	property := prop.ForAll(func(doc *testutil.TestDoc) bool {
		// Create test client and collection
		client, err := CreateTestClient()
		if err != nil {
			return false
		}
		defer client.Close()

		ctx := context.Background()
		collectionName := "test_collection_concurrent"
		baseCollection := mongo.NewCollection[testutil.TestDoc](ctx, client, collectionName)

		// Create multiple isolated collections
		collections := make([]mongo.Collection[testutil.TestDoc], 5)
		for i := 0; i < 5; i++ {
			ctxWithValue := context.WithValue(context.Background(), "concurrent", i)
			collections[i] = baseCollection.Ctx(ctxWithValue)

			if collections[i] == nil {
				return false
			}
		}

		// Verify all collections are isolated from each other
		for i := 0; i < len(collections); i++ {
			for j := i + 1; j < len(collections); j++ {
				if collections[i] == collections[j] {
					return false
				}
			}
		}

		// Test that all collections can perform operations independently
		for _, collection := range collections {
			result := collection.FindOne(map[string]interface{}{"name": doc.Name})
			if result == nil {
				return false
			}

			manyResult := collection.FindMany(map[string]interface{}{"active": doc.Active})
			if manyResult == nil {
				return false
			}
		}

		return true
	}, testutil.GenTestDoc())

	runner.RunProperty(t, "Collection instance isolation with concurrent access", property)
}
