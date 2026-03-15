// FILE: internal/mongo/integration/crud_integration_test.go
// PACKAGE: mongo_integration_test
// PURPOSE: Behavioral correctness tests for Create, FindOne, FindMany, and related read operations.

package mongo_integration_test

import (
	"context"
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mg "go.mongodb.org/mongo-driver/mongo"
)

func TestCreate_SingleDoc(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "crud_create")

	doc := integDoc{Name: "alpha", Value: 10, Active: true}
	if err := coll.Create(&doc); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var fetched integDoc
	if err := coll.FindOne(bson.M{"_id": doc.ID}).Result(&fetched); err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}

	if fetched.ID != doc.ID {
		t.Fatalf("expected ID %v, got %v", doc.ID, fetched.ID)
	}
	if fetched.Name != "alpha" {
		t.Fatalf("expected Name 'alpha', got %q", fetched.Name)
	}
	if fetched.Value != 10 {
		t.Fatalf("expected Value 10, got %d", fetched.Value)
	}
	if fetched.Active != true {
		t.Fatal("expected Active true")
	}
	if fetched.ID.IsZero() {
		t.Fatal("expected non-zero ID")
	}
	if fetched.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}
	if fetched.UpdatedAt.IsZero() {
		t.Fatal("expected non-zero UpdatedAt")
	}
}

func TestCreate_NilDoc(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "crud_create_nil")

	err := coll.Create(nil)
	if err == nil {
		t.Fatal("expected error for nil doc, got nil")
	}
	if err.Error() != "doc must not be nil" {
		t.Fatalf("expected 'doc must not be nil', got %q", err.Error())
	}
}

func TestCreateMany_MultipleDocs(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "crud_create_many")

	docs := []integDoc{
		{Name: "one", Value: 1},
		{Name: "two", Value: 2},
		{Name: "three", Value: 3},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	all, err := coll.FindMany(bson.M{}).All()
	if err != nil {
		t.Fatalf("FindMany failed: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 docs, got %d", len(all))
	}
	for i, d := range all {
		if d.ID.IsZero() {
			t.Fatalf("doc[%d] has zero ID", i)
		}
		if d.CreatedAt.IsZero() {
			t.Fatalf("doc[%d] has zero CreatedAt", i)
		}
		if d.UpdatedAt.IsZero() {
			t.Fatalf("doc[%d] has zero UpdatedAt", i)
		}
	}
}

func TestCreateMany_NilSlice(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "crud_create_many_nil")

	err := coll.CreateMany(nil)
	if err == nil {
		t.Fatal("expected error for nil docs, got nil")
	}
	if err.Error() != "docs must not be nil" {
		t.Fatalf("expected 'docs must not be nil', got %q", err.Error())
	}
}

func TestCreateMany_EmptySlice(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "crud_create_many_empty")

	err := coll.CreateMany(&[]integDoc{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}

	count, err := coll.Count(nil)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected count 0, got %d", count)
	}
}

func TestFindOne_ExistingDoc(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "crud_find_one")

	doc := integDoc{Name: "find_me", Value: 42}
	if err := coll.Create(&doc); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var fetched integDoc
	if err := coll.FindOne(bson.M{"_id": doc.ID}).Result(&fetched); err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}
	if fetched.ID != doc.ID {
		t.Fatalf("expected ID %v, got %v", doc.ID, fetched.ID)
	}
	if fetched.Name != doc.Name {
		t.Fatalf("expected Name %q, got %q", doc.Name, fetched.Name)
	}
}

func TestFindOne_NoMatch(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "crud_find_one_nomatch")

	var doc integDoc
	err := coll.FindOne(bson.M{"_id": primitive.NewObjectID()}).Result(&doc)
	if err != mg.ErrNoDocuments {
		t.Fatalf("expected ErrNoDocuments, got %v", err)
	}
}

func TestFindOne_WithProjection(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "crud_find_one_proj")

	doc := integDoc{Name: "proj_test", Value: 99}
	if err := coll.Create(&doc); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var fetched integDoc
	if err := coll.FindOne(bson.M{"_id": doc.ID}).Proj(bson.M{"name": 1, "_id": 1}).Result(&fetched); err != nil {
		t.Fatalf("FindOne with projection failed: %v", err)
	}
	if fetched.Name != "proj_test" {
		t.Fatalf("expected Name 'proj_test', got %q", fetched.Name)
	}
	if fetched.Value != 0 {
		t.Fatalf("expected Value 0 (excluded by projection), got %d", fetched.Value)
	}
}

