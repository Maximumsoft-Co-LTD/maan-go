// FILE: internal/mongo/integration/text_regex_integration_test.go
// PACKAGE: mongo_integration_test
// PURPOSE: Behavioral correctness tests for RegexFields and TxtFind operations.

package mongo_integration_test

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	mg "go.mongodb.org/mongo-driver/mongo"
)

func TestRegexFields_CaseInsensitive(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "regex_case")

	docs := []integDoc{
		{Name: "Alpha"},
		{Name: "BETA"},
		{Name: "alpha_2"},
		{Name: "gamma"},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	results, err := coll.RegexFields("alpha", "name")
	if err != nil {
		t.Fatalf("RegexFields failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}
}

func TestRegexFields_MultipleFields(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "regex_multi")

	docs := []integDoc{
		{Name: "test_item"},
		{Name: "other", Tags: []string{"test_tag"}},
		{Name: "nothing", Tags: []string{"unrelated"}},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	results, err := coll.RegexFields("test", "name", "tags")
	if err != nil {
		t.Fatalf("RegexFields failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}
}

func TestRegexFields_NoMatch(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "regex_nomatch")

	docs := []integDoc{
		{Name: "alpha"},
		{Name: "beta"},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	results, err := coll.RegexFields("zzz_no_match_xyz", "name")
	if err != nil {
		t.Fatalf("RegexFields failed: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(results))
	}
}

func TestTxtFind_WithTextIndex(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "txtfind")

	_, err := coll.Idx().CreateOne(mg.IndexModel{
		Keys: bson.D{{Key: "name", Value: "text"}},
	})
	if err != nil {
		t.Fatalf("CreateOne text index failed: %v", err)
	}

	docs := []integDoc{
		{Name: "golang developer"},
		{Name: "python developer"},
		{Name: "rust engineer"},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	results, err := coll.TxtFind("developer")
	if err != nil {
		t.Fatalf("TxtFind failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}
}

func TestTxtFind_NoMatch(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "txtfind_nomatch")

	_, err := coll.Idx().CreateOne(mg.IndexModel{
		Keys: bson.D{{Key: "name", Value: "text"}},
	})
	if err != nil {
		t.Fatalf("CreateOne text index failed: %v", err)
	}

	docs := []integDoc{
		{Name: "golang developer"},
		{Name: "python developer"},
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	results, err := coll.TxtFind("zzz_no_match_xyz")
	if err != nil {
		t.Fatalf("TxtFind failed: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(results))
	}
}
