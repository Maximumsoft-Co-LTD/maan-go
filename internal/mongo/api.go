package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mg "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client represents a MongoDB client that knows how to provide read/write connections
// and the database name it is scoped to.
type Client interface {
	// Write returns the underlying *mongo.Client configured for write operations.
	Write() *mg.Client
	// Read returns the underlying *mongo.Client configured for read operations.
	// Returns the same client as Write when no separate read URI was configured.
	Read() *mg.Client
	// DbName returns the logical database name this client is scoped to.
	DbName() string
	// Close disconnects both write and read clients. Safe to call even when
	// read and write share the same underlying connection.
	Close() error
	// WithTx runs fn inside an automatically managed transaction.
	// Commits on nil error; rolls back otherwise.
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
	// StartTx begins a manual transaction and returns a TxSession.
	// Call tx.Close(&err) (usually via defer) to commit or rollback.
	StartTx(ctx context.Context) (TxSession, error)
}

// Collection is the fluent entry point for CRUD and aggregation operations on strongly typed documents.
type Collection[T any] interface {
	// Ctx returns a shallow copy of the collection bound to the given context.
	// Use this to attach a request-scoped deadline or cancellation to an existing collection instance.
	Ctx(ctx context.Context) Collection[T]
	// Build returns an ExtendedCollection bound to ctx for chainable dynamic queries.
	Build(ctx context.Context) ExtendedCollection[T]
	// Name returns the MongoDB collection name.
	Name() string
	// Create inserts doc as a single document. Returns error if doc is nil.
	// Automatically calls model-default hooks (DefaultId, DefaultCreatedAt, DefaultUpdatedAt).
	Create(doc *T, opts ...*options.InsertOneOptions) error
	// CreateMany inserts every element of docs. Returns error if docs is nil.
	// Automatically calls model-default hooks on each element before inserting.
	CreateMany(docs *[]T, opts ...*options.InsertManyOptions) error
	// Find returns a ManyResult builder for filter. Alias for FindMany.
	Find(filter any) ManyResult[T]
	// FindOne returns a SingleResult builder for query.
	FindOne(query any) SingleResult[T]
	// FindMany returns a ManyResult builder for filter.
	FindMany(filter any) ManyResult[T]
	// Save performs an upsert: updates the first matching document or inserts if none found.
	// Automatically injects updated_at into the $set clause.
	Save(filter any, update any, opts ...*options.UpdateOptions) error
	// SaveMany performs a multi-document upsert for all documents matching filter.
	SaveMany(filter any, update any, opts ...*options.UpdateOptions) error
	// Upd updates the first document matching filter. Does NOT insert when no match exists.
	Upd(filter any, update any, opts ...*options.UpdateOptions) error
	// UpdMany updates all documents matching filter. Does NOT insert new documents.
	UpdMany(filter any, update any, opts ...*options.UpdateOptions) error
	// Del deletes the first document matching filter.
	Del(filter any, opts ...*options.DeleteOptions) error
	// DelMany deletes all documents matching filter.
	DelMany(filter any, opts ...*options.DeleteOptions) error
	// FindOneAndUpd atomically finds the first document matching filter, applies update, and decodes the result into out.
	// Automatically injects updated_at into the $set clause.
	FindOneAndUpd(filter any, update any, out *T, opts ...*options.FindOneAndUpdateOptions) error
	// FindOneAndDel atomically finds the first document matching filter, deletes it, and decodes it into out.
	FindOneAndDel(filter any, out *T, opts ...*options.FindOneAndDeleteOptions) error
	// Distinct returns the distinct values for the given field across all documents matching filter.
	// Pass nil filter to match all documents.
	Distinct(field string, filter any) ([]any, error)
	// Count returns the number of documents matching filter.
	// Pass nil filter to count all documents.
	Count(filter any) (int64, error)
	// Agg returns an Aggregate builder for the given aggregation pipeline.
	Agg(pipeline any) Aggregate[T]
	// RegexFields performs a case-insensitive regex search across the given field names.
	RegexFields(q string, fields ...string) ([]T, error)
	// TxtFind performs a MongoDB full-text search ($text / $search).
	// Requires a text index on the collection.
	TxtFind(q string) ([]T, error)
	// WithTx runs fn inside an automatically managed transaction.
	// Commits on nil error; rolls back otherwise.
	WithTx(fn func(ctx context.Context) error) error
	// StartTx begins a manual transaction and returns a TxSession.
	// Call tx.Close(&err) (usually via defer) to commit or rollback.
	StartTx() (TxSession, error)
	// Watch returns a ChangeStream builder for real-time change events on the collection.
	// The collection's bound context (set via Ctx) controls the stream lifetime.
	// Requires a MongoDB replica set or sharded cluster.
	Watch(pipeline ...any) ChangeStream[T]
	// Idx returns an IndexManager for managing indexes on this collection.
	Idx() IndexManager
}

