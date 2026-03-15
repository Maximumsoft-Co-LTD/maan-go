// FILE: internal/mongo/integration/distinct_count_integration_test.go
// PACKAGE: mongo_integration_test
// PURPOSE: Behavioral correctness tests for Distinct and Count operations.

package mongo_integration_test

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestDistinct_ReturnsUniqueValues(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "distinct_unique")

	docs := []integDoc{
		{Name: "alpha", Value: 1},
		{Name: "beta", Value: 2},
		{Name: "alpha", Value: 3},
		{Name: "gamma", Value: 4},
		{Name: "beta", Value: 5},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	vals, err := coll.Distinct("name", nil)
	if err != nil {
		t.Fatalf("Distinct failed: %v", err)
	}
	if len(vals) != 3 {
		t.Fatalf("expected 3 unique values, got %d: %v", len(vals), vals)
	}
}

func TestDistinct_WithFilter(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "distinct_filter")

	docs := []integDoc{
		{Name: "a", Active: true},
		{Name: "b", Active: true},
		{Name: "a", Active: false},
		{Name: "c", Active: false},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	vals, err := coll.Distinct("name", bson.M{"active": true})
	if err != nil {
		t.Fatalf("Distinct with filter failed: %v", err)
	}
	if len(vals) != 2 {
		t.Fatalf("expected 2 unique values, got %d: %v", len(vals), vals)
	}
}

func TestDistinct_EmptyCollection(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "distinct_empty")

	vals, err := coll.Distinct("name", nil)
	if err != nil {
		t.Fatalf("Distinct on empty collection failed: %v", err)
	}
	if len(vals) != 0 {
		t.Fatalf("expected 0 values, got %d", len(vals))
	}
}

func TestCount_AllDocuments(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "count_all")

	docs := []integDoc{
		{Name: "a"}, {Name: "b"}, {Name: "c"}, {Name: "d"}, {Name: "e"},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	count, err := coll.Count(nil)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 5 {
		t.Fatalf("expected count 5, got %d", count)
	}
}

func TestCount_WithFilter(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "count_filter")

	docs := []integDoc{
		{Name: "a", Active: true},
		{Name: "b", Active: true},
		{Name: "c", Active: true},
		{Name: "d", Active: false},
		{Name: "e", Active: false},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	count, err := coll.Count(bson.M{"active": true})
	if err != nil {
		t.Fatalf("Count with filter failed: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected count 3, got %d", count)
	}
}

func TestCount_EmptyCollection(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "count_empty")

	count, err := coll.Count(nil)
	if err != nil {
		t.Fatalf("Count on empty collection failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected count 0, got %d", count)
	}
}
