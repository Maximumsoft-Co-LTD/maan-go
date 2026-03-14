// FILE: internal/mongo/integration/atomic_ops_integration_test.go
// PACKAGE: mongo_integration_test
// PURPOSE: Behavioral correctness tests for FindOneAndUpd, FindOneAndDel, Del, and DelMany.

package mongo_integration_test

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mg "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestFindOneAndUpd_ReturnsUpdatedDoc(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "atomic_fau")

	doc := integDoc{Name: "fau_test", Value: 10}
	if err := coll.Create(&doc); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var out integDoc
	err := coll.FindOneAndUpd(
		bson.M{"name": "fau_test"},
		bson.M{"$set": bson.M{"value": 99}},
		&out,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	if err != nil {
		t.Fatalf("FindOneAndUpd failed: %v", err)
	}
	if out.Value != 99 {
		t.Fatalf("expected Value 99, got %d", out.Value)
	}
	if out.Name != "fau_test" {
		t.Fatalf("expected Name 'fau_test', got %q", out.Name)
	}
}

func TestFindOneAndUpd_NoMatch(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "atomic_fau_nomatch")

	var out integDoc
	err := coll.FindOneAndUpd(bson.M{"_id": primitive.NewObjectID()}, bson.M{"$set": bson.M{"value": 1}}, &out)
	if err != mg.ErrNoDocuments {
		t.Fatalf("expected ErrNoDocuments, got %v", err)
	}
}

func TestFindOneAndUpd_InjectsUpdatedAt(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "atomic_fau_ts")

	doc := integDoc{Name: "fau_ts", Value: 1}
	if err := coll.Create(&doc); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	before := time.Now().UTC().Add(-1 * time.Minute)
	var out integDoc
	err := coll.FindOneAndUpd(
		bson.M{"_id": doc.ID},
		bson.M{"$set": bson.M{"value": 2}},
		&out,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	if err != nil {
		t.Fatalf("FindOneAndUpd failed: %v", err)
	}
	if out.UpdatedAt.IsZero() {
		t.Fatal("expected non-zero UpdatedAt")
	}
	if out.UpdatedAt.Before(before) {
		t.Fatalf("UpdatedAt %v is before test start %v", out.UpdatedAt, before)
	}
}

func TestFindOneAndDel_ReturnsDeletedDoc(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "atomic_fad")

	doc := integDoc{Name: "fad_test", Value: 42}
	if err := coll.Create(&doc); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var out integDoc
	if err := coll.FindOneAndDel(bson.M{"name": "fad_test"}, &out); err != nil {
		t.Fatalf("FindOneAndDel failed: %v", err)
	}
	if out.Name != "fad_test" {
		t.Fatalf("expected Name 'fad_test', got %q", out.Name)
	}
	if out.Value != 42 {
		t.Fatalf("expected Value 42, got %d", out.Value)
	}

	count, err := coll.Count(nil)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected count 0 after delete, got %d", count)
	}
}

func TestFindOneAndDel_NoMatch(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "atomic_fad_nomatch")

	var out integDoc
	err := coll.FindOneAndDel(bson.M{"_id": primitive.NewObjectID()}, &out)
	if err != mg.ErrNoDocuments {
		t.Fatalf("expected ErrNoDocuments, got %v", err)
	}
}

func TestDel_DeletesOneDoc(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "atomic_del")

	docs := []integDoc{
		{Name: "a", Value: 1, Active: true},
		{Name: "b", Value: 2, Active: true},
		{Name: "c", Value: 3, Active: false},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	if err := coll.Del(bson.M{"active": true}); err != nil {
		t.Fatalf("Del failed: %v", err)
	}

	total, err := coll.Count(nil)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total count 2, got %d", total)
	}

	activeCount, err := coll.Count(bson.M{"active": true})
	if err != nil {
		t.Fatalf("Count active failed: %v", err)
	}
	if activeCount != 1 {
		t.Fatalf("expected 1 active doc remaining, got %d", activeCount)
	}
}

func TestDelMany_DeletesAllMatching(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "atomic_delmany")

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

	if err := coll.DelMany(bson.M{"active": true}); err != nil {
		t.Fatalf("DelMany failed: %v", err)
	}

	total, err := coll.Count(nil)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total count 2, got %d", total)
	}

	activeCount, err := coll.Count(bson.M{"active": true})
	if err != nil {
		t.Fatalf("Count active failed: %v", err)
	}
	if activeCount != 0 {
		t.Fatalf("expected 0 active docs, got %d", activeCount)
	}
}
