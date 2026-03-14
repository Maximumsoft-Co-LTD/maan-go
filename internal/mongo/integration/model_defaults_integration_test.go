// FILE: internal/mongo/integration/model_defaults_integration_test.go
// PACKAGE: mongo_integration_test
// PURPOSE: Behavioral correctness tests for model-default hooks (DefaultId, DefaultCreatedAt, DefaultUpdatedAt).

package mongo_integration_test

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestModelDefaults_Create_PopulatesIDAndTimestamps(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "defaults_create")

	doc := &integDoc{Name: "defaults_test", Value: 1}
	if err := coll.Create(doc); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if doc.ID.IsZero() {
		t.Fatal("expected non-zero ID after Create")
	}

	var fetched integDoc
	if err := coll.FindOne(bson.M{"_id": doc.ID}).Result(&fetched); err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}
	if fetched.ID != doc.ID {
		t.Fatalf("expected ID %v, got %v", doc.ID, fetched.ID)
	}

	oneMinuteAgo := time.Now().UTC().Add(-1 * time.Minute)
	if fetched.CreatedAt.IsZero() || fetched.CreatedAt.Before(oneMinuteAgo) {
		t.Fatalf("expected recent CreatedAt, got %v", fetched.CreatedAt)
	}
	if fetched.UpdatedAt.IsZero() || fetched.UpdatedAt.Before(oneMinuteAgo) {
		t.Fatalf("expected recent UpdatedAt, got %v", fetched.UpdatedAt)
	}
}

func TestModelDefaults_Create_PreservesExistingID(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "defaults_preserve_id")

	presetID := primitive.NewObjectID()
	doc := &integDoc{ID: presetID, Name: "preset_id"}
	if err := coll.Create(doc); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if doc.ID != presetID {
		t.Fatalf("expected ID %v (not overwritten), got %v", presetID, doc.ID)
	}

	var fetched integDoc
	if err := coll.FindOne(bson.M{"_id": presetID}).Result(&fetched); err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}
	if fetched.Name != "preset_id" {
		t.Fatalf("expected Name 'preset_id', got %q", fetched.Name)
	}
}

func TestModelDefaults_CreateMany_PopulatesAllDocs(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "defaults_create_many")

	docs := []integDoc{
		{Name: "many_1"},
		{Name: "many_2"},
		{Name: "many_3"},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	for i, doc := range docs {
		if doc.ID.IsZero() {
			t.Fatalf("docs[%d] has zero ID", i)
		}
		if doc.CreatedAt.IsZero() {
			t.Fatalf("docs[%d] has zero CreatedAt", i)
		}
		if doc.UpdatedAt.IsZero() {
			t.Fatalf("docs[%d] has zero UpdatedAt", i)
		}
	}

	all, err := coll.FindMany(nil).All()
	if err != nil {
		t.Fatalf("FindMany failed: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 docs, got %d", len(all))
	}
}

func TestModelDefaults_Save_InjectsUpdatedAt(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "defaults_save_ts")

	doc := &integDoc{Name: "save_ts_test", Value: 1}
	if err := coll.Create(doc); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var original integDoc
	if err := coll.FindOne(bson.M{"_id": doc.ID}).Result(&original); err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	if err := coll.Save(bson.M{"_id": doc.ID}, bson.M{"$set": bson.M{"name": "updated_name"}}); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	var updated integDoc
	if err := coll.FindOne(bson.M{"_id": doc.ID}).Result(&updated); err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}

	if !updated.UpdatedAt.After(original.UpdatedAt) {
		t.Fatalf("expected UpdatedAt to be after original, got %v vs %v", updated.UpdatedAt, original.UpdatedAt)
	}
	if updated.CreatedAt != original.CreatedAt {
		t.Fatalf("expected CreatedAt unchanged, got %v vs %v", updated.CreatedAt, original.CreatedAt)
	}
}
