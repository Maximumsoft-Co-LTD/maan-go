package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo/testutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newTestChangeStream(t *testing.T) ChangeStream[testutil.TestDoc] {
	t.Helper()
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	testutil.AssertNoError(t, err, "Failed to create fake client")
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	coll := NewCollection[testutil.TestDoc](ctx, client, testutil.TestCollectionName)
	return coll.Watch(ctx)
}

// TestChangeStreamOpFilters tests OnIst / OnUpd / OnDel / OnRep / OnIstAndUpd builder methods.
func TestChangeStreamOpFilters(t *testing.T) {
	base := newTestChangeStream(t)

	t.Run("OnIst adds insert op", func(t *testing.T) {
		cs := base.OnIst().(*changeStream[testutil.TestDoc])
		testutil.AssertContains(t, cs.onOps, "insert", "OnIst should add 'insert'")
	})

	t.Run("OnUpd adds update op", func(t *testing.T) {
		cs := base.OnUpd().(*changeStream[testutil.TestDoc])
		testutil.AssertContains(t, cs.onOps, "update", "OnUpd should add 'update'")
	})

	t.Run("OnDel adds delete op", func(t *testing.T) {
		cs := base.OnDel().(*changeStream[testutil.TestDoc])
		testutil.AssertContains(t, cs.onOps, "delete", "OnDel should add 'delete'")
	})

	t.Run("OnRep adds replace op", func(t *testing.T) {
		cs := base.OnRep().(*changeStream[testutil.TestDoc])
		testutil.AssertContains(t, cs.onOps, "replace", "OnRep should add 'replace'")
	})

	t.Run("OnIstAndUpd adds insert and update ops", func(t *testing.T) {
		cs := base.OnIstAndUpd().(*changeStream[testutil.TestDoc])
		testutil.AssertContains(t, cs.onOps, "insert", "OnIstAndUpd should add 'insert'")
		testutil.AssertContains(t, cs.onOps, "update", "OnIstAndUpd should add 'update'")
	})

	t.Run("chaining ops accumulates all", func(t *testing.T) {
		cs := base.OnIst().OnDel().(*changeStream[testutil.TestDoc])
		testutil.AssertContains(t, cs.onOps, "insert", "Chained OnIst should have 'insert'")
		testutil.AssertContains(t, cs.onOps, "delete", "Chained OnDel should have 'delete'")
	})
}

// TestChangeStreamFullDocOptions tests FullDoc / UpdLookup / FullDocRequired.
func TestChangeStreamFullDocOptions(t *testing.T) {
	base := newTestChangeStream(t)

	t.Run("FullDoc sets fullDoc field", func(t *testing.T) {
		cs := base.FullDoc("whenAvailable").(*changeStream[testutil.TestDoc])
		testutil.AssertEqual(t, "whenAvailable", cs.fullDoc, "FullDoc should set fullDoc")
	})

	t.Run("UpdLookup sets updateLookup", func(t *testing.T) {
		cs := base.UpdLookup().(*changeStream[testutil.TestDoc])
		testutil.AssertEqual(t, "updateLookup", cs.fullDoc, "UpdLookup should set 'updateLookup'")
	})

	t.Run("FullDocRequired sets required", func(t *testing.T) {
		cs := base.FullDocRequired().(*changeStream[testutil.TestDoc])
		testutil.AssertEqual(t, "required", cs.fullDoc, "FullDocRequired should set 'required'")
	})
}

// TestChangeStreamResumeTokens tests ResumeAfter and StartAfter.
func TestChangeStreamResumeTokens(t *testing.T) {
	base := newTestChangeStream(t)
	token := bson.M{"_data": "abc123"}

	t.Run("ResumeAfter sets resumeAfter", func(t *testing.T) {
		cs := base.ResumeAfter(token).(*changeStream[testutil.TestDoc])
		testutil.AssertEqual(t, token, cs.resumeAfter, "ResumeAfter should set token")
	})

	t.Run("StartAfter sets startAfter", func(t *testing.T) {
		cs := base.StartAfter(token).(*changeStream[testutil.TestDoc])
		testutil.AssertEqual(t, token, cs.startAfter, "StartAfter should set token")
	})
}

// TestChangeStreamMiscOptions tests Bsz, MaxAwait, and Comment builder methods.
func TestChangeStreamMiscOptions(t *testing.T) {
	base := newTestChangeStream(t)

	t.Run("Bsz sets batchSize", func(t *testing.T) {
		cs := base.Bsz(50).(*changeStream[testutil.TestDoc])
		testutil.AssertNotNil(t, cs.batchSize, "Bsz should set batchSize")
		testutil.AssertEqual(t, int32(50), *cs.batchSize, "Bsz should set correct value")
	})

	t.Run("MaxAwait sets maxAwait", func(t *testing.T) {
		d := 5 * time.Second
		cs := base.MaxAwait(d).(*changeStream[testutil.TestDoc])
		testutil.AssertNotNil(t, cs.maxAwait, "MaxAwait should set maxAwait")
		testutil.AssertEqual(t, d, *cs.maxAwait, "MaxAwait should set correct duration")
	})

	t.Run("Comment sets comment", func(t *testing.T) {
		cs := base.Comment("my stream").(*changeStream[testutil.TestDoc])
		testutil.AssertEqual(t, "my stream", cs.comment, "Comment should set comment")
	})
}

