// FILE: internal/mongo/integration/change_stream_integration_test.go
// PACKAGE: mongo_integration_test
// PURPOSE: Behavioral correctness tests for ChangeStream (Watch) operations.
// NOTE: Requires MongoDB replica set.

package mongo_integration_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo"

	"go.mongodb.org/mongo-driver/bson"
)

func TestChangeStream_ReceivesInsertEvent(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "cs_insert")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var received []mongo.ChangeEvent[integDoc]
	var mu sync.Mutex

	errCh := make(chan error, 1)
	go func() {
		errCh <- coll.Ctx(ctx).Watch().OnIst().Stream(func(st mongo.CsEvt[integDoc]) error {
			mu.Lock()
			received = append(received, st.ChangeEvent)
			mu.Unlock()
			cancel()
			return nil
		})
	}()

	time.Sleep(500 * time.Millisecond)
	if err := coll.Create(&integDoc{Name: "stream-insert"}); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err := <-errCh
	if err != nil && err != context.Canceled {
		t.Fatalf("Stream error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(received) < 1 {
		t.Fatal("expected at least 1 event")
	}
	if received[0].OperationType != "insert" {
		t.Fatalf("expected 'insert', got %q", received[0].OperationType)
	}
	if received[0].FullDocument == nil {
		t.Fatal("expected non-nil FullDocument")
	}
	if received[0].FullDocument.Name != "stream-insert" {
		t.Fatalf("expected Name 'stream-insert', got %q", received[0].FullDocument.Name)
	}
}

func TestChangeStream_ReceivesUpdateEvent(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "cs_update")

	doc := integDoc{Name: "cs_upd_target", Value: 1}
	if err := coll.Create(&doc); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var received []mongo.ChangeEvent[integDoc]
	var mu sync.Mutex

	errCh := make(chan error, 1)
	go func() {
		errCh <- coll.Ctx(ctx).Watch().OnUpd().UpdLookup().Stream(func(st mongo.CsEvt[integDoc]) error {
			mu.Lock()
			received = append(received, st.ChangeEvent)
			mu.Unlock()
			cancel()
			return nil
		})
	}()

	time.Sleep(500 * time.Millisecond)
	if err := coll.Upd(bson.M{"name": doc.Name}, bson.M{"$set": bson.M{"value": 999}}); err != nil {
		t.Fatalf("Upd failed: %v", err)
	}

	err := <-errCh
	if err != nil && err != context.Canceled {
		t.Fatalf("Stream error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(received) < 1 {
		t.Fatal("expected at least 1 event")
	}
	if received[0].OperationType != "update" {
		t.Fatalf("expected 'update', got %q", received[0].OperationType)
	}
	if received[0].FullDocument == nil {
		t.Fatal("expected non-nil FullDocument (UpdLookup)")
	}
}

func TestChangeStream_ReceivesDeleteEvent(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "cs_delete")

	doc := integDoc{Name: "cs_del_target", Value: 1}
	if err := coll.Create(&doc); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var received []mongo.ChangeEvent[integDoc]
	var mu sync.Mutex

	errCh := make(chan error, 1)
	go func() {
		errCh <- coll.Ctx(ctx).Watch().OnDel().Stream(func(st mongo.CsEvt[integDoc]) error {
			mu.Lock()
			received = append(received, st.ChangeEvent)
			mu.Unlock()
			cancel()
			return nil
		})
	}()

	time.Sleep(500 * time.Millisecond)
	if err := coll.Del(bson.M{"_id": doc.ID}); err != nil {
		t.Fatalf("Del failed: %v", err)
	}

	err := <-errCh
	if err != nil && err != context.Canceled {
		t.Fatalf("Stream error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(received) < 1 {
		t.Fatal("expected at least 1 event")
	}
	if received[0].OperationType != "delete" {
		t.Fatalf("expected 'delete', got %q", received[0].OperationType)
	}
	if received[0].DocumentKey == nil {
		t.Fatal("expected non-nil DocumentKey")
	}
}

func TestChangeStream_FiltersByOpType(t *testing.T) {
	t.Parallel()
	client := connectTestClient(t)
	coll, _ := newColl[integDoc](t, client, "cs_filter_op")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var received []mongo.ChangeEvent[integDoc]
	var mu sync.Mutex

	errCh := make(chan error, 1)
	go func() {
		errCh <- coll.Ctx(ctx).Watch().OnIst().Stream(func(st mongo.CsEvt[integDoc]) error {
			mu.Lock()
			received = append(received, st.ChangeEvent)
			if len(received) >= 2 {
				mu.Unlock()
				cancel()
				return nil
			}
			mu.Unlock()
			return nil
		})
	}()

	time.Sleep(500 * time.Millisecond)

	// Create 2 docs (insert events)
	doc1 := integDoc{Name: "filter_a", Value: 1}
	doc2 := integDoc{Name: "filter_b", Value: 2}
	if err := coll.Create(&doc1); err != nil {
		t.Fatalf("Create doc1 failed: %v", err)
	}
	if err := coll.Create(&doc2); err != nil {
		t.Fatalf("Create doc2 failed: %v", err)
	}

	// Update one doc (update event -- should NOT be received)
	_ = coll.Upd(bson.M{"_id": doc1.ID}, bson.M{"$set": bson.M{"value": 999}})
	// Delete one doc (delete event -- should NOT be received)
	_ = coll.Del(bson.M{"_id": doc2.ID})

	err := <-errCh
	if err != nil && err != context.Canceled {
		t.Fatalf("Stream error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	for i, evt := range received {
		if evt.OperationType != "insert" {
			t.Fatalf("event[%d] expected 'insert', got %q", i, evt.OperationType)
		}
	}
}
