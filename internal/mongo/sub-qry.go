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
	m        Client
	collName string
	coll     *mongo.Collection
	filter   any
	proj     any
	sort     any
	hint     any
}

// newQueryCore initializes the common fields for a query builder.
func newQueryCore(ctx context.Context, m Client, collName string, filter any) queryCore {
	if filter == nil {
		filter = bson.M{}
	}
	return queryCore{
		ctx:      ctx,
		m:        m,
		collName: collName,
		coll:     m.Read().Database(m.DbName()).Collection(collName),
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
func NewSingle[T any](ctx context.Context, m Client, collName string, query any) SingleResult[T] {
	return &single[T]{
		queryCore: newQueryCore(ctx, m, collName, query),
	}
}

func (s *single[T]) Proj(p any) SingleResult[T]                     { s.proj = p; return s }
func (s *single[T]) Sort(v any) SingleResult[T]                     { s.sort = v; return s }
func (s *single[T]) Hint(v any) SingleResult[T]                     { s.hint = v; return s }
func (s *single[T]) Opts(o *options.FindOneOptions) SingleResult[T] { s.extra = o; return s }
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
		if s.extra.Projection != nil {
			fo.SetProjection(s.extra.Projection)
		}
		if s.extra.Sort != nil {
			fo.SetSort(s.extra.Sort)
		}
		if s.extra.Hint != nil {
			fo.SetHint(s.extra.Hint)
		}
	}
	return fo
}
func (s *single[T]) Res(out *T) error {
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
func NewMany[T any](ctx context.Context, m Client, collName string, filter any) ManyResult[T] {
	return &many[T]{
		queryCore: newQueryCore(ctx, m, collName, filter),
	}
}

func (m *many[T]) Proj(p any) ManyResult[T]                   { m.proj = p; return m }
func (m *many[T]) Sort(s any) ManyResult[T]                   { m.sort = s; return m }
func (m *many[T]) Hint(h any) ManyResult[T]                   { m.hint = h; return m }
func (m *many[T]) Lim(n int64) ManyResult[T]                  { m.limit = n; return m }
func (m *many[T]) Skp(n int64) ManyResult[T]                  { m.skip = n; return m }
func (m *many[T]) Bsz(n int32) ManyResult[T]                  { m.batch = &n; return m }
func (m *many[T]) Opts(fo *options.FindOptions) ManyResult[T] { m.extra = fo; return m }
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
		if m.extra.Limit != nil {
			fo.SetLimit(*m.extra.Limit)
		}
		if m.extra.Skip != nil {
			fo.SetSkip(*m.extra.Skip)
		}
		if m.extra.Sort != nil {
			fo.SetSort(m.extra.Sort)
		}
		if m.extra.Projection != nil {
			fo.SetProjection(m.extra.Projection)
		}
		if m.extra.Hint != nil {
			fo.SetHint(m.extra.Hint)
		}
		if m.extra.BatchSize != nil {
			fo.SetBatchSize(*m.extra.BatchSize)
		}
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

func (m *many[T]) Res(out *[]T) error {
	items, err := m.All()
	if err != nil {
		return err
	}
	*out = items
	return nil
}

func (m *many[T]) Strm(fn func(ctx context.Context, doc T) error) error {
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

func (m *many[T]) Each(fn func(ctx context.Context, doc T) error) error { return m.Strm(fn) }
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

//#endregion
