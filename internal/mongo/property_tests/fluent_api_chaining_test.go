package property_tests

import (
	"context"
	"testing"

	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo"
	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo/testutil"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
	"go.mongodb.org/mongo-driver/bson"
)

// TestFluentAPIMethodChainingConsistency tests Property 2: Fluent API method chaining consistency
// **Feature: unit-testing, Property 2: Fluent API method chaining consistency**
// **Validates: Requirements 1.4**
func TestFluentAPIMethodChainingConsistency(t *testing.T) {
	runner := NewPropertyTestRunner()

	// Test Collection method chaining
	runner.RunProperty(t, "Collection method chaining consistency",
		testCollectionMethodChaining())

	// Test SingleResult method chaining
	runner.RunProperty(t, "SingleResult method chaining consistency",
		testSingleResultMethodChaining())

	// Test ManyResult method chaining
	runner.RunProperty(t, "ManyResult method chaining consistency",
		testManyResultMethodChaining())

	// Test ExtendedCollection method chaining
	runner.RunProperty(t, "ExtendedCollection method chaining consistency",
		testExtendedCollectionMethodChaining())

	// Test Aggregate method chaining
	runner.RunProperty(t, "Aggregate method chaining consistency",
		testAggregateMethodChaining())
}

// testCollectionMethodChaining tests that Collection methods return the correct interface types
func testCollectionMethodChaining() gopter.Prop {
	return prop.ForAll(func(doc *testutil.TestDoc) bool {
		client, err := CreateTestClient()
		if err != nil {
			return false
		}
		defer client.Close()

		ctx := context.Background()
		collection := mongo.NewCollection[testutil.TestDoc](ctx, client, "test_collection")

		// Test that Ctx() returns Collection[T]
		ctxCollection := collection.Ctx(ctx)
		if !isCollectionOfType[testutil.TestDoc](ctxCollection) {
			return false
		}

		// Test that chained Ctx() calls work
		doubleCtxCollection := collection.Ctx(ctx).Ctx(ctx)
		if !isCollectionOfType[testutil.TestDoc](doubleCtxCollection) {
			return false
		}

		// Test that Build() returns ExtendedCollection[T]
		extendedCollection := collection.Build(ctx)
		if !isExtendedCollectionOfType[testutil.TestDoc](extendedCollection) {
			return false
		}

		// Test that chained Build() after Ctx() works
		chainedExtended := collection.Ctx(ctx).Build(ctx)
		if !isExtendedCollectionOfType[testutil.TestDoc](chainedExtended) {
			return false
		}

		// Test that FindOne() returns SingleResult[T]
		singleResult := collection.FindOne(bson.M{"name": doc.Name})
		if !isSingleResultOfType[testutil.TestDoc](singleResult) {
			return false
		}

		// Test that chained FindOne() after Ctx() works
		chainedSingle := collection.Ctx(ctx).FindOne(bson.M{"name": doc.Name})
		if !isSingleResultOfType[testutil.TestDoc](chainedSingle) {
			return false
		}

		// Test that FindMany() returns ManyResult[T]
		manyResult := collection.FindMany(bson.M{"active": doc.Active})
		if !isManyResultOfType[testutil.TestDoc](manyResult) {
			return false
		}

		// Test that chained FindMany() after Ctx() works
		chainedMany := collection.Ctx(ctx).FindMany(bson.M{"active": doc.Active})
		if !isManyResultOfType[testutil.TestDoc](chainedMany) {
			return false
		}

		// Test that Agg() returns Aggregate[T]
		pipeline := bson.A{bson.M{"$match": bson.M{"value": doc.Value}}}
		aggregate := collection.Agg(pipeline)
		if !isAggregateOfType[testutil.TestDoc](aggregate) {
			return false
		}

		// Test that chained Agg() after Ctx() works
		chainedAgg := collection.Ctx(ctx).Agg(pipeline)
		if !isAggregateOfType[testutil.TestDoc](chainedAgg) {
			return false
		}

		return true
	}, testutil.GenTestDoc())
}

