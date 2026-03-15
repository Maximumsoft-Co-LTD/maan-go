//go:build load_test

// FILE: internal/mongo/bench/load_test.go
// PACKAGE: mongo_bench_test
// PURPOSE: Concurrency / load tests. Run with: go test -tags load_test -race -timeout 120s ./internal/mongo/bench/

package mongo_bench_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func TestLoad_ConcurrentCreate_100(t *testing.T) {
	client := connectLoadClient(t)
	coll, _ := newLoadColl[integDoc](t, client, "load_create")

	var wg sync.WaitGroup
	errCh := make(chan error, 100*100)

	wg.Add(100)
	for g := 0; g < 100; g++ {
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				doc := &integDoc{Name: fmt.Sprintf("g%d_d%d", gid, i), Value: i}
				if err := coll.Create(doc); err != nil {
					errCh <- err
				}
			}
		}(g)
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatalf("concurrent Create error: %v", err)
	}

	count, err := coll.Count(nil)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 10000 {
		t.Fatalf("expected count 10000, got %d", count)
	}
}

func TestLoad_ConcurrentFind_100(t *testing.T) {
	client := connectLoadClient(t)
	coll, _ := newLoadColl[integDoc](t, client, "load_find")
	seedDocsT(t, coll, 1000)

	var wg sync.WaitGroup
	errCh := make(chan error, 100)

	for g := 0; g < 100; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			var doc integDoc
			name := fmt.Sprintf("load_%d", gid%1000)
			if err := coll.FindOne(bson.M{"name": name}).Result(&doc); err != nil {
				errCh <- err
			}
		}(g)
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatalf("concurrent Find error: %v", err)
	}
}

func TestLoad_ConcurrentMixedOps(t *testing.T) {
	client := connectLoadClient(t)
	coll, _ := newLoadColl[integDoc](t, client, "load_mixed")
	seedDocsT(t, coll, 100)

	var wg sync.WaitGroup
	for g := 0; g < 50; g++ {
		wg.Add(2)
		// Writer
		go func(gid int) {
			defer wg.Done()
			doc := &integDoc{Name: fmt.Sprintf("writer_%d", gid), Value: gid}
			_ = coll.Create(doc)
		}(g)
		// Reader
		go func(gid int) {
			defer wg.Done()
			_, _ = coll.FindMany(bson.M{}).Limit(10).All()
		}(g)
	}
	wg.Wait()
}

func TestLoad_TransactionContention(t *testing.T) {
	client := connectLoadClient(t)
	coll, _ := newLoadColl[integDoc](t, client, "load_tx")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			_ = coll.Ctx(ctx).WithTx(func(txCtx context.Context) error {
				txColl := coll.Ctx(txCtx)
				doc := &integDoc{Name: fmt.Sprintf("tx_%d", gid), Value: gid}
				return txColl.Create(doc)
			})
		}(g)
	}
	wg.Wait()

	count, err := coll.Count(nil)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 10 {
		t.Fatalf("expected count 10, got %d", count)
	}
}
