package property_tests

import (
	"testing"
	"time"

	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo"
	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo/testutil"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// PropertyTestRunner provides a standardized way to run property-based tests
type PropertyTestRunner struct {
	config testutil.PropertyTestConfig
}

// NewPropertyTestRunner creates a new property test runner with default configuration
func NewPropertyTestRunner() *PropertyTestRunner {
	return &PropertyTestRunner{
		config: testutil.DefaultPropertyTestConfig(),
	}
}

// WithConfig sets custom configuration for the property test runner
func (r *PropertyTestRunner) WithConfig(config testutil.PropertyTestConfig) *PropertyTestRunner {
	r.config = config
	return r
}

// RunProperty executes a property-based test with the configured parameters
func (r *PropertyTestRunner) RunProperty(t *testing.T, name string, property gopter.Prop) {
	t.Helper()

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = r.config.Iterations
	parameters.MaxShrinkCount = r.config.MaxShrinks
	parameters.Rng.Seed(r.config.RandomSeed)

	// Set timeout for the entire test
	if r.config.Timeout > 0 {
		timer := time.NewTimer(r.config.Timeout)
		defer timer.Stop()

		done := make(chan bool, 1)
		go func() {
			defer func() { done <- true }()

			properties := gopter.NewProperties(parameters)
			properties.Property(name, property)

			if !properties.Run(gopter.ConsoleReporter(false)) {
				t.Errorf("Property %s failed", name)
			}
		}()

		select {
		case <-done:
			// Test completed successfully
		case <-timer.C:
			t.Errorf("Property test %s timed out after %v", name, r.config.Timeout)
		}
	} else {
		properties := gopter.NewProperties(parameters)
		properties.Property(name, property)

		if !properties.Run(gopter.ConsoleReporter(false)) {
			t.Errorf("Property %s failed", name)
		}
	}
}

// CreateTestClient creates a fake client for property testing
func CreateTestClient() (mongo.Client, error) {
	return mongo.NewFakeClient(
		mongo.WithFakeDatabase(testutil.TestDatabaseName),
		mongo.WithFakeURI("mongodb://localhost:27017"),
	)
}

// PropertyTestHelper provides common utilities for property tests
type PropertyTestHelper struct {
	client mongo.Client
}

// NewPropertyTestHelper creates a new property test helper
func NewPropertyTestHelper() (*PropertyTestHelper, error) {
	client, err := CreateTestClient()
	if err != nil {
		return nil, err
	}

	return &PropertyTestHelper{
		client: client,
	}, nil
}

// Client returns the test client
func (h *PropertyTestHelper) Client() mongo.Client {
	return h.client
}

// Cleanup cleans up test resources
func (h *PropertyTestHelper) Cleanup() error {
	if h.client != nil {
		return h.client.Close()
	}
	return nil
}

// Common property test patterns

// ForAllValid creates a property that tests all valid inputs
func ForAllValid[T any](gen gopter.Gen, predicate func(T) bool) gopter.Prop {
	return prop.ForAll(predicate, gen)
}

// ForAllValidPair creates a property that tests all valid input pairs
func ForAllValidPair[T, U any](genT gopter.Gen, genU gopter.Gen, predicate func(T, U) bool) gopter.Prop {
	return prop.ForAll(predicate, genT, genU)
}

// ForAllValidTriple creates a property that tests all valid input triples
func ForAllValidTriple[T, U, V any](genT gopter.Gen, genU gopter.Gen, genV gopter.Gen, predicate func(T, U, V) bool) gopter.Prop {
	return prop.ForAll(predicate, genT, genU, genV)
}
