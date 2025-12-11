package mongo

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo/testutil"
)

// TestModelDefaultsWithCreate tests that model defaults are applied during Create operations
func TestModelDefaultsWithCreate(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	coll := NewCollection[testutil.DefaultTestDoc](ctx, client, "test_defaults")

	// Create a document without defaults set
	doc := &testutil.DefaultTestDoc{
		Name: "test_document",
		// ID, CreatedAt, UpdatedAt are zero values
	}

	// Store original state
	originalID := doc.ID
	originalCreatedAt := doc.CreatedAt
	originalUpdatedAt := doc.UpdatedAt

	// Create the document (will fail with fake client, but defaults should still be applied)
	err = coll.Create(doc)
	// Error is expected with fake client, but defaults should have been applied before the operation

	// Verify that defaults were applied
	if originalID.IsZero() && doc.ID.IsZero() {
		t.Error("Expected ID to be generated, but it's still zero")
	}
	if originalCreatedAt.IsZero() && doc.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set, but it's still zero")
	}
	if originalUpdatedAt.IsZero() && doc.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set, but it's still zero")
	}

	// Verify that the generated values are reasonable
	now := time.Now().UTC()
	if !doc.CreatedAt.IsZero() {
		timeDiff := now.Sub(doc.CreatedAt)
		if timeDiff < 0 || timeDiff > time.Minute {
			t.Errorf("CreatedAt timestamp seems unreasonable: %v (diff from now: %v)", doc.CreatedAt, timeDiff)
		}
	}
	if !doc.UpdatedAt.IsZero() {
		timeDiff := now.Sub(doc.UpdatedAt)
		if timeDiff < 0 || timeDiff > time.Minute {
			t.Errorf("UpdatedAt timestamp seems unreasonable: %v (diff from now: %v)", doc.UpdatedAt, timeDiff)
		}
	}
}

// TestModelDefaultsWithCreateMany tests that model defaults are applied during CreateMany operations
func TestModelDefaultsWithCreateMany(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	coll := NewCollection[testutil.DefaultTestDoc](ctx, client, "test_defaults")

	// Create multiple documents without defaults set
	docs := []testutil.DefaultTestDoc{
		{Name: "doc1"},
		{Name: "doc2"},
		{Name: "doc3"},
	}

	// Store original states
	originalStates := make([]struct {
		ID        primitive.ObjectID
		CreatedAt time.Time
		UpdatedAt time.Time
	}, len(docs))

	for i, doc := range docs {
		originalStates[i] = struct {
			ID        primitive.ObjectID
			CreatedAt time.Time
			UpdatedAt time.Time
		}{doc.ID, doc.CreatedAt, doc.UpdatedAt}
	}

	// Create the documents (will fail with fake client, but defaults should still be applied)
	err = coll.CreateMany(&docs)
	// Error is expected with fake client, but defaults should have been applied before the operation

	// Verify that defaults were applied to all documents
	for i, doc := range docs {
		original := originalStates[i]

		if original.ID.IsZero() && doc.ID.IsZero() {
			t.Errorf("Document %d: Expected ID to be generated, but it's still zero", i)
		}
		if original.CreatedAt.IsZero() && doc.CreatedAt.IsZero() {
			t.Errorf("Document %d: Expected CreatedAt to be set, but it's still zero", i)
		}
		if original.UpdatedAt.IsZero() && doc.UpdatedAt.IsZero() {
			t.Errorf("Document %d: Expected UpdatedAt to be set, but it's still zero", i)
		}

		// Verify that each document has unique IDs
		for j, otherDoc := range docs {
			if i != j && doc.ID == otherDoc.ID && !doc.ID.IsZero() {
				t.Errorf("Documents %d and %d have the same ID: %v", i, j, doc.ID)
			}
		}
	}
}

// TestModelDefaultsPreserveExistingValues tests that existing values are not overwritten
func TestModelDefaultsPreserveExistingValues(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	coll := NewCollection[testutil.DefaultTestDoc](ctx, client, "test_defaults")

	// Create a document with existing values
	existingID := primitive.NewObjectID()
	existingCreatedAt := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	existingUpdatedAt := time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC)

	doc := &testutil.DefaultTestDoc{
		ID:        existingID,
		Name:      "test_document",
		CreatedAt: existingCreatedAt,
		UpdatedAt: existingUpdatedAt,
	}

	// Create the document (will fail with fake client, but defaults should still be applied)
	err = coll.Create(doc)
	// Error is expected with fake client, but defaults should have been applied before the operation

	// Verify that existing values were preserved
	if doc.ID != existingID {
		t.Errorf("Expected ID to be preserved as %v, but got %v", existingID, doc.ID)
	}
	if doc.CreatedAt != existingCreatedAt {
		t.Errorf("Expected CreatedAt to be preserved as %v, but got %v", existingCreatedAt, doc.CreatedAt)
	}
	if doc.UpdatedAt != existingUpdatedAt {
		t.Errorf("Expected UpdatedAt to be preserved as %v, but got %v", existingUpdatedAt, doc.UpdatedAt)
	}
}

