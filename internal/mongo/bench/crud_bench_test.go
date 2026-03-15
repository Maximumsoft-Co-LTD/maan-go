// FILE: internal/mongo/bench/crud_bench_test.go
// PACKAGE: mongo_bench_test
// PURPOSE: Benchmarks for CRUD operations.

package mongo_bench_test

import (
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func BenchmarkCreate(b *testing.B) {
	client := connectBenchClient(b)
	coll, _ := newBenchColl[integDoc](b, client, "bench_create")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc := &integDoc{Name: fmt.Sprintf("bench_%d", i), Value: i}
		if err := coll.Create(doc); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateMany_10(b *testing.B) {
	client := connectBenchClient(b)
	coll, _ := newBenchColl[integDoc](b, client, "bench_create_many_10")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		docs := make([]integDoc, 10)
		for j := range docs {
			docs[j] = integDoc{Name: fmt.Sprintf("batch_%d_%d", i, j), Value: j}
		}
		if err := coll.CreateMany(&docs); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateMany_100(b *testing.B) {
	client := connectBenchClient(b)
	coll, _ := newBenchColl[integDoc](b, client, "bench_create_many_100")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		docs := make([]integDoc, 100)
		for j := range docs {
			docs[j] = integDoc{Name: fmt.Sprintf("batch_%d_%d", i, j), Value: j}
		}
		if err := coll.CreateMany(&docs); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFindOne(b *testing.B) {
	client := connectBenchClient(b)
	coll, _ := newBenchColl[integDoc](b, client, "bench_find_one")
	seedDocs(b, coll, 1000)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var doc integDoc
		if err := coll.FindOne(bson.M{"name": "bench_0"}).Result(&doc); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFindMany_All(b *testing.B) {
	client := connectBenchClient(b)
	coll, _ := newBenchColl[integDoc](b, client, "bench_find_many_all")
	seedDocs(b, coll, 100)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := coll.FindMany(bson.M{}).All()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFindMany_WithFilter(b *testing.B) {
	client := connectBenchClient(b)
	coll, _ := newBenchColl[integDoc](b, client, "bench_find_many_filter")
	seedDocs(b, coll, 1000)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := coll.FindMany(bson.M{"active": true}).All()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSave(b *testing.B) {
	client := connectBenchClient(b)
	coll, _ := newBenchColl[integDoc](b, client, "bench_save")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := fmt.Sprintf("save_%d", i)
		err := coll.Save(bson.M{"name": name}, bson.M{"$set": bson.M{"name": name, "value": i}})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpd(b *testing.B) {
	client := connectBenchClient(b)
	coll, _ := newBenchColl[integDoc](b, client, "bench_upd")
	seedDocs(b, coll, 1)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := coll.Upd(bson.M{"name": "bench_0"}, bson.M{"$set": bson.M{"value": i}})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDel(b *testing.B) {
	client := connectBenchClient(b)
	coll, _ := newBenchColl[integDoc](b, client, "bench_del")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc := &integDoc{Name: fmt.Sprintf("del_%d", i), Value: i}
		_ = coll.Create(doc)
		if err := coll.Del(bson.M{"_id": doc.ID}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFindOneAndUpd(b *testing.B) {
	client := connectBenchClient(b)
	coll, _ := newBenchColl[integDoc](b, client, "bench_fau")
	seedDocs(b, coll, 1)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var out integDoc
		err := coll.FindOneAndUpd(
			bson.M{"name": "bench_0"},
			bson.M{"$set": bson.M{"value": i}},
			&out,
			options.FindOneAndUpdate().SetReturnDocument(options.After),
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFindOneAndDel(b *testing.B) {
	client := connectBenchClient(b)
	coll, _ := newBenchColl[integDoc](b, client, "bench_fad")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc := &integDoc{Name: fmt.Sprintf("fad_%d", i), Value: i}
		_ = coll.Create(doc)
		var out integDoc
		err := coll.FindOneAndDel(bson.M{"_id": doc.ID}, &out)
		if err != nil {
			b.Fatal(err)
		}
	}
}
