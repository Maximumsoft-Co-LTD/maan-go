package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	mg "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type indexManager struct {
	ctx  context.Context
	view mg.IndexView
}

var _ IndexManager = (*indexManager)(nil)

// CreateOne creates a single index on the collection and returns the index name.
func (m *indexManager) CreateOne(model mg.IndexModel, opts ...*options.CreateIndexesOptions) (string, error) {
	return m.view.CreateOne(m.ctx, model, opts...)
}

// CreateMany creates multiple indexes on the collection and returns their names.
func (m *indexManager) CreateMany(models []mg.IndexModel, opts ...*options.CreateIndexesOptions) ([]string, error) {
	return m.view.CreateMany(m.ctx, models, opts...)
}

// DropOne drops the index with the given name.
func (m *indexManager) DropOne(name string, opts ...*options.DropIndexesOptions) error {
	_, err := m.view.DropOne(m.ctx, name, opts...)
	return err
}

// DropAll drops all non-_id indexes on the collection.
func (m *indexManager) DropAll(opts ...*options.DropIndexesOptions) error {
	_, err := m.view.DropAll(m.ctx, opts...)
	return err
}

// List returns all indexes on the collection as a slice of bson.M documents.
func (m *indexManager) List(opts ...*options.ListIndexesOptions) ([]bson.M, error) {
	cur, err := m.view.List(m.ctx, opts...)
	if err != nil {
		return nil, err
	}
	var result []bson.M
	return result, cur.All(m.ctx, &result)
}
