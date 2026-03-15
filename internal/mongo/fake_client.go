package mongo

import (
	"context"
	"errors"
	"time"

	mg "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FakeClientOption customizes behavior of NewFakeClient.
type FakeClientOption func(*fakeClientConfig)

type fakeClientConfig struct {
	dbName string
	uri    string
}

// WithFakeDatabase overrides the default database name for NewFakeClient.
func WithFakeDatabase(name string) FakeClientOption {
	return func(cfg *fakeClientConfig) {
		if name != "" {
			cfg.dbName = name
		}
	}
}

// WithFakeURI overrides the default MongoDB URI used when building fake clients.
// Note: the fake client never connects to the URI; it is only stored inside the client options.
func WithFakeURI(uri string) FakeClientOption {
	return func(cfg *fakeClientConfig) {
		if uri != "" {
			cfg.uri = uri
		}
	}
}

// NewFakeClient returns a Client implementation backed by disconnected mongo.Client instances.
// It is intended for unit tests that exercise builders and helpers without touching a live MongoDB.
func NewFakeClient(opts ...FakeClientOption) (Client, error) {
	cfg := fakeClientConfig{
		dbName: "testdb",
		uri:    "mongodb://localhost:27017",
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	build := func() (*mg.Client, error) {
		// ServerSelectionTimeout is set very short so unit-test operations fail fast
		// instead of hanging 30 s waiting for a server that is not running.
		return mg.Connect(context.Background(), options.Client().
			ApplyURI(cfg.uri).
			SetServerSelectionTimeout(5 * time.Millisecond))
	}
	write, err := build()
	if err != nil {
		return nil, err
	}
	read, err := build()
	if err != nil {
		return nil, err
	}
	return &fakeClient{
		read: read, write: write, db: cfg.dbName,
	}, nil
}

type fakeClient struct {
	read  *mg.Client
	write *mg.Client
	db    string
}

// Write returns the disconnected write client.
func (f *fakeClient) Write() *mg.Client { return f.write }

// Read returns the disconnected read client.
func (f *fakeClient) Read() *mg.Client { return f.read }

// DbName returns the configured test database name.
func (f *fakeClient) DbName() string { return f.db }

// Close disconnects both the fake read and write clients.
func (f *fakeClient) Close() error {
	var err error
	if f.read != nil {
		err = errors.Join(err, f.read.Disconnect(context.Background()))
	}
	if f.write != nil {
		err = errors.Join(err, f.write.Disconnect(context.Background()))
	}
	return err
}
