// FILE: internal/mongo/bench/extended_collection_bench_test.go
// PACKAGE: mongo_bench_test
// PURPOSE: Benchmarks for ExtendedCollection (Build, By, Where, First, Many, Count).

package mongo_bench_test

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func BenchmarkExtColl_ByFirst(b *testing.B) {
	client := connectBenchClient(b)
	coll, _ := newBenchColl[integDoc](b, client, "bench_ext_by_first")
	seedDocs(b, coll, 1000)

	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var doc integDoc
		err := coll.Build(ctx).By("Name", "bench_0").First(&doc)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExtColl_WhereMany(b *testing.B) {
	client := connectBenchClient(b)
	coll, _ := newBenchColl[integDoc](b, client, "bench_ext_where_many")
	seedDocs(b, coll, 100)

	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var docs []integDoc
		err := coll.Build(ctx).Where(bson.M{"active": true}).Many(&docs)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExtColl_Count(b *testing.B) {
	client := connectBenchClient(b)
	coll, _ := newBenchColl[integDoc](b, client, "bench_ext_count")
	seedDocs(b, coll, 1000)

	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := coll.Build(ctx).Count()
		if err != nil {
			b.Fatal(err)
		}
	}
}
