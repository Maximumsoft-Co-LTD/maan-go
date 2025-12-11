package mongo

import (
	"context"
	"testing"

	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo/testutil"
	"go.mongodb.org/mongo-driver/bson"
)

// TestExtendedCollectionCreation tests ExtendedCollection creation via Build()
func TestExtendedCollectionCreation(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	collection := NewCollection[testutil.TestDoc](ctx, client, testutil.TestCollectionName)

	// Test Build() creates ExtendedCollection
	extendedCollection := collection.Build(ctx)
	testutil.AssertNotNil(t, extendedCollection, "Build() should return an ExtendedCollection")

	// Test initial filter is empty
	filter := extendedCollection.GetFilter()
	testutil.AssertNotNil(t, filter, "GetFilter() should return a filter")
}

// TestExtendedCollectionByMethod tests the By() method for field-value filtering
func TestExtendedCollectionByMethod(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	collection := NewCollection[testutil.TestDoc](ctx, client, testutil.TestCollectionName)
	extendedCollection := collection.Build(ctx)

	// Test By() with string field
	byName := extendedCollection.By("Name", "test_name")
	testutil.AssertNotNil(t, byName, "By() should return an ExtendedCollection")

	filter := byName.GetFilter().(bson.M)
	testutil.AssertEqual(t, "test_name", filter["name"], "By() should set name field in filter")

	// Test By() with integer field
	byValue := extendedCollection.By("Value", 42)
	filter = byValue.GetFilter().(bson.M)
	testutil.AssertEqual(t, 42, filter["value"], "By() should set value field in filter")

	// Test By() with boolean field
	byActive := extendedCollection.By("Active", true)
	filter = byActive.GetFilter().(bson.M)
	testutil.AssertEqual(t, true, filter["active"], "By() should set active field in filter")
}

// TestExtendedCollectionWhereMethod tests the Where() method for complex filtering
func TestExtendedCollectionWhereMethod(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	collection := NewCollection[testutil.TestDoc](ctx, client, testutil.TestCollectionName)
	extendedCollection := collection.Build(ctx)

	// Test Where() with simple filter
	simpleFilter := bson.M{"name": "test"}
	whereSimple := extendedCollection.Where(simpleFilter)
	testutil.AssertNotNil(t, whereSimple, "Where() should return an ExtendedCollection")

	filter := whereSimple.GetFilter().(bson.M)
	testutil.AssertEqual(t, "test", filter["name"], "Where() should set filter fields")

	// Test Where() with complex filter
	complexFilter := bson.M{
		"$and": []bson.M{
			{"name": "test"},
			{"value": bson.M{"$gt": 10}},
		},
	}
	whereComplex := extendedCollection.Where(complexFilter)
	filter = whereComplex.GetFilter().(bson.M)
	testutil.AssertNotNil(t, filter["$and"], "Where() should preserve complex filter structure")
}

// TestExtendedCollectionMethodChaining tests chaining By() and Where() methods
func TestExtendedCollectionMethodChaining(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	collection := NewCollection[testutil.TestDoc](ctx, client, testutil.TestCollectionName)
	extendedCollection := collection.Build(ctx)

	// Test chaining By() methods
	chained := extendedCollection.By("Name", "test").By("Value", 42).By("Active", true)
	testutil.AssertNotNil(t, chained, "Chained By() calls should work")

	filter := chained.GetFilter().(bson.M)
	testutil.AssertEqual(t, "test", filter["name"], "Chained By() should preserve name")
	testutil.AssertEqual(t, 42, filter["value"], "Chained By() should preserve value")
	testutil.AssertEqual(t, true, filter["active"], "Chained By() should preserve active")

	// Test chaining By() and Where()
	mixedChain := extendedCollection.By("Name", "test").Where(bson.M{"value": bson.M{"$gt": 10}})
	filter = mixedChain.GetFilter().(bson.M)
	testutil.AssertEqual(t, "test", filter["name"], "Mixed chain should preserve By() field")
	testutil.AssertNotNil(t, filter["value"], "Mixed chain should preserve Where() field")
}

