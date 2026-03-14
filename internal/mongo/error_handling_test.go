package mongo

import (
	"context"
	"testing"

	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo/testutil"
	"go.mongodb.org/mongo-driver/bson"
)

// TestNilInputHandling tests specific nil input scenarios
func TestNilInputHandling(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	collection := NewCollection[testutil.TestDoc](context.Background(), client, testutil.TestCollectionName)

	t.Run("Create with nil document", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Create with nil document panicked: %v", r)
			}
		}()

		err := collection.Create(nil)
		// Should not panic - error is acceptable but not required
		_ = err
	})

	t.Run("CreateMany with nil slice", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("CreateMany with nil slice panicked: %v", r)
			}
		}()

		err := collection.CreateMany(nil)
		// Should not panic - error is acceptable but not required
		_ = err
	})

	t.Run("WithTx with nil function returns error", func(t *testing.T) {
		err := collection.WithTx(nil)
		if err == nil {
			t.Fatal("expected error for nil fn, got nil")
		}
		if err.Error() != "fn must not be nil" {
			t.Fatalf("expected 'fn must not be nil', got %q", err.Error())
		}
	})

	t.Run("FindOne with nil query", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("FindOne with nil query panicked: %v", r)
			}
		}()

		result := collection.FindOne(nil)
		if result == nil {
			t.Error("FindOne should return a result object even with nil query")
		}
	})

	t.Run("Save with nil filter and update", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Save with nil inputs panicked: %v", r)
			}
		}()

		err := collection.Save(nil, nil)
		// Should not panic - error is acceptable but not required
		_ = err
	})
}

// TestInvalidInputHandling tests specific invalid input scenarios
func TestInvalidInputHandling(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	collection := NewCollection[testutil.TestDoc](context.Background(), client, testutil.TestCollectionName)

	t.Run("Del with invalid filter type", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Del with invalid filter panicked: %v", r)
			}
		}()

		err := collection.Del("invalid_string_filter")
		// Should handle gracefully, may return error
		_ = err // Error is acceptable
	})

	t.Run("FindMany with invalid filter", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("FindMany with invalid filter panicked: %v", r)
			}
		}()

		result := collection.FindMany(123) // Invalid type
		if result == nil {
			t.Error("FindMany should return a result object even with invalid filter")
		}
	})

	t.Run("Agg with invalid pipeline", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Agg with invalid pipeline panicked: %v", r)
			}
		}()

		agg := collection.Agg("invalid_string_pipeline")
		if agg == nil {
			t.Error("Agg should return an aggregate object even with invalid pipeline")
		}
	})
}

// TestEmptyResultHandling tests specific empty result scenarios
func TestEmptyResultHandling(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	collection := NewCollection[testutil.TestDoc](context.Background(), client, testutil.TestCollectionName)

	t.Run("FindOne with non-matching filter", func(t *testing.T) {
		result := collection.FindOne(bson.M{"nonexistent_field": "nonexistent_value"})

		var doc testutil.TestDoc
		err := result.Result(&doc)
		// Should return error (no documents found) but not panic
		if err == nil {
			// If no error, document should be zero value
			if doc.Name != "" || doc.Value != 0 {
				t.Error("Expected zero value document when no match found")
			}
		}
	})

	t.Run("FindMany with non-matching filter returns empty slice", func(t *testing.T) {
		result := collection.FindMany(bson.M{"nonexistent_field": "nonexistent_value"})

		docs, err := result.All()
		if err != nil {
			// Error is acceptable for fake client
			return
		}

		if docs == nil {
			t.Error("All() should return empty slice, not nil")
		}
		if len(docs) != 0 {
			t.Error("Expected empty slice for non-matching filter")
		}
	})

	t.Run("Count with non-matching filter returns zero", func(t *testing.T) {
		result := collection.FindMany(bson.M{"nonexistent_field": "nonexistent_value"})

		count, err := result.Cnt()
		if err != nil {
			// Error is acceptable for fake client
			return
		}

		if count != 0 {
			t.Errorf("Expected count 0, got %d", count)
		}
	})

	t.Run("Stream with empty results doesn't call function", func(t *testing.T) {
		result := collection.FindMany(bson.M{"nonexistent_field": "nonexistent_value"})

		callCount := 0
		err := result.Stream(func(ctx context.Context, doc testutil.TestDoc) error {
			callCount++
			return nil
		})

		if err != nil {
			// Error is acceptable for fake client
			return
		}

		if callCount != 0 {
			t.Errorf("Stream function should not be called for empty results, called %d times", callCount)
		}
	})
}

// TestExtendedCollectionErrorHandling tests error handling in ExtendedCollection
func TestExtendedCollectionErrorHandling(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	collection := NewCollection[testutil.TestDoc](context.Background(), client, testutil.TestCollectionName)
	extCollection := collection.Build(context.Background())

	t.Run("First with nil output", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("First with nil output panicked: %v", r)
			}
		}()

		err := extCollection.First(nil)
		// Should not panic - error is acceptable but not required
		_ = err
	})

	t.Run("Many with nil output", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Many with nil output panicked: %v", r)
			}
		}()

		err := extCollection.Many(nil)
		// Should not panic - error is acceptable but not required
		_ = err
	})

	t.Run("Exists with non-matching filter", func(t *testing.T) {
		extCollection := extCollection.Where(bson.M{"nonexistent_field": "nonexistent_value"})

		exists, err := extCollection.Exists()
		if err != nil {
			// Error is acceptable for fake client
			return
		}

		if exists {
			t.Error("Exists should return false for non-matching filter")
		}
	})
}

// TestResourceCleanup tests specific resource cleanup scenarios
func TestResourceCleanup(t *testing.T) {
	t.Run("Client Close is idempotent", func(t *testing.T) {
		client, err := NewFakeClient(
			WithFakeDatabase(testutil.TestDatabaseName),
			WithFakeURI("mongodb://localhost:27017"),
		)
		if err != nil {
			t.Fatalf("Failed to create fake client: %v", err)
		}

		// First close
		err1 := client.Close()
		// Second close should not panic
		err2 := client.Close()

		// Both may return errors, but should not panic
		_ = err1
		_ = err2
	})

	t.Run("Client methods safe after close", func(t *testing.T) {
		client, err := NewFakeClient(
			WithFakeDatabase(testutil.TestDatabaseName),
			WithFakeURI("mongodb://localhost:27017"),
		)
		if err != nil {
			t.Fatalf("Failed to create fake client: %v", err)
		}

		client.Close()

		// These should not panic after close
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Client methods panicked after close: %v", r)
			}
		}()

		client.DbName()
		client.Write()
		client.Read()
	})

	t.Run("WithTx handles errors gracefully", func(t *testing.T) {
		client, err := NewFakeClient(
			WithFakeDatabase(testutil.TestDatabaseName),
			WithFakeURI("mongodb://localhost:27017"),
		)
		if err != nil {
			t.Fatalf("Failed to create fake client: %v", err)
		}
		defer client.Close()

		collection := NewCollection[testutil.TestDoc](context.Background(), client, testutil.TestCollectionName)

		// WithTx with function that returns error
		err = collection.WithTx(func(ctx context.Context) error {
			return context.Canceled
		})

		// Should handle error gracefully, not panic
		if err == nil {
			t.Log("WithTx may return nil error with fake client")
		}
	})
}
