// FILE: internal/mongo/integration/transaction_integration_test.go
// PACKAGE: mongo_integration_test
// PURPOSE: Behavioral correctness tests for WithTx (automatic) and StartTx (manual) transaction handling.
// NOTE: Transactions require a MongoDB replica set.

package mongo_integration_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestWithTx_CommitOnSuccess(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "tx_commit")

	err := coll.WithTx(func(ctx context.Context) error {
		txColl := coll.Ctx(ctx)
		return txColl.Create(&integDoc{Name: "in-tx", Value: 1})
	})
	if err != nil {
		t.Fatalf("WithTx failed: %v", err)
	}

	count, err := coll.Count(bson.M{"name": "in-tx"})
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count 1 (committed), got %d", count)
	}
}

func TestWithTx_RollbackOnError(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "tx_rollback")

	err := coll.WithTx(func(ctx context.Context) error {
		txColl := coll.Ctx(ctx)
		if err := txColl.Create(&integDoc{Name: "should-not-persist"}); err != nil {
			return err
		}
		return errors.New("intentional rollback")
	})
	if err == nil {
		t.Fatal("expected error from WithTx, got nil")
	}

	count, err := coll.Count(bson.M{"name": "should-not-persist"})
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected count 0 (rolled back), got %d", count)
	}
}

func TestWithTx_RollbackOnPanic(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "tx_panic")

	err := coll.WithTx(func(ctx context.Context) error {
		txColl := coll.Ctx(ctx)
		_ = txColl.Create(&integDoc{Name: "should-not-persist-panic"})
		panic("boom")
	})
	if err == nil {
		t.Fatal("expected error from WithTx after panic, got nil")
	}
	if !strings.Contains(err.Error(), "transaction panic") {
		t.Fatalf("expected error to contain 'transaction panic', got %q", err.Error())
	}

	count, err := coll.Count(bson.M{"name": "should-not-persist-panic"})
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected count 0 (rolled back), got %d", count)
	}
}

func TestWithTx_NilFn(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "tx_nil_fn")

	err := coll.WithTx(nil)
	if err == nil {
		t.Fatal("expected error for nil fn, got nil")
	}
	if err.Error() != "fn must not be nil" {
		t.Fatalf("expected 'fn must not be nil', got %q", err.Error())
	}
}

func TestStartTx_ManualCommit(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "tx_manual_commit")

	tx, err := coll.StartTx()
	if err != nil {
		t.Fatalf("StartTx failed: %v", err)
	}

	txColl := coll.Ctx(tx.Ctx())
	err = txColl.Create(&integDoc{Name: "manual-commit"})
	if err != nil {
		t.Fatalf("Create in tx failed: %v", err)
	}

	var commitErr error
	tx.Close(&commitErr)
	if commitErr != nil {
		t.Fatalf("tx.Close commit failed: %v", commitErr)
	}

	count, err := coll.Count(bson.M{"name": "manual-commit"})
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}
}

func TestStartTx_ManualRollback(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "tx_manual_rollback")

	tx, err := coll.StartTx()
	if err != nil {
		t.Fatalf("StartTx failed: %v", err)
	}

	txColl := coll.Ctx(tx.Ctx())
	_ = txColl.Create(&integDoc{Name: "manual-rollback"})

	rollbackErr := errors.New("force rollback")
	tx.Close(&rollbackErr)

	count, err := coll.Count(bson.M{"name": "manual-rollback"})
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected count 0 (rolled back), got %d", count)
	}
}