// TestChangeStreamImmutability tests that each builder method returns a new instance
// without mutating the original.
func TestChangeStreamImmutability(t *testing.T) {
	base := newTestChangeStream(t).(*changeStream[testutil.TestDoc])

	t.Run("OnIst does not mutate original", func(t *testing.T) {
		_ = base.OnIst()
		testutil.AssertEqual(t, 0, len(base.onOps), "Original should not have onOps after OnIst")
	})

	t.Run("FullDoc does not mutate original", func(t *testing.T) {
		_ = base.FullDoc("updateLookup")
		testutil.AssertEqual(t, "", base.fullDoc, "Original fullDoc should remain empty")
	})

	t.Run("Bsz does not mutate original", func(t *testing.T) {
		_ = base.Bsz(100)
		testutil.AssertEqual(t, (*int32)(nil), base.batchSize, "Original batchSize should remain nil")
	})

	t.Run("Comment does not mutate original", func(t *testing.T) {
		_ = base.Comment("test")
		testutil.AssertEqual(t, "", base.comment, "Original comment should remain empty")
	})
}

// TestChangeStreamNilCallback tests that Stream(nil) and Each(nil) return an error.
func TestChangeStreamNilCallback(t *testing.T) {
	cs := newTestChangeStream(t)

	t.Run("Stream(nil) returns error", func(t *testing.T) {
		err := cs.Stream(nil)
		testutil.AssertError(t, err, "Stream(nil) should return an error")
	})

	t.Run("Each(nil) returns error", func(t *testing.T) {
		err := cs.Each(nil)
		testutil.AssertError(t, err, "Each(nil) should return an error")
	})
}

// TestChangeStreamBuildPipeline tests that build() prepends $match on operationType
// when onOps is set.
func TestChangeStreamBuildPipeline(t *testing.T) {
	base := newTestChangeStream(t)

	t.Run("no onOps — pipeline has no extra stage", func(t *testing.T) {
		cs := base.(*changeStream[testutil.TestDoc])
		_, pipeline := cs.build()
		testutil.AssertEqual(t, 0, len(pipeline), "Empty pipeline should have no stages")
	})

	t.Run("OnIst — pipeline prepends $match stage", func(t *testing.T) {
		cs := base.OnIst().(*changeStream[testutil.TestDoc])
		_, pipeline := cs.build()
		testutil.AssertEqual(t, 1, len(pipeline), "Pipeline should have 1 stage after OnIst")

		stage, ok := pipeline[0].(bson.M)
		testutil.AssertTrue(t, ok, "$match stage should be bson.M")

		match, ok := stage["$match"].(bson.M)
		testutil.AssertTrue(t, ok, "$match value should be bson.M")

		inExpr, ok := match["operationType"].(bson.M)
		testutil.AssertTrue(t, ok, "operationType filter should be bson.M")

		ops, ok := inExpr["$in"].([]string)
		testutil.AssertTrue(t, ok, "$in should be []string")
		testutil.AssertContains(t, ops, "insert", "$in should contain 'insert'")
	})

	t.Run("user pipeline stages are appended after $match", func(t *testing.T) {
		userStage := bson.M{"$project": bson.M{"_id": 0}}
		collRaw := newTestChangeStream(t).(*changeStream[testutil.TestDoc])
		collRaw.pipeline = []any{userStage}

		cs := collRaw.OnUpd()
		_, pipeline := cs.(*changeStream[testutil.TestDoc]).build()

		testutil.AssertEqual(t, 2, len(pipeline), "Pipeline should have $match + user stage")

		// First stage is the $match filter
		_, isMatch := pipeline[0].(bson.M)
		testutil.AssertTrue(t, isMatch, "First stage should be $match bson.M")

		// Second stage is the user's projection
		testutil.AssertEqual(t, userStage, pipeline[1], "Second stage should be user stage")
	})
}

// TestChangeStreamOptsBuildOverride tests that Opts() fields override builder fields.
func TestChangeStreamOptsBuildOverride(t *testing.T) {
	base := newTestChangeStream(t)

	t.Run("Opts overrides Bsz", func(t *testing.T) {
		override := int32(999)
		cs := base.Bsz(10).Opts(options.ChangeStream().SetBatchSize(override))
		opts, _ := cs.(*changeStream[testutil.TestDoc]).build()
		testutil.AssertNotNil(t, opts.BatchSize, "BatchSize should be set")
		testutil.AssertEqual(t, override, *opts.BatchSize, "Opts should override Bsz")
	})

	t.Run("Opts overrides fullDocument", func(t *testing.T) {
		override := options.FullDocument("required")
		cs := base.UpdLookup().Opts(options.ChangeStream().SetFullDocument(override))
		opts, _ := cs.(*changeStream[testutil.TestDoc]).build()
		testutil.AssertNotNil(t, opts.FullDocument, "FullDocument should be set")
		testutil.AssertEqual(t, override, *opts.FullDocument, "Opts should override FullDoc")
	})

	t.Run("builder fields preserved when Opts does not set them", func(t *testing.T) {
		token := bson.M{"_data": "xyz"}
		cs := base.ResumeAfter(token).Opts(options.ChangeStream().SetBatchSize(5))
		opts, _ := cs.(*changeStream[testutil.TestDoc]).build()

		testutil.AssertNotNil(t, opts.ResumeAfter, "ResumeAfter should be preserved from builder")
		testutil.AssertNotNil(t, opts.BatchSize, "BatchSize should come from Opts")
	})
}

// TestChangeStreamStartAtTime tests StartAtTime builder method.
func TestChangeStreamStartAtTime(t *testing.T) {
	base := newTestChangeStream(t)
	ts := &primitive.Timestamp{T: uint32(time.Now().Unix()), I: 1}

	cs := base.StartAtTime(ts).(*changeStream[testutil.TestDoc])
	testutil.AssertNotNil(t, cs.startAtTime, "StartAtTime should set startAtTime")
	testutil.AssertEqual(t, ts, cs.startAtTime, "StartAtTime should store the given timestamp")
}
