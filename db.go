package maango

import (
	"context"
	"reflect"

	mg "github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo"
)

// collInit is an internal interface used by DB to initialize struct fields via reflection.
type collInit interface {
	initColl(ctx context.Context, client Client, name string)
}

// initColl initializes a *Coll[T] field with a new Collection.
func (c *Coll[T]) initColl(ctx context.Context, client Client, name string) {
	c.Collection = mg.NewCollection[T](ctx, client, name)
}

// initColl initializes a *ExColl[T] field with a new ExtendedCollection.
func (c *ExColl[T]) initColl(ctx context.Context, client Client, name string) {
	dbName := client.DbName()
	read := client.Read().Database(dbName).Collection(name)
	write := client.Write().Database(dbName).Collection(name)
	c.ExtendedCollection = mg.NewExtendedCollection[T](ctx, read, write, name)
}

// DB creates an instance of T and auto-initializes every field that carries a
// `collection_name` struct tag. Fields must be of type Coll[T] or ExColl[T].
//
// Example:
//
//	type MyDB struct {
//	    Users    maango.Coll[User]    `collection_name:"users"`
//	    Products maango.Coll[Product] `collection_name:"products"`
//	}
//
//	db := maango.DB[MyDB](ctx, client)
//	db.Users.Ctx(ctx).FindOne(bson.M{"_id": id})
func DB[T any](ctx context.Context, client Client) *T {
	result := new(T)
	val := reflect.ValueOf(result).Elem()
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		collName, ok := typ.Field(i).Tag.Lookup("collection_name")
		if !ok || collName == "" {
			continue
		}
		fieldVal := val.Field(i)
		if !fieldVal.CanAddr() {
			continue
		}
		if init, ok := fieldVal.Addr().Interface().(collInit); ok {
			init.initColl(ctx, client, collName)
		}
	}
	return result
}
