package mongo

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ctxKey string

type testDoc struct {
	ID     string `bson:"_id"`
	Name   string `bson:"name"`
	Active bool   `bson:"active"`
}

func TestCollectionCtxReturnsIsolatedCopies(t *testing.T) {
	client, err := NewFakeClient()
	if err != nil {
		t.Fatalf("fake client: %v", err)
	}
	base := NewCollection[testDoc](context.Background(), client, "docs").(*collection[testDoc])

	ctxA := context.WithValue(context.Background(), ctxKey("req"), "A")
	ctxB := context.WithValue(context.Background(), ctxKey("req"), "B")

	colA := base.Ctx(ctxA).(*collection[testDoc])
	colB := base.Ctx(ctxB).(*collection[testDoc])

	if base == colA || base == colB {
		t.Fatalf("expected Ctx() to return a new collection instance")
	}

	if got := colA.getCtx().Value(ctxKey("req")); got != "A" {
		t.Fatalf("colA context mismatch, want A got %v", got)
	}
	if got := colB.getCtx().Value(ctxKey("req")); got != "B" {
		t.Fatalf("colB context mismatch, want B got %v", got)
	}
	if base.getCtx().Value(ctxKey("req")) != nil {
		t.Fatalf("base context should remain unchanged")
	}
}

func TestExtendedCollectionFiltersAreIsolated(t *testing.T) {
	client, err := NewFakeClient()
	if err != nil {
		t.Fatalf("fake client: %v", err)
	}
	base := NewCollection[testDoc](context.Background(), client, "docs")

	builder := base.Build(context.Background())
	first := builder.By("Name", "foo")
	second := builder.By("Name", "bar").Where(bson.M{"active": true})

	if len(builder.GetFilter().(bson.M)) != 0 {
		t.Fatalf("expected base builder filter to remain empty")
	}
	if val := first.GetFilter().(bson.M)["name"]; val != "foo" {
		t.Fatalf("first filter mismatch, got %v", val)
	}
	secondFilter := second.GetFilter().(bson.M)
	if secondFilter["name"] != "bar" || secondFilter["active"] != true {
		t.Fatalf("second filter mismatch: %#v", secondFilter)
	}
}

func TestSingleBuildHonorsExtraOptions(t *testing.T) {
	client, err := NewFakeClient()
	if err != nil {
		t.Fatalf("fake client: %v", err)
	}
	coll := NewCollection[testDoc](context.Background(), client, "docs")
	collation := &options.Collation{Locale: "th"}
	maxTime := 5 * time.Second
	extra := options.FindOne().SetCollation(collation).SetMaxTime(maxTime)

	builder := coll.FindOne(bson.M{"name": "foo"}).
		Proj(bson.M{"name": 1}).
		Sort(bson.M{"name": 1}).
		Hint(bson.M{"name": 1}).
		Opts(extra)

	internal := builder.(*single[testDoc])
	opts := internal.build()

	if opts.Collation != collation {
		t.Fatalf("expected collation to be preserved")
	}
	if opts.MaxTime == nil || *opts.MaxTime != maxTime {
		t.Fatalf("maxTime not propagated")
	}
}

func TestManyBuildHonorsExtraOptions(t *testing.T) {
	client, err := NewFakeClient()
	if err != nil {
		t.Fatalf("fake client: %v", err)
	}
	coll := NewCollection[testDoc](context.Background(), client, "docs")
	comment := "demo"
	extra := options.Find().
		SetComment(comment).
		SetCollation(&options.Collation{Locale: "en"}).
		SetBatchSize(10)

	builder := coll.FindMany(bson.M{}).
		Limit(5).
		Skip(2).
		Bsz(3).
		Opts(extra)

	internal := builder.(*many[testDoc])
	opts := internal.build()

	if opts.Comment == nil || *opts.Comment != comment {
		t.Fatalf("expected comment to propagate")
	}
	if opts.Collation == nil || opts.Collation.Locale != "en" {
		t.Fatalf("expected collation to propagate")
	}
	if opts.BatchSize == nil || *opts.BatchSize != 10 {
		t.Fatalf("expected batch size from extra to win")
	}
	if opts.Limit == nil || *opts.Limit != 5 {
		t.Fatalf("limit not applied")
	}
}

func TestAggregateBuildHonorsExtraOptions(t *testing.T) {
	client, err := NewFakeClient()
	if err != nil {
		t.Fatalf("fake client: %v", err)
	}
	hint := bson.D{{Key: "name", Value: 1}}
	extra := options.Aggregate().SetComment("agg").SetHint(hint)
	dbName := client.DbName()
	builder := NewAgg[testDoc](context.Background(), client.Read().Database(dbName).Collection("docs"), "docs", bson.A{}).
		Disk(true).
		Bsz(7).
		Opts(extra)

	internal := builder.(*agg[testDoc])
	opts := internal.build()

	if opts.Comment == nil || *opts.Comment != "agg" {
		t.Fatalf("expected comment to propagate")
	}
	if opts.Hint == nil {
		t.Fatalf("expected hint to propagate")
	}
	if opts.AllowDiskUse == nil || *opts.AllowDiskUse != true {
		t.Fatalf("allowDiskUse not set")
	}
	if opts.BatchSize == nil || *opts.BatchSize != 7 {
		t.Fatalf("batch size not set")
	}
}

func TestEnsureUpdateHasTimestamp(t *testing.T) {
	nowUpdate := ensureUpdateHasTimestamp(bson.M{"name": "foo"})
	set := nowUpdate["$set"].(bson.M)
	if _, ok := set["updated_at"]; !ok {
		t.Fatalf("missing updated_at for plain doc")
	}

	updateWithSet := ensureUpdateHasTimestamp(bson.M{"$set": bson.M{"bar": "baz"}})
	set = updateWithSet["$set"].(bson.M)
	if set["bar"] != "baz" {
		t.Fatalf("existing $set data lost")
	}
	if _, ok := set["updated_at"]; !ok {
		t.Fatalf("missing updated_at in $set path")
	}
}
