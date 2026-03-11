package maango_test

import (
	"context"
	"testing"

	maango "github.com/Maximumsoft-Co-LTD/maan-go"
)

type dbTestDoc struct {
	ID   string `bson:"_id"`
	Name string `bson:"name"`
}

type dbTestDoc2 struct {
	ID    string `bson:"_id"`
	Value int    `bson:"value"`
}

type testDB struct {
	Docs    maango.Coll[dbTestDoc]  `collection_name:"docs"`
	Items   maango.Coll[dbTestDoc2] `collection_name:"items"`
	NoTag   maango.Coll[dbTestDoc]
	EmptyTag maango.Coll[dbTestDoc] `collection_name:""`
}

type testDBWithExColl struct {
	Docs  maango.Coll[dbTestDoc]   `collection_name:"docs"`
	Items maango.ExColl[dbTestDoc2] `collection_name:"items"`
}

func TestDB_InitializesTaggedCollFields(t *testing.T) {
	ctx := context.Background()
	client, err := maango.NewFakeClient()
	if err != nil {
		t.Fatalf("NewFakeClient: %v", err)
	}

	db := maango.DB[testDB](ctx, client)
	if db == nil {
		t.Fatal("DB returned nil")
	}

	// Tagged fields must be initialized.
	if db.Docs.Collection == nil {
		t.Error("Docs.Collection should not be nil")
	}
	if db.Items.Collection == nil {
		t.Error("Items.Collection should not be nil")
	}

	// Collection names must match the tags.
	if got := db.Docs.Collection.Name(); got != "docs" {
		t.Errorf("Docs.Name() = %q, want %q", got, "docs")
	}
	if got := db.Items.Collection.Name(); got != "items" {
		t.Errorf("Items.Name() = %q, want %q", got, "items")
	}

	// Field without tag must remain uninitialized.
	if db.NoTag.Collection != nil {
		t.Error("NoTag.Collection should remain nil (no tag)")
	}

	// Field with empty tag must remain uninitialized.
	if db.EmptyTag.Collection != nil {
		t.Error("EmptyTag.Collection should remain nil (empty tag)")
	}
}

func TestDB_InitializesExCollFields(t *testing.T) {
	ctx := context.Background()
	client, err := maango.NewFakeClient()
	if err != nil {
		t.Fatalf("NewFakeClient: %v", err)
	}

	db := maango.DB[testDBWithExColl](ctx, client)
	if db == nil {
		t.Fatal("DB returned nil")
	}

	if db.Docs.Collection == nil {
		t.Error("Docs.Collection should not be nil")
	}
	if db.Items.ExtendedCollection == nil {
		t.Error("Items.ExtendedCollection should not be nil")
	}
}

func TestDB_ReturnsNonNilForEmptyStruct(t *testing.T) {
	ctx := context.Background()
	client, err := maango.NewFakeClient()
	if err != nil {
		t.Fatalf("NewFakeClient: %v", err)
	}

	type emptyDB struct{}
	db := maango.DB[emptyDB](ctx, client)
	if db == nil {
		t.Error("DB should return non-nil pointer even for empty struct")
	}
}
