package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// #region (aggregate) this
type agg[T any] struct {
	ctx       context.Context
	coll      *mongo.Collection
	collName  string
	pipeline  any
	allowDisk *bool
	batch     *int32
	extra     *options.AggregateOptions
}

var _ Aggregate[any] = (*agg[any])(nil)

func NewAgg[T any](ctx context.Context, coll *mongo.Collection, collName string, pipeline any) Aggregate[T] {
	return &agg[T]{
		ctx:      normalizeCtx(ctx),
		collName: collName,
		coll:     coll,
		pipeline: pipeline,
	}
}

func (a *agg[T]) getCtx() context.Context {
	return normalizeCtx(a.ctx)
}

func (a *agg[T]) Disk(b bool) Aggregate[T] {
	next := *a
	next.allowDisk = boolPtr(b)
	return &next
}
func (a *agg[T]) Bsz(n int32) Aggregate[T] {
	next := *a
	next.batch = int32Ptr(n)
	return &next
}
func (a *agg[T]) Opts(ao *options.AggregateOptions) Aggregate[T] {
	next := *a
	next.extra = ao
	return &next
}

func (a *agg[T]) Result(out *[]T) error {
	if out == nil {
		return nil
	}
	items, err := a.All()
	if err != nil {
		return err
	}
	*out = items
	return nil
}

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

func (a *agg[T]) Stream(fn func(ctx context.Context, doc T) error) error {
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

func (a *agg[T]) Each(fn func(ctx context.Context, doc T) error) error { return a.Stream(fn) }

func (a *agg[T]) streamRaw(fn func(ctx context.Context, doc bson.M) error) error {
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

func (a *agg[T]) EachRaw(fn func(ctx context.Context, doc bson.M) error) error {
	return a.streamRaw(fn)
}

func (a *agg[T]) build() *options.AggregateOptions {
	ao := options.Aggregate()
	if a.extra != nil {
		*ao = *a.extra
	}
	if a.allowDisk != nil {
		ao.SetAllowDiskUse(*a.allowDisk)
	}
	if a.batch != nil {
		ao.SetBatchSize(*a.batch)
	}
	return ao
}

//#endregion

func boolPtr(v bool) *bool {
	val := v
	return &val
}

func int32Ptr(v int32) *int32 {
	val := v
	return &val
}
