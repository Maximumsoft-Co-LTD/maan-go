package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type defaultable interface {
	DefaultId() primitive.ObjectID
	DefaultCreatedAt() time.Time
	DefaultUpdatedAt() time.Time
}

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
