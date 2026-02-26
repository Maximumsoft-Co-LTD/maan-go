package mongo

import (
	"context"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// #region (queryCore)

// queryCore holds the common state for MongoDB query builders.
// It is not exported and is meant to be embedded by other query structs.
type queryCore struct {
	ctx      context.Context
	collName string
	coll     *mongo.Collection
	filter   any
	proj     any
	sort     any
	hint     any
}

// newQueryCore initializes the common fields for a query builder.
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

// NewSingle creates a new query builder for a single document.
func NewSingle[T any](ctx context.Context, coll *mongo.Collection, collName string, query any) SingleResult[T] {
	return &single[T]{
		queryCore: newQueryCore(ctx, coll, collName, query),
	}
}

func (s *single[T]) Proj(p any) SingleResult[T]                     { next := *s; next.proj = p; return &next }
func (s *single[T]) Sort(v any) SingleResult[T]                     { next := *s; next.sort = v; return &next }
func (s *single[T]) Hint(v any) SingleResult[T]                     { next := *s; next.hint = v; return &next }
func (s *single[T]) Opts(o *options.FindOneOptions) SingleResult[T] { next := *s; next.extra = o; return &next }
func (s *single[T]) build() *options.FindOneOptions {
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
		return mergeFindOneOptions(fo, s.extra)
	}
	return fo
}
func (s *single[T]) Result(out *T) error {
	return s.coll.FindOne(s.ctx, s.filter, s.build()).Decode(out)
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

// NewMany creates a new query builder for multiple documents.
func NewMany[T any](ctx context.Context, coll *mongo.Collection, collName string, filter any) ManyResult[T] {
	return &many[T]{
		queryCore: newQueryCore(ctx, coll, collName, filter),
	}
}

func (m *many[T]) Proj(p any) ManyResult[T]                   { next := *m; next.proj = p; return &next }
func (m *many[T]) Sort(s any) ManyResult[T]                   { next := *m; next.sort = s; return &next }
func (m *many[T]) Hint(h any) ManyResult[T]                   { next := *m; next.hint = h; return &next }
func (m *many[T]) Limit(n int64) ManyResult[T]                { next := *m; next.limit = n; return &next }
func (m *many[T]) Skip(n int64) ManyResult[T]                 { next := *m; next.skip = n; return &next }
func (m *many[T]) Bsz(n int32) ManyResult[T]                  { next := *m; next.batch = &n; return &next }
func (m *many[T]) Opts(fo *options.FindOptions) ManyResult[T] { next := *m; next.extra = fo; return &next }
func (m *many[T]) build() *options.FindOptions {
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
		return mergeFindOptions(fo, m.extra)
	}
	return fo
}
func (m *many[T]) All() ([]T, error) {
	var out []T
	cur, err := m.coll.Find(m.ctx, m.filter, m.build())
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

func (m *many[T]) Result(out *[]T) error {
	items, err := m.All()
	if err != nil {
		return err
	}
	*out = items
	return nil
}

func (m *many[T]) Stream(fn func(ctx context.Context, doc T) error) error {
	cur, err := m.coll.Find(m.ctx, m.filter, m.build())
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

func (m *many[T]) Each(fn func(ctx context.Context, doc T) error) error { return m.Stream(fn) }
func (m *many[T]) Cnt() (int64, error)                                  { return m.coll.CountDocuments(m.ctx, m.filter) }

//#endregion

// #region helper functions

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
	set[updatedAt] = now
	u["$set"] = set

	return u
}

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

func mergeFindOptions(fo *options.FindOptions, opts ...*options.FindOptions) *options.FindOptions {
	if opts == nil {
		return fo
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}

		if opt.AllowDiskUse != nil {
			fo.AllowDiskUse = opt.AllowDiskUse
		}
		if opt.AllowPartialResults != nil {
			fo.AllowPartialResults = opt.AllowPartialResults
		}
		if opt.BatchSize != nil {
			fo.BatchSize = opt.BatchSize
		}
		if opt.Collation != nil {
			fo.Collation = opt.Collation
		}
		if opt.Comment != nil {
			fo.Comment = opt.Comment
		}
		if opt.CursorType != nil {
			fo.CursorType = opt.CursorType
		}
		if opt.Let != nil {
			fo.Let = opt.Let
		}
		if opt.Max != nil {
			fo.Max = opt.Max
		}
		if opt.MaxAwaitTime != nil {
			fo.MaxAwaitTime = opt.MaxAwaitTime
		}
		if opt.MaxTime != nil {
			fo.MaxTime = opt.MaxTime
		}
		if opt.Min != nil {
			fo.Min = opt.Min
		}
		if opt.NoCursorTimeout != nil {
			fo.NoCursorTimeout = opt.NoCursorTimeout
		}
		if opt.ReturnKey != nil {
			fo.ReturnKey = opt.ReturnKey
		}
		if opt.ShowRecordID != nil {
			fo.ShowRecordID = opt.ShowRecordID
		}
		if opt.Sort != nil {
			fo.Sort = opt.Sort
		}
		if opt.Projection != nil {
			fo.Projection = opt.Projection
		}
		if opt.Hint != nil {
			fo.Hint = opt.Hint
		}
		if opt.Limit != nil {
			fo.Limit = opt.Limit
		}
		if opt.Skip != nil {
			fo.Skip = opt.Skip
		}
	}
	return fo
}

func mergeFindOneOptions(fo *options.FindOneOptions, opts ...*options.FindOneOptions) *options.FindOneOptions {
	if opts == nil {
		return fo
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if opt.AllowPartialResults != nil {
			fo.AllowPartialResults = opt.AllowPartialResults
		}
		if opt.Collation != nil {
			fo.Collation = opt.Collation
		}
		if opt.Comment != nil {
			fo.Comment = opt.Comment
		}
		if opt.Max != nil {
			fo.Max = opt.Max
		}
		if opt.MaxTime != nil {
			fo.MaxTime = opt.MaxTime
		}
		if opt.Min != nil {
			fo.Min = opt.Min
		}
		if opt.ReturnKey != nil {
			fo.ReturnKey = opt.ReturnKey
		}
		if opt.ShowRecordID != nil {
			fo.ShowRecordID = opt.ShowRecordID
		}
		if opt.Sort != nil {
			fo.Sort = opt.Sort
		}
		if opt.Projection != nil {
			fo.Projection = opt.Projection
		}
		if opt.Hint != nil {
			fo.Hint = opt.Hint
		}
	}

	return fo
}

//#endregion
