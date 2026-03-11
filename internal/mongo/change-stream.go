// Package mongo — change-stream.go
//
// changeStream[T] implements the ChangeStream[T] interface.
// It is an immutable fluent builder: every option method returns a shallow copy so
// the same base stream can be forked into multiple configurations without side effects.
package mongo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mg "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// changeStream[T] holds the configuration for a MongoDB collection-level change stream.
// Fields are set via the fluent builder methods and consumed by build() / Stream().
type changeStream[T any] struct {
	ctx            context.Context              // lifetime of the stream; comes from Watch(ctx)
	coll           *mg.Collection               // read client collection to call Watch on
	collName       string                       // stored for diagnostics / error messages
	pipeline       []any                        // user-supplied aggregation stages (optional)
	onOps          []string                     // prepended as a $match on operationType when non-empty
	fullDoc        string                       // fullDocument option ("updateLookup", "required", etc.)
	resumeAfter    bson.M                       // resume token — stream restarts after this event
	startAfter     bson.M                       // like resumeAfter but supports post-invalidate resume
	extra          *options.ChangeStreamOptions // raw options merged last; override builder settings
	batchSize      *int32
	collation      *options.Collation
	comment        string
	fullDocBefore  string
	maxAwait       *time.Duration
	showExpanded   *bool
	startAtTime    *primitive.Timestamp
	custom         bson.M
	customPipeline bson.M
}

// compile-time check: changeStream[T] must satisfy ChangeStream[T].
var _ ChangeStream[any] = (*changeStream[any])(nil)

// NewChangeStream creates a new ChangeStream builder bound to the given collection.
// Normally called via Collection.Watch(ctx, pipeline...) rather than directly.
func NewChangeStream[T any](ctx context.Context, coll *mg.Collection, collName string, pipeline []any) ChangeStream[T] {
	return &changeStream[T]{
		ctx:      normalizeCtx(ctx),
		coll:     coll,
		collName: collName,
		pipeline: pipeline,
	}
}

// On filters events to the specified operation types and returns a new builder.
// Valid values: "insert", "update", "replace", "delete", "drop", "rename", "invalidate".
func (cs *changeStream[T]) On(operationTypes ...string) ChangeStream[T] {
	next := *cs
	next.onOps = append(append([]string{}, cs.onOps...), operationTypes...)
	return &next
}

// OnIst filters for "insert" events only.
func (cs *changeStream[T]) OnIst() ChangeStream[T] { return cs.On("insert") }

// OnUpd filters for "update" events only.
func (cs *changeStream[T]) OnUpd() ChangeStream[T] { return cs.On("update") }

// OnDel filters for "delete" events only.
func (cs *changeStream[T]) OnDel() ChangeStream[T] { return cs.On("delete") }

// OnRep filters for "replace" events only.
func (cs *changeStream[T]) OnRep() ChangeStream[T] { return cs.On("replace") }

// OnIstAndUpd filters for "insert" and "update" events (most common combination).
func (cs *changeStream[T]) OnIstAndUpd() ChangeStream[T] { return cs.On("insert", "update") }

// FullDoc sets the fullDocument option. Valid values: "default", "updateLookup", "whenAvailable", "required".
func (cs *changeStream[T]) FullDoc(option string) ChangeStream[T] {
	next := *cs
	next.fullDoc = option
	return &next
}

// UpdLookup sets fullDocument to "updateLookup" so the driver fetches the full document
// after every update. FullDocument will be non-nil for update events.
func (cs *changeStream[T]) UpdLookup() ChangeStream[T] { return cs.FullDoc("updateLookup") }

// FullDocRequired sets fullDocument to "required". The server errors if the full document
// cannot be provided (e.g. the document was deleted between the change and the lookup).
func (cs *changeStream[T]) FullDocRequired() ChangeStream[T] { return cs.FullDoc("required") }

// ResumeAfter resumes the stream starting after the given resume token.
// Use st.ChangeEvent.ResumeToken from a previous callback to obtain the token.
func (cs *changeStream[T]) ResumeAfter(token bson.M) ChangeStream[T] {
	next := *cs
	next.resumeAfter = token
	return &next
}

// StartAfter is like ResumeAfter but can resume even after an "invalidate" event.
func (cs *changeStream[T]) StartAfter(token bson.M) ChangeStream[T] {
	next := *cs
	next.startAfter = token
	return &next
}

