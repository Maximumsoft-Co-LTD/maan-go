package repository

import (
	"context"

	"potja-mongo/internal/entities"
	"potja-mongo/pkg/mongo"
)

const (
	bankConfigCollName = "bank_config"
)

var _ RepoReader = (*MongoRepo)(nil)

type MongoRepo struct {
	ctx        context.Context
	bankConfig mongo.Collection[entities.BankConfig]
}

type RepoReader interface {
	BankConfig(ctx context.Context) mongo.Collection[entities.BankConfig]
	Close() error
}

// NewMongoRepo constructs a repository reader backed by MongoDB collections.
func NewMongoRepo(ctx context.Context, mc mongo.Client) RepoReader {
	return &MongoRepo{ctx: ctx, bankConfig: mongo.NewCollection[entities.BankConfig](ctx, mc, bankConfigCollName)}
}

func (r *MongoRepo) getCtx() context.Context {
	if r.ctx == nil {
		r.ctx = context.Background()
	}
	return r.ctx
}

// BankRegister returns the bank register collection scoped to the provided context.
func (r *MongoRepo) BankConfig(ctx context.Context) mongo.Collection[entities.BankConfig] {
	r.ctx = ctx
	return r.bankConfig.Ctx(r.getCtx())
}

func (r *MongoRepo) Close() error {
	r.ctx = nil
	r.bankConfig = nil
	return nil
}
