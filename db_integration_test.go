// FILE: db_integration_test.go (root package level)
// PACKAGE: maango_test
// PURPOSE: Integration tests for DB[T] reflection-based auto-initialization.

package maango_test

import (
	"context"
	"os"
	"testing"
	"time"

	maango "github.com/Maximumsoft-Co-LTD/maan-go"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const dbIntegTestDBName = "maango_integration_test"

// integDBDoc is a document type for DB reflection integration tests.
// Named differently from dbTestDoc in db_test.go to avoid redeclaration.
type integDBDoc struct {
	ID        primitive.ObjectID `bson:"_id"`
	Name      string        `bson:"name"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at"`
}

func (d *integDBDoc) DefaultId() primitive.ObjectID {
	if d.ID.IsZero() {
		d.ID = primitive.NewObjectID()
	}
	return d.ID
}

func (d *integDBDoc) DefaultCreatedAt() time.Time {
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now().UTC()
	}
	return d.CreatedAt
}

func (d *integDBDoc) DefaultUpdatedAt() time.Time {
	if d.UpdatedAt.IsZero() {
		d.UpdatedAt = time.Now().UTC()
	}
	return d.UpdatedAt
}

// integDBSchema demonstrates the DB[T] pattern with Coll fields.
type integDBSchema struct {
	Users    maango.Coll[integDBDoc] `collection_name:"integ_db_users"`
	Products maango.Coll[integDBDoc] `collection_name:"integ_db_products"`
}

// integDBSchemaEx demonstrates the DB[T] pattern with an ExColl field.
type integDBSchemaEx struct {
	Items maango.ExColl[integDBDoc] `collection_name:"integ_db_items"`
}

func connectDBIntegClient(t *testing.T) maango.Client {
	t.Helper()
	uri := os.Getenv("MONGO_INTEGRATION_URI")
	if uri == "" {
		t.Skip("MONGO_INTEGRATION_URI not set; skipping integration test")
	}
	client, err := maango.NewClient(context.Background(),
		maango.WithWriteURI(uri),
		maango.WithDatabase(dbIntegTestDBName),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestDB_ReflectionInitialization(t *testing.T) {
	client := connectDBIntegClient(t)

	t.Cleanup(func() {
		db := client.Write().Database(dbIntegTestDBName)
		_ = db.Collection("integ_db_users").Drop(context.Background())
		_ = db.Collection("integ_db_products").Drop(context.Background())
	})

	schema := maango.DB[integDBSchema](context.Background(), client)
	if schema == nil {
		t.Fatal("expected non-nil schema")
	}
	if schema.Users.Name() != "integ_db_users" {
		t.Fatalf("expected Users collection name 'integ_db_users', got %q", schema.Users.Name())
	}
	if schema.Products.Name() != "integ_db_products" {
		t.Fatalf("expected Products collection name 'integ_db_products', got %q", schema.Products.Name())
	}

	// Create and verify via Users
	userDoc := &integDBDoc{Name: "db_reflection_user"}
	if err := schema.Users.Create(userDoc); err != nil {
		t.Fatalf("Users.Create failed: %v", err)
	}

	var fetchedUser integDBDoc
	if err := schema.Users.FindOne(bson.M{"_id": userDoc.ID}).Result(&fetchedUser); err != nil {
		t.Fatalf("Users.FindOne failed: %v", err)
	}
	if fetchedUser.Name != "db_reflection_user" {
		t.Fatalf("expected Name 'db_reflection_user', got %q", fetchedUser.Name)
	}

	// Create and verify via Products
	prodDoc := &integDBDoc{Name: "db_reflection_product"}
	if err := schema.Products.Create(prodDoc); err != nil {
		t.Fatalf("Products.Create failed: %v", err)
	}

	var fetchedProd integDBDoc
	if err := schema.Products.FindOne(bson.M{"_id": prodDoc.ID}).Result(&fetchedProd); err != nil {
		t.Fatalf("Products.FindOne failed: %v", err)
	}
	if fetchedProd.Name != "db_reflection_product" {
		t.Fatalf("expected Name 'db_reflection_product', got %q", fetchedProd.Name)
	}
}

func TestDB_ExCollField(t *testing.T) {
	client := connectDBIntegClient(t)

	t.Cleanup(func() {
		_ = client.Write().Database(dbIntegTestDBName).Collection("integ_db_items").Drop(context.Background())
	})

	schema := maango.DB[integDBSchemaEx](context.Background(), client)
	if schema == nil {
		t.Fatal("expected non-nil schema")
	}

	// ExColl doesn't have Create, so use a Coll handle for insertion
	writeColl := maango.NewColl[integDBDoc](context.Background(), client, "integ_db_items")
	doc := &integDBDoc{Name: "excoll_item"}
	if err := writeColl.Create(doc); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Query via schema.Items (ExColl)
	var fetched integDBDoc
	err := schema.Items.By("Name", "excoll_item").First(&fetched)
	if err != nil {
		t.Fatalf("ExColl By/First failed: %v", err)
	}
	if fetched.Name != "excoll_item" {
		t.Fatalf("expected Name 'excoll_item', got %q", fetched.Name)
	}
}
