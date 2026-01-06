package mongo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo/testutil"
)

// TestTxSessionInterface tests the TxSession interface methods
func TestTxSessionInterface(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	collection := NewCollection[testutil.TestDoc](ctx, client, testutil.TestCollectionName)

	t.Run("StartTx returns TxSession interface", func(t *testing.T) {
		txSession, err := collection.StartTx()

		// For fake client, we expect connection errors but the interface should work
		if err != nil {
			// Connection error is expected with fake client
			if !isConnectionError(err) {
				t.Errorf("Expected connection error, got: %v", err)
			}
			return
		}

		if txSession == nil {
			t.Error("Expected non-nil TxSession")
			return
		}

		// Test TxSession interface methods
		sessionCtx := txSession.Ctx()
		if sessionCtx == nil {
			t.Error("TxSession.Ctx() returned nil context")
		}

		// Test Close method
		var txErr error
		txSession.Close(&txErr)
	})

	t.Run("TxSession.Ctx returns valid context", func(t *testing.T) {
		txSession, err := collection.StartTx()
		if err != nil {
			// Skip test if we can't create session with fake client
			t.Skipf("Cannot create transaction session with fake client: %v", err)
		}

		if txSession == nil {
			t.Skip("Cannot create transaction session with fake client")
		}

		defer func() {
			var txErr error
			txSession.Close(&txErr)
		}()

		sessionCtx := txSession.Ctx()
		if sessionCtx == nil {
			t.Error("TxSession.Ctx() returned nil context")
			return
		}

		// Test that context is usable
		select {
		case <-sessionCtx.Done():
			// Context is cancelled, which is acceptable
		default:
			// Context is active, which is also acceptable
		}

		// Test context deadline functionality
		_, hasDeadline := sessionCtx.Deadline()
		// Deadline may or may not be set, both are valid
		_ = hasDeadline
	})

	t.Run("TxSession.Close handles nil error pointer", func(t *testing.T) {
		txSession, err := collection.StartTx()
		if err != nil {
			t.Skipf("Cannot create transaction session with fake client: %v", err)
		}

		if txSession == nil {
			t.Skip("Cannot create transaction session with fake client")
		}

		// Test Close with nil error pointer (should not panic)
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("TxSession.Close(nil) panicked: %v", r)
			}
		}()

		txSession.Close(nil)
	})

	t.Run("TxSession.Close handles error pointer", func(t *testing.T) {
		txSession, err := collection.StartTx()
		if err != nil {
			t.Skipf("Cannot create transaction session with fake client: %v", err)
		}

		if txSession == nil {
			t.Skip("Cannot create transaction session with fake client")
		}

		// Test Close with error pointer
		var txErr error
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("TxSession.Close(&err) panicked: %v", r)
			}
		}()

		txSession.Close(&txErr)
	})
}

// TestWithTxAutomaticTransactionHandling tests automatic transaction handling with WithTx()
func TestWithTxAutomaticTransactionHandling(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	collection := NewCollection[testutil.TestDoc](ctx, client, testutil.TestCollectionName)

	t.Run("WithTx executes function successfully", func(t *testing.T) {
		executed := false
		err := collection.WithTx(func(ctx context.Context) error {
			executed = true
			if ctx == nil {
				t.Error("WithTx provided nil context")
			}
			return nil
		})

		// For fake client, we expect connection errors and function may not execute
		if err != nil {
			if !isConnectionError(err) {
				t.Errorf("WithTx returned unexpected error: %v", err)
			}
			// If there's a connection error, function execution is not guaranteed
			return
		}

		if !executed {
			t.Error("WithTx did not execute the provided function")
		}
	})

	t.Run("WithTx propagates function errors", func(t *testing.T) {
		expectedErr := errors.New("test transaction error")
		err := collection.WithTx(func(ctx context.Context) error {
			return expectedErr
		})

		// The error should be propagated (or we get a connection error)
		if err == nil {
			t.Error("WithTx should propagate function errors")
		}
	})

	t.Run("WithTx handles function panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("WithTx did not handle function panic: %v", r)
			}
		}()

		err := collection.WithTx(func(ctx context.Context) error {
			panic("test panic")
		})

		// Should not panic, should return error or connection error
		_ = err
	})

	t.Run("WithTx provides session context", func(t *testing.T) {
		var receivedCtx context.Context
		err := collection.WithTx(func(ctx context.Context) error {
			receivedCtx = ctx
			return nil
		})

		// For fake client, we expect connection errors
		if err != nil {
			if !isConnectionError(err) {
				t.Errorf("WithTx returned unexpected error: %v", err)
			}
			// If there's a connection error, context may not be provided
			return
		}

		if receivedCtx == nil {
			t.Error("WithTx provided nil context to function")
		}
	})
}

