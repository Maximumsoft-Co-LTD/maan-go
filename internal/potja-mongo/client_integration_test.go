package mongo_test

import (
	"context"
	"os"
	"testing"
	"time"

	"potja-mongo/pkg/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type integrationDoc struct {
	ID        primitive.ObjectID `bson:"_id"`
	Name      string             `bson:"name"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

func (d *integrationDoc) DefaultId() primitive.ObjectID {
	if d.ID.IsZero() {
		d.ID = primitive.NewObjectID()
	}
	return d.ID
}

func (d *integrationDoc) DefaultCreatedAt() time.Time {
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now().UTC()
	}
	return d.CreatedAt
}

func (d *integrationDoc) DefaultUpdatedAt() time.Time {
	if d.UpdatedAt.IsZero() {
		d.UpdatedAt = time.Now().UTC()
	}
	return d.UpdatedAt
}

func TestClientRoundTrip(t *testing.T) {
	uri := os.Getenv("MONGO_INTEGRATION_URI")
	if uri == "" {
		t.Skip("MONGO_INTEGRATION_URI not set; skipping integration test")
	}

	ctx := context.Background()
	client, err := mongo.NewClient(
		ctx,
		mongo.WithWriteURI(uri),
		mongo.WithDatabase("potja_mongo_integration"),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})

	const collName = "client_round_trip"
	coll := mongo.NewCollection[integrationDoc](ctx, client, collName)

	doc := &integrationDoc{Name: "roundtrip"}
	if err := coll.Create(doc); err != nil {
		t.Fatalf("failed to insert document: %v", err)
	}
	t.Cleanup(func() {
		_ = coll.Del(bson.M{"_id": doc.ID})
	})

	var fetched integrationDoc
	if err := coll.FindOne(bson.M{"_id": doc.ID}).Res(&fetched); err != nil {
		t.Fatalf("failed to fetch document: %v", err)
	}

	if fetched.ID != doc.ID {
		t.Fatalf("expected ID %s but got %s", doc.ID.Hex(), fetched.ID.Hex())
	}
	if fetched.Name != doc.Name {
		t.Fatalf("expected name %q but got %q", doc.Name, fetched.Name)
	}
	if fetched.CreatedAt.IsZero() || fetched.UpdatedAt.IsZero() {
		t.Fatalf("timestamps were not set: %+v", fetched)
	}
}
