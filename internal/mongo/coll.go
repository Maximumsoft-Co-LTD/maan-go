package mongo

import (
	"context"
	"errors"
	"strings"

	"go.mongodb.org/mongo-driver/bson"

	mg "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type collection[T any] struct {
	ctx      context.Context
	client   Client
	collName string
	read     *mg.Collection
	write    *mg.Collection
}

var _ Collection[any] = (*collection[any])(nil)

// NewCollection creates a strongly typed collection wrapper scoped to the provided context.
func NewCollection[T any](ctx context.Context, client Client, name string) Collection[T] {
	dbName := client.DbName()
	readDB := client.Read().Database(dbName)
	writeDB := client.Write().Database(dbName)
	return &collection[T]{
		ctx:      normalizeCtx(ctx),
		client:   client,
		collName: name,
		read:     readDB.Collection(name),
		write:    writeDB.Collection(name),
	}
}

// Ctx returns a shallow copy of the collection bound to ctx.
// The original collection is unchanged.
func (c *collection[T]) Ctx(ctx context.Context) Collection[T] {
	next := *c
	next.ctx = normalizeCtx(ctx)
	return &next
}

// getCtx returns the collection's context, falling back to context.Background() when unset.
func (c *collection[T]) getCtx() context.Context {
	if c == nil {
		return context.Background()
	}
	return normalizeCtx(c.ctx)
}

