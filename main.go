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
	Client  = mg.Client
	Option  = mg.Option
	Session = mg.TxSession
)
type (
	// Collection is the fluent entry point for CRUD and aggregation operations on strongly typed documents.
	Coll[T any] mg.Collection[T]
	// ExtendedCollection supports building reusable dynamic queries that can be chained.
	ExColl[T any] mg.ExtendedCollection[T]
	// SingleResult models a find-one query with optional modifiers.
	SingleResult[T any] mg.SingleResult[T]
	// ManyResult models a find-many query with modifiers and streaming helpers.
	ManyResult[T any] mg.ManyResult[T]
	// Aggregate wraps aggregation pipelines with streaming helpers.
	Aggregate[T any] mg.Aggregate[T]
	// BsonM is a map of string keys to any values.
	BsonM bson.M
	// BsonA is a slice of any values.
	BsonA bson.A
	// BsonE is a key-value pair.
	BsonE bson.E
	// BsonD is a map of string keys to any values.
	BsonD bson.D
)

// NewClient proxies to pkg/mongo.NewClient.
func NewClient(ctx context.Context, opts ...Option) (Client, error) {
	return mg.NewClient(ctx, opts...)
}

// Option helpers re-exported from the mg package.
// WithWriteURI sets the write URI for the client.
func WithWriteURI(uri string) Option { return mg.WithWriteURI(uri) }

// WithReadURI sets the read URI for the client.
func WithReadURI(uri string) Option { return mg.WithReadURI(uri) }

// WithDatabase sets the database name for the client.
func WithDatabase(name string) Option { return mg.WithDatabase(name) }

// WithTimeout sets the timeout for the client.
func WithTimeout(d time.Duration) Option { return mg.WithTimeout(d) }

// WithReadPreference sets the read preference for the client.
func WithReadPreference(rp *readpref.ReadPref) Option { return mg.WithReadPreference(rp) }

// WithWriteConcern sets the write concern for the client.
func WithWriteConcern(wc *writeconcern.WriteConcern) Option { return mg.WithWriteConcern(wc) }

// WithClientOptions sets the client options for the client.
func WithClientOptions(mutator func(*options.ClientOptions)) Option {
	return mg.WithClientOptions(mutator)
}

// Collection constructors and helpers.
func NewColl[T any](ctx context.Context, client Client, name string) Coll[T] {
	return mg.NewCollection[T](ctx, client, name)
}

func NewExColl[T any](ctx context.Context, read, write *mongo.Collection, name string) ExColl[T] {
	return mg.NewExtendedCollection[T](ctx, read, write, name)
}

func NewSingle[T any](ctx context.Context, coll *mongo.Collection, collName string, query any) SingleResult[T] {
	return mg.NewSingle[T](ctx, coll, collName, query)
}

func NewMany[T any](ctx context.Context, coll *mongo.Collection, collName string, filter any) ManyResult[T] {
	return mg.NewMany[T](ctx, coll, collName, filter)
}

func NewAgg[T any](ctx context.Context, coll *mongo.Collection, collName string, pipeline any) Aggregate[T] {
	return mg.NewAgg[T](ctx, coll, collName, pipeline)
}
