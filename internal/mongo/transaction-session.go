package mongo

import (
	"context"
	"sync"

	mg "go.mongodb.org/mongo-driver/mongo"
)

type transaction struct {
	session    mg.Session
	sessionCtx mg.SessionContext
	ctx        context.Context
	client     Client
	read       *mg.Collection
	write      *mg.Collection
	endOnce    sync.Once
}

func newTransaction(ctx context.Context, m Client, read *mg.Collection, write *mg.Collection) (TxSession, error) {
	session, err := m.Write().StartSession()
	if err != nil {
		return nil, err
	}
	if err := session.StartTransaction(); err != nil {
		session.EndSession(ctx)
		return nil, err
	}
	sessionCtx := mg.NewSessionContext(ctx, session)
	return &transaction{
		session:    session,
		sessionCtx: sessionCtx,
		ctx:        ctx,
		client:     m,
		read:       read,
		write:      write,
	}, nil
}

var _ TxSession = (*transaction)(nil)

func (tx *transaction) SessionCtx() context.Context {
	return tx.sessionCtx
}

func (tx *transaction) Commit() error {
	defer tx.end()
	if err := tx.session.CommitTransaction(tx.sessionCtx); err != nil {
		_ = tx.session.AbortTransaction(tx.sessionCtx)
		return err
	}
	return nil
}

func (tx *transaction) Rollback(err error) error {
	tx.session.AbortTransaction(tx.sessionCtx)
	tx.end()
	return err
}

func (tx *transaction) end() {
	tx.endOnce.Do(func() {
		tx.session.EndSession(tx.ctx)
	})
	tx.client = nil
	tx.read = nil
	tx.write = nil
	tx.session = nil
	tx.sessionCtx = nil
	tx.ctx = nil
	tx.endOnce = sync.Once{}
}