// TestExtendedCollectionQueryExecution tests query execution methods
func TestExtendedCollectionQueryExecution(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	collection := NewCollection[testutil.TestDoc](ctx, client, testutil.TestCollectionName)
	extendedCollection := collection.Build(ctx).By("Name", "test")

	// Test First() method
	var result testutil.TestDoc
	err = extendedCollection.First(&result)
	// Error expected with fake client, but should not panic

	// Test Many() method
	var results []testutil.TestDoc
	err = extendedCollection.Many(&results)
	// Error expected with fake client, but should not panic

	// Test Count() method
	count, err := extendedCollection.Count()
	// Error expected with fake client, but should not panic
	_ = count

	// Test Exists() method
	exists, err := extendedCollection.Exists()
	// Error expected with fake client, but should not panic
	_ = exists
}

// TestExtendedCollectionUpdateOperations tests update operations
func TestExtendedCollectionUpdateOperations(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	collection := NewCollection[testutil.TestDoc](ctx, client, testutil.TestCollectionName)
	extendedCollection := collection.Build(ctx).By("Name", "test")

	// Test Save() method
	update := bson.M{"$set": bson.M{"value": 100}}
	err = extendedCollection.Save(update)
	// Error expected with fake client, but should not panic

	// Test SaveMany() method
	err = extendedCollection.SaveMany(update)
	// Error expected with fake client, but should not panic

	// Test Delete() method
	err = extendedCollection.Delete()
	// Error expected with fake client, but should not panic
}

// TestExtendedCollectionFilterComposition tests filter composition behavior
func TestExtendedCollectionFilterComposition(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	collection := NewCollection[testutil.TestDoc](ctx, client, testutil.TestCollectionName)
	extendedCollection := collection.Build(ctx)

	// Test that multiple By() calls accumulate in filter
	step1 := extendedCollection.By("Name", "test")
	filter1 := step1.GetFilter().(bson.M)
	testutil.AssertEqual(t, 1, len(filter1), "First By() should add one field")

	step2 := step1.By("Value", 42)
	filter2 := step2.GetFilter().(bson.M)
	testutil.AssertEqual(t, 2, len(filter2), "Second By() should add another field")
	testutil.AssertEqual(t, "test", filter2["name"], "Previous field should be preserved")
	testutil.AssertEqual(t, 42, filter2["value"], "New field should be added")

	// Test that Where() merges with existing filter
	step3 := step2.Where(bson.M{"active": true})
	filter3 := step3.GetFilter().(bson.M)
	testutil.AssertEqual(t, 3, len(filter3), "Where() should merge with existing filter")
	testutil.AssertEqual(t, "test", filter3["name"], "By() fields should be preserved")
	testutil.AssertEqual(t, 42, filter3["value"], "By() fields should be preserved")
	testutil.AssertEqual(t, true, filter3["active"], "Where() field should be added")
}

// TestExtendedCollectionIsolation tests that ExtendedCollection instances are isolated
func TestExtendedCollectionIsolation(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	collection := NewCollection[testutil.TestDoc](ctx, client, testutil.TestCollectionName)
	extendedCollection := collection.Build(ctx)

	// Create two separate chains
	chain1 := extendedCollection.By("Name", "test1")
	chain2 := extendedCollection.By("Name", "test2")

	// Verify they have different filters
	filter1 := chain1.GetFilter().(bson.M)
	filter2 := chain2.GetFilter().(bson.M)

	testutil.AssertEqual(t, "test1", filter1["name"], "Chain1 should have its own filter")
	testutil.AssertEqual(t, "test2", filter2["name"], "Chain2 should have its own filter")

	// Verify original is unchanged
	originalFilter := extendedCollection.GetFilter().(bson.M)
	testutil.AssertEqual(t, 0, len(originalFilter), "Original ExtendedCollection should remain unchanged")
}

// TestExtendedCollectionFieldNameMapping tests field name to BSON tag mapping
func TestExtendedCollectionFieldNameMapping(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	collection := NewCollection[testutil.TestDoc](ctx, client, testutil.TestCollectionName)
	extendedCollection := collection.Build(ctx)

	// Test that struct field names are mapped to BSON field names
	byName := extendedCollection.By("Name", "test")
	filter := byName.GetFilter().(bson.M)

	// Should use BSON tag "name" not struct field "Name"
	_, hasName := filter["name"]
	_, hasNameCapital := filter["Name"]

	testutil.AssertTrue(t, hasName, "Filter should use BSON tag 'name'")
	testutil.AssertFalse(t, hasNameCapital, "Filter should not use struct field 'Name'")

	// Test with ID field (should map to _id)
	byID := extendedCollection.By("ID", "test_id")
	filter = byID.GetFilter().(bson.M)

	_, hasID := filter["_id"]
	testutil.AssertTrue(t, hasID, "ID field should map to '_id' BSON tag")
}
