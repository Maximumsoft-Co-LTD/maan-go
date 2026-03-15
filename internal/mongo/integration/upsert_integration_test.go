// FILE: internal/mongo/integration/upsert_integration_test.go
// PACKAGE: mongo_integration_test
// PURPOSE: Behavioral correctness tests for Save (upsert), SaveMany, Upd, and UpdMany.

package mongo_integration_test

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func TestSave_CreatesNewDoc(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "upsert_save_new")

	err := coll.Save(bson.M{"name": "new"}, bson.M{"$set": bson.M{"name": "new", "value": 42}})
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	count, err := coll.Count(nil)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}

	var doc integDoc
	if err := coll.FindOne(bson.M{"name": "new"}).Result(&doc); err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}
	if doc.Value != 42 {
		t.Fatalf("expected Value 42, got %d", doc.Value)
	}
}

func TestSave_UpdatesExistingDoc(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "upsert_save_upd")

	doc := integDoc{Name: "existing", Value: 10}
	if err := coll.Create(&doc); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err := coll.Save(bson.M{"name": "existing"}, bson.M{"$set": bson.M{"value": 99}})
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	var fetched integDoc
	if err := coll.FindOne(bson.M{"name": "existing"}).Result(&fetched); err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}
	if fetched.Value != 99 {
		t.Fatalf("expected Value 99, got %d", fetched.Value)
	}

	count, err := coll.Count(nil)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count 1 (not duplicated), got %d", count)
	}
}

func TestSave_InjectsUpdatedAt(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "upsert_save_ts")

	before := time.Now().UTC().Add(-1 * time.Minute)
	err := coll.Save(bson.M{"name": "ts_test"}, bson.M{"$set": bson.M{"name": "ts_test", "value": 1}})
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	var doc integDoc
	if err := coll.FindOne(bson.M{"name": "ts_test"}).Result(&doc); err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}
	if doc.UpdatedAt.IsZero() {
		t.Fatal("expected non-zero UpdatedAt")
	}
	if doc.UpdatedAt.Before(before) {
		t.Fatalf("UpdatedAt %v is before test start %v", doc.UpdatedAt, before)
	}
}

func TestSaveMany_CreatesMultiple(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "upsert_save_many_new")

	err := coll.SaveMany(bson.M{"active": true}, bson.M{"$set": bson.M{"name": "bulk", "active": true}})
	if err != nil {
		t.Fatalf("SaveMany failed: %v", err)
	}

	count, err := coll.Count(nil)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}
}

func TestSaveMany_UpdatesMultiple(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "upsert_save_many_upd")

	docs := []integDoc{
		{Name: "a", Value: 1, Active: true},
		{Name: "b", Value: 2, Active: true},
		{Name: "c", Value: 3, Active: true},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	err := coll.SaveMany(bson.M{"active": true}, bson.M{"$set": bson.M{"value": 999}})
	if err != nil {
		t.Fatalf("SaveMany failed: %v", err)
	}

	updated, err := coll.FindMany(bson.M{"value": 999}).All()
	if err != nil {
		t.Fatalf("FindMany failed: %v", err)
	}
	if len(updated) != 3 {
		t.Fatalf("expected 3 updated docs, got %d", len(updated))
	}
}

func TestUpd_UpdatesExistingDoc(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "upsert_upd")

	doc := integDoc{Name: "upd_test", Value: 10}
	if err := coll.Create(&doc); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := coll.Upd(bson.M{"name": "upd_test"}, bson.M{"$set": bson.M{"value": 50}}); err != nil {
		t.Fatalf("Upd failed: %v", err)
	}

	var fetched integDoc
	if err := coll.FindOne(bson.M{"name": "upd_test"}).Result(&fetched); err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}
	if fetched.Value != 50 {
		t.Fatalf("expected Value 50, got %d", fetched.Value)
	}
}

func TestUpd_NoInsertOnNoMatch(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "upsert_upd_noinsert")

	err := coll.Upd(bson.M{"name": "nonexistent"}, bson.M{"$set": bson.M{"value": 99}})
	if err != nil {
		t.Fatalf("Upd with no match should not error, got %v", err)
	}

	count, err := coll.Count(nil)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected count 0 (Upd does NOT upsert), got %d", count)
	}
}

func TestUpd_InjectsUpdatedAt(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "upsert_upd_ts")

	doc := integDoc{Name: "upd_ts_test", Value: 10}
	if err := coll.Create(&doc); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var original integDoc
	if err := coll.FindOne(bson.M{"_id": doc.ID}).Result(&original); err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	if err := coll.Upd(bson.M{"_id": doc.ID}, bson.M{"$set": bson.M{"value": 77}}); err != nil {
		t.Fatalf("Upd failed: %v", err)
	}

	var updated integDoc
	if err := coll.FindOne(bson.M{"_id": doc.ID}).Result(&updated); err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}
	if !updated.UpdatedAt.After(original.UpdatedAt) && !updated.UpdatedAt.Equal(original.UpdatedAt) {
		t.Fatalf("expected UpdatedAt >= original, got %v vs %v", updated.UpdatedAt, original.UpdatedAt)
	}
}

func TestUpdMany_UpdatesAllMatching(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "upsert_updmany")

	docs := []integDoc{
		{Name: "a", Value: 1, Active: true},
		{Name: "b", Value: 2, Active: true},
		{Name: "c", Value: 3, Active: true},
		{Name: "d", Value: 4, Active: false},
		{Name: "e", Value: 5, Active: false},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	if err := coll.UpdMany(bson.M{"active": true}, bson.M{"$set": bson.M{"value": 888}}); err != nil {
		t.Fatalf("UpdMany failed: %v", err)
	}

	updated, err := coll.FindMany(bson.M{"value": 888}).All()
	if err != nil {
		t.Fatalf("FindMany failed: %v", err)
	}
	if len(updated) != 3 {
		t.Fatalf("expected 3 updated docs, got %d", len(updated))
	}
}

func TestUpdMany_NoInsertOnNoMatch(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "upsert_updmany_noinsert")

	err := coll.UpdMany(bson.M{"name": "nonexistent"}, bson.M{"$set": bson.M{"value": 1}})
	if err != nil {
		t.Fatalf("UpdMany with no match should not error, got %v", err)
	}

	count, err := coll.Count(nil)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected count 0, got %d", count)
	}
}
