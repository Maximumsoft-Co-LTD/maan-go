package mongo

import (
	"context"
	"fmt"
	"sync"

	mg "go.mongodb.org/mongo-driver/mongo"
)

type transactionSession struct {
	session    mg.Session
	sessionCtx mg.SessionContext
	ctx        context.Context
	once       sync.Once
}

// NewTransactionSession is a helper method to start a new transaction session.
// Example:
// tx, err := NewTransactionSession(context.Background(), client)
// tx.Ctx()
// NewTransactionSession will return a new transaction session instance with the context set.
// The transaction session is isolated and will not affect the original collection.
func NewTransactionSession(ctx context.Context, client Client) (TxSession, error) {
	s, err := client.Write().StartSession()
	if err != nil {
		return nil, err
	}
	if err := s.StartTransaction(); err != nil {
		s.EndSession(ctx)
		return nil, err
	}
	return &transactionSession{
		session:    s,
		sessionCtx: mg.NewSessionContext(ctx, s),
		ctx:        ctx,
	}, nil
}

// Ctx is a helper method to get the context of the transaction session.
// Example:
// ctx := tx.Ctx()
// Ctx will return the context of the transaction session with the context set.
// The context is isolated and will not affect the original collection.
// The context is the session context.
func (s *transactionSession) Ctx() context.Context { return s.sessionCtx }

// Close is a helper method to close the transaction session.
// Example:
// err := tx.Close(&err)
// Close will return a new error with the context set.
// The error is isolated and will not affect the original collection.
// The error is committed if *errp is nil, otherwise it is aborted.
// The transaction session is closed and the context is reset.
func (s *transactionSession) Close(errp *error) {
	defer func() {
		if p := recover(); p != nil {
			if s.session != nil {
				s.session.AbortTransaction(s.sessionCtx)
				s.session.EndSession(s.ctx)
			}
			s.session, s.sessionCtx, s.ctx = nil, nil, nil
			panic(p)
		}
	}()

	s.once.Do(func() {
		if s.session == nil {
			return
		}
		if errp != nil && *errp == nil {
			if err := s.session.CommitTransaction(s.sessionCtx); err != nil {
				s.session.AbortTransaction(s.sessionCtx)
				if *errp == nil {
					*errp = fmt.Errorf("commit failed: %w", err)
				}
			}
		} else {
			s.session.AbortTransaction(s.sessionCtx)
		}
		s.session.EndSession(s.ctx)
		s.session, s.sessionCtx, s.ctx = nil, nil, nil
	})
}