// testSingleResultMethodChaining tests that SingleResult methods return the correct interface types
func testSingleResultMethodChaining() gopter.Prop {
	return prop.ForAll(func(doc *testutil.TestDoc) bool {
		client, err := CreateTestClient()
		if err != nil {
			return false
		}
		defer client.Close()

		ctx := context.Background()
		collection := mongo.NewCollection[testutil.TestDoc](ctx, client, "test_collection")

		// Start with a SingleResult
		singleResult := collection.FindOne(bson.M{"name": doc.Name})

		// Test that Proj() returns SingleResult[T]
		projResult := singleResult.Proj(bson.M{"name": 1})
		if !isSingleResultOfType[testutil.TestDoc](projResult) {
			return false
		}

		// Test that Sort() returns SingleResult[T]
		sortResult := singleResult.Sort(bson.M{"name": 1})
		if !isSingleResultOfType[testutil.TestDoc](sortResult) {
			return false
		}

		// Test that Hint() returns SingleResult[T]
		hintResult := singleResult.Hint(bson.M{"name": 1})
		if !isSingleResultOfType[testutil.TestDoc](hintResult) {
			return false
		}

		// Test method chaining
		chainedResult := singleResult.Proj(bson.M{"name": 1}).Sort(bson.M{"name": 1}).Hint(bson.M{"name": 1})
		if !isSingleResultOfType[testutil.TestDoc](chainedResult) {
			return false
		}

		// Test that chaining preserves type through multiple operations
		complexChain := collection.FindOne(bson.M{"name": doc.Name}).
			Proj(bson.M{"name": 1, "value": 1}).
			Sort(bson.M{"name": 1}).
			Hint(bson.M{"name": 1})
		if !isSingleResultOfType[testutil.TestDoc](complexChain) {
			return false
		}

		return true
	}, testutil.GenTestDoc())
}

// testManyResultMethodChaining tests that ManyResult methods return the correct interface types
func testManyResultMethodChaining() gopter.Prop {
	return prop.ForAll(func(doc *testutil.TestDoc) bool {
		client, err := CreateTestClient()
		if err != nil {
			return false
		}
		defer client.Close()

		ctx := context.Background()
		collection := mongo.NewCollection[testutil.TestDoc](ctx, client, "test_collection")

		// Start with a ManyResult
		manyResult := collection.FindMany(bson.M{"active": doc.Active})

		// Test that Proj() returns ManyResult[T]
		projResult := manyResult.Proj(bson.M{"name": 1})
		if !isManyResultOfType[testutil.TestDoc](projResult) {
			return false
		}

		// Test that Sort() returns ManyResult[T]
		sortResult := manyResult.Sort(bson.M{"name": 1})
		if !isManyResultOfType[testutil.TestDoc](sortResult) {
			return false
		}

		// Test that Hint() returns ManyResult[T]
		hintResult := manyResult.Hint(bson.M{"name": 1})
		if !isManyResultOfType[testutil.TestDoc](hintResult) {
			return false
		}

		// Test that Limit() returns ManyResult[T]
		limitResult := manyResult.Limit(10)
		if !isManyResultOfType[testutil.TestDoc](limitResult) {
			return false
		}

		// Test that Skip() returns ManyResult[T]
		skipResult := manyResult.Skip(5)
		if !isManyResultOfType[testutil.TestDoc](skipResult) {
			return false
		}

		// Test that Bsz() returns ManyResult[T]
		bszResult := manyResult.Bsz(100)
		if !isManyResultOfType[testutil.TestDoc](bszResult) {
			return false
		}

		// Test method chaining
		chainedResult := manyResult.
			Proj(bson.M{"name": 1}).
			Sort(bson.M{"name": 1}).
			Hint(bson.M{"name": 1}).
			Limit(10).
			Skip(5).
			Bsz(100)
		if !isManyResultOfType[testutil.TestDoc](chainedResult) {
			return false
		}

		// Test that chaining preserves type through multiple operations
		complexChain := collection.FindMany(bson.M{"active": doc.Active}).
			Proj(bson.M{"name": 1, "value": 1}).
			Sort(bson.M{"name": 1}).
			Limit(20).
			Skip(10)
		if !isManyResultOfType[testutil.TestDoc](complexChain) {
			return false
		}

		return true
	}, testutil.GenTestDoc())
}

