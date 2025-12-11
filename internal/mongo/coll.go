package mongo

import (
	"context"
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

// Ctx is a helper method to set the context for the collection.
// Example:
// col := NewCollection[testDoc](context.Background(), client, "docs").Ctx(context.Background())
// col.FindOne(bson.M{"name": "foo"})
// Ctx will return a new collection instance with the context set.
// The context is isolated and will not affect the original collection.
func (c *collection[T]) Ctx(ctx context.Context) Collection[T] {
	next := *c
	next.ctx = normalizeCtx(ctx)
	return &next
}

func (c *collection[T]) getCtx() context.Context {
	if c == nil {
		return context.Background()
	}
	return normalizeCtx(c.ctx)
}

func normalizeCtx(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

// Build is a helper method to build a new extended collection.
// Example:
// builder := col.Build(context.Background())
// builder.By("Name", "foo")
// builder.Where(bson.M{"active": true})
// Build will return a new extended collection instance with the context set.
// The extended collection is isolated and will not affect the original collection.
func (c *collection[T]) Build(ctx context.Context) ExtendedCollection[T] {
	return NewExtendedCollection[T](ctx, c.read, c.write, c.collName)
}

// FindOne is a helper method to find a single document in the collection.
// Example:
// result := col.FindOne(bson.M{"name": "foo"})
// result.Result(&result)
// FindOne will return a new single result instance with the context set.
// The single result is isolated and will not affect the original collection.
func (c *collection[T]) FindOne(query any) SingleResult[T] {
	if query == nil {
		query = bson.M{}
	}
	return NewSingle[T](c.getCtx(), c.read, c.collName, query)
}

// FindMany is a helper method to find many documents in the collection.
// Example:
// results := col.FindMany(bson.M{"name": "foo"})
// results.Result(&results)
// FindMany will return a new many result instance with the context set.
// The many result is isolated and will not affect the original collection.
func (c *collection[T]) FindMany(filter any) ManyResult[T] {
	if filter == nil {
		filter = bson.M{}
	}
	return NewMany[T](c.getCtx(), c.read, c.collName, filter)
}

// Create is a helper method to create a new document in the collection.
// Example:
// err := col.Create(&testDoc{Name: "foo"})
// Create will return a new document in the collection with the context set.
// The document is isolated and will not affect the original collection.
func (c *collection[T]) Create(doc *T) error {
	if doc == nil {
		return nil
	}
	applyModelDefaults(doc)
	_, err := c.write.InsertOne(c.getCtx(), doc)
	return err
}

// CreateMany is a helper method to create many documents in the collection.
// Example:
// err := col.CreateMany(&[]testDoc{{Name: "foo"}, {Name: "bar"}})
// CreateMany will return a new documents in the collection with the context set.
// The documents are isolated and will not affect the original collection.
func (c *collection[T]) CreateMany(docs *[]T) error {
	if docs == nil || len(*docs) == 0 {
		return nil
	}
	items := *docs
	nd := make([]any, len(items))
	for i := range items {
		applyModelDefaults(&items[i])
		nd[i] = items[i]
	}
	_, err := c.write.InsertMany(c.getCtx(), nd)
	return err
}

// Save is a helper method to save (upsert) a document in the collection.
// If document exists, it will be updated. If not, it will be inserted.
// Example:
// err := col.Save(bson.M{"name": "foo"}, bson.M{"$set": bson.M{"name": "bar"}})
// Save will upsert a document in the collection with the context set.
// The document is isolated and will not affect the original collection.
func (c *collection[T]) Save(filter any, update any) error {
	opt := options.Update().SetUpsert(true)
	_, err := c.write.UpdateOne(c.getCtx(), filter, ensureUpdateHasTimestamp(update), opt)
	return err
}

// SaveMany is a helper method to save (upsert) many documents in the collection.
// If documents exist, they will be updated. If not, they will be inserted.
// Example:
// err := col.SaveMany(bson.M{"name": "foo"}, bson.M{"$set": bson.M{"name": "bar"}})
// SaveMany will upsert documents in the collection with the context set.
// The documents are isolated and will not affect the original collection.
func (c *collection[T]) SaveMany(filter any, update any) error {
	opt := options.Update().SetUpsert(true)
	_, err := c.write.UpdateMany(c.getCtx(), filter, ensureUpdateHasTimestamp(update), opt)
	return err
}

// Upd is a helper method to update an existing document in the collection.
// Will only update documents that already exist, will NOT insert new documents.
// Example:
// err := col.Upd(bson.M{"name": "foo"}, bson.M{"$set": bson.M{"name": "bar"}})
// Upd will update existing documents only in the collection with the context set.
// The document is isolated and will not affect the original collection.
func (c *collection[T]) Upd(filter any, update any) error {
	_, err := c.write.UpdateOne(c.getCtx(), filter, ensureUpdateHasTimestamp(update))
	return err
}

// UpdMany is a helper method to update many existing documents in the collection.
// Will only update documents that already exist, will NOT insert new documents.
// Example:
// err := col.UpdMany(bson.M{"name": "foo"}, bson.M{"$set": bson.M{"name": "bar"}})
// UpdMany will update existing documents only in the collection with the context set.
// The documents are isolated and will not affect the original collection.
func (c *collection[T]) UpdMany(filter any, update any) error {
	_, err := c.write.UpdateMany(c.getCtx(), filter, ensureUpdateHasTimestamp(update))
	return err
}

// Del is a helper method to delete a document in the collection.
// Example:
// err := col.Del(bson.M{"name": "foo"})
// Del will return a new document in the collection with the context set.
// The document is isolated and will not affect the original collection.
func (c *collection[T]) Del(filter any) error {
	_, err := c.write.DeleteOne(c.getCtx(), filter)
	return err
}

// Agg is a helper method to create a new aggregate.
// Example:
// agg := col.Agg(bson.M{"$match": bson.M{"name": "foo"}})
// agg.Disk(true)
// agg.Bsz(100)
// Agg will return a new aggregate instance with the context set.
// The aggregate is isolated and will not affect the original collection.
func (c *collection[T]) Agg(pipeline any) Aggregate[T] {
	return NewAgg[T](c.getCtx(), c.read, c.collName, pipeline)
}

// RegexFields is a helper method to find documents in the collection by regex.
// Example:
// results := col.RegexFields("foo", "name", "email")
// RegexFields will return a new documents in the collection with the context set.
// The documents are isolated and will not affect the original collection.
func (c *collection[T]) RegexFields(q string, fields ...string) ([]T, error) {
	ors := make([]bson.M, 0, len(fields))
	for _, f := range fields {
		ors = append(ors, bson.M{f: bson.M{"$regex": q, "$options": "i"}})
	}
	return c.Find(bson.M{"$or": ors}).All()
}

// TxtFind is a helper method to find documents in the collection by text search.
// Example:
// results := col.TxtFind("foo")
// TxtFind will return a new documents in the collection with the context set.
// The documents are isolated and will not affect the original collection.
func (c *collection[T]) TxtFind(q string) ([]T, error) {
	return c.Find(bson.M{"$text": bson.M{"$search": q}}).All()
}

// Name is a helper method to get the name of the collection.
// Example:
// name := col.Name()
// Name will return the name of the collection with the context set.
// The name is isolated and will not affect the original collection.
func (c *collection[T]) Name() string {
	return c.collName
}

// StartTx is a helper method to start a new transaction.
// Example:
// tx, err := col.StartTx()
// tx.Ctx()
// StartTx will return a new transaction instance with the context set.
// The transaction is isolated and will not affect the original collection.
func (c *collection[T]) StartTx() (TxSession, error) {
	return NewTransactionSession(c.getCtx(), c.client)
}

// WithTx is a helper method to execute a function with a transaction.
// Example:
//
//	err := col.WithTx(func(ctx context.Context) error {
//		return nil
//	})
//
// WithTx will return a new error with the context set.
// The error is isolated and will not affect the original collection.
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

func (c *collection[T]) Find(filter any) ManyResult[T] {
	if filter == nil {
		filter = bson.M{}
	}
	return NewMany[T](c.getCtx(), c.read, c.collName, filter)
}

func copyFilter(original bson.M) bson.M {
	newFilter := bson.M{}
	for k, v := range original {
		newFilter[k] = v
	}
	return newFilter
}

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
