package maango

import (
	"context"
	"time"

	maango "maan-go/pkg/mongo"

	mg "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// Re-export core types so users can depend on package maango directly.
type (
	Client    = maango.Client
	Option    = maango.Option
	TxSession = maango.TxSession
)
type (
	Collection[T any]         maango.Collection[T]
	ExtendedCollection[T any] maango.ExtendedCollection[T]
	SingleResult[T any]       maango.SingleResult[T]
	ManyResult[T any]         maango.ManyResult[T]
	Aggregate[T any]          maango.Aggregate[T]
)

// NewClient proxies to pkg/mongo.NewClient.
func NewClient(ctx context.Context, opts ...Option) (Client, error) {
	return maango.NewClient(ctx, opts...)
}

// Option helpers re-exported from the maango package.
func WithWriteURI(uri string) Option { return maango.WithWriteURI(uri) }

func WithReadURI(uri string) Option { return maango.WithReadURI(uri) }

func WithDatabase(name string) Option { return maango.WithDatabase(name) }

func WithTimeout(d time.Duration) Option { return maango.WithTimeout(d) }

func WithReadPreference(rp *readpref.ReadPref) Option { return maango.WithReadPreference(rp) }

func WithWriteConcern(wc *writeconcern.WriteConcern) Option { return maango.WithWriteConcern(wc) }

func WithClientOptions(mutator func(*options.ClientOptions)) Option {
	return maango.WithClientOptions(mutator)
}

// Collection constructors and helpers.
func NewCollection[T any](ctx context.Context, client Client, name string) Collection[T] {
	return maango.NewCollection[T](ctx, client, name)
}

func NewExtendedCollection[T any](ctx context.Context, client Client, read, write *mg.Collection, name string) ExtendedCollection[T] {
	return maango.NewExtendedCollection[T](ctx, client, read, write, name)
}

func NewSingle[T any](ctx context.Context, client Client, collName string, query any) SingleResult[T] {
	return maango.NewSingle[T](ctx, client, collName, query)
}

func NewMany[T any](ctx context.Context, client Client, collName string, filter any) ManyResult[T] {
	return maango.NewMany[T](ctx, client, collName, filter)
}

func NewAgg[T any](ctx context.Context, client Client, collName string, pipeline any) Aggregate[T] {
	return maango.NewAgg[T](ctx, client, collName, pipeline)
}