// IndexManager provides index management operations on a MongoDB collection.
type IndexManager interface {
	// CreateOne creates a single index on the collection and returns the index name.
	CreateOne(model mg.IndexModel, opts ...*options.CreateIndexesOptions) (string, error)
	// CreateMany creates multiple indexes on the collection and returns their names.
	CreateMany(models []mg.IndexModel, opts ...*options.CreateIndexesOptions) ([]string, error)
	// DropOne drops the index with the given name.
	DropOne(name string, opts ...*options.DropIndexesOptions) error
	// DropAll drops all non-_id indexes on the collection.
	DropAll(opts ...*options.DropIndexesOptions) error
	// List returns all indexes on the collection as a slice of bson.M documents.
	List(opts ...*options.ListIndexesOptions) ([]bson.M, error)
}

// ExtendedCollection supports building reusable dynamic queries that can be chained.
type ExtendedCollection[T any] interface {
	// By adds an equality condition on the named struct field.
	// The field name is resolved to a BSON key via the bson tag or snake_case conversion.
	By(string, any) ExtendedCollection[T]
	// Where merges the provided BSON filter into the current accumulated filter.
	Where(bson.M) ExtendedCollection[T]
	// First decodes the first document matching the accumulated filter into result.
	First(*T) error
	// Many decodes all documents matching the accumulated filter into results.
	Many(*[]T) error
	// Save performs a single-document update using the accumulated filter.
	// Automatically injects updated_at.
	Save(any, ...*options.UpdateOptions) error
	// SaveMany performs a multi-document update using the accumulated filter.
	SaveMany(any, ...*options.UpdateOptions) error
	// Del deletes the first document matching the accumulated filter.
	Del(...*options.DeleteOptions) error
	// Delete is an alias for Del.
	Delete(...*options.DeleteOptions) error
	// DelMany deletes all documents matching the accumulated filter.
	DelMany(...*options.DeleteOptions) error
	// DeleteMany is an alias for DelMany.
	DeleteMany(...*options.DeleteOptions) error
	// Count returns the number of documents matching the accumulated filter.
	Count() (int64, error)
	// Exists returns true when at least one document matches the accumulated filter.
	Exists() (bool, error)
	// GetFilter returns the accumulated BSON filter as built by By/Where calls.
	GetFilter() any
}

// SingleResult models a find-one query with optional modifiers.
type SingleResult[T any] interface {
	// Proj sets the projection document (which fields to include or exclude).
	Proj(p any) SingleResult[T]
	// Sort sets the sort document.
	Sort(s any) SingleResult[T]
	// Hint sets the index hint.
	Hint(h any) SingleResult[T]
	// Opts merges raw FindOneOptions on top of the builder settings.
	Opts(fo *options.FindOneOptions) SingleResult[T]
	// Result executes the query and decodes the result into out.
	// Returns mongo.ErrNoDocuments if no matching document is found.
	Result(out *T) error
}

// ManyResult models a find-many query with modifiers and streaming helpers.
type ManyResult[T any] interface {
	// Proj sets the projection document.
	Proj(p any) ManyResult[T]
	// Sort sets the sort document.
	Sort(s any) ManyResult[T]
	// Hint sets the index hint.
	Hint(h any) ManyResult[T]
	// Limit caps the number of documents returned.
	Limit(n int64) ManyResult[T]
	// Skip skips the first n documents in the result set.
	Skip(n int64) ManyResult[T]
	// Bsz sets the cursor batch size.
	Bsz(n int32) ManyResult[T]
	// Opts merges raw FindOptions on top of the builder settings.
	Opts(fo *options.FindOptions) ManyResult[T]
	// All executes the query and returns all matching documents.
	All() ([]T, error)
	// Result executes the query and decodes all matching documents into out.
	Result(out *[]T) error
	// Stream executes the query and calls fn for each document.
	// Stops and returns the first non-nil error from fn.
	Stream(fn func(ctx context.Context, doc T) error) error
	// Each is an alias for Stream.
	Each(fn func(ctx context.Context, doc T) error) error
	// Cnt returns the count of documents matching the filter (ignores Limit/Skip).
	Cnt() (int64, error)
}

