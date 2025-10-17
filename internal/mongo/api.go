package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client represents a MongoDB client that knows how to provide read/write connections
// and the database name it is scoped to.
type Client interface {
	Write() *mongo.Client
	Read() *mongo.Client
	DbName() string
	Close() error
}

// Collection is the fluent entry point for CRUD and aggregation operations on strongly typed documents.
type Collection[T any] interface {
	Agg(pipeline any) Aggregate[T]
	Build(ctx context.Context) ExtendedCollection[T]
	Create(doc *T) error
	CreateMany(docs *[]T) error
	Ctx(ctx context.Context) Collection[T]
	Del(filter any) error
	FindOne(query any) SingleResult[T]
	FindMany(filter any) ManyResult[T]
	Name() string
	ReFind(q string, fields ...string) ([]T, error)
	Save(filter any, update any) error
	SaveMany(filter any, update any) error
	TxtFind(q string) ([]T, error)
	WithTx(fn func(ctx context.Context) error) error
	StartTx() (TxSession, error)
}

// ExtendedCollection supports building reusable dynamic queries that can be chained.
type ExtendedCollection[T any] interface {
	By(string, any) ExtendedCollection[T]
	Where(bson.M) ExtendedCollection[T]
	First(*T) error
	Many(*[]T) error
	Save(any) error
	SaveMany(any) error
	Delete() error
	Count() (int64, error)
	Exists() (bool, error)
	GetFilter() any
}

// SingleResult models a find-one query with optional modifiers.
type SingleResult[T any] interface {
	Proj(p any) SingleResult[T]
	Sort(s any) SingleResult[T]
	Hint(h any) SingleResult[T]
	Opts(fo *options.FindOneOptions) SingleResult[T]
	Res(out *T) error
}

// ManyResult models a find-many query with modifiers and streaming helpers.
type ManyResult[T any] interface {
	Proj(p any) ManyResult[T]
	Sort(s any) ManyResult[T]
	Hint(h any) ManyResult[T]
	Lim(n int64) ManyResult[T]
	Skp(n int64) ManyResult[T]
	Bsz(n int32) ManyResult[T]
	Opts(fo *options.FindOptions) ManyResult[T]
	All() ([]T, error)
	Res(out *[]T) error
	Strm(fn func(ctx context.Context, doc T) error) error
	Each(fn func(ctx context.Context, doc T) error) error
	Cnt() (int64, error)
}

// Aggregate wraps aggregation pipelines with streaming helpers.
type Aggregate[T any] interface {
	Disk(b bool) Aggregate[T]
	Bsz(n int32) Aggregate[T]
	Opts(ao *options.AggregateOptions) Aggregate[T]
	All() ([]T, error)
	Raw() ([]bson.M, error)
	Strm(fn func(ctx context.Context, doc T) error) error
	Each(fn func(ctx context.Context, doc T) error) error
}

// TxSession exposes a MongoDB session used for manual transaction control.
type TxSession interface {
	SessionCtx() context.Context
	Commit() error
	Rollback()
}
