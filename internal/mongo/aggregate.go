package mongo

import (
	"context"
	"errors"

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

// NewAgg creates an Aggregate builder for pipeline on coll. Normally called via Collection.Agg.
func NewAgg[T any](ctx context.Context, coll *mongo.Collection, collName string, pipeline any) Aggregate[T] {
	return &agg[T]{
		ctx:      normalizeCtx(ctx),
		collName: collName,
		coll:     coll,
		pipeline: pipeline,
	}
}

// getCtx returns the aggregate's context, falling back to context.Background() when unset.
func (a *agg[T]) getCtx() context.Context {
	return normalizeCtx(a.ctx)
}

// Disk enables (true) or disables (false) writing temporary aggregation data to disk.
func (a *agg[T]) Disk(b bool) Aggregate[T] {
	next := *a
	next.allowDisk = boolPtr(b)
	return &next
}

// Bsz sets the cursor batch size for the aggregation result.
func (a *agg[T]) Bsz(n int32) Aggregate[T] {
	next := *a
	next.batch = int32Ptr(n)
	return &next
}

// Opts merges raw AggregateOptions on top of the builder settings (Disk, Bsz).
func (a *agg[T]) Opts(ao *options.AggregateOptions) Aggregate[T] {
	next := *a
	next.extra = ao
	return &next
}

// Result executes the pipeline and decodes all results into out.
// Returns an error if out is nil.
func (a *agg[T]) Result(out *[]T) error {
	if out == nil {
		return errors.New("out must not be nil")
	}
	items, err := a.All()
	if err != nil {
		return err
	}
	*out = items
	return nil
}

// openCursor runs the aggregation pipeline and returns the raw cursor.
func (a *agg[T]) openCursor() (*mongo.Cursor, error) {
	return a.coll.Aggregate(a.getCtx(), a.pipeline, a.build()...)
}

// All executes the pipeline and returns all typed results.
func (a *agg[T]) All() ([]T, error) {
	var out []T
	cur, err := a.openCursor()
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

// Raw executes the pipeline and returns raw bson.M documents.
func (a *agg[T]) Raw() ([]bson.M, error) {
	var out []bson.M
	cur, err := a.openCursor()
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

// Stream executes the pipeline and calls fn for each typed document.
// Stops on the first non-nil error returned by fn.
func (a *agg[T]) Stream(fn func(ctx context.Context, doc T) error) error {
	if fn == nil {
		return errors.New("fn must not be nil")
	}
	cur, err := a.openCursor()
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

// Each is an alias for Stream.
func (a *agg[T]) Each(fn func(ctx context.Context, doc T) error) error { return a.Stream(fn) }

// streamRaw executes the pipeline and calls fn for each raw bson.M document.
func (a *agg[T]) streamRaw(fn func(ctx context.Context, doc bson.M) error) error {
	if fn == nil {
		return errors.New("fn must not be nil")
	}
	cur, err := a.openCursor()
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

// EachRaw executes the pipeline and calls fn for each raw bson.M document. Alias for streamRaw.
func (a *agg[T]) EachRaw(fn func(ctx context.Context, doc bson.M) error) error {
	return a.streamRaw(fn)
}

// build assembles the final AggregateOptions from builder state.
// Builder fields are applied first; extra (Opts) is appended as a second Lister,
// so it acts as a final override without wiping builder settings.
func (a *agg[T]) build() []*options.AggregateOptions {
	opts := options.Aggregate()
	if a.allowDisk != nil {
		opts.SetAllowDiskUse(*a.allowDisk)
	}
	if a.batch != nil {
		opts.SetBatchSize(*a.batch)
	}
	if a.extra != nil {
		return []*options.AggregateOptions{opts, a.extra}
	}
	return []*options.AggregateOptions{opts}
}

//#endregion

// boolPtr returns a pointer to v.
func boolPtr(v bool) *bool {
	val := v
	return &val
}

// int32Ptr returns a pointer to v.
func int32Ptr(v int32) *int32 {
	val := v
	return &val
}