// Aggregate wraps aggregation pipelines with streaming helpers.
type Aggregate[T any] interface {
	// Disk enables (true) or disables (false) writing temporary aggregation data to disk.
	Disk(b bool) Aggregate[T]
	// Bsz sets the cursor batch size for the aggregation result.
	Bsz(n int32) Aggregate[T]
	// Opts merges raw AggregateOptions on top of the builder settings.
	Opts(ao *options.AggregateOptions) Aggregate[T]
	// All executes the pipeline and returns all typed results.
	All() ([]T, error)
	// Result executes the pipeline and decodes all results into out.
	Result(out *[]T) error
	// Raw executes the pipeline and returns raw bson.M documents.
	Raw() ([]bson.M, error)
	// Stream executes the pipeline and calls fn for each typed document.
	// Stops on the first non-nil error returned by fn.
	Stream(fn func(ctx context.Context, doc T) error) error
	// Each is an alias for Stream.
	Each(fn func(ctx context.Context, doc T) error) error
	// EachRaw executes the pipeline and calls fn for each raw bson.M document.
	EachRaw(fn func(ctx context.Context, doc bson.M) error) error
}

// TxSession exposes a MongoDB session used for manual transaction control.
type TxSession interface {
	// Ctx returns the session-aware context that must be passed to collection operations.
	Ctx() context.Context
	// Close commits the transaction when *err == nil, otherwise aborts it.
	// Always call via `defer tx.Close(&err)` to ensure proper cleanup.
	Close(err *error)
}

// ChangeEventNamespace holds the database and collection name where a change occurred.
type ChangeEventNamespace struct {
	// DB is the name of the database.
	DB string
	// Coll is the name of the collection.
	Coll string
}

// ChangeUpdateDesc describes which fields were modified in an "update" operation.
type ChangeUpdateDesc struct {
	// UpdatedFields is a map of field paths that were set or changed.
	UpdatedFields bson.M
	// RemovedFields is the list of field paths that were unset.
	RemovedFields []string
}

// CsEvt (Change Stream Event) is the single argument passed to Stream/Each callbacks.
// It bundles the typed change event data together with the streaming context so the
// callback has everything it needs without managing two separate parameters.
//
// Example:
//
//	coll.Watch().OnIstAndUpd().UpdLookup().
//	    Stream(func(st maango.CsEvt[Order]) error {
//	        log.Printf("op=%s id=%v", st.ChangeEvent.OperationType, st.ChangeEvent.DocumentKey)
//	        if doc := st.ChangeEvent.FullDocument; doc != nil {
//	            log.Printf("doc=%+v", *doc)
//	        }
//	        log.Printf("ctx deadline: %v", st.Ctx())
//	        return nil
//	    })
type CsEvt[T any] struct {
	// ChangeEvent contains all event data: operation type, document, keys, etc.
	ChangeEvent ChangeEvent[T]
	ctx         context.Context
}

// Ctx returns the context that controls the lifetime of the change stream.
// Use it to propagate deadlines or pass to downstream calls inside the callback.
func (s CsEvt[T]) Ctx() context.Context { return s.ctx }

// ChangeEvent represents a single MongoDB change stream event with a typed FullDocument.
// The event mirrors the MongoDB change event document returned by the $changeStream aggregation stage.
//
// Fields availability by operation type:
//
//	insert  → FullDocument is always set
//	update  → FullDocument is nil unless UpdLookup() / FullDocRequired() is used; UpdateDesc is set
//	replace → FullDocument is always set
//	delete  → FullDocument is always nil; DocumentKey contains the deleted _id
type ChangeEvent[T any] struct {
	// ResumeToken is the opaque token that can be passed to ResumeAfter/StartAfter
	// to restart the stream from exactly this point after an interruption.
	ResumeToken bson.M
	// OperationType indicates what happened: "insert", "update", "replace",
	// "delete", "drop", "rename", "dropDatabase", or "invalidate".
	OperationType string
	// FullDocument is the typed version of the document after the change.
	// It is nil for "delete" events and for "update" events unless
	// UpdLookup() or FullDocRequired() was set on the stream builder.
	FullDocument *T
	// DocumentKey holds the _id (and shard key if applicable) of the affected document.
	DocumentKey bson.M
	// Namespace identifies which database and collection produced this event.
	Namespace ChangeEventNamespace
	// UpdateDesc is only set for "update" operations. It lists the fields that
	// were changed (UpdatedFields) and the fields that were removed (RemovedFields).
	UpdateDesc *ChangeUpdateDesc
}

