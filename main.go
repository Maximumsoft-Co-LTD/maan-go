package maango

import (
	"context"
	"time"

	mg "github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// Re-export core types so users can depend on package maango directly.
type (
	Client           = mg.Client
	Option           = mg.Option
	Session          = mg.TxSession
	FakeClientOption = mg.FakeClientOption
)
type (
	// Coll is a concrete struct wrapping Collection[T] to support DB auto-initialization via reflection.
	// All Collection[T] methods are accessible through embedding.
	Coll[T any] struct{ mg.Collection[T] }
	// ExColl is a concrete struct wrapping ExtendedCollection[T] to support DB auto-initialization via reflection.
	// All ExtendedCollection[T] methods are accessible through embedding.
	ExColl[T any] struct{ mg.ExtendedCollection[T] }
	// SingleResult models a find-one query with optional modifiers.
	SingleResult[T any] mg.SingleResult[T]
	// ManyResult models a find-many query with modifiers and streaming helpers.
	ManyResult[T any] mg.ManyResult[T]
	// Aggregate wraps aggregation pipelines with streaming helpers.
	Aggregate[T any] mg.Aggregate[T]
	// ChangeStream is a fluent builder for watching real-time change events on a collection.
	ChangeStream[T any] mg.ChangeStream[T]
	// ChangeEvent is a typed MongoDB change stream event.
	ChangeEvent[T any] mg.ChangeEvent[T]
	// CsEvt (Change Stream Event) bundles a change event and its context into a single
	// callback argument. Use st.ChangeEvent to access event data and st.Ctx() to get the context.
	CsEvt[T any] mg.CsEvt[T]
	// ChangeEventNamespace holds the database and collection name of a change event.
	ChangeEventNamespace = mg.ChangeEventNamespace
	// ChangeUpdateDesc describes the fields modified in an update operation.
	ChangeUpdateDesc = mg.ChangeUpdateDesc
	// BsonM is a map of string keys to any values.
	BsonM bson.M
	// BsonA is a slice of any values.
	BsonA bson.A
	// BsonE is a key-value pair.
	BsonE bson.E
	// BsonD is a map of string keys to any values.
	BsonD bson.D
)

// NewClient creates a MongoDB client pair with optional read/write separation.
// Requires at least WithWriteURI and WithDatabase options.
func NewClient(ctx context.Context, opts ...Option) (Client, error) {
	return mg.NewClient(ctx, opts...)
}

// NewFakeClient returns a Client backed by disconnected mongo.Client instances for unit testing.
// No real MongoDB connection is established.
func NewFakeClient(opts ...mg.FakeClientOption) (Client, error) {
	return mg.NewFakeClient(opts...)
}

// WithFakeDatabase overrides the default database name used by NewFakeClient.
func WithFakeDatabase(name string) FakeClientOption { return mg.WithFakeDatabase(name) }

// WithFakeURI overrides the MongoDB URI stored inside NewFakeClient (never actually dialed).
func WithFakeURI(uri string) FakeClientOption { return mg.WithFakeURI(uri) }

// WithWriteURI sets the URI used for write operations (required).
func WithWriteURI(uri string) Option { return mg.WithWriteURI(uri) }

// WithReadURI sets the URI used for read operations. Falls back to the write URI when omitted.
func WithReadURI(uri string) Option { return mg.WithReadURI(uri) }

// WithDatabase specifies the logical database name (required).
func WithDatabase(name string) Option { return mg.WithDatabase(name) }

// WithTimeout overrides the connection timeout (default 60 s).
func WithTimeout(d time.Duration) Option { return mg.WithTimeout(d) }

// WithReadPreference overrides the read preference for the read client.
func WithReadPreference(rp *readpref.ReadPref) Option { return mg.WithReadPreference(rp) }

// WithWriteConcern overrides the write concern for the write client.
func WithWriteConcern(wc *writeconcern.WriteConcern) Option { return mg.WithWriteConcern(wc) }

// WithClientOptions allows callers to mutate the underlying mongo client options before dialing.
func WithClientOptions(mutator func(*options.ClientOptions)) Option {
	return mg.WithClientOptions(mutator)
}

// NewColl creates a strongly typed collection wrapper for the given collection name.
func NewColl[T any](ctx context.Context, client Client, name string) Coll[T] {
	return Coll[T]{mg.NewCollection[T](ctx, client, name)}
}

// NewExColl creates an ExtendedCollection from raw *mongo.Collection handles.
// Prefer NewExCollFromClient when a Client is available.
func NewExColl[T any](ctx context.Context, read, write *mongo.Collection, name string) ExColl[T] {
	return ExColl[T]{mg.NewExtendedCollection[T](ctx, read, write, name)}
}

// NewExCollFromClient is the preferred way to create an ExtendedCollection.
// It derives the read/write collection handles from Client automatically.
func NewExCollFromClient[T any](ctx context.Context, client Client, name string) ExColl[T] {
	dbName := client.DbName()
	read := client.Read().Database(dbName).Collection(name)
	write := client.Write().Database(dbName).Collection(name)
	return ExColl[T]{mg.NewExtendedCollection[T](ctx, read, write, name)}
}

// NewSingle creates a SingleResult builder. Normally obtained via Collection.FindOne.
func NewSingle[T any](ctx context.Context, coll *mongo.Collection, collName string, query any) SingleResult[T] {
	return mg.NewSingle[T](ctx, coll, collName, query)
}

// NewMany creates a ManyResult builder. Normally obtained via Collection.Find/FindMany.
func NewMany[T any](ctx context.Context, coll *mongo.Collection, collName string, filter any) ManyResult[T] {
	return mg.NewMany[T](ctx, coll, collName, filter)
}

// NewAgg creates an Aggregate builder. Normally obtained via Collection.Agg.
func NewAgg[T any](ctx context.Context, coll *mongo.Collection, collName string, pipeline any) Aggregate[T] {
	return mg.NewAgg[T](ctx, coll, collName, pipeline)
}