// testExtendedCollectionMethodChaining tests that ExtendedCollection methods return the correct interface types
func testExtendedCollectionMethodChaining() gopter.Prop {
	return prop.ForAll(func(doc *testutil.TestDoc) bool {
		client, err := CreateTestClient()
		if err != nil {
			return false
		}
		defer client.Close()

		ctx := context.Background()
		collection := mongo.NewCollection[testutil.TestDoc](ctx, client, "test_collection")

		// Start with an ExtendedCollection
		extendedCollection := collection.Build(ctx)

		// Test that By() returns ExtendedCollection[T]
		byResult := extendedCollection.By("name", doc.Name)
		if !isExtendedCollectionOfType[testutil.TestDoc](byResult) {
			return false
		}

		// Test that Where() returns ExtendedCollection[T]
		whereResult := extendedCollection.Where(bson.M{"active": doc.Active})
		if !isExtendedCollectionOfType[testutil.TestDoc](whereResult) {
			return false
		}

		// Test method chaining
		chainedResult := extendedCollection.
			By("name", doc.Name).
			Where(bson.M{"active": doc.Active}).
			By("value", doc.Value)
		if !isExtendedCollectionOfType[testutil.TestDoc](chainedResult) {
			return false
		}

		// Test that chaining preserves type through multiple operations
		complexChain := collection.Build(ctx).
			By("name", doc.Name).
			Where(bson.M{"active": doc.Active}).
			By("value", doc.Value).
			Where(bson.M{"name": bson.M{"$ne": ""}})
		if !isExtendedCollectionOfType[testutil.TestDoc](complexChain) {
			return false
		}

		return true
	}, testutil.GenTestDoc())
}

// testAggregateMethodChaining tests that Aggregate methods return the correct interface types
func testAggregateMethodChaining() gopter.Prop {
	return prop.ForAll(func(doc *testutil.TestDoc) bool {
		client, err := CreateTestClient()
		if err != nil {
			return false
		}
		defer client.Close()

		ctx := context.Background()
		collection := mongo.NewCollection[testutil.TestDoc](ctx, client, "test_collection")

		// Start with an Aggregate
		pipeline := bson.A{bson.M{"$match": bson.M{"value": doc.Value}}}
		aggregate := collection.Agg(pipeline)

		// Test that Disk() returns Aggregate[T]
		diskResult := aggregate.Disk(true)
		if !isAggregateOfType[testutil.TestDoc](diskResult) {
			return false
		}

		// Test that Bsz() returns Aggregate[T]
		bszResult := aggregate.Bsz(100)
		if !isAggregateOfType[testutil.TestDoc](bszResult) {
			return false
		}

		// Test method chaining
		chainedResult := aggregate.Disk(true).Bsz(100)
		if !isAggregateOfType[testutil.TestDoc](chainedResult) {
			return false
		}

		// Test that chaining preserves type through multiple operations
		complexChain := collection.Agg(pipeline).
			Disk(true).
			Bsz(50).
			Disk(false).
			Bsz(200)
		if !isAggregateOfType[testutil.TestDoc](complexChain) {
			return false
		}

		return true
	}, testutil.GenTestDoc())
}

// TestFluentAPIMethodChainingIsolation tests that method chaining creates isolated instances
func TestFluentAPIMethodChainingIsolation(t *testing.T) {
	runner := NewPropertyTestRunner()

	runner.RunProperty(t, "Method chaining creates isolated instances", prop.ForAll(func(doc *testutil.TestDoc) bool {
		client, err := CreateTestClient()
		if err != nil {
			return false
		}
		defer client.Close()

		ctx := context.Background()
		collection := mongo.NewCollection[testutil.TestDoc](ctx, client, "test_collection")

		// Test that Ctx() creates a new instance
		originalCollection := collection
		ctxCollection := collection.Ctx(ctx)

		// They should be different instances but same type
		if originalCollection == ctxCollection {
			return false // Should be different instances
		}

		// Test that Build() creates a new instance
		extendedCollection1 := collection.Build(ctx)
		extendedCollection2 := collection.Build(ctx)

		// They should be different instances but same type
		if extendedCollection1 == extendedCollection2 {
			return false // Should be different instances
		}

		// Test that query builders create isolated instances
		singleResult1 := collection.FindOne(bson.M{"name": doc.Name})
		singleResult2 := collection.FindOne(bson.M{"name": doc.Name})

		// They should be different instances but same type
		if singleResult1 == singleResult2 {
			return false // Should be different instances
		}

		return true
	}, testutil.GenTestDoc()))
}