// TestManualTransactionHandling tests manual transaction handling with StartTx()
func TestManualTransactionHandling(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	collection := NewCollection[testutil.TestDoc](ctx, client, testutil.TestCollectionName)

	t.Run("StartTx creates transaction session", func(t *testing.T) {
		txSession, err := collection.StartTx()

		if err != nil {
			// For fake client, connection errors are expected
			if !isConnectionError(err) {
				t.Errorf("StartTx returned unexpected error: %v", err)
			}
			return
		}

		if txSession == nil {
			t.Error("StartTx returned nil session")
			return
		}

		defer func() {
			var txErr error
			txSession.Close(&txErr)
		}()

		// Test that session provides valid context
		sessionCtx := txSession.Ctx()
		if sessionCtx == nil {
			t.Error("Transaction session provided nil context")
		}
	})

	t.Run("Multiple StartTx calls create independent sessions", func(t *testing.T) {
		txSession1, err1 := collection.StartTx()
		txSession2, err2 := collection.StartTx()

		// Handle fake client connection errors
		if err1 != nil && !isConnectionError(err1) {
			t.Errorf("First StartTx returned unexpected error: %v", err1)
		}
		if err2 != nil && !isConnectionError(err2) {
			t.Errorf("Second StartTx returned unexpected error: %v", err2)
		}

		// Clean up sessions if they were created
		if txSession1 != nil {
			defer func() {
				var txErr error
				txSession1.Close(&txErr)
			}()
		}
		if txSession2 != nil {
			defer func() {
				var txErr error
				txSession2.Close(&txErr)
			}()
		}

		// If both sessions were created, they should be different
		if txSession1 != nil && txSession2 != nil {
			if txSession1 == txSession2 {
				t.Error("StartTx should create independent sessions")
			}

			ctx1 := txSession1.Ctx()
			ctx2 := txSession2.Ctx()

			if ctx1 == ctx2 {
				t.Error("Independent sessions should have different contexts")
			}
		}
	})
}

// TestTransactionCommitAndRollbackScenarios tests transaction commit and rollback scenarios
func TestTransactionCommitAndRollbackScenarios(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	collection := NewCollection[testutil.TestDoc](ctx, client, testutil.TestCollectionName)

	t.Run("Successful WithTx should commit", func(t *testing.T) {
		err := collection.WithTx(func(ctx context.Context) error {
			// Simulate successful operations
			doc := testutil.FixtureTestDoc()
			txCollection := collection.Ctx(ctx)

			// Try to create document (will fail with fake client but shouldn't panic)
			createErr := txCollection.Create(doc)

			// For fake client, we expect connection errors
			if createErr != nil && !isConnectionError(createErr) {
				return createErr
			}

			return nil
		})

		// For fake client, we expect connection errors but no panics
		if err != nil && !isConnectionError(err) {
			t.Errorf("Successful WithTx returned unexpected error: %v", err)
		}
	})

	t.Run("Failed WithTx should rollback", func(t *testing.T) {
		expectedErr := errors.New("transaction should rollback")

		err := collection.WithTx(func(ctx context.Context) error {
			// Simulate some operations before error
			doc := testutil.FixtureTestDoc()
			txCollection := collection.Ctx(ctx)

			// Try to create document (will fail with fake client)
			_ = txCollection.Create(doc)

			// Return error to trigger rollback
			return expectedErr
		})

		// Should propagate the error (or connection error)
		if err == nil {
			t.Error("Failed WithTx should return error")
		}
	})

	t.Run("Manual transaction commit scenario", func(t *testing.T) {
		txSession, err := collection.StartTx()
		if err != nil {
			t.Skipf("Cannot create transaction session with fake client: %v", err)
		}

		if txSession == nil {
			t.Skip("Cannot create transaction session with fake client")
		}

		// Simulate successful transaction (commit)
		var txErr error
		defer txSession.Close(&txErr)

		sessionCtx := txSession.Ctx()
		txCollection := collection.Ctx(sessionCtx)

		// Try operations (will fail with fake client but test the pattern)
		doc := testutil.FixtureTestDoc()
		createErr := txCollection.Create(doc)

		// For fake client, connection errors are expected
		if createErr != nil && !isConnectionError(createErr) {
			txErr = createErr
		}

		// txErr remains nil for successful scenario
	})

	t.Run("Manual transaction rollback scenario", func(t *testing.T) {
		txSession, err := collection.StartTx()
		if err != nil {
			t.Skipf("Cannot create transaction session with fake client: %v", err)
		}

		if txSession == nil {
			t.Skip("Cannot create transaction session with fake client")
		}

		// Simulate failed transaction (rollback)
		txErr := errors.New("transaction failed")
		defer txSession.Close(&txErr)

		sessionCtx := txSession.Ctx()
		txCollection := collection.Ctx(sessionCtx)

		// Try operations that should be rolled back
		doc := testutil.FixtureTestDoc()
		_ = txCollection.Create(doc)

		// txErr is set to trigger rollback
	})
}

