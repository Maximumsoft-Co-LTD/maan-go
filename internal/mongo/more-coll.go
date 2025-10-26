package mongo

import (
	"context"
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	mg "go.mongodb.org/mongo-driver/mongo"
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

func (c *extendedCollection[T]) clone() *extendedCollection[T] {
	next := *c
	next.filter = copyFilter(c.filter)
	next.ctx = normalizeCtx(c.ctx)
	return &next
}

// Dynamic query methods
// By is a helper method to set a field value in the filter.
// Example:
// builder.By("Name", "foo") will set the filter to {"name": "foo"}
func (c *extendedCollection[T]) By(field string, value any) ExtendedCollection[T] {
	mongoField := c.getMongoFieldName(field)
	next := c.clone()
	next.filter[mongoField] = value
	return next
}

// Where is a helper method to set a filter in the query.
// Example:
// builder.Where(bson.M{"name": "foo"}) will set the filter to {"name": "foo"}
func (c *extendedCollection[T]) Where(filter bson.M) ExtendedCollection[T] {
	next := c.clone()
	for k, v := range filter {
		next.filter[k] = v
	}
	return next
}

// First is a helper method to find the first document in the collection.
// Example:
// var result testDoc
//
//	if err := builder.First(&result); err != nil {
//		return err
//	}
//
// First will return the first document in the collection that matches the filter.
func (c *extendedCollection[T]) First(result *T) error {
	return c.read.FindOne(c.ctx, c.filter).Decode(result)
}

// Many is a helper method to find many documents in the collection.
// Example:
// var results []testDoc
//
//	if err := builder.Many(&results); err != nil {
//		return err
//	}
//
// Many will return all documents in the collection that matches the filter.
func (c *extendedCollection[T]) Many(results *[]T) error {
	cur, err := c.read.Find(c.ctx, c.filter)
	if err != nil {
		return err
	}
	defer cur.Close(c.ctx)
	return cur.All(c.ctx, results)
}

// Save is a helper method to save a document to the collection.
// Example:
//
//	if err := builder.Save(bson.M{"name": "foo"}); err != nil {
//		return err
//	}
//
// Save will update the document in the collection that matches the filter.
func (c *extendedCollection[T]) Save(doc any) error {
	_, err := c.write.UpdateOne(c.ctx, c.filter, doc)
	return err
}

// SaveMany is a helper method to save many documents to the collection.
// Example:
//
//	if err := builder.SaveMany(bson.M{"name": "foo"}); err != nil {
//		return err
//	}
//
// SaveMany will update many documents in the collection that matches the filter.
func (c *extendedCollection[T]) SaveMany(update any) error {
	_, err := c.write.UpdateMany(c.ctx, c.filter, update)
	return err
}

// Delete is a helper method to delete a document from the collection.
// Example:
//
//	if err := builder.Delete(); err != nil {
//		return err
//	}
//
// Delete will delete the document in the collection that matches the filter.
func (c *extendedCollection[T]) Delete() error {
	_, err := c.write.DeleteOne(c.ctx, c.filter)
	return err
}

// Count is a helper method to count the number of documents in the collection.
// Example:
//
//	count, err := builder.Count(); err != nil {
//		return err
//	}
//
// Count will return the number of documents in the collection that matches the filter.
func (c *extendedCollection[T]) Count() (int64, error) {
	var count int64
	count, err := c.read.CountDocuments(c.ctx, c.filter)
	return count, err
}

// Exists is a helper method to check if a document exists in the collection.
// Example:
//
//	exists, err := builder.Exists(); err != nil {
//		return err
//	}
//
// Exists will return true if a document exists in the collection that matches the filter.
func (c *extendedCollection[T]) Exists() (bool, error) {
	count, err := c.Count()
	return count > 0, err
}

// GetFilter is a helper method to get the filter in the query.
// Example:
// filter := builder.GetFilter();
// GetFilter will return the filter in the query.
func (c *extendedCollection[T]) GetFilter() any {
	return c.filter
}

func (c *extendedCollection[T]) getMongoFieldName(fieldName string) string {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

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
