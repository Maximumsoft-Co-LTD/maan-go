// FILE: internal/mongo/integration/aggregation_integration_test.go
// PACKAGE: mongo_integration_test
// PURPOSE: Behavioral correctness tests for Aggregate pipeline operations.

package mongo_integration_test

import (
	"context"
	"testing"

	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

// aggResult is a typed output for $group aggregation stages.
type aggResult struct {
	ID    interface{} `bson:"_id"`
	Total int         `bson:"total"`
}

func TestAgg_MatchAndProject(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "agg_match_proj")

	docs := []integDoc{
		{Name: "a", Value: 10, Active: true},
		{Name: "b", Value: 20, Active: true},
		{Name: "c", Value: 30, Active: true},
		{Name: "d", Value: 40, Active: false},
		{Name: "e", Value: 50, Active: false},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	pipeline := bson.A{
		bson.M{"$match": bson.M{"active": true}},
		bson.M{"$project": bson.M{"name": 1, "value": 1, "_id": 0}},
	}
	results, err := coll.Agg(pipeline).All()
	if err != nil {
		t.Fatalf("Agg.All failed: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}

func TestAgg_GroupAndSort(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)

	// We need to insert integDocs but read aggResults. Use two handles to same collection.
	name := uniqueColl("agg_group_sort")
	dropColl(t, client, name)
	writeColl := mongo.NewCollection[integDoc](context.Background(), client, name)

	docs := []integDoc{
		{Name: "a", Value: 10, Active: true},
		{Name: "b", Value: 20, Active: true},
		{Name: "c", Value: 30, Active: false},
	}
	if err := writeColl.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	pipeline := bson.A{
		bson.M{"$group": bson.M{"_id": "$active", "total": bson.M{"$sum": "$value"}}},
		bson.M{"$sort": bson.M{"_id": 1}},
	}
	rawResults, err := writeColl.Agg(pipeline).Raw()
	if err != nil {
		t.Fatalf("Agg.Raw failed: %v", err)
	}
	if len(rawResults) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(rawResults))
	}

	// Sorted by _id: false (first), true (second)
	for _, r := range rawResults {
		active, ok := r["_id"].(bool)
		if !ok {
			t.Fatalf("unexpected _id type: %T", r["_id"])
		}
		total, ok := r["total"].(int32)
		if !ok {
			// Try int64
			total64, ok := r["total"].(int64)
			if !ok {
				t.Fatalf("unexpected total type: %T", r["total"])
			}
			total = int32(total64)
		}
		if active && total != 30 {
			t.Fatalf("expected active=true total 30, got %d", total)
		}
		if !active && total != 30 {
			t.Fatalf("expected active=false total 30, got %d", total)
		}
	}
}

func TestAgg_Result(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "agg_result")

	docs := []integDoc{
		{Name: "a", Value: 1},
		{Name: "b", Value: 2},
		{Name: "c", Value: 3},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	pipeline := bson.A{bson.M{"$match": bson.M{}}}
	var out []integDoc
	if err := coll.Agg(pipeline).Result(&out); err != nil {
		t.Fatalf("Agg.Result failed: %v", err)
	}
	if len(out) != 3 {
		t.Fatalf("expected 3 results, got %d", len(out))
	}
}

func TestAgg_Raw(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "agg_raw")

	docs := []integDoc{
		{Name: "raw1", Value: 1},
		{Name: "raw2", Value: 2},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	pipeline := bson.A{bson.M{"$match": bson.M{}}}
	rawDocs, err := coll.Agg(pipeline).Raw()
	if err != nil {
		t.Fatalf("Agg.Raw failed: %v", err)
	}
	if len(rawDocs) != 2 {
		t.Fatalf("expected 2 raw docs, got %d", len(rawDocs))
	}
	for i, doc := range rawDocs {
		if _, ok := doc["_id"]; !ok {
			t.Fatalf("rawDoc[%d] missing _id key", i)
		}
		if _, ok := doc["name"]; !ok {
			t.Fatalf("rawDoc[%d] missing name key", i)
		}
	}
}

func TestAgg_Stream(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "agg_stream")

	docs := []integDoc{
		{Name: "s1", Value: 1},
		{Name: "s2", Value: 2},
		{Name: "s3", Value: 3},
		{Name: "s4", Value: 4},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	pipeline := bson.A{bson.M{"$match": bson.M{}}}
	var collected []integDoc
	err := coll.Agg(pipeline).Stream(func(ctx context.Context, doc integDoc) error {
		collected = append(collected, doc)
		return nil
	})
	if err != nil {
		t.Fatalf("Agg.Stream failed: %v", err)
	}
	if len(collected) != 4 {
		t.Fatalf("expected 4 streamed docs, got %d", len(collected))
	}
}

func TestAgg_Each(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "agg_each")

	docs := []integDoc{
		{Name: "e1", Value: 1},
		{Name: "e2", Value: 2},
		{Name: "e3", Value: 3},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	pipeline := bson.A{bson.M{"$match": bson.M{}}}
	counter := 0
	err := coll.Agg(pipeline).Each(func(ctx context.Context, doc integDoc) error {
		counter++
		return nil
	})
	if err != nil {
		t.Fatalf("Agg.Each failed: %v", err)
	}
	if counter != 3 {
		t.Fatalf("expected 3 calls, got %d", counter)
	}
}

func TestAgg_EachRaw(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "agg_each_raw")

	docs := []integDoc{
		{Name: "r1", Value: 1},
		{Name: "r2", Value: 2},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	pipeline := bson.A{bson.M{"$match": bson.M{}}}
	var rawDocs []bson.M
	err := coll.Agg(pipeline).EachRaw(func(ctx context.Context, doc bson.M) error {
		rawDocs = append(rawDocs, doc)
		return nil
	})
	if err != nil {
		t.Fatalf("Agg.EachRaw failed: %v", err)
	}
	if len(rawDocs) != 2 {
		t.Fatalf("expected 2 raw docs, got %d", len(rawDocs))
	}
}

func TestAgg_Disk(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "agg_disk")

	docs := []integDoc{
		{Name: "d1", Value: 1},
		{Name: "d2", Value: 2},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	pipeline := bson.A{bson.M{"$match": bson.M{}}}
	results, err := coll.Agg(pipeline).Disk(true).All()
	if err != nil {
		t.Fatalf("Agg.Disk.All failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestAgg_Bsz(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "agg_bsz")

	docs := []integDoc{
		{Name: "b1", Value: 1},
		{Name: "b2", Value: 2},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	pipeline := bson.A{bson.M{"$match": bson.M{}}}
	results, err := coll.Agg(pipeline).Bsz(10).All()
	if err != nil {
		t.Fatalf("Agg.Bsz.All failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}
