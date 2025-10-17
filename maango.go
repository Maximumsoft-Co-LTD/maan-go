package maango

import (
	"context"
	"time"

	maangoMongo "github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo"

	mg "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// Re-export core types so users can depend on package maango directly.
type (
	Client    = maangoMongo.Client
	Option    = maangoMongo.Option
	TxSession = maangoMongo.TxSession
)
type (
	Collection[T any]         maangoMongo.Collection[T]
	ExtendedCollection[T any] maangoMongo.ExtendedCollection[T]
	SingleResult[T any]       maangoMongo.SingleResult[T]
	ManyResult[T any]         maangoMongo.ManyResult[T]
	Aggregate[T any]          maangoMongo.Aggregate[T]
)

// NewClient proxies to pkg/mongo.NewClient.
func NewClient(ctx context.Context, opts ...Option) (Client, error) {
	return maangoMongo.NewClient(ctx, opts...)
}

// Option helpers re-exported from the maangoMongo package.
func WithWriteURI(uri string) Option { return maangoMongo.WithWriteURI(uri) }

func WithReadURI(uri string) Option { return maangoMongo.WithReadURI(uri) }

func WithDatabase(name string) Option { return maangoMongo.WithDatabase(name) }

func WithTimeout(d time.Duration) Option { return maangoMongo.WithTimeout(d) }

func WithReadPreference(rp *readpref.ReadPref) Option { return maangoMongo.WithReadPreference(rp) }

func WithWriteConcern(wc *writeconcern.WriteConcern) Option { return maangoMongo.WithWriteConcern(wc) }

func WithClientOptions(mutator func(*options.ClientOptions)) Option {
	return maangoMongo.WithClientOptions(mutator)
}

// Collection constructors and helpers.
func NewCollection[T any](ctx context.Context, client Client, name string) Collection[T] {
	return maangoMongo.NewCollection[T](ctx, client, name)
}

func NewExtendedCollection[T any](ctx context.Context, client Client, read, write *mg.Collection, name string) ExtendedCollection[T] {
	return maangoMongo.NewExtendedCollection[T](ctx, client, read, write, name)
}

func NewSingle[T any](ctx context.Context, client Client, collName string, query any) SingleResult[T] {
	return maangoMongo.NewSingle[T](ctx, client, collName, query)
}

func NewMany[T any](ctx context.Context, client Client, collName string, filter any) ManyResult[T] {
	return maangoMongo.NewMany[T](ctx, client, collName, filter)
}

func NewAgg[T any](ctx context.Context, client Client, collName string, pipeline any) Aggregate[T] {
	return maangoMongo.NewAgg[T](ctx, client, collName, pipeline)
}
