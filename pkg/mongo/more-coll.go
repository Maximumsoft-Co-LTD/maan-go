package mongo

import (
	"context"
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	mg "go.mongodb.org/mongo-driver/mongo"
)

//#region morecoll

type extendedCollection[T any] struct {
	ctx     context.Context
	connect Client
	read    *mg.Collection
	write   *mg.Collection
	name    string
	filter  bson.M
}

var _ ExtendedCollection[any] = (*extendedCollection[any])(nil)

// NewExtendedCollection creates a chainable query builder bound to the provided Mongo clients.
func NewExtendedCollection[T any](ctx context.Context, connect Client, read *mg.Collection, write *mg.Collection, name string) ExtendedCollection[T] {
	return &extendedCollection[T]{ctx: ctx, connect: connect, read: read, write: write, name: name, filter: bson.M{}}
}

// Dynamic query methods
func (c *extendedCollection[T]) By(field string, value any) ExtendedCollection[T] {
	mongoField := c.getMongoFieldName(field)
	newFilter := copyFilter(c.filter)
	newFilter[mongoField] = value
	c.filter = newFilter
	return c
}

func (c *extendedCollection[T]) Where(filter bson.M) ExtendedCollection[T] {
	newFilter := copyFilter(c.filter)
	for k, v := range filter {
		newFilter[k] = v
	}
	c.filter = newFilter
	return c
}

func (c *extendedCollection[T]) First(result *T) error {
	return c.read.FindOne(c.ctx, c.filter).Decode(result)
}

func (c *extendedCollection[T]) Many(results *[]T) error {
	cur, err := c.read.Find(c.ctx, c.filter)
	if err != nil {
		return err
	}
	defer cur.Close(c.ctx)
	return cur.All(c.ctx, results)
}

func (c *extendedCollection[T]) Save(doc any) error {
	_, err := c.write.UpdateOne(c.ctx, c.filter, doc)
	return err
}

func (c *extendedCollection[T]) SaveMany(update any) error {
	_, err := c.write.UpdateMany(c.ctx, c.filter, update)
	return err
}

func (c *extendedCollection[T]) Delete() error {
	_, err := c.write.DeleteOne(c.ctx, c.filter)
	return err
}

func (c *extendedCollection[T]) Count() (int64, error) {
	var count int64
	count, err := c.read.CountDocuments(c.ctx, c.filter)
	return count, err
}

func (c *extendedCollection[T]) Exists() (bool, error) {
	count, err := c.Count()
	return count > 0, err
}

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
