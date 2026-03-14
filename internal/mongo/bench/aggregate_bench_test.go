// FILE: internal/mongo/bench/aggregate_bench_test.go
// PACKAGE: mongo_bench_test
// PURPOSE: Benchmarks for aggregation pipeline operations.

package mongo_bench_test

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func BenchmarkAgg_SimpleMatch(b *testing.B) {
	client := connectBenchClient(b)
	coll, _ := newBenchColl[integDoc](b, client, "bench_agg_match")
	seedDocs(b, coll, 1000)

	pipeline := bson.A{bson.M{"$match": bson.M{"active": true}}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := coll.Agg(pipeline).All()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAgg_GroupCount(b *testing.B) {
	client := connectBenchClient(b)
	coll, _ := newBenchColl[integDoc](b, client, "bench_agg_group")
	seedDocs(b, coll, 1000)

	pipeline := bson.A{
		bson.M{"$group": bson.M{"_id": "$active", "total": bson.M{"$sum": 1}}},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := coll.Agg(pipeline).Raw()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAgg_Stream(b *testing.B) {
	client := connectBenchClient(b)
	coll, _ := newBenchColl[integDoc](b, client, "bench_agg_stream")
	seedDocs(b, coll, 100)

	pipeline := bson.A{bson.M{"$match": bson.M{}}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := coll.Agg(pipeline).Stream(func(ctx context.Context, doc integDoc) error {
			return nil
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}