func TestFindOne_WithSort(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "crud_find_one_sort")

	docs := []integDoc{
		{Name: "a", Value: 10},
		{Name: "b", Value: 30},
		{Name: "c", Value: 20},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	var fetched integDoc
	if err := coll.FindOne(bson.M{}).Sort(bson.M{"value": -1}).Result(&fetched); err != nil {
		t.Fatalf("FindOne with sort failed: %v", err)
	}
	if fetched.Value != 30 {
		t.Fatalf("expected Value 30 (highest), got %d", fetched.Value)
	}
}

func TestFindMany_All(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "crud_find_many_all")

	docs := make([]integDoc, 5)
	for i := range docs {
		docs[i] = integDoc{Name: fmt.Sprintf("doc_%d", i), Value: i}
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	all, err := coll.FindMany(bson.M{}).All()
	if err != nil {
		t.Fatalf("FindMany.All failed: %v", err)
	}
	if len(all) != 5 {
		t.Fatalf("expected 5 docs, got %d", len(all))
	}
}

func TestFindMany_WithFilter(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "crud_find_many_filter")

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

	active, err := coll.FindMany(bson.M{"active": true}).All()
	if err != nil {
		t.Fatalf("FindMany with filter failed: %v", err)
	}
	if len(active) != 3 {
		t.Fatalf("expected 3 active docs, got %d", len(active))
	}
}

func TestFindMany_WithLimitSkipSort(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "crud_find_many_lss")

	docs := make([]integDoc, 10)
	for i := range docs {
		docs[i] = integDoc{Name: fmt.Sprintf("doc_%d", i), Value: i}
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	result, err := coll.Find(nil).Sort(bson.M{"value": 1}).Limit(3).Skip(2).All()
	if err != nil {
		t.Fatalf("Find with limit/skip/sort failed: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 docs, got %d", len(result))
	}
	expected := []int{2, 3, 4}
	for i, doc := range result {
		if doc.Value != expected[i] {
			t.Fatalf("result[%d].Value: expected %d, got %d", i, expected[i], doc.Value)
		}
	}
}

func TestFindMany_Stream(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "crud_find_many_stream")

	docs := make([]integDoc, 5)
	for i := range docs {
		docs[i] = integDoc{Name: fmt.Sprintf("doc_%d", i), Value: i}
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	var collected []integDoc
	err := coll.Find(nil).Stream(func(ctx context.Context, doc integDoc) error {
		collected = append(collected, doc)
		return nil
	})
	if err != nil {
		t.Fatalf("Stream failed: %v", err)
	}
	if len(collected) != 5 {
		t.Fatalf("expected 5 streamed docs, got %d", len(collected))
	}
}

func TestFindMany_Each(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "crud_find_many_each")

	docs := make([]integDoc, 5)
	for i := range docs {
		docs[i] = integDoc{Name: fmt.Sprintf("doc_%d", i), Value: i}
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	counter := 0
	err := coll.Find(nil).Each(func(ctx context.Context, doc integDoc) error {
		counter++
		return nil
	})
	if err != nil {
		t.Fatalf("Each failed: %v", err)
	}
	if counter != 5 {
		t.Fatalf("expected 5 calls, got %d", counter)
	}
}

func TestFindMany_Result(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "crud_find_many_result")

	docs := []integDoc{
		{Name: "a", Value: 1},
		{Name: "b", Value: 2},
		{Name: "c", Value: 3},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	var out []integDoc
	if err := coll.Find(nil).Result(&out); err != nil {
		t.Fatalf("Result failed: %v", err)
	}
	if len(out) != 3 {
		t.Fatalf("expected 3 docs, got %d", len(out))
	}
}

func TestFindMany_Cnt(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "crud_find_many_cnt")

	docs := make([]integDoc, 5)
	for i := range docs {
		docs[i] = integDoc{Name: fmt.Sprintf("doc_%d", i), Value: i}
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	count, err := coll.Find(nil).Cnt()
	if err != nil {
		t.Fatalf("Cnt failed: %v", err)
	}
	if count != 5 {
		t.Fatalf("expected count 5, got %d", count)
	}
}

func TestFind_AliasForFindMany(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "crud_find_alias")

	docs := []integDoc{
		{Name: "x", Value: 1, Active: true},
		{Name: "y", Value: 2, Active: true},
		{Name: "z", Value: 3, Active: true},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	a, err := coll.Find(bson.M{"active": true}).All()
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	b, err := coll.FindMany(bson.M{"active": true}).All()
	if err != nil {
		t.Fatalf("FindMany failed: %v", err)
	}
	if len(a) != len(b) {
		t.Fatalf("Find and FindMany returned different lengths: %d vs %d", len(a), len(b))
	}
}
