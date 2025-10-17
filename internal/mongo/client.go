package mongo

import (
	"context"
	"errors"
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

	baseCtx := ctx
	if baseCtx == nil {
		baseCtx = context.Background()
	}

	timedCtx, cancel := context.WithTimeout(baseCtx, cfg.timeout)
	defer cancel()

	writeCli, err := connect(timedCtx, cfg.writeURI, &cfg, kindWrite)
	if err != nil {
		return nil, err
	}

	readCli := writeCli
	if readURI != cfg.writeURI {
		readCli, err = connect(timedCtx, readURI, &cfg, kindRead)
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

func connect(timedCtx context.Context, uri string, cfg *clientConfig, kind connectionKind) (*mg.Client, error) {
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

	cli, err := mg.Connect(timedCtx, clientOpts)
	if err != nil {
		return nil, err
	}
	return cli, nil
}

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
