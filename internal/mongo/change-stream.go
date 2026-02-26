// Package mongo — change-stream.go
//
// changeStream[T] implements the ChangeStream[T] interface.
// It is an immutable fluent builder: every option method returns a shallow copy so
// the same base stream can be forked into multiple configurations without side effects.
package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	mg "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// changeStream[T] holds the configuration for a MongoDB collection-level change stream.
// Fields are set via the fluent builder methods and consumed by build() / Stream().
type changeStream[T any] struct {
	ctx         context.Context          // lifetime of the stream; comes from Watch(ctx)
	coll        *mg.Collection           // read client collection to call Watch on
	collName    string                   // stored for diagnostics / error messages
	pipeline    []any                    // user-supplied aggregation stages (optional)
	onOps       []string                 // prepended as a $match on operationType when non-empty
	fullDoc     string                   // fullDocument option ("updateLookup", "required", etc.)
	resumeAfter bson.M                   // resume token — stream restarts after this event
	startAfter  bson.M                   // like resumeAfter but supports post-invalidate resume
	extra       *options.ChangeStreamOptions // raw options merged last; override builder settings
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

func (cs *changeStream[T]) On(operationTypes ...string) ChangeStream[T] {
	next := *cs
	next.onOps = append(append([]string{}, cs.onOps...), operationTypes...)
	return &next
}

func (cs *changeStream[T]) OnIst() ChangeStream[T]        { return cs.On("insert") }
func (cs *changeStream[T]) OnUpd() ChangeStream[T]        { return cs.On("update") }
func (cs *changeStream[T]) OnDel() ChangeStream[T]        { return cs.On("delete") }
func (cs *changeStream[T]) OnRep() ChangeStream[T]        { return cs.On("replace") }
func (cs *changeStream[T]) OnIstAndUpd() ChangeStream[T]  { return cs.On("insert", "update") }

func (cs *changeStream[T]) FullDoc(option string) ChangeStream[T] {
	next := *cs
	next.fullDoc = option
	return &next
}

func (cs *changeStream[T]) UpdLookup() ChangeStream[T]      { return cs.FullDoc("updateLookup") }
func (cs *changeStream[T]) FullDocRequired() ChangeStream[T] { return cs.FullDoc("required") }

func (cs *changeStream[T]) ResumeAfter(token bson.M) ChangeStream[T] {
	next := *cs
	next.resumeAfter = token
	return &next
}

func (cs *changeStream[T]) StartAfter(token bson.M) ChangeStream[T] {
	next := *cs
	next.startAfter = token
	return &next
}

func (cs *changeStream[T]) Opts(o *options.ChangeStreamOptions) ChangeStream[T] {
	next := *cs
	next.extra = o
	return &next
}

func (cs *changeStream[T]) Stream(fn func(st CsEvt[T]) error) error {
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

func (cs *changeStream[T]) Each(fn func(st CsEvt[T]) error) error {
	return cs.Stream(fn)
}

// build assembles the final ChangeStreamOptions and the aggregation pipeline.
// If onOps is non-empty, a $match stage is prepended to filter by operationType.
// The user-supplied pipeline stages follow after that filter.
func (cs *changeStream[T]) build() (*options.ChangeStreamOptions, []any) {
	opts := options.ChangeStream()
	if cs.extra != nil {
		*opts = *cs.extra
	}
	if cs.fullDoc != "" {
		opts.SetFullDocument(options.FullDocument(cs.fullDoc))
	}
	if cs.resumeAfter != nil {
		opts.SetResumeAfter(cs.resumeAfter)
	}
	if cs.startAfter != nil {
		opts.SetStartAfter(cs.startAfter)
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