// Opts applies raw ChangeStreamOptions, merged on top of any builder settings (escape hatch).
func (cs *changeStream[T]) Opts(o *options.ChangeStreamOptions) ChangeStream[T] {
	next := *cs
	next.extra = o
	return &next
}

// Bsz sets the batch size for the change stream cursor.
func (cs *changeStream[T]) Bsz(n int32) ChangeStream[T] {
	next := *cs; next.batchSize = &n; return &next
}

// Collation sets the collation for the change stream.
func (cs *changeStream[T]) Collation(c *options.Collation) ChangeStream[T] {
	next := *cs; next.collation = c; return &next
}

// Comment attaches a comment to the change stream command (visible in profiler / logs).
func (cs *changeStream[T]) Comment(s string) ChangeStream[T] {
	next := *cs; next.comment = s; return &next
}

// FullDocBefore sets the fullDocumentBeforeChange option.
// Valid values: "off", "whenAvailable", "required".
func (cs *changeStream[T]) FullDocBefore(option string) ChangeStream[T] {
	next := *cs; next.fullDocBefore = option; return &next
}

// MaxAwait sets the maximum time the server waits for new data before returning an empty batch.
func (cs *changeStream[T]) MaxAwait(d time.Duration) ChangeStream[T] {
	next := *cs; next.maxAwait = &d; return &next
}

// ShowExpanded enables expanded change stream events (MongoDB 6.0+).
func (cs *changeStream[T]) ShowExpanded(b bool) ChangeStream[T] {
	next := *cs; next.showExpanded = &b; return &next
}

// StartAtTime sets the operation time at which the change stream should start.
func (cs *changeStream[T]) StartAtTime(t *primitive.Timestamp) ChangeStream[T] {
	next := *cs; next.startAtTime = t; return &next
}

// Custom sets a custom BSON document to be added to the change stream command.
func (cs *changeStream[T]) Custom(m bson.M) ChangeStream[T] {
	next := *cs; next.custom = m; return &next
}

// CustomPipeline sets a custom BSON document to be added to the change stream aggregation pipeline.
func (cs *changeStream[T]) CustomPipeline(m bson.M) ChangeStream[T] {
	next := *cs; next.customPipeline = m; return &next
}

// Stream opens the change stream and calls fn for every incoming event until fn returns
// a non-nil error or the context passed to Watch is cancelled.
func (cs *changeStream[T]) Stream(fn func(st CsEvt[T]) error) error {
	if fn == nil {
		return errors.New("fn must not be nil")
	}
	opts, pipeline := cs.build()
	stream, err := cs.coll.Watch(cs.ctx, pipeline, opts)
	if err != nil {
		return err
	}
	defer stream.Close(cs.ctx)

	for stream.Next(cs.ctx) {
		event, err := decodeChangeEvent[T](stream)
		if err != nil {
			return err
		}
		if err := fn(CsEvt[T]{ChangeEvent: event, ctx: cs.ctx}); err != nil {
			return err
		}
	}
	return stream.Err()
}

// Each is an alias for Stream.
func (cs *changeStream[T]) Each(fn func(st CsEvt[T]) error) error {
	return cs.Stream(fn)
}

// build assembles the final ChangeStreamOptions and the aggregation pipeline.
// If onOps is non-empty, a $match stage is prepended to filter by operationType.
// The user-supplied pipeline stages follow after that filter.
// Builder fields are applied first; Opts() fields are merged on top so that
// Opts() acts as a final escape-hatch override without wiping builder settings.
func (cs *changeStream[T]) build() (*options.ChangeStreamOptions, []any) {
	opts := options.ChangeStream()
	if cs.fullDoc != "" {
		opts.SetFullDocument(options.FullDocument(cs.fullDoc))
	}
	if cs.resumeAfter != nil {
		opts.SetResumeAfter(cs.resumeAfter)
	}
	if cs.startAfter != nil {
		opts.SetStartAfter(cs.startAfter)
	}
	if cs.batchSize != nil {
		opts.SetBatchSize(*cs.batchSize)
	}
	if cs.collation != nil {
		opts.SetCollation(*cs.collation)
	}
	if cs.comment != "" {
		opts.SetComment(cs.comment)
	}
	if cs.fullDocBefore != "" {
		opts.SetFullDocumentBeforeChange(options.FullDocument(cs.fullDocBefore))
	}
	if cs.maxAwait != nil {
		opts.SetMaxAwaitTime(*cs.maxAwait)
	}
	if cs.showExpanded != nil {
		opts.SetShowExpandedEvents(*cs.showExpanded)
	}
	if cs.startAtTime != nil {
		opts.SetStartAtOperationTime(cs.startAtTime)
	}
	if cs.custom != nil {
		opts.SetCustom(cs.custom)
	}
	if cs.customPipeline != nil {
		opts.SetCustomPipeline(cs.customPipeline)
	}
	if cs.extra != nil {
		mergeChangeStreamOpts(opts, cs.extra)
	}

	pipeline := make([]any, 0, len(cs.pipeline)+1)
	if len(cs.onOps) > 0 {
		pipeline = append(pipeline, bson.M{
			"$match": bson.M{"operationType": bson.M{"$in": cs.onOps}},
		})
	}
	pipeline = append(pipeline, cs.pipeline...)

	return opts, pipeline
}

