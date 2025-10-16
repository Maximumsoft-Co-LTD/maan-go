package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BankConfig struct {
	ID        primitive.ObjectID `bson:"_id"`
	Name      string             `bson:"name"`
	Code      string             `bson:"code"`
	Logo      string             `bson:"logo"`
	Url       string             `bson:"url"`
	Status    string             `bson:"status"`
	IsDefault bool               `bson:"is_default"`
	IsActive  bool               `bson:"is_active"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

func (b *BankConfig) DefaultId() primitive.ObjectID {
	if b.ID.IsZero() {
		b.ID = primitive.NewObjectID()
	}
	return b.ID
}

func (b *BankConfig) DefaultCreatedAt() time.Time {
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now().UTC()
	}
	return b.CreatedAt
}

func (b *BankConfig) DefaultUpdatedAt() time.Time {
	if b.UpdatedAt.IsZero() {
		b.UpdatedAt = time.Now().UTC()
	}
	return b.UpdatedAt
}
