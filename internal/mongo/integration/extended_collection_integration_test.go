// FILE: internal/mongo/integration/extended_collection_integration_test.go
// PACKAGE: mongo_integration_test
// PURPOSE: Behavioral correctness tests for ExtendedCollection (Build, By, Where, First, Many, Count, Exists, Save, Del, DelMany, GetFilter).

package mongo_integration_test

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestExtColl_ByAndFirst(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "ext_by_first")

	docs := []integDoc{
		{Name: "alice", Value: 1},
		{Name: "bob", Value: 2},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	var doc integDoc
	err := coll.Build(context.Background()).By("Name", "alice").First(&doc)
	if err != nil {
		t.Fatalf("By/First failed: %v", err)
	}
	if doc.Name != "alice" {
		t.Fatalf("expected Name 'alice', got %q", doc.Name)
	}
	if doc.Value != 1 {
		t.Fatalf("expected Value 1, got %d", doc.Value)
	}
}

func TestExtColl_WhereAndMany(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "ext_where_many")

	allDocs := []integDoc{
		{Name: "a", Active: true},
		{Name: "b", Active: true},
		{Name: "c", Active: true},
		{Name: "d", Active: false},
		{Name: "e", Active: false},
	}
	if err := coll.CreateMany(&allDocs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	var docs []integDoc
	err := coll.Build(context.Background()).Where(bson.M{"active": true}).Many(&docs)
	if err != nil {
		t.Fatalf("Where/Many failed: %v", err)
	}
	if len(docs) != 3 {
		t.Fatalf("expected 3 docs, got %d", len(docs))
	}
}

func TestExtColl_ChainedByWhere(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "ext_chain")

	allDocs := []integDoc{
		{Name: "x", Active: true, Value: 5},
		{Name: "x", Active: true, Value: 15},
		{Name: "x", Active: false, Value: 20},
	}
	if err := coll.CreateMany(&allDocs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	var docs []integDoc
	err := coll.Build(context.Background()).
		By("Active", true).
		Where(bson.M{"value": bson.M{"$gt": 10}}).
		Many(&docs)
	if err != nil {
		t.Fatalf("Chained By/Where/Many failed: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 doc (Active=true AND value>10), got %d", len(docs))
	}
	if docs[0].Value != 15 {
		t.Fatalf("expected Value 15, got %d", docs[0].Value)
	}
}

func TestExtColl_Count(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "ext_count")

	docs := []integDoc{
		{Name: "a"}, {Name: "b"}, {Name: "c"}, {Name: "d"}, {Name: "e"},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	count, err := coll.Build(context.Background()).Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 5 {
		t.Fatalf("expected count 5, got %d", count)
	}
}

func TestExtColl_Exists_True(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "ext_exists_true")

	doc := integDoc{Name: "exists_test"}
	if err := coll.Create(&doc); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	exists, err := coll.Build(context.Background()).By("Name", "exists_test").Exists()
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Fatal("expected Exists to return true")
	}
}

func TestExtColl_Exists_False(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "ext_exists_false")

	exists, err := coll.Build(context.Background()).By("Name", "nonexistent").Exists()
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Fatal("expected Exists to return false")
	}
}

func TestExtColl_Save(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "ext_save")

	doc := integDoc{Name: "save_ext", Value: 10}
	if err := coll.Create(&doc); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err := coll.Build(context.Background()).By("Name", "save_ext").Save(bson.M{"$set": bson.M{"value": 77}})
	if err != nil {
		t.Fatalf("ExtColl Save failed: %v", err)
	}

	var fetched integDoc
	if err := coll.FindOne(bson.M{"name": "save_ext"}).Result(&fetched); err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}
	if fetched.Value != 77 {
		t.Fatalf("expected Value 77, got %d", fetched.Value)
	}
}

func TestExtColl_Del(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "ext_del")

	doc := integDoc{Name: "del_ext"}
	if err := coll.Create(&doc); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err := coll.Build(context.Background()).By("Name", "del_ext").Del()
	if err != nil {
		t.Fatalf("ExtColl Del failed: %v", err)
	}

	count, err := coll.Count(nil)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected count 0, got %d", count)
	}
}

func TestExtColl_DelMany(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "ext_delmany")

	docs := []integDoc{
		{Name: "a", Active: true},
		{Name: "b", Active: true},
		{Name: "c", Active: false},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	err := coll.Build(context.Background()).Where(bson.M{"active": true}).DelMany()
	if err != nil {
		t.Fatalf("ExtColl DelMany failed: %v", err)
	}

	count, err := coll.Count(nil)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}
}

func TestExtColl_GetFilter(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "ext_getfilter")

	ext := coll.Build(context.Background()).By("Name", "x").Where(bson.M{"active": true})
	filter := ext.GetFilter()

	bsonFilter, ok := filter.(bson.M)
	if !ok {
		t.Fatalf("expected bson.M, got %T", filter)
	}
	if bsonFilter["name"] != "x" {
		t.Fatalf("expected name='x', got %v", bsonFilter["name"])
	}
	if bsonFilter["active"] != true {
		t.Fatalf("expected active=true, got %v", bsonFilter["active"])
	}
}
