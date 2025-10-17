package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// #region (aggregate) this
type agg[T any] struct {
	ctx       *context.Context
	client    Client
	coll      *mongo.Collection
	collName  string
	pipeline  any
	allowDisk *bool
	batch     *int32
	extra     *options.AggregateOptions
}

var _ Aggregate[any] = (*agg[any])(nil)

func NewAgg[T any](ctx context.Context, client Client, collName string, pipeline any) Aggregate[T] {
	return &agg[T]{
		ctx:      &ctx,
		client:   client,
		collName: collName,
		coll:     client.Read().Database(client.DbName()).Collection(collName),
		pipeline: pipeline,
	}
}

func (a *agg[T]) getCtx() context.Context {
	if a.ctx == nil {
		return context.Background()
	}
	return *a.ctx
}

func (a *agg[T]) Disk(b bool) Aggregate[T]                       { a.allowDisk = &b; return a }
func (a *agg[T]) Bsz(n int32) Aggregate[T]                       { a.batch = &n; return a }
func (a *agg[T]) Opts(ao *options.AggregateOptions) Aggregate[T] { a.extra = ao; return a }

func (a *agg[T]) All() ([]T, error) {
	var out []T
	cur, err := a.coll.Aggregate(a.getCtx(), a.pipeline, a.build())
	if err != nil {
		return nil, err
	}
	defer cur.Close(a.getCtx())
	for cur.Next(a.getCtx()) {
		var v T
		if err := cur.Decode(&v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (a *agg[T]) Raw() ([]bson.M, error) {
	var out []bson.M
	cur, err := a.coll.Aggregate(a.getCtx(), a.pipeline, a.build())
	if err != nil {
		return nil, err
	}
	defer cur.Close(a.getCtx())
	for cur.Next(a.getCtx()) {
		var v bson.M
		if err := cur.Decode(&v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (a *agg[T]) Strm(fn func(ctx context.Context, doc T) error) error {
	cur, err := a.coll.Aggregate(a.getCtx(), a.pipeline, a.build())
	if err != nil {
		return err
	}
	defer cur.Close(a.getCtx())
	for cur.Next(a.getCtx()) {
		var v T
		if err := cur.Decode(&v); err != nil {
			return err
		}
		if err := fn(a.getCtx(), v); err != nil {
			return err
		}
	}
	return cur.Err()
}

func (a *agg[T]) Each(fn func(ctx context.Context, doc T) error) error { return a.Strm(fn) }

func (a *agg[T]) strmRaw(fn func(ctx context.Context, doc bson.M) error) error {
	cur, err := a.coll.Aggregate(a.getCtx(), a.pipeline, a.build())
	if err != nil {
		return err
	}
	defer cur.Close(a.getCtx())
	for cur.Next(a.getCtx()) {
		var v bson.M
		if err := cur.Decode(&v); err != nil {
			return err
		}
		if err := fn(a.getCtx(), v); err != nil {
			return err
		}
	}
	return cur.Err()
}

func (a *agg[T]) EachRaw(fn func(ctx context.Context, doc bson.M) error) error { return a.strmRaw(fn) }

func (a *agg[T]) build() *options.AggregateOptions {
	ao := options.AggregateOptions{}
	if a.allowDisk != nil {
		ao.SetAllowDiskUse(*a.allowDisk)
	}
	if a.batch != nil {
		ao.SetBatchSize(*a.batch)
	}
	if a.extra != nil {
		if a.extra.AllowDiskUse != nil {
			ao.SetAllowDiskUse(*a.extra.AllowDiskUse)
		}
		if a.extra.BatchSize != nil {
			ao.SetBatchSize(*a.extra.BatchSize)
		}
	}
	return &ao
}

//#endregion
