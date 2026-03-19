package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	mg "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

const (
	defaultConnTimeout = 60 * time.Second
)

type client struct {
	write *mg.Client
	read  *mg.Client
	db    string
}

func (c *client) Write() *mg.Client { return c.write }
func (c *client) Read() *mg.Client  { return c.read }
func (c *client) DbName() string    { return c.db }

type clientConfig struct {
	writeURI     string
	readURI      string
	dbName       string
	timeout      time.Duration
	readPref     *readpref.ReadPref
	writeConcern *writeconcern.WriteConcern
	clientOpts   []func(*options.ClientOptions)
}

// Option configures a MongoDB client.
type Option func(*clientConfig)

// WithWriteURI sets the URI used for write operations (required).
func WithWriteURI(uri string) Option {
	return func(cfg *clientConfig) {
		cfg.writeURI = uri
	}
}

// WithReadURI sets the URI used for read operations. If not provided, the write URI is reused.
func WithReadURI(uri string) Option {
	return func(cfg *clientConfig) {
		cfg.readURI = uri
	}
}

// WithDatabase specifies the logical database name to use (required).
func WithDatabase(name string) Option {
	return func(cfg *clientConfig) {
		cfg.dbName = name
	}
}

// WithTimeout changes the connection timeout applied while establishing connections.
func WithTimeout(d time.Duration) Option {
	return func(cfg *clientConfig) {
		cfg.timeout = d
	}
}

// WithReadPreference overrides the read preference for the read client.
func WithReadPreference(rp *readpref.ReadPref) Option {
	return func(cfg *clientConfig) {
		cfg.readPref = rp
	}
}

// WithWriteConcern overrides the write concern for the write client.
func WithWriteConcern(wc *writeconcern.WriteConcern) Option {
	return func(cfg *clientConfig) {
		cfg.writeConcern = wc
	}
}

// WithClientOptions allows callers to tweak the underlying mongo client options before connection.
func WithClientOptions(mutator func(*options.ClientOptions)) Option {
	return func(cfg *clientConfig) {
		if mutator != nil {
			cfg.clientOpts = append(cfg.clientOpts, mutator)
		}
	}
}

// NewClient creates a MongoDB client pair with optional read/write separation.
func NewClient(ctx context.Context, opts ...Option) (Client, error) {
	cfg := clientConfig{
		timeout: defaultConnTimeout,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.writeURI == "" {
		return nil, errors.New("mongo: write URI is required")
	}
	if cfg.dbName == "" {
		return nil, errors.New("mongo: database name is required")
	}

	readURI := cfg.readURI
	if readURI == "" {
		readURI = cfg.writeURI
	}

	writeCli, err := connect(ctx, cfg.writeURI, &cfg, kindWrite)
	if err != nil {
		return nil, err
	}

	readCli := writeCli
	if readURI != cfg.writeURI {
		readCli, err = connect(ctx, readURI, &cfg, kindRead)
		if err != nil {
			_ = writeCli.Disconnect(context.Background())
			return nil, err
		}
	}

	return &client{write: writeCli, read: readCli, db: cfg.dbName}, nil
}

type connectionKind int

const (
	kindWrite connectionKind = iota
	kindRead
)

// connect dials a MongoDB server at uri and applies kind-specific options (read preference / write concern).
func connect(ctx context.Context, uri string, cfg *clientConfig, kind connectionKind) (*mg.Client, error) {
	clientOpts := options.Client().ApplyURI(uri)
	for _, mutator := range cfg.clientOpts {
		mutator(clientOpts)
	}
	switch kind {
	case kindRead:
		if cfg.readPref != nil {
			clientOpts.SetReadPreference(cfg.readPref)
		}
	case kindWrite:
		if cfg.writeConcern != nil {
			clientOpts.SetWriteConcern(cfg.writeConcern)
		}
	}

	cli, err := mg.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}
	return cli, nil
}

// Close disconnects both write and read clients. When read and write share the
// same underlying *mg.Client, only one disconnect call is issued.
func (c *client) Close() error {
	var closeErr error
	if c.write != nil {
		if err := c.write.Disconnect(context.Background()); err != nil {
			closeErr = errors.Join(closeErr, err)
		}
	}
	if c.read != nil && c.read != c.write {
		if err := c.read.Disconnect(context.Background()); err != nil {
			closeErr = errors.Join(closeErr, err)
		}
	}
	return closeErr
}

// StartTx begins a manual transaction and returns a TxSession.
func (c *client) StartTx(ctx context.Context) (TxSession, error) {
	return NewTransactionSession(ctx, c)
}

// WithTx runs fn inside an automatically managed transaction.
func (c *client) WithTx(ctx context.Context, fn func(ctx context.Context) error) (retErr error) {
	if fn == nil {
		return errors.New("fn must not be nil")
	}

	// Reuse existing session if detected (no nested transactions)
	if sess := mg.SessionFromContext(ctx); sess != nil {
		return fn(ctx)
	}

	sess, err := c.write.StartSession()
	if err != nil {
		return err
	}
	defer sess.EndSession(ctx)

	if err = sess.StartTransaction(); err != nil {
		return err
	}
	txCtx := mg.NewSessionContext(ctx, sess)
	defer func() {
		if r := recover(); r != nil {
			_ = sess.AbortTransaction(txCtx)
			retErr = fmt.Errorf("transaction panic: %v", r)
		}
	}()
	if err = fn(txCtx); err != nil {
		_ = sess.AbortTransaction(txCtx)
		return err
	}
	return sess.CommitTransaction(txCtx)
}
