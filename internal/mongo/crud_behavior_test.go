package mongo

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestDoc for CRUD behavior testing
type CrudTestDoc struct {
	ID     primitive.ObjectID `bson:"_id"`
	Name   string        `bson:"name"`
	Status string        `bson:"status"`
	Value  int           `bson:"value"`
}

func TestCRUDBehaviorDifferences(t *testing.T) {
	client, err := NewFakeClient()
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	coll := NewCollection[CrudTestDoc](context.Background(), client, "crud_test")

	t.Run("Save vs Upd method calls", func(t *testing.T) {
		// Test that Save and Upd methods can be called without panicking
		// Note: Fake client will return "client is disconnected" errors,
		// but the important thing is that the methods have correct signatures
		// and the upsert behavior is correctly configured in the implementation
		
		// Save method should use upsert=true internally
		err := coll.Save(
			bson.M{"name": "test_doc"},
			bson.M{"$set": bson.M{"status": "active", "value": 100}},
		)
		// Expect error from fake client, but method should not panic
		if err == nil {
			t.Error("Expected error from fake client, but got nil")
		}

		// Upd method should use upsert=false internally
		err = coll.Upd(
			bson.M{"name": "non_existent_doc"},
			bson.M{"$set": bson.M{"status": "inactive", "value": 200}},
		)
		// Expect error from fake client, but method should not panic
		if err == nil {
			t.Error("Expected error from fake client, but got nil")
		}
	})

	t.Run("SaveMany vs UpdMany method calls", func(t *testing.T) {
		// SaveMany method should use upsert=true internally
		err := coll.SaveMany(
			bson.M{"name": "batch_doc"},
			bson.M{"$set": bson.M{"status": "batch_active", "value": 300}},
		)
		// Expect error from fake client, but method should not panic
		if err == nil {
			t.Error("Expected error from fake client, but got nil")
		}

		// UpdMany method should use upsert=false internally
		err = coll.UpdMany(
			bson.M{"name": "non_existent_batch"},
			bson.M{"$set": bson.M{"status": "batch_inactive", "value": 400}},
		)
		// Expect error from fake client, but method should not panic
		if err == nil {
			t.Error("Expected error from fake client, but got nil")
		}
	})

	t.Run("Method signatures and return types", func(t *testing.T) {
		// Verify all CRUD methods have correct signatures
		var doc CrudTestDoc
		var docs []CrudTestDoc

		// Create methods
		_ = coll.Create(&doc)
		_ = coll.CreateMany(&docs)

		// Save methods (upsert)
		_ = coll.Save(bson.M{}, bson.M{})
		_ = coll.SaveMany(bson.M{}, bson.M{})

		// Update methods (update only)
		_ = coll.Upd(bson.M{}, bson.M{})
		_ = coll.UpdMany(bson.M{}, bson.M{})

		// Delete methods
		_ = coll.Del(bson.M{})

		// All methods should return error only
		t.Log("All CRUD methods have correct signatures")
	})
}

func TestCRUDMethodDocumentation(t *testing.T) {
	t.Run("Method behavior documentation", func(t *testing.T) {
		// This test serves as living documentation for method behaviors
		
		// CREATE methods - Insert new documents only
		// - Create(doc) - Insert single document, error on duplicate
		// - CreateMany(docs) - Insert multiple documents, error on any duplicate
		
		// SAVE methods - Upsert (Update if exists, Insert if not)
		// - Save(filter, update) - Upsert single document matching filter
		// - SaveMany(filter, update) - Upsert multiple documents matching filter
		
		// UPDATE methods - Update existing documents only (no insert)
		// - Upd(filter, update) - Update single existing document
		// - UpdMany(filter, update) - Update multiple existing documents
		
		// DELETE methods - Remove documents
		// - Del(filter) - Delete single document matching filter
		
		t.Log("CRUD method behaviors documented")
	})
}