// mergeChangeStreamOpts copies non-nil fields from src into dst.
// Only fields explicitly set in src override the corresponding dst fields,
// preserving any values already set by the fluent builder methods.
func mergeChangeStreamOpts(dst, src *options.ChangeStreamOptions) {
	if src.BatchSize != nil {
		dst.BatchSize = src.BatchSize
	}
	if src.Collation != nil {
		dst.Collation = src.Collation
	}
	if src.Comment != nil {
		dst.Comment = src.Comment
	}
	if src.FullDocument != nil {
		dst.FullDocument = src.FullDocument
	}
	if src.FullDocumentBeforeChange != nil {
		dst.FullDocumentBeforeChange = src.FullDocumentBeforeChange
	}
	if src.MaxAwaitTime != nil {
		dst.MaxAwaitTime = src.MaxAwaitTime
	}
	if src.ResumeAfter != nil {
		dst.ResumeAfter = src.ResumeAfter
	}
	if src.ShowExpandedEvents != nil {
		dst.ShowExpandedEvents = src.ShowExpandedEvents
	}
	if src.StartAtOperationTime != nil {
		dst.StartAtOperationTime = src.StartAtOperationTime
	}
	if src.StartAfter != nil {
		dst.StartAfter = src.StartAfter
	}
	if src.Custom != nil {
		dst.Custom = src.Custom
	}
	if src.CustomPipeline != nil {
		dst.CustomPipeline = src.CustomPipeline
	}
}

// decodeChangeEvent decodes the current cursor position of a *mongo.ChangeStream into a
// strongly typed ChangeEvent[T]. It first decodes the raw BSON into an internal struct,
// then unmarshals the fullDocument field separately into *T so the caller gets a typed value.
func decodeChangeEvent[T any](stream *mg.ChangeStream) (ChangeEvent[T], error) {
	type rawNs struct {
		DB   string `bson:"db"`
		Coll string `bson:"coll"`
	}
	type rawUpdateDesc struct {
		UpdatedFields bson.M   `bson:"updatedFields"`
		RemovedFields []string `bson:"removedFields"`
	}
	type rawEvent struct {
		ResumeToken   bson.M         `bson:"_id"`
		OperationType string         `bson:"operationType"`
		FullDocument  bson.Raw       `bson:"fullDocument"`
		DocumentKey   bson.M         `bson:"documentKey"`
		Ns            rawNs          `bson:"ns"`
		UpdateDesc    *rawUpdateDesc `bson:"updateDescription"`
	}

	var raw rawEvent
	if err := stream.Decode(&raw); err != nil {
		return ChangeEvent[T]{}, err
	}

	event := ChangeEvent[T]{
		ResumeToken:   raw.ResumeToken,
		OperationType: raw.OperationType,
		DocumentKey:   raw.DocumentKey,
		Namespace: ChangeEventNamespace{
			DB:   raw.Ns.DB,
			Coll: raw.Ns.Coll,
		},
	}

	if raw.UpdateDesc != nil {
		event.UpdateDesc = &ChangeUpdateDesc{
			UpdatedFields: raw.UpdateDesc.UpdatedFields,
			RemovedFields: raw.UpdateDesc.RemovedFields,
		}
	}

	if len(raw.FullDocument) > 0 {
		var doc T
		if err := bson.Unmarshal(raw.FullDocument, &doc); err != nil {
			return event, err
		}
		event.FullDocument = &doc
	}

	return event, nil
}
