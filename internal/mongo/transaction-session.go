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
	client     Client
	read       *mg.Collection
	write      *mg.Collection
	once       sync.Once
}

func NewTransactionSession(ctx context.Context, client *mg.Client, read, write *mg.Collection) (TxSession, error) {
	s, err := client.StartSession()
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
		read:       read,
		write:      write,
	}, nil
}

func (s *transactionSession) Ctx() context.Context { return s.sessionCtx }

func (s *transactionSession) Close(errp *error) {
	defer func() {
		if p := recover(); p != nil {
			if s.session != nil {
				_ = s.session.AbortTransaction(s.sessionCtx)
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
				_ = s.session.AbortTransaction(s.sessionCtx)
				if *errp == nil {
					*errp = fmt.Errorf("commit failed: %w", err)
				}
			}
		} else {
			_ = s.session.AbortTransaction(s.sessionCtx)
		}
		s.session.EndSession(s.ctx)
		s.session, s.sessionCtx, s.ctx = nil, nil, nil
	})
}
