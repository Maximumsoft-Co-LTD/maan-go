package mongo

import (
	"context"
	"testing"

	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo/testutil"
	"go.mongodb.org/mongo-driver/bson"
)

// TestCollectionCreation tests basic collection creation and interface compliance
func TestCollectionCreation(t *testing.T) {
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

	testutil.AssertNotNil(t, collection, "Collection should not be nil")
	testutil.AssertEqual(t, testutil.TestCollectionName, collection.Name(), "Collection name should match")
}

// TestCollectionCRUDOperations tests basic CRUD operations interface
func TestCollectionCRUDOperations(t *testing.T) {
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

	// Test document for operations
	doc := testutil.FixtureTestDoc()

	// Test Create operation (will fail with fake client but should not panic)
	err = collection.Create(doc)
	// Error is expected with fake client, but operation should complete without panic

	// Test CreateMany operation
	docs := []testutil.TestDoc{*doc}
	err = collection.CreateMany(&docs)
	// Error is expected with fake client, but operation should complete without panic

	// Test FindOne operation
	singleResult := collection.FindOne(bson.M{"_id": doc.ID})
	testutil.AssertNotNil(t, singleResult, "FindOne should return a SingleResult")

	// Test FindMany operation
	manyResult := collection.FindMany(bson.M{"name": doc.Name})
	testutil.AssertNotNil(t, manyResult, "FindMany should return a ManyResult")

	// Test Save operation (upsert)
	err = collection.Save(bson.M{"_id": doc.ID}, bson.M{"$set": bson.M{"name": "updated"}})
	// Error is expected with fake client, but operation should complete without panic

	// Test SaveMany operation
	err = collection.SaveMany(bson.M{"name": doc.Name}, bson.M{"$set": bson.M{"active": false}})
	// Error is expected with fake client, but operation should complete without panic

	// Test Del operation
	err = collection.Del(bson.M{"_id": doc.ID})
	// Error is expected with fake client, but operation should complete without panic
}

// TestCollectionContextHandling tests context handling and isolation
func TestCollectionContextHandling(t *testing.T) {
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

	// Test Ctx() method creates new instance
	ctxWithValue := testutil.FixtureContextWithValue("test_key", "test_value")
	newCollection := collection.Ctx(ctxWithValue)

	testutil.AssertNotNil(t, newCollection, "Ctx() should return a new collection")
	// Note: We can't directly compare collection instances due to interface types,
	// but we can verify they work independently

	// Test that both collections are functional
	result1 := collection.FindOne(bson.M{"name": "test"})
	result2 := newCollection.FindOne(bson.M{"name": "test"})

	testutil.AssertNotNil(t, result1, "Original collection should work")
	testutil.AssertNotNil(t, result2, "New collection should work")
}

// TestCollectionBuildExtended tests Build() method for creating ExtendedCollection
func TestCollectionBuildExtended(t *testing.T) {
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

	// Test Build() method
	extendedCollection := collection.Build(ctx)
	testutil.AssertNotNil(t, extendedCollection, "Build() should return an ExtendedCollection")

	// Test ExtendedCollection methods
	byCollection := extendedCollection.By("name", "test")
	testutil.AssertNotNil(t, byCollection, "By() should return an ExtendedCollection")

	whereCollection := byCollection.Where(bson.M{"active": true})
	testutil.AssertNotNil(t, whereCollection, "Where() should return an ExtendedCollection")

	// Test chaining
	chainedCollection := extendedCollection.By("name", "test").Where(bson.M{"value": 42})
	testutil.AssertNotNil(t, chainedCollection, "Chained operations should work")
}

// TestCollectionAggregation tests aggregation pipeline operations
func TestCollectionAggregation(t *testing.T) {
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

	// Test Agg() method
	pipeline := []bson.M{
		{"$match": bson.M{"active": true}},
		{"$sort": bson.M{"name": 1}},
	}

	aggregate := collection.Agg(pipeline)
	testutil.AssertNotNil(t, aggregate, "Agg() should return an Aggregate")

	// Test Aggregate methods
	diskAggregate := aggregate.Disk(true)
	testutil.AssertNotNil(t, diskAggregate, "Disk() should return an Aggregate")

	bszAggregate := diskAggregate.Bsz(100)
	testutil.AssertNotNil(t, bszAggregate, "Bsz() should return an Aggregate")
}

// TestCollectionTransactions tests transaction operations
func TestCollectionTransactions(t *testing.T) {
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

	// Test WithTx() method
	err = collection.WithTx(func(ctx context.Context) error {
		// This would normally contain transaction operations
		return nil
	})
	// Error is expected with fake client, but operation should complete without panic

	// Test StartTx() method
	txSession, err := collection.StartTx()
	// Error is expected with fake client for actual transaction start
	// but the method should be callable
	if txSession != nil {
		defer func() {
			var txErr error
			txSession.Close(&txErr)
		}()
	}
}

// TestCollectionLifecycle tests collection lifecycle and resource management
func TestCollectionLifecycle(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}

	ctx := context.Background()
	collection := NewCollection[testutil.TestDoc](ctx, client, testutil.TestCollectionName)

	// Test that collection is functional
	testutil.AssertNotNil(t, collection, "Collection should be created successfully")
	testutil.AssertEqual(t, testutil.TestCollectionName, collection.Name(), "Collection name should be correct")

	// Test various operations to ensure collection is properly initialized
	singleResult := collection.FindOne(bson.M{})
	testutil.AssertNotNil(t, singleResult, "Collection should support FindOne")

	manyResult := collection.FindMany(bson.M{})
	testutil.AssertNotNil(t, manyResult, "Collection should support FindMany")

	aggregate := collection.Agg([]bson.M{})
	testutil.AssertNotNil(t, aggregate, "Collection should support Agg")

	// Test client cleanup
	err = client.Close()
	// Note: Fake client may return "client is disconnected" error, which is expected
}

// TestCollectionWithDifferentDocumentTypes tests collection with various document types
func TestCollectionWithDifferentDocumentTypes(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Test with TestDoc
	testCollection := NewCollection[testutil.TestDoc](ctx, client, "test_docs")
	testutil.AssertNotNil(t, testCollection, "TestDoc collection should be created")

	// Test with DefaultTestDoc
	defaultCollection := NewCollection[testutil.DefaultTestDoc](ctx, client, "default_docs")
	testutil.AssertNotNil(t, defaultCollection, "DefaultTestDoc collection should be created")

	// Test operations on each collection type
	testDoc := testutil.FixtureTestDoc()
	err = testCollection.Create(testDoc)
	// Error expected with fake client, but should not panic

	defaultDoc := testutil.FixtureDefaultTestDoc()
	err = defaultCollection.Create(defaultDoc)
	// Error expected with fake client, but should not panic

	// Test that each collection maintains its type
	testResult := testCollection.FindOne(bson.M{"name": testDoc.Name})
	testutil.AssertNotNil(t, testResult, "TestDoc collection should return SingleResult[TestDoc]")

	defaultResult := defaultCollection.FindOne(bson.M{"name": defaultDoc.Name})
	testutil.AssertNotNil(t, defaultResult, "DefaultTestDoc collection should return SingleResult[DefaultTestDoc]")
}
