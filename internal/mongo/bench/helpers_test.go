// FILE: internal/mongo/bench/helpers_test.go
// PACKAGE: mongo_bench_test
// PURPOSE: Shared helpers for benchmark tests.

package mongo_bench_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const benchDBName = "maango_bench_test"

// integDoc is the standard benchmark document with model-default hooks.
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

// uniqueColl returns a collection name guaranteed unique per benchmark run.
func uniqueColl(base string) string {
	return fmt.Sprintf("%s_%s", base, primitive.NewObjectID().Hex())
}

// connectBenchClient creates a real mongo.Client for benchmark tests.
// Calls b.Skip if MONGO_INTEGRATION_URI is not set.
func connectBenchClient(b *testing.B) mongo.Client {
	b.Helper()
	uri := os.Getenv("MONGO_INTEGRATION_URI")
	if uri == "" {
		b.Skip("MONGO_INTEGRATION_URI not set; skipping benchmark")
	}
	client, err := mongo.NewClient(context.Background(),
		mongo.WithWriteURI(uri),
		mongo.WithDatabase(benchDBName),
	)
	if err != nil {
		b.Fatalf("failed to create client: %v", err)
	}
	b.Cleanup(func() { _ = client.Close() })
	return client
}

// connectLoadClient creates a real mongo.Client for load tests (testing.T).
func connectLoadClient(t *testing.T) mongo.Client {
	t.Helper()
	uri := os.Getenv("MONGO_INTEGRATION_URI")
	if uri == "" {
		t.Skip("MONGO_INTEGRATION_URI not set; skipping load test")
	}
	client, err := mongo.NewClient(context.Background(),
		mongo.WithWriteURI(uri),
		mongo.WithDatabase(benchDBName),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// newBenchColl creates a uniquely-named collection and registers cleanup.
func newBenchColl[T any](b *testing.B, client mongo.Client, base string) (mongo.Collection[T], string) {
	b.Helper()
	name := uniqueColl(base)
	b.Cleanup(func() {
		_ = client.Write().Database(benchDBName).Collection(name).Drop(context.Background())
	})
	coll := mongo.NewCollection[T](context.Background(), client, name)
	return coll, name
}

// newLoadColl creates a uniquely-named collection for load tests and registers cleanup.
func newLoadColl[T any](t *testing.T, client mongo.Client, base string) (mongo.Collection[T], string) {
	t.Helper()
	name := uniqueColl(base)
	t.Cleanup(func() {
		_ = client.Write().Database(benchDBName).Collection(name).Drop(context.Background())
	})
	coll := mongo.NewCollection[T](context.Background(), client, name)
	return coll, name
}

// seedDocs inserts n documents into the collection for read benchmarks.
func seedDocs(b *testing.B, coll mongo.Collection[integDoc], n int) {
	b.Helper()
	docs := make([]integDoc, n)
	for i := range docs {
		docs[i] = integDoc{Name: fmt.Sprintf("bench_%d", i), Value: i, Active: i%2 == 0}
	}
	if err := coll.CreateMany(&docs); err != nil {
		b.Fatalf("failed to seed docs: %v", err)
	}
}

// seedDocsT inserts n documents using testing.T (for load tests).
func seedDocsT(t *testing.T, coll mongo.Collection[integDoc], n int) {
	t.Helper()
	docs := make([]integDoc, n)
	for i := range docs {
		docs[i] = integDoc{Name: fmt.Sprintf("load_%d", i), Value: i, Active: i%2 == 0}
	}
	if err := coll.CreateMany(&docs); err != nil {
		t.Fatalf("failed to seed docs: %v", err)
	}
}
