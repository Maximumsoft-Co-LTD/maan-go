package mongo

import (
	"context"
	"strings"

	"go.mongodb.org/mongo-driver/bson"

	mg "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type collection[T any] struct {
	ctx      *context.Context
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
		ctx:      &ctx,
		client:   client,
		collName: name,
		read:     readDB.Collection(name),
		write:    writeDB.Collection(name),
	}
}

func (c *collection[T]) Ctx(ctx context.Context) Collection[T] {
	c.ctx = &ctx
	return Collection[T](c) // ใช้ type assertion เพื่อส่งคืนประเภท Collection[T] ที่ถูกต้อง
}

func (c *collection[T]) getCtx() context.Context {
	if c.ctx == nil {
		return context.Background()
	}
	return *c.ctx
}

func (c *collection[T]) Build(ctx context.Context) ExtendedCollection[T] {
	return NewExtendedCollection[T](ctx, c.client, c.read, c.write, c.collName)
}

func (c *collection[T]) FindOne(query any) SingleResult[T] {
	if query == nil {
		query = bson.M{}
	}
	return NewSingle[T](c.getCtx(), c.client, c.collName, query)
}

func (c *collection[T]) FindMany(filter any) ManyResult[T] {
	if filter == nil {
		filter = bson.M{}
	}
	return NewMany[T](c.getCtx(), c.client, c.collName, filter)
}

func (c *collection[T]) Create(doc *T) error {
	if doc == nil {
		return nil
	}
	applyModelDefaults(doc)
	_, err := c.write.InsertOne(c.getCtx(), doc)
	return err
}

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

func (c *collection[T]) Save(filter any, update any) error {
	opt := options.Update().SetUpsert(true)
	_, err := c.write.UpdateOne(c.getCtx(), filter, ensureUpdateHasTimestamp(update), opt)
	return err
}

func (c *collection[T]) SaveMany(filter any, update any) error {
	opt := options.Update().SetUpsert(true)
	_, err := c.write.UpdateMany(c.getCtx(), filter, ensureUpdateHasTimestamp(update), opt)
	return err
}

func (c *collection[T]) Upd(filter any, update any) error {
	_, err := c.write.UpdateOne(c.getCtx(), filter, ensureUpdateHasTimestamp(update))
	return err
}

func (c *collection[T]) UpdMany(filter any, update any) error {
	_, err := c.write.UpdateMany(c.getCtx(), filter, ensureUpdateHasTimestamp(update))
	return err
}

func (c *collection[T]) Del(filter any) error {
	_, err := c.write.DeleteOne(c.getCtx(), filter)
	return err
}

func (c *collection[T]) Agg(pipeline any) Aggregate[T] {
	return NewAgg[T](c.getCtx(), c.client, c.collName, pipeline)
}

func (c *collection[T]) ReFind(q string, fields ...string) ([]T, error) {
	ors := make([]bson.M, 0, len(fields))
	for _, f := range fields {
		ors = append(ors, bson.M{f: bson.M{"$regex": q, "$options": "i"}})
	}
	return c.Find(bson.M{"$or": ors}).All()
}

func (c *collection[T]) TxtFind(q string) ([]T, error) {
	return c.Find(bson.M{"$text": bson.M{"$search": q}}).All()
}

func (c *collection[T]) Name() string {
	return c.collName
}

func (c *collection[T]) StartTx() (TxSession, error) {
	return NewTransactionSession(c.getCtx(), c.client.Write(), c.read, c.write)
}

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
	return NewMany[T](c.getCtx(), c.client, c.collName, filter)
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
// 			q = q.Lim(opt.Limit)
// 		}
// 		if opt.Skip > 0 {
// 			q = q.Skp(opt.Skip)
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
