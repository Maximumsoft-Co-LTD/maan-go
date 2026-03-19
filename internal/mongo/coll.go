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
	return NewSingle[T](c.getCtx(), c.getReadColl(), c.collName, query)
}

// FindMany returns a ManyResult builder for the given filter.
// Pass nil to match all documents.
func (c *collection[T]) FindMany(filter any) ManyResult[T] {
	if filter == nil {
		filter = bson.M{}
	}
	return NewMany[T](c.getCtx(), c.getReadColl(), c.collName, filter)
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
	upsertOpt := options.Update().SetUpsert(true)
	allOpts := append([]*options.UpdateOptions{upsertOpt}, opts...)
	_, err := c.write.UpdateOne(c.getCtx(), filter, ensureUpdateHasTimestamp(update), allOpts...)
	return err
}

// SaveMany performs a multi-document upsert for all documents matching filter.
// Automatically injects updated_at into the $set clause.
func (c *collection[T]) SaveMany(filter any, update any, opts ...*options.UpdateOptions) error {
	upsertOpt := options.Update().SetUpsert(true)
	allOpts := append([]*options.UpdateOptions{upsertOpt}, opts...)
	_, err := c.write.UpdateMany(c.getCtx(), filter, ensureUpdateHasTimestamp(update), allOpts...)
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

// DelMany deletes all documents matching filter.
func (c *collection[T]) DelMany(filter any, opts ...*options.DeleteOptions) error {
	_, err := c.write.DeleteMany(c.getCtx(), filter, opts...)
	return err
}

// FindOneAndUpd atomically finds the first document matching filter, applies update, and decodes the result into out.
// Automatically injects updated_at into the $set clause.
func (c *collection[T]) FindOneAndUpd(filter any, update any, out *T, opts ...*options.FindOneAndUpdateOptions) error {
	return c.write.FindOneAndUpdate(c.getCtx(), filter, ensureUpdateHasTimestamp(update), opts...).Decode(out)
}

// FindOneAndDel atomically finds the first document matching filter, deletes it, and decodes it into out.
func (c *collection[T]) FindOneAndDel(filter any, out *T, opts ...*options.FindOneAndDeleteOptions) error {
	return c.write.FindOneAndDelete(c.getCtx(), filter, opts...).Decode(out)
}

// Distinct returns the distinct values for the given field across all documents matching filter.
// Pass nil filter to match all documents.
func (c *collection[T]) Distinct(field string, filter any) ([]any, error) {
	if filter == nil {
		filter = bson.M{}
	}
	return c.getReadColl().Distinct(c.getCtx(), field, filter)
}

// Count returns the number of documents matching filter.
// Pass nil filter to count all documents.
func (c *collection[T]) Count(filter any) (int64, error) {
	if filter == nil {
		filter = bson.M{}
	}
	return c.getReadColl().CountDocuments(c.getCtx(), filter)
}

// Agg returns an Aggregate builder for the given aggregation pipeline.
func (c *collection[T]) Agg(pipeline any) Aggregate[T] {
	return NewAgg[T](c.getCtx(), c.getReadColl(), c.collName, pipeline)
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
	return c.client.StartTx(c.getCtx())
}

// WithTx runs fn inside an automatically managed transaction.
// Commits when fn returns nil; rolls back otherwise. Panics inside fn are caught and rolled back.
func (c *collection[T]) WithTx(fn func(ctx context.Context) error) (retErr error) {
	return c.client.WithTx(c.getCtx(), fn)
}

func (c *collection[T]) getReadColl() *mg.Collection {
	if mg.SessionFromContext(c.getCtx()) != nil { // Check has transaction
		return c.write
	}
	return c.read
}

// Find returns a ManyResult builder for filter. Alias for FindMany.
func (c *collection[T]) Find(filter any) ManyResult[T] {
	if filter == nil {
		filter = bson.M{}
	}
	return NewMany[T](c.getCtx(), c.getReadColl(), c.collName, filter)
}

// Watch returns a ChangeStream builder for real-time change events on the collection.
// The collection's bound context (set via Ctx) controls the stream lifetime.
// Requires a MongoDB replica set or sharded cluster.
func (c *collection[T]) Watch(pipeline ...any) ChangeStream[T] {
	return NewChangeStream[T](c.getCtx(), c.getReadColl(), c.collName, pipeline)
}

// Idx returns an IndexManager for managing indexes on this collection.
func (c *collection[T]) Idx() IndexManager {
	return &indexManager{ctx: c.getCtx(), view: c.write.Indexes()}
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