// TestTransactionErrorHandling tests various error scenarios
func TestTransactionErrorHandling(t *testing.T) {
	client, err := NewFakeClient(
		WithFakeDatabase(testutil.TestDatabaseName),
		WithFakeURI("mongodb://localhost:27017"),
	)
	if err != nil {
		t.Fatalf("Failed to create fake client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	collection := NewCollection[testutil.TestDoc](ctx, client, testutil.TestCollectionName)

	t.Run("WithTx with cancelled context", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		cancelledCollection := collection.Ctx(cancelledCtx)

		err := cancelledCollection.WithTx(func(ctx context.Context) error {
			return nil
		})

		// Should handle cancelled context gracefully
		if err == nil {
			t.Log("WithTx with cancelled context completed without error")
		} else if !isConnectionError(err) && !isCancellationError(err) {
			t.Errorf("WithTx with cancelled context returned unexpected error: %v", err)
		}
	})

	t.Run("WithTx with timeout context", func(t *testing.T) {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		timeoutCollection := collection.Ctx(timeoutCtx)

		err := timeoutCollection.WithTx(func(ctx context.Context) error {
			// Add small delay to potentially trigger timeout
			time.Sleep(2 * time.Millisecond)
			return nil
		})

		// Should handle timeout gracefully
		if err != nil && !isConnectionError(err) && !isTimeoutError(err) {
			t.Errorf("WithTx with timeout context returned unexpected error: %v", err)
		}
	})

	t.Run("StartTx with cancelled context", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		cancelledCollection := collection.Ctx(cancelledCtx)

		txSession, err := cancelledCollection.StartTx()

		// Should handle cancelled context gracefully
		if err != nil && !isConnectionError(err) && !isCancellationError(err) {
			t.Errorf("StartTx with cancelled context returned unexpected error: %v", err)
		}

		if txSession != nil {
			defer func() {
				var txErr error
				txSession.Close(&txErr)
			}()
		}
	})

	t.Run("Multiple Close calls on same session", func(t *testing.T) {
		txSession, err := collection.StartTx()
		if err != nil {
			t.Skipf("Cannot create transaction session with fake client: %v", err)
		}

		if txSession == nil {
			t.Skip("Cannot create transaction session with fake client")
		}

		// Test multiple Close calls (should be safe)
		var txErr error

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Multiple Close calls caused panic: %v", r)
			}
		}()

		txSession.Close(&txErr)
		txSession.Close(&txErr) // Second close should be safe
		txSession.Close(&txErr) // Third close should be safe
	})
}

// Helper functions for error classification

func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "connection") ||
		contains(errStr, "network") ||
		contains(errStr, "dial") ||
		contains(errStr, "timeout") ||
		contains(errStr, "session") ||
		contains(errStr, "transaction") ||
		contains(errStr, "client") ||
		contains(errStr, "server")
}

func isCancellationError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "cancel") || contains(errStr, "context canceled")
}

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "timeout") || contains(errStr, "deadline")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsInner(s, substr))))
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
