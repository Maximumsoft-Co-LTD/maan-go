package mongo

import (
	"context"
	"errors"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// #region (queryCore)

// queryCore holds the shared state for single/many query builders.
// It is unexported and embedded by single[T] and many[T].
type queryCore struct {
	ctx      context.Context
	collName string
	coll     *mongo.Collection
	filter   any
	proj     any
	sort     any
	hint     any
}

// newQueryCore initializes the shared fields for a single/many query builder.
func newQueryCore(ctx context.Context, coll *mongo.Collection, collName string, filter any) queryCore {
	if filter == nil {
		filter = bson.M{}
	}
	return queryCore{
		ctx:      ctx,
		collName: collName,
		coll:     coll,
		filter:   filter,
	}
}

//#endregion

// #region (single)

type single[T any] struct {
	queryCore
	extra *options.FindOneOptions
}

var _ SingleResult[any] = (*single[any])(nil)

// NewSingle creates a SingleResult builder for query on coll. Normally called via Collection.FindOne.
func NewSingle[T any](ctx context.Context, coll *mongo.Collection, collName string, query any) SingleResult[T] {
	return &single[T]{
		queryCore: newQueryCore(ctx, coll, collName, query),
	}
}

// Proj sets the projection document and returns a new builder.
func (s *single[T]) Proj(p any) SingleResult[T] { next := *s; next.proj = p; return &next }

// Sort sets the sort document and returns a new builder.
func (s *single[T]) Sort(v any) SingleResult[T] { next := *s; next.sort = v; return &next }

// Hint sets the index hint and returns a new builder.
func (s *single[T]) Hint(v any) SingleResult[T] { next := *s; next.hint = v; return &next }

// Opts merges raw FindOneOptions on top of the builder settings.
func (s *single[T]) Opts(o *options.FindOneOptions) SingleResult[T] {
	next := *s; next.extra = o; return &next
}

// buildOpts assembles the final FindOneOptions from builder state, applying extra as override.
func (s *single[T]) buildOpts() []*options.FindOneOptions {
	fo := options.FindOne()
	if s.proj != nil {
		fo.SetProjection(s.proj)
	}
	if s.sort != nil {
		fo.SetSort(s.sort)
	}
	if s.hint != nil {
		fo.SetHint(s.hint)
	}
	if s.extra != nil {
		return []*options.FindOneOptions{fo, s.extra}
	}
	return []*options.FindOneOptions{fo}
}
// Result executes the find-one query and decodes the result into out.
// Returns mongo.ErrNoDocuments when no document matches the filter.
func (s *single[T]) Result(out *T) error {
	return s.coll.FindOne(s.ctx, s.filter, s.buildOpts()...).Decode(out)
}

//#endregion

// #region (many)
type many[T any] struct {
	queryCore
	limit int64
	skip  int64
	batch *int32
	extra *options.FindOptions
}

var _ ManyResult[any] = (*many[any])(nil)

// NewMany creates a ManyResult builder for filter on coll. Normally called via Collection.Find/FindMany.
func NewMany[T any](ctx context.Context, coll *mongo.Collection, collName string, filter any) ManyResult[T] {
	return &many[T]{
		queryCore: newQueryCore(ctx, coll, collName, filter),
	}
}

// Proj sets the projection document and returns a new builder.
func (m *many[T]) Proj(p any) ManyResult[T] { next := *m; next.proj = p; return &next }

// Sort sets the sort document and returns a new builder.
func (m *many[T]) Sort(s any) ManyResult[T] { next := *m; next.sort = s; return &next }

// Hint sets the index hint and returns a new builder.
func (m *many[T]) Hint(h any) ManyResult[T] { next := *m; next.hint = h; return &next }

// Limit caps the number of documents returned and returns a new builder.
func (m *many[T]) Limit(n int64) ManyResult[T] { next := *m; next.limit = n; return &next }

// Skip skips the first n documents and returns a new builder.
func (m *many[T]) Skip(n int64) ManyResult[T] { next := *m; next.skip = n; return &next }

// Bsz sets the cursor batch size and returns a new builder.
func (m *many[T]) Bsz(n int32) ManyResult[T] { next := *m; next.batch = &n; return &next }

// Opts merges raw FindOptions on top of the builder settings.
func (m *many[T]) Opts(fo *options.FindOptions) ManyResult[T] {
	next := *m; next.extra = fo; return &next
}

// buildOpts assembles the final FindOptions from builder state, applying extra as override.
func (m *many[T]) buildOpts() []*options.FindOptions {
	fo := options.Find()
	if m.limit > 0 {
		fo.SetLimit(m.limit)
	}
	if m.skip > 0 {
		fo.SetSkip(m.skip)
	}
	if m.sort != nil {
		fo.SetSort(m.sort)
	}
	if m.proj != nil {
		fo.SetProjection(m.proj)
	}
	if m.hint != nil {
		fo.SetHint(m.hint)
	}
	if m.batch != nil {
		fo.SetBatchSize(*m.batch)
	}
	if m.extra != nil {
		return []*options.FindOptions{fo, m.extra}
	}
	return []*options.FindOptions{fo}
}
// All executes the query and returns all matching documents.
func (m *many[T]) All() ([]T, error) {
	var out []T
	cur, err := m.coll.Find(m.ctx, m.filter, m.buildOpts()...)
	if err != nil {
		return nil, err
	}
	defer cur.Close(m.ctx)
	for cur.Next(m.ctx) {
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

// Result executes the query and decodes all matching documents into out.
func (m *many[T]) Result(out *[]T) error {
	items, err := m.All()
	if err != nil {
		return err
	}
	*out = items
	return nil
}

// Stream executes the query and calls fn for each document.
// Stops and returns the first non-nil error from fn.
func (m *many[T]) Stream(fn func(ctx context.Context, doc T) error) error {
	if fn == nil {
		return errors.New("fn must not be nil")
	}
	cur, err := m.coll.Find(m.ctx, m.filter, m.buildOpts()...)
	if err != nil {
		return err
	}
	defer cur.Close(m.ctx)
	for cur.Next(m.ctx) {
		var v T
		if err := cur.Decode(&v); err != nil {
			return err
		}
		if err := fn(m.ctx, v); err != nil {
			return err
		}
	}
	return cur.Err()
}

// Each is an alias for Stream.
func (m *many[T]) Each(fn func(ctx context.Context, doc T) error) error { return m.Stream(fn) }

// Cnt returns the count of documents matching the filter (ignores Limit/Skip).
func (m *many[T]) Cnt() (int64, error) { return m.coll.CountDocuments(m.ctx, m.filter) }

//#endregion

// #region helper functions

// ensureUpdateHasTimestamp wraps a raw update document in $set if needed and injects
// updated_at = now when the caller has not already set it.
func ensureUpdateHasTimestamp(update any) bson.M {
	const updatedAt = "updated_at"
	now := time.Now().UTC()

	u := toBsonM(update)

	hasOp := false
	for k := range u {
		if len(k) > 0 && k[0] == '$' {
			hasOp = true
			break
		}
	}
	if !hasOp {
		u = bson.M{"$set": u}
	}

	set := toBsonM(u["$set"])
	if _, alreadySet := set[updatedAt]; !alreadySet {
		set[updatedAt] = now
	}
	u["$set"] = set

	return u
}

// toBsonM converts common BSON/map types to bson.M. Returns an empty bson.M for nil or unsupported types.
func toBsonM(v any) bson.M {
	switch m := v.(type) {
	case nil:
		return bson.M{}
	case bson.M:
		return m
	case map[string]any:
		return bson.M(m)
	case bson.D:
		return m.Map()
	default:
		rv := reflect.ValueOf(v)
		if rv.IsValid() && rv.Kind() == reflect.Map && rv.Type().Key().Kind() == reflect.String {
			out := bson.M{}
			iter := rv.MapRange()
			for iter.Next() {
				out[iter.Key().String()] = iter.Value().Interface()
			}
			return out
		}
		return bson.M{}
	}
}

//#endregion
