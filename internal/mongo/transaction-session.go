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

// NewTransactionSession starts a MongoDB session and immediately begins a transaction.
// The returned TxSession must be closed via tx.Close(&err) to commit or rollback.
// Normally called via Collection.StartTx rather than directly.
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

// Ctx returns the session-aware context that must be passed to all collection operations
// that should participate in this transaction.
func (s *transactionSession) Ctx() context.Context { return s.sessionCtx }

// Close commits the transaction when *errp == nil, otherwise aborts it.
// Always call via defer to guarantee cleanup:
//
//	tx, err := coll.StartTx()
//	if err != nil { return err }
//	defer tx.Close(&err)
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
				*errp = fmt.Errorf("commit failed: %w", err)
			}
		} else {
			s.session.AbortTransaction(s.sessionCtx)
		}
		s.session.EndSession(s.ctx)
		s.session, s.sessionCtx, s.ctx = nil, nil, nil
	})
}