// TestModelDefaultsWithRegularDocuments tests that documents without default interfaces are unaffected
func TestModelDefaultsWithRegularDocuments(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	coll := NewCollection[testutil.TestDoc](ctx, client, "test_regular")

	// Create a regular document (doesn't implement default interfaces)
	doc := &testutil.TestDoc{
		Name:   "test_document",
		Value:  42,
		Active: true,
		// ID is zero value
	}

	// Store original state
	originalID := doc.ID

	// Create the document (will fail with fake client, but should not apply defaults)
	err = coll.Create(doc)
	// Error is expected with fake client

	// Verify that the ID remains zero (no defaults applied)
	if doc.ID != originalID {
		t.Errorf("Expected ID to remain unchanged as %v, but got %v", originalID, doc.ID)
	}
	if !doc.ID.IsZero() {
		t.Error("Expected ID to remain zero for regular documents, but it was set")
	}
}

// TestApplyModelDefaultsDirectly tests the applyModelDefaults function directly
func TestApplyModelDefaultsDirectly(t *testing.T) {
	t.Run("with nil document", func(t *testing.T) {
		// Should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("applyModelDefaults panicked with nil: %v", r)
			}
		}()

		applyModelDefaults(nil)
	})

	t.Run("with document implementing defaults", func(t *testing.T) {
		doc := &testutil.DefaultTestDoc{
			Name: "test",
		}

		// Store original state
		originalID := doc.ID
		originalCreatedAt := doc.CreatedAt
		originalUpdatedAt := doc.UpdatedAt

		applyModelDefaults(doc)

		// Verify defaults were applied
		if originalID.IsZero() && doc.ID.IsZero() {
			t.Error("Expected ID to be generated")
		}
		if originalCreatedAt.IsZero() && doc.CreatedAt.IsZero() {
			t.Error("Expected CreatedAt to be set")
		}
		if originalUpdatedAt.IsZero() && doc.UpdatedAt.IsZero() {
			t.Error("Expected UpdatedAt to be set")
		}
	})

	t.Run("with document not implementing defaults", func(t *testing.T) {
		doc := &testutil.TestDoc{
			Name: "test",
		}

		// Store original state
		originalID := doc.ID

		applyModelDefaults(doc)

		// Verify nothing changed
		if doc.ID != originalID {
			t.Error("Expected ID to remain unchanged for non-default document")
		}
	})

	t.Run("with existing values", func(t *testing.T) {
		existingID := primitive.NewObjectID()
		existingCreatedAt := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		existingUpdatedAt := time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC)

		doc := &testutil.DefaultTestDoc{
			ID:        existingID,
			Name:      "test",
			CreatedAt: existingCreatedAt,
			UpdatedAt: existingUpdatedAt,
		}

		applyModelDefaults(doc)

		// Verify existing values were preserved
		if doc.ID != existingID {
			t.Error("Expected existing ID to be preserved")
		}
		if doc.CreatedAt != existingCreatedAt {
			t.Error("Expected existing CreatedAt to be preserved")
		}
		if doc.UpdatedAt != existingUpdatedAt {
			t.Error("Expected existing UpdatedAt to be preserved")
		}
	})
}

// TestModelDefaultsIntegration tests model defaults integration with collection interface
func TestModelDefaultsIntegration(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	coll := NewCollection[testutil.DefaultTestDoc](ctx, client, "test_integration")

	// Create a document
	doc := &testutil.DefaultTestDoc{
		Name: "integration_test",
	}

	// Store original state
	originalID := doc.ID
	originalCreatedAt := doc.CreatedAt
	originalUpdatedAt := doc.UpdatedAt

	// Create the document (will fail with fake client, but defaults should be applied)
	err = coll.Create(doc)
	// Error is expected with fake client

	// Verify the document was populated with defaults
	if originalID.IsZero() && doc.ID.IsZero() {
		t.Error("Expected ID to be generated")
	}
	if originalCreatedAt.IsZero() && doc.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	if originalUpdatedAt.IsZero() && doc.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}

	// Test that collection interface methods don't panic with default documents
	singleResult := coll.FindOne(map[string]interface{}{"name": "integration_test"})
	if singleResult == nil {
		t.Error("Expected FindOne to return a result interface")
	}

	manyResult := coll.FindMany(map[string]interface{}{"name": "integration_test"})
	if manyResult == nil {
		t.Error("Expected FindMany to return a result interface")
	}

	// Test with extended collection interface
	extColl := coll.Build(ctx)
	if extColl == nil {
		t.Error("Expected Build to return an extended collection interface")
	}

	// Test that extended collection methods don't panic
	chainedColl := extColl.By("name", "integration_test")
	if chainedColl == nil {
		t.Error("Expected By to return a chained collection interface")
	}
}