// normalizeCtx returns ctx unchanged, or context.Background() when ctx is nil.
func normalizeCtx(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

// Build returns an ExtendedCollection bound to ctx, sharing the same underlying
// read and write handles as this collection.
func (c *collection[T]) Build(ctx context.Context) ExtendedCollection[T] {
	return NewExtendedCollection[T](ctx, c.read, c.write, c.collName)
}

// FindOne returns a SingleResult builder for the given query filter.
// Pass nil to match any document.
func (c *collection[T]) FindOne(query any) SingleResult[T] {
	if query == nil {
		query = bson.M{}
	}
	return NewSingle[T](c.getCtx(), c.read, c.collName, query)
}

// FindMany returns a ManyResult builder for the given filter.
// Pass nil to match all documents.
func (c *collection[T]) FindMany(filter any) ManyResult[T] {
	if filter == nil {
		filter = bson.M{}
	}
	return NewMany[T](c.getCtx(), c.read, c.collName, filter)
}

// Create inserts doc as a single document. Returns error if doc is nil.
// Automatically calls model-default hooks before inserting.
func (c *collection[T]) Create(doc *T, opts ...*options.InsertOneOptions) error {
	if doc == nil {
		return errors.New("doc must not be nil")
	}
	applyModelDefaults(doc)
	_, err := c.write.InsertOne(c.getCtx(), doc, opts...)
	return err
}

// CreateMany inserts every element of docs. Returns error if docs is nil.
// Calls model-default hooks on each element before inserting.
func (c *collection[T]) CreateMany(docs *[]T, opts ...*options.InsertManyOptions) error {
	if docs == nil {
		return errors.New("docs must not be nil")
	}
	if len(*docs) == 0 {
		return nil
	}
	items := *docs
	nd := make([]any, len(items))
	for i := range items {
		applyModelDefaults(&items[i])
		nd[i] = items[i]
	}
	_, err := c.write.InsertMany(c.getCtx(), nd, opts...)
	return err
}

// Save performs an upsert: updates the first matching document or inserts if none found.
// Automatically injects updated_at into the $set clause.
func (c *collection[T]) Save(filter any, update any, opts ...*options.UpdateOptions) error {
	_, err := c.write.UpdateOne(c.getCtx(), filter, ensureUpdateHasTimestamp(update), opts...)
	return err
}

// SaveMany performs a multi-document upsert for all documents matching filter.
// Automatically injects updated_at into the $set clause.
func (c *collection[T]) SaveMany(filter any, update any, opts ...*options.UpdateOptions) error {
	_, err := c.write.UpdateMany(c.getCtx(), filter, ensureUpdateHasTimestamp(update), opts...)
	return err
}

// Upd updates the first document matching filter. Does NOT insert when no match exists.
func (c *collection[T]) Upd(filter any, update any, opts ...*options.UpdateOptions) error {
	_, err := c.write.UpdateOne(c.getCtx(), filter, ensureUpdateHasTimestamp(update), opts...)
	return err
}

// UpdMany updates all documents matching filter. Does NOT insert new documents.
func (c *collection[T]) UpdMany(filter any, update any, opts ...*options.UpdateOptions) error {
	_, err := c.write.UpdateMany(c.getCtx(), filter, ensureUpdateHasTimestamp(update), opts...)
	return err
}

// Del deletes the first document matching filter.
func (c *collection[T]) Del(filter any, opts ...*options.DeleteOptions) error {
	_, err := c.write.DeleteOne(c.getCtx(), filter, opts...)
	return err
}

// Agg returns an Aggregate builder for the given aggregation pipeline.
func (c *collection[T]) Agg(pipeline any) Aggregate[T] {
	return NewAgg[T](c.getCtx(), c.read, c.collName, pipeline)
}

// RegexFields performs a case-insensitive regex search across the given field names.
// Equivalent to { $or: [ {field: {$regex: q, $options: "i"}} ... ] }.
func (c *collection[T]) RegexFields(q string, fields ...string) ([]T, error) {
	ors := make([]bson.M, 0, len(fields))
	for _, f := range fields {
		ors = append(ors, bson.M{f: bson.M{"$regex": q, "$options": "i"}})
	}
	return c.Find(bson.M{"$or": ors}).All()
}

// TxtFind performs a MongoDB full-text search ($text / $search).
// Requires a text index on the collection.
func (c *collection[T]) TxtFind(q string) ([]T, error) {
	return c.Find(bson.M{"$text": bson.M{"$search": q}}).All()
}

// Name returns the MongoDB collection name.
func (c *collection[T]) Name() string {
	return c.collName
}

// StartTx begins a manual transaction and returns a TxSession.
// Call tx.Close(&err) (usually via defer) to commit or rollback.
func (c *collection[T]) StartTx() (TxSession, error) {
	return NewTransactionSession(c.getCtx(), c.client)
}

// WithTx runs fn inside an automatically managed transaction.
// Commits when fn returns nil; rolls back otherwise.
func (c *collection[T]) WithTx(fn func(ctx context.Context) error) error {
	sess, err := c.client.Write().StartSession()
	if err != nil {
		return err
	}
	defer sess.EndSession(c.getCtx())

	_, err = sess.WithTransaction(c.getCtx(), func(sc mg.SessionContext) (any, error) {
		return nil, fn(sc)
	})
	return err
}

// Find returns a ManyResult builder for filter. Alias for FindMany.
func (c *collection[T]) Find(filter any) ManyResult[T] {
	if filter == nil {
		filter = bson.M{}
	}
	return NewMany[T](c.getCtx(), c.read, c.collName, filter)
}
// Watch returns a ChangeStream builder for real-time change events on the collection.
// Requires a MongoDB replica set or sharded cluster.
func (c *collection[T]) Watch(ctx context.Context, pipeline ...any) ChangeStream[T] {
	return NewChangeStream[T](ctx, c.read, c.collName, pipeline)
}

// copyFilter returns a shallow copy of original to avoid mutation of shared filter maps.
func copyFilter(original bson.M) bson.M {
	newFilter := bson.M{}
	for k, v := range original {
		newFilter[k] = v
	}
	return newFilter
}

// toSnakeCase converts a CamelCase string to snake_case (e.g. "UserName" → "user_name").
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// func (c *coll[T]) FindPg(filter any, opt *portOut.PageOpt) ([]T, int64, error) {
// 	q := c.Find(filter)
// 	if opt != nil {
// 		if opt.Limit > 0 {
// 			q = q.Limit(opt.Limit)
// 		}
// 		if opt.Skip > 0 {
// 			q = q.Skip(opt.Skip)
// 		}
// 		if opt.Sort != nil {
// 			q = q.Sort(opt.Sort)
// 		}
// 		if opt.Projection != nil {
// 			q = q.Proj(opt.Projection)
// 		}
// 		if opt.Hint != nil {
// 			q = q.Hint(opt.Hint)
// 		}
// 	}
// 	items, err := q.All()
// 	if err != nil {
// 		return nil, 0, err
// 	}
// 	total, err := c.read.CountDocuments(c.ctx, filter)
// 	if err != nil {
// 		return nil, 0, err
// 	}
// 	return items, total, nil
// }
