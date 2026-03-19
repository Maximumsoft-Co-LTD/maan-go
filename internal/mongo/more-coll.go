package mongo

import (
	"context"
	"reflect"
	"strings"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	mg "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//#region more-coll

type extendedCollection[T any] struct {
	ctx    context.Context
	read   *mg.Collection
	write  *mg.Collection
	name   string
	filter bson.M
}

var _ ExtendedCollection[any] = (*extendedCollection[any])(nil)

// NewExtendedCollection creates a chainable query builder bound to the provided Mongo clients.
func NewExtendedCollection[T any](ctx context.Context, collRead *mg.Collection, collWrite *mg.Collection, name string) ExtendedCollection[T] {
	return &extendedCollection[T]{ctx: normalizeCtx(ctx), read: collRead, write: collWrite, name: name, filter: bson.M{}}
}

// clone returns a shallow copy of c with its filter deep-copied to avoid mutations.
func (c *extendedCollection[T]) clone() *extendedCollection[T] {
	next := *c
	next.filter = copyFilter(c.filter)
	next.ctx = normalizeCtx(c.ctx)
	return &next
}

// By adds an equality condition on field (resolved via bson tag or snake_case) and returns a new builder.
func (c *extendedCollection[T]) By(field string, value any) ExtendedCollection[T] {
	mongoField := c.getMongoFieldName(field)
	next := c.clone()
	next.filter[mongoField] = value
	return next
}

// Where merges filter into the current accumulated filter and returns a new builder.
func (c *extendedCollection[T]) Where(filter bson.M) ExtendedCollection[T] {
	next := c.clone()
	for k, v := range filter {
		next.filter[k] = v
	}
	return next
}

// First decodes the first document matching the accumulated filter into result.
func (c *extendedCollection[T]) First(result *T) error {
	return c.getReadColl().FindOne(c.ctx, c.filter).Decode(result)
}

// Many decodes all documents matching the accumulated filter into results.
func (c *extendedCollection[T]) Many(results *[]T) error {
	cur, err := c.getReadColl().Find(c.ctx, c.filter)
	if err != nil {
		return err
	}
	defer cur.Close(c.ctx)
	return cur.All(c.ctx, results)
}

// Save performs a single-document update using the accumulated filter.
// Automatically injects updated_at into the $set clause.
func (c *extendedCollection[T]) Save(doc any, opts ...*options.UpdateOptions) error {
	_, err := c.write.UpdateOne(c.ctx, c.filter, ensureUpdateHasTimestamp(doc), opts...)
	return err
}

// SaveMany performs a multi-document update using the accumulated filter.
// Automatically injects updated_at into the $set clause.
func (c *extendedCollection[T]) SaveMany(update any, opts ...*options.UpdateOptions) error {
	_, err := c.write.UpdateMany(c.ctx, c.filter, ensureUpdateHasTimestamp(update), opts...)
	return err
}

// Delete deletes the first document matching the accumulated filter. Alias for Del.
func (c *extendedCollection[T]) Delete(opts ...*options.DeleteOptions) error {
	_, err := c.write.DeleteOne(c.ctx, c.filter, opts...)
	return err
}

// Del deletes the first document matching the accumulated filter.
func (c *extendedCollection[T]) Del(opts ...*options.DeleteOptions) error {
	_, err := c.write.DeleteOne(c.ctx, c.filter, opts...)
	return err
}

// DelMany deletes all documents matching the accumulated filter.
func (c *extendedCollection[T]) DelMany(opts ...*options.DeleteOptions) error {
	_, err := c.write.DeleteMany(c.ctx, c.filter, opts...)
	return err
}

// DeleteMany deletes all documents matching the accumulated filter. Alias for DelMany.
func (c *extendedCollection[T]) DeleteMany(opts ...*options.DeleteOptions) error {
	_, err := c.write.DeleteMany(c.ctx, c.filter, opts...)
	return err
}

// Count returns the number of documents matching the accumulated filter.
func (c *extendedCollection[T]) Count() (int64, error) {
	count, err := c.getReadColl().CountDocuments(c.ctx, c.filter)
	return count, err
}

// Exists returns true when at least one document matches the accumulated filter.
func (c *extendedCollection[T]) Exists() (bool, error) {
	count, err := c.Count()
	return count > 0, err
}

func (c *extendedCollection[T]) getReadColl() *mg.Collection {
	if mg.SessionFromContext(c.ctx) != nil {
		return c.write
	}
	return c.read
}

// GetFilter returns the accumulated BSON filter as built by By/Where calls.
func (c *extendedCollection[T]) GetFilter() any {
	return c.filter
}

// Ctx returns a new ExtendedCollection instance with the given context.
func (c *extendedCollection[T]) Ctx(ctx context.Context) ExtendedCollection[T] {
	next := c.clone()
	next.ctx = normalizeCtx(ctx)
	return next
}

type fieldCacheKey struct {
	typ   reflect.Type
	field string
}

var fieldNameCache sync.Map // key: fieldCacheKey, value: string

// getMongoFieldName resolves fieldName to a BSON key for type T.
// Results are cached in fieldNameCache to avoid repeated reflection.
func (c *extendedCollection[T]) getMongoFieldName(fieldName string) string {
	var zero T
	t := reflect.TypeOf(zero)
	if t == nil {
		return toSnakeCase(fieldName)
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	key := fieldCacheKey{t, fieldName}
	if cached, ok := fieldNameCache.Load(key); ok {
		return cached.(string)
	}

	result := computeMongoFieldName(t, fieldName)
	fieldNameCache.Store(key, result)
	return result
}

// computeMongoFieldName walks the fields of t looking for fieldName.
// Returns the value of the first bson tag part when found, otherwise the snake_case name.
func computeMongoFieldName(t reflect.Type, fieldName string) string {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name == fieldName {
			if bsonTag := field.Tag.Get("bson"); bsonTag != "" {
				parts := strings.Split(bsonTag, ",")
				if parts[0] != "" && parts[0] != "-" {
					return parts[0]
				}
			}
			return toSnakeCase(field.Name)
		}
	}
	return toSnakeCase(fieldName)
}

//#endregion
