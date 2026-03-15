// FILE: internal/mongo/integration/index_integration_test.go
// PACKAGE: mongo_integration_test
// PURPOSE: Behavioral correctness tests for IndexManager (Idx) operations.

package mongo_integration_test

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	mg "go.mongodb.org/mongo-driver/mongo"
)

func TestIdx_CreateOne(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "idx_create_one")

	name, err := coll.Idx().CreateOne(mg.IndexModel{
		Keys: bson.D{{Key: "name", Value: 1}},
	})
	if err != nil {
		t.Fatalf("CreateOne failed: %v", err)
	}
	if name == "" {
		t.Fatal("expected non-empty index name")
	}
}

func TestIdx_CreateMany(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "idx_create_many")

	names, err := coll.Idx().CreateMany([]mg.IndexModel{
		{Keys: bson.D{{Key: "name", Value: 1}}},
		{Keys: bson.D{{Key: "value", Value: -1}}},
	})
	if err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 index names, got %d", len(names))
	}
}

func TestIdx_List(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "idx_list")

	_, err := coll.Idx().CreateOne(mg.IndexModel{
		Keys: bson.D{{Key: "name", Value: 1}},
	})
	if err != nil {
		t.Fatalf("CreateOne failed: %v", err)
	}

	indexes, err := coll.Idx().List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(indexes) < 2 {
		t.Fatalf("expected at least 2 indexes (default _id + name), got %d", len(indexes))
	}

	found := false
	for _, idx := range indexes {
		if key, ok := idx["key"]; ok {
			if keyDoc, ok := key.(bson.M); ok {
				if _, hasName := keyDoc["name"]; hasName {
					found = true
					break
				}
			}
		}
	}
	if !found {
		t.Fatal("expected to find 'name' index in list")
	}
}

func TestIdx_DropOne(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "idx_drop_one")

	name, err := coll.Idx().CreateOne(mg.IndexModel{
		Keys: bson.D{{Key: "name", Value: 1}},
	})
	if err != nil {
		t.Fatalf("CreateOne failed: %v", err)
	}

	if err := coll.Idx().DropOne(name); err != nil {
		t.Fatalf("DropOne failed: %v", err)
	}

	indexes, err := coll.Idx().List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(indexes) != 1 {
		t.Fatalf("expected only _id index (1), got %d", len(indexes))
	}
}

func TestIdx_DropAll(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "idx_drop_all")

	_, err := coll.Idx().CreateMany([]mg.IndexModel{
		{Keys: bson.D{{Key: "name", Value: 1}}},
		{Keys: bson.D{{Key: "value", Value: -1}}},
	})
	if err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	if err := coll.Idx().DropAll(); err != nil {
		t.Fatalf("DropAll failed: %v", err)
	}

	indexes, err := coll.Idx().List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(indexes) != 1 {
		t.Fatalf("expected only _id index (1), got %d", len(indexes))
	}
}
