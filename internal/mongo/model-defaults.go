package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// defaultable is implemented by document structs that want the library to auto-populate
// ID, created_at, and updated_at fields on insert.
// The methods are responsible for populating the respective fields on the receiver.
type defaultable interface {
	DefaultId() primitive.ObjectID
	DefaultCreatedAt() time.Time
	DefaultUpdatedAt() time.Time
}

// applyModelDefaults calls the defaultable hooks on doc when doc implements the interface.
// It is a no-op when doc is nil or does not implement defaultable.
func applyModelDefaults(doc any) {
	if doc == nil {
		return
	}
	if d, ok := doc.(defaultable); ok {
		d.DefaultId()
		d.DefaultCreatedAt()
		d.DefaultUpdatedAt()
	}
}
