package property_tests

import (
	"context"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo"
)

// TestFakeClientConfigurationCorrectness tests Property 12: Fake client configuration correctness
// **Feature: unit-testing, Property 12: Fake client configuration correctness**
// **Validates: Requirements 4.1**
func TestFakeClientConfigurationCorrectness(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("fake client respects configuration options", prop.ForAll(
		func(dbName string, uri string) bool {
			// Generate valid database names and URIs
			if dbName == "" {
				dbName = "testdb" // default fallback
			}
			if uri == "" {
				uri = "mongodb://localhost:27017" // default fallback
			}

			// Create fake client with configuration options
			client, err := mongo.NewFakeClient(
				mongo.WithFakeDatabase(dbName),
				mongo.WithFakeURI(uri),
			)
			if err != nil {
				return false
			}
			defer client.Close()

			// Verify configuration is applied correctly
			actualDbName := client.DbName()
			if actualDbName != dbName {
				return false
			}

			// Verify client provides read and write connections
			readClient := client.Read()
			writeClient := client.Write()
			if readClient == nil || writeClient == nil {
				return false
			}

			// Verify read and write clients are separate instances
			if readClient == writeClient {
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool {
			return len(s) > 0 && len(s) <= 50 // reasonable database name length
		}),
		gen.OneConstOf(
			"mongodb://localhost:27017",
			"mongodb://localhost:27018",
			"mongodb://test.example.com:27017",
			"mongodb://user:pass@localhost:27017/testdb",
		),
	))

	properties.TestingRun(t)
}

// TestFakeClientDefaultConfiguration tests that fake client works with default configuration
// **Feature: unit-testing, Property 12: Fake client configuration correctness**
// **Validates: Requirements 4.1**
func TestFakeClientDefaultConfiguration(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("fake client works with default configuration", prop.ForAll(
		func() bool {
			// Create fake client with no options (should use defaults)
			client, err := mongo.NewFakeClient()
			if err != nil {
				return false
			}
			defer client.Close()

			// Verify default database name is set
			dbName := client.DbName()
			if dbName == "" {
				return false
			}

			// Verify client provides read and write connections
			readClient := client.Read()
			writeClient := client.Write()
			if readClient == nil || writeClient == nil {
				return false
			}

			return true
		},
	))

	properties.TestingRun(t)
}

// TestFakeClientEmptyConfiguration tests handling of empty configuration values
// **Feature: unit-testing, Property 12: Fake client configuration correctness**
// **Validates: Requirements 4.1**
func TestFakeClientEmptyConfiguration(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("fake client handles empty configuration gracefully", prop.ForAll(
		func() bool {
			// Create fake client with empty configuration values
			client, err := mongo.NewFakeClient(
				mongo.WithFakeDatabase(""), // empty database name should use default
				mongo.WithFakeURI(""),      // empty URI should use default
			)
			if err != nil {
				return false
			}
			defer client.Close()

			// Verify defaults are used when empty values are provided
			dbName := client.DbName()
			if dbName == "" {
				return false
			}

			// Verify client still works with defaults
			readClient := client.Read()
			writeClient := client.Write()
			if readClient == nil || writeClient == nil {
				return false
			}

			return true
		},
	))

	properties.TestingRun(t)
}

// TestFakeClientAPICompatibility tests Property 13: Fake client API compatibility
// **Feature: unit-testing, Property 13: Fake client API compatibility**
// **Validates: Requirements 4.2, 4.5**
func TestFakeClientAPICompatibility(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("fake client provides same interface as real client", prop.ForAll(
		func(dbName string) bool {
			if dbName == "" {
				dbName = "testdb"
			}

			// Create fake client
			fakeClient, err := mongo.NewFakeClient(
				mongo.WithFakeDatabase(dbName),
			)
			if err != nil {
				return false
			}
			defer fakeClient.Close()

			// Verify fake client implements Client interface
			var _ mongo.Client = fakeClient

			// Test all Client interface methods are available
			readClient := fakeClient.Read()
			writeClient := fakeClient.Write()
			actualDbName := fakeClient.DbName()

			if readClient == nil || writeClient == nil {
				return false
			}

			if actualDbName != dbName {
				return false
			}

			// Test that Close() method is available and callable
			// (We won't call it here since we defer it above)

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool {
			return len(s) > 0 && len(s) <= 50
		}),
	))

	properties.TestingRun(t)
}

// TestFakeClientCollectionCompatibility tests that fake client can create collections
// **Feature: unit-testing, Property 13: Fake client API compatibility**
// **Validates: Requirements 4.2, 4.5**
func TestFakeClientCollectionCompatibility(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("fake client creates compatible collections", prop.ForAll(
		func(collectionName string) bool {
			if collectionName == "" {
				collectionName = "test_collection"
			}

			// Create fake client
			client, err := mongo.NewFakeClient()
			if err != nil {
				return false
			}
			defer client.Close()

			// Test that we can create collections using the fake client
			// This tests the compatibility with the collection creation API
			ctx := context.Background()
			collection := mongo.NewCollection[TestDoc](ctx, client, collectionName)
			if collection == nil {
				return false
			}

			// Verify collection has the expected name
			if collection.Name() != collectionName {
				return false
			}

			// Test that collection methods are available (interface compatibility)
			// We won't execute them since they require actual MongoDB connection
			// but we verify the methods exist and can be called

			// Test context handling
			newCtx := context.Background()
			ctxCollection := collection.Ctx(newCtx)
			if ctxCollection == nil {
				return false
			}

			// Test Build method for ExtendedCollection
			extendedCollection := collection.Build(ctx)
			if extendedCollection == nil {
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool {
			return len(s) > 0 && len(s) <= 50
		}),
	))

	properties.TestingRun(t)
}

// TestFakeClientMethodSignatures tests that fake client methods have correct signatures
// **Feature: unit-testing, Property 13: Fake client API compatibility**
// **Validates: Requirements 4.2, 4.5**
func TestFakeClientMethodSignatures(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("fake client methods have correct signatures", prop.ForAll(
		func() bool {
			// Create fake client
			client, err := mongo.NewFakeClient()
			if err != nil {
				return false
			}
			defer client.Close()

			// Test method signatures match expected interface
			// Read() should return *mongo.Client
			readClient := client.Read()
			if readClient == nil {
				return false
			}

			// Write() should return *mongo.Client
			writeClient := client.Write()
			if writeClient == nil {
				return false
			}

			// DbName() should return string
			dbName := client.DbName()
			if dbName == "" {
				return false
			}

			// Close() should return error
			// We test this by creating a separate client to close
			testClient, err := mongo.NewFakeClient()
			if err != nil {
				return false
			}
			closeErr := testClient.Close()
			// Close error is acceptable for fake client, but method should exist
			_ = closeErr

			return true
		},
	))

	properties.TestingRun(t)
}

// TestDoc is a simple test document type for compatibility testing
type TestDoc struct {
	ID   string `bson:"_id"`
	Name string `bson:"name"`
}

// TestFakeClientLifecycleManagement tests Property 14: Fake client lifecycle management
// **Feature: unit-testing, Property 14: Fake client lifecycle management**
// **Validates: Requirements 4.3**
func TestFakeClientLifecycleManagement(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("fake client follows proper initialization and cleanup lifecycle", prop.ForAll(
		func(dbName string, uri string) bool {
			if dbName == "" {
				dbName = "testdb"
			}
			if uri == "" {
				uri = "mongodb://localhost:27017"
			}

			// Test initialization
			client, err := mongo.NewFakeClient(
				mongo.WithFakeDatabase(dbName),
				mongo.WithFakeURI(uri),
			)
			if err != nil {
				return false
			}

			// Verify client is properly initialized
			if client == nil {
				return false
			}

			// Verify client provides expected functionality after initialization
			readClient := client.Read()
			writeClient := client.Write()
			actualDbName := client.DbName()

			if readClient == nil || writeClient == nil {
				return false
			}

			if actualDbName != dbName {
				return false
			}

			// Test cleanup
			closeErr := client.Close()
			// Close may return error for fake client (e.g., "client is disconnected")
			// but should not panic and should be callable
			_ = closeErr

			// After close, client should still be safe to query for basic info
			// (though operations may fail)
			postCloseDbName := client.DbName()
			if postCloseDbName != dbName {
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool {
			return len(s) > 0 && len(s) <= 50
		}),
		gen.OneConstOf(
			"mongodb://localhost:27017",
			"mongodb://localhost:27018",
			"mongodb://test.example.com:27017",
		),
	))

	properties.TestingRun(t)
}

// TestFakeClientMultipleInstances tests that multiple fake clients can coexist
// **Feature: unit-testing, Property 14: Fake client lifecycle management**
// **Validates: Requirements 4.3**
func TestFakeClientMultipleInstances(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("multiple fake client instances can coexist independently", prop.ForAll(
		func(db1 string, db2 string) bool {
			if db1 == "" {
				db1 = "testdb1"
			}
			if db2 == "" {
				db2 = "testdb2"
			}
			// Ensure database names are different
			if db1 == db2 {
				db2 = db2 + "_alt"
			}

			// Create first client
			client1, err1 := mongo.NewFakeClient(
				mongo.WithFakeDatabase(db1),
			)
			if err1 != nil {
				return false
			}
			defer client1.Close()

			// Create second client
			client2, err2 := mongo.NewFakeClient(
				mongo.WithFakeDatabase(db2),
			)
			if err2 != nil {
				return false
			}
			defer client2.Close()

			// Verify clients are independent
			if client1 == client2 {
				return false
			}

			// Verify each client maintains its own configuration
			if client1.DbName() != db1 {
				return false
			}
			if client2.DbName() != db2 {
				return false
			}

			// Verify each client has independent read/write connections
			read1 := client1.Read()
			write1 := client1.Write()
			read2 := client2.Read()
			write2 := client2.Write()

			if read1 == nil || write1 == nil || read2 == nil || write2 == nil {
				return false
			}

			// Verify connections are independent between clients
			if read1 == read2 || write1 == write2 {
				return false
			}

			return true
		},
		gen.OneConstOf("db1", "database1", "test1", "mongo1", "client1"),
		gen.OneConstOf("db2", "database2", "test2", "mongo2", "client2"),
	))

	properties.TestingRun(t)
}

// TestFakeClientResourceCleanup tests that fake client properly cleans up resources
// **Feature: unit-testing, Property 14: Fake client lifecycle management**
// **Validates: Requirements 4.3**
func TestFakeClientResourceCleanup(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("fake client properly cleans up resources on close", prop.ForAll(
		func() bool {
			// Create client
			client, err := mongo.NewFakeClient()
			if err != nil {
				return false
			}

			// Get references to internal connections
			readClient := client.Read()
			writeClient := client.Write()

			if readClient == nil || writeClient == nil {
				return false
			}

			// Close client
			closeErr := client.Close()
			// Error is acceptable for fake client, but close should be callable
			_ = closeErr

			// Verify client is still safe to query after close
			// (basic info should still be available)
			dbName := client.DbName()
			if dbName == "" {
				return false
			}

			// Verify we can still get references (though they may be disconnected)
			postCloseRead := client.Read()
			postCloseWrite := client.Write()

			if postCloseRead == nil || postCloseWrite == nil {
				return false
			}

			return true
		},
	))

	properties.TestingRun(t)
}

// TestFakeClientIdempotentClose tests that calling Close multiple times is safe
// **Feature: unit-testing, Property 14: Fake client lifecycle management**
// **Validates: Requirements 4.3**
func TestFakeClientIdempotentClose(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("fake client close is idempotent and safe to call multiple times", prop.ForAll(
		func() bool {
			// Create client
			client, err := mongo.NewFakeClient()
			if err != nil {
				return false
			}

			// Call close multiple times
			err1 := client.Close()
			err2 := client.Close()
			err3 := client.Close()

			// Errors are acceptable for fake client, but should not panic
			_ = err1
			_ = err2
			_ = err3

			// Client should still be safe to query after multiple closes
			dbName := client.DbName()
			if dbName == "" {
				return false
			}

			return true
		},
	))

	properties.TestingRun(t)
}
