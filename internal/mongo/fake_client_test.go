package mongo

import (
	"context"
	"testing"

	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo/testutil"
)

// TestFakeClientCreation tests basic fake client creation
func TestFakeClientCreation(t *testing.T) {
	client, err := NewFakeClient()
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	// Test default database name
	dbName := client.DbName()
	if dbName == "" {
		t.Error("Expected non-empty default database name")
	}

	// Test read and write clients
	readClient := client.Read()
	writeClient := client.Write()

	if readClient == nil {
		t.Error("Expected non-nil read client")
	}
	if writeClient == nil {
		t.Error("Expected non-nil write client")
	}
	if readClient == writeClient {
		t.Error("Expected separate read and write client instances")
	}
}

// TestFakeClientWithOptions tests fake client creation with options
func TestFakeClientWithOptions(t *testing.T) {
	testDB := "custom_test_db"
	testURI := "mongodb://custom:27017"

	client, err := NewFakeClient(
		WithFakeDatabase(testDB),
		WithFakeURI(testURI),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client with options: %v", err)
	}
	defer client.Close()

	// Verify custom database name
	if client.DbName() != testDB {
		t.Errorf("Expected database name %s, got %s", testDB, client.DbName())
	}
}

// TestFakeClientEmptyOptions tests fake client with empty option values
func TestFakeClientEmptyOptions(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(""), // empty should use default
		WithFakeURI(""),      // empty should use default
	)
	if err != nil {
		t.Fatalf("Failed to create fake client with empty options: %v", err)
	}
	defer client.Close()

	// Should use defaults when empty values provided
	dbName := client.DbName()
	if dbName == "" {
		t.Error("Expected default database name when empty string provided")
	}
}

// TestFakeClientInterfaceCompliance tests that fake client implements Client interface
func TestFakeClientInterfaceCompliance(t *testing.T) {
	client, err := NewFakeClient()
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	// Verify interface compliance
	var _ Client = client

	// Test all interface methods are callable
	readClient := client.Read()
	writeClient := client.Write()
	dbName := client.DbName()
	closeErr := client.Close()

	if readClient == nil {
		t.Error("Read() should return non-nil client")
	}
	if writeClient == nil {
		t.Error("Write() should return non-nil client")
	}
	if dbName == "" {
		t.Error("DbName() should return non-empty string")
	}
	// Close error is acceptable for fake client
	_ = closeErr
}

// TestFakeClientCollectionCreation tests creating collections with fake client
func TestFakeClientCollectionCreation(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	collectionName := "test_collection"

	// Test creating a collection
	collection := NewCollection[testutil.TestDoc](ctx, client, collectionName)
	if collection == nil {
		t.Fatal("Expected non-nil collection")
	}

	// Verify collection name
	if collection.Name() != collectionName {
		t.Errorf("Expected collection name %s, got %s", collectionName, collection.Name())
	}

	// Test that collection methods are available (interface compliance)
	ctxCollection := collection.Ctx(ctx)
	if ctxCollection == nil {
		t.Error("Expected non-nil context collection")
	}

	extendedCollection := collection.Build(ctx)
	if extendedCollection == nil {
		t.Error("Expected non-nil extended collection")
	}
}

// TestFakeClientMultipleInstances tests creating multiple fake client instances
func TestFakeClientMultipleInstances(t *testing.T) {
	client1, err1 := NewFakeClient(WithFakeDatabase("db1"))
	if err1 != nil {
		t.Fatalf("Failed to create first fake client: %v", err1)
	}
	defer client1.Close()

	client2, err2 := NewFakeClient(WithFakeDatabase("db2"))
	if err2 != nil {
		t.Fatalf("Failed to create second fake client: %v", err2)
	}
	defer client2.Close()

	// Verify clients are independent
	if client1 == client2 {
		t.Error("Expected different client instances")
	}

	if client1.DbName() == client2.DbName() {
		t.Error("Expected different database names")
	}

	// Verify independent connections
	if client1.Read() == client2.Read() {
		t.Error("Expected independent read clients")
	}
	if client1.Write() == client2.Write() {
		t.Error("Expected independent write clients")
	}
}

// TestFakeClientClose tests fake client close behavior
func TestFakeClientClose(t *testing.T) {
	client, err := NewFakeClient()
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}

	// Test close
	closeErr := client.Close()
	// Error is acceptable for fake client
	_ = closeErr

	// Test that client is still safe to query after close
	dbName := client.DbName()
	if dbName == "" {
		t.Error("Expected database name to still be available after close")
	}

	// Test multiple closes are safe
	closeErr2 := client.Close()
	closeErr3 := client.Close()
	_ = closeErr2
	_ = closeErr3
}

// TestFakeClientResourceManagement tests resource management patterns
func TestFakeClientResourceManagement(t *testing.T) {
	// Test that we can create and close many clients without issues
	for i := 0; i < 10; i++ {
		client, err := NewFakeClient()
		if err != nil {
			t.Fatalf("Failed to create fake client %d: %v", i, err)
		}

		// Verify client works
		if client.DbName() == "" {
			t.Errorf("Client %d has empty database name", i)
		}
		if client.Read() == nil || client.Write() == nil {
			t.Errorf("Client %d has nil connections", i)
		}

		// Close client
		client.Close()
	}
}

// TestFakeClientConfigurationEdgeCases tests edge cases in configuration
func TestFakeClientConfigurationEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		dbName   string
		uri      string
		expected string
	}{
		{
			name:     "normal values",
			dbName:   "testdb",
			uri:      "mongodb://localhost:27017",
			expected: "testdb",
		},
		{
			name:     "empty database name",
			dbName:   "",
			uri:      "mongodb://localhost:27017",
			expected: "testdb", // should use default
		},
		{
			name:     "whitespace database name",
			dbName:   "   ",
			uri:      "mongodb://localhost:27017",
			expected: "   ", // should preserve whitespace
		},
		{
			name:     "special characters in database name",
			dbName:   "test-db_123",
			uri:      "mongodb://localhost:27017",
			expected: "test-db_123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewFakeClient(
				WithFakeDatabase(tc.dbName),
				WithFakeURI(tc.uri),
			)
			if err != nil {
				t.Fatalf("Failed to create fake client: %v", err)
			}
			defer client.Close()

			actualDbName := client.DbName()
			if actualDbName != tc.expected {
				t.Errorf("Expected database name %q, got %q", tc.expected, actualDbName)
			}
		})
	}
}