// ChangeStream is a fluent builder for watching real-time change events on a MongoDB collection.
// Call Watch on a Collection to obtain a ChangeStream, chain option methods, then call
// Stream or Each to begin consuming events.
//
// Requires MongoDB replica set or sharded cluster (change streams are not available on standalone).
//
// Basic example — react to every change:
//
//	err := coll.Ctx(ctx).Watch().
//	    Stream(func(st maango.CsEvt[Order]) error {
//	        fmt.Println(st.ChangeEvent.OperationType)
//	        return nil
//	    })
//
// Filtered example — inserts and updates, with full document on updates:
//
//	err := coll.Ctx(ctx).Watch().
//	    OnIstAndUpd().
//	    UpdLookup().
//	    Stream(func(st maango.CsEvt[Order]) error {
//	        fmt.Printf("op=%s doc=%+v\n",
//	            st.ChangeEvent.OperationType,
//	            st.ChangeEvent.FullDocument)
//	        return nil
//	    })
//
// Resume after interruption:
//
//	var lastToken bson.M
//	coll.Ctx(ctx).Watch().Stream(func(st maango.CsEvt[Order]) error {
//	    lastToken = st.ChangeEvent.ResumeToken
//	    return nil
//	})
//	// ... restart later:
//	coll.Ctx(ctx).Watch().ResumeAfter(lastToken).Stream(handler)
type ChangeStream[T any] interface {
	// On filters events to the specified operation types.
	// Valid values: "insert", "update", "replace", "delete", "drop", "rename", "invalidate".
	//
	// Example: .On("insert", "delete")
	On(operationTypes ...string) ChangeStream[T]

	// OnIst filters for "insert" events only.
	OnIst() ChangeStream[T]
	// OnUpd filters for "update" events only.
	OnUpd() ChangeStream[T]
	// OnDel filters for "delete" events only.
	OnDel() ChangeStream[T]
	// OnRep filters for "replace" events only.
	OnRep() ChangeStream[T]
	// OnIstAndUpd filters for "insert" and "update" events (most common combination).
	OnIstAndUpd() ChangeStream[T]

	// FullDoc sets the fullDocument option controlling when the full document is returned.
	// Valid values: "default", "updateLookup", "whenAvailable", "required".
	FullDoc(option string) ChangeStream[T]
	// UpdLookup sets fullDocument to "updateLookup": the driver performs an extra read to
	// fetch the full document for every update event. FullDocument will be non-nil for updates.
	UpdLookup() ChangeStream[T]
	// FullDocRequired sets fullDocument to "required": the server errors if the full document
	// cannot be provided (e.g. document was deleted between the change and the lookup).
	FullDocRequired() ChangeStream[T]

	// ResumeAfter resumes the stream starting after the given resume token.
	// Use st.ChangeEvent.ResumeToken from a previous callback to obtain the token.
	ResumeAfter(token bson.M) ChangeStream[T]
	// StartAfter is like ResumeAfter but can resume even after an "invalidate" event.
	StartAfter(token bson.M) ChangeStream[T]
	// Opts applies raw ChangeStreamOptions, merged on top of any builder settings.
	Opts(o *options.ChangeStreamOptions) ChangeStream[T]

	// Bsz sets the batch size for the change stream cursor.
	Bsz(n int32) ChangeStream[T]
	// Collation sets the collation for the change stream.
	Collation(c *options.Collation) ChangeStream[T]
	// Comment sets a comment for the change stream.
	Comment(s string) ChangeStream[T]
	// FullDocBefore sets the fullDocumentBeforeChange option.
	// Valid values: "off", "whenAvailable", "required".
	FullDocBefore(option string) ChangeStream[T]
	// MaxAwait sets the maximum amount of time the server will wait for new data before returning an empty batch.
	MaxAwait(d time.Duration) ChangeStream[T]
	// ShowExpanded enables expanded change stream events (MongoDB 6.0+).
	ShowExpanded(b bool) ChangeStream[T]
	// StartAtTime sets the operation time to start the change stream from.
	StartAtTime(t *primitive.Timestamp) ChangeStream[T]
	// Custom sets a custom BSON document to be added to the change stream command.
	Custom(m bson.M) ChangeStream[T]
	// CustomPipeline sets a custom BSON document to be added to the change stream aggregation pipeline.
	CustomPipeline(m bson.M) ChangeStream[T]

	// Stream opens the change stream and calls fn for every incoming event until fn
	// returns a non-nil error or the collection's context is cancelled.
	//
	// Example:
	//
	//	err := coll.Ctx(ctx).Watch().OnIstAndUpd().UpdLookup().
	//	    Stream(func(st maango.CsEvt[Order]) error {
	//	        fmt.Println(st.ChangeEvent.OperationType, st.ChangeEvent.FullDocument)
	//	        return nil
	//	    })
	Stream(fn func(st CsEvt[T]) error) error

	// Each is an alias for Stream.
	Each(fn func(st CsEvt[T]) error) error
}
