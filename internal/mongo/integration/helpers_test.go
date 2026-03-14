// FILE: internal/mongo/integration/helpers_test.go
// PACKAGE: mongo_integration_test
// PURPOSE: Shared helpers, test document type, and client setup for all integration tests.

package mongo_integration_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const testDBName = "maango_integration_test"

// integDoc is the standard integration test document with model-default hooks.
type integDoc struct {
	ID        primitive.ObjectID `bson:"_id"`
	Name      string        `bson:"name"`
	Value     int           `bson:"value"`
	Active    bool          `bson:"active"`
	Tags      []string      `bson:"tags,omitempty"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at"`
}

func (d *integDoc) DefaultId() primitive.ObjectID {
	if d.ID.IsZero() {
		d.ID = primitive.NewObjectID()
	}
	return d.ID
}

func (d *integDoc) DefaultCreatedAt() time.Time {
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now().UTC()
	}
	return d.CreatedAt
}

func (d *integDoc) DefaultUpdatedAt() time.Time {
	if d.UpdatedAt.IsZero() {
		d.UpdatedAt = time.Now().UTC()
	}
	return d.UpdatedAt
}

// uniqueColl returns a collection name guaranteed unique per test run.
func uniqueColl(base string) string {
	return fmt.Sprintf("%s_%s", base, primitive.NewObjectID().Hex())
}

// connectTestClient creates a real mongo.Client from MONGO_INTEGRATION_URI.
// Calls t.Skip if the env var is missing. Registers t.Cleanup to close the client.
func connectTestClient(t *testing.T) mongo.Client {
	t.Helper()
	uri := os.Getenv("MONGO_INTEGRATION_URI")
	if uri == "" {
		t.Skip("MONGO_INTEGRATION_URI not set; skipping integration test")
	}
	client, err := mongo.NewClient(context.Background(),
		mongo.WithWriteURI(uri),
		mongo.WithDatabase(testDBName),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// dropColl drops a collection after test completes via t.Cleanup.
func dropColl(t *testing.T, client mongo.Client, name string) {
	t.Helper()
	t.Cleanup(func() {
		_ = client.Write().Database(testDBName).Collection(name).Drop(context.Background())
	})
}

// newColl creates a uniquely-named collection and registers cleanup.
// Returns the collection and its generated name.
func newColl[T any](t *testing.T, client mongo.Client, base string) (mongo.Collection[T], string) {
	t.Helper()
	name := uniqueColl(base)
	dropColl(t, client, name)
	coll := mongo.NewCollection[T](context.Background(), client, name)
	return coll, name
}
