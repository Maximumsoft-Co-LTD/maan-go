package testutil

import (
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AssertEqual checks if two values are equal and reports test failure if not
func AssertEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, but got %v. %v", expected, actual, msgAndArgs)
	}
}

// AssertNotEqual checks if two values are not equal and reports test failure if they are
func AssertNotEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected values to be different, but both were %v. %v", expected, msgAndArgs)
	}
}

// AssertNil checks if a value is nil and reports test failure if not
func AssertNil(t *testing.T, value interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if value != nil {
		t.Errorf("Expected nil, but got %v. %v", value, msgAndArgs)
	}
}

// AssertNotNil checks if a value is not nil and reports test failure if it is
func AssertNotNil(t *testing.T, value interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if value == nil {
		t.Errorf("Expected non-nil value, but got nil. %v", msgAndArgs)
	}
}

// AssertNoError checks if an error is nil and reports test failure if not
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err != nil {
		t.Errorf("Expected no error, but got %v. %v", err, msgAndArgs)
	}
}

// AssertError checks if an error is not nil and reports test failure if it is
func AssertError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err == nil {
		t.Errorf("Expected an error, but got nil. %v", msgAndArgs)
	}
}

// AssertTrue checks if a condition is true and reports test failure if not
func AssertTrue(t *testing.T, condition bool, msgAndArgs ...interface{}) {
	t.Helper()
	if !condition {
		t.Errorf("Expected condition to be true, but it was false. %v", msgAndArgs)
	}
}

// AssertFalse checks if a condition is false and reports test failure if not
func AssertFalse(t *testing.T, condition bool, msgAndArgs ...interface{}) {
	t.Helper()
	if condition {
		t.Errorf("Expected condition to be false, but it was true. %v", msgAndArgs)
	}
}

// AssertContains checks if a slice contains a specific element
func AssertContains(t *testing.T, slice interface{}, element interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	sliceValue := reflect.ValueOf(slice)
	if sliceValue.Kind() != reflect.Slice {
		t.Errorf("Expected slice, but got %T. %v", slice, msgAndArgs)
		return
	}

	for i := 0; i < sliceValue.Len(); i++ {
		if reflect.DeepEqual(sliceValue.Index(i).Interface(), element) {
			return
		}
	}
	t.Errorf("Expected slice to contain %v, but it didn't. Slice: %v. %v", element, slice, msgAndArgs)
}

// AssertNotContains checks if a slice does not contain a specific element
func AssertNotContains(t *testing.T, slice interface{}, element interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	sliceValue := reflect.ValueOf(slice)
	if sliceValue.Kind() != reflect.Slice {
		t.Errorf("Expected slice, but got %T. %v", slice, msgAndArgs)
		return
	}

	for i := 0; i < sliceValue.Len(); i++ {
		if reflect.DeepEqual(sliceValue.Index(i).Interface(), element) {
			t.Errorf("Expected slice not to contain %v, but it did. Slice: %v. %v", element, slice, msgAndArgs)
			return
		}
	}
}

// AssertLen checks if a slice/map/string has the expected length
func AssertLen(t *testing.T, object interface{}, expectedLen int, msgAndArgs ...interface{}) {
	t.Helper()
	objectValue := reflect.ValueOf(object)
	actualLen := objectValue.Len()
	if actualLen != expectedLen {
		t.Errorf("Expected length %d, but got %d. Object: %v. %v", expectedLen, actualLen, object, msgAndArgs)
	}
}

// AssertGreater checks if a value is greater than another
func AssertGreater(t *testing.T, actual, expected interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	actualValue := reflect.ValueOf(actual)
	expectedValue := reflect.ValueOf(expected)

	if actualValue.Kind() != expectedValue.Kind() {
		t.Errorf("Cannot compare different types: %T vs %T. %v", actual, expected, msgAndArgs)
		return
	}

	switch actualValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if actualValue.Int() <= expectedValue.Int() {
			t.Errorf("Expected %v to be greater than %v. %v", actual, expected, msgAndArgs)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if actualValue.Uint() <= expectedValue.Uint() {
			t.Errorf("Expected %v to be greater than %v. %v", actual, expected, msgAndArgs)
		}
	case reflect.Float32, reflect.Float64:
		if actualValue.Float() <= expectedValue.Float() {
			t.Errorf("Expected %v to be greater than %v. %v", actual, expected, msgAndArgs)
		}
	default:
		t.Errorf("Cannot compare type %T. %v", actual, msgAndArgs)
	}
}

// AssertDocumentEqual compares two documents for equality, handling MongoDB-specific types
func AssertDocumentEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()

	// Handle primitive.ObjectID comparison
	if expectedID, ok := expected.(primitive.ObjectID); ok {
		if actualID, ok := actual.(primitive.ObjectID); ok {
			if expectedID != actualID {
				t.Errorf("Expected ObjectID %v, but got %v. %v", expectedID.Hex(), actualID.Hex(), msgAndArgs)
			}
			return
		}
	}

	// Handle time.Time comparison with tolerance
	if expectedTime, ok := expected.(time.Time); ok {
		if actualTime, ok := actual.(time.Time); ok {
			tolerance := time.Second
			if expectedTime.Sub(actualTime).Abs() > tolerance {
				t.Errorf("Expected time %v, but got %v (tolerance: %v). %v", expectedTime, actualTime, tolerance, msgAndArgs)
			}
			return
		}
	}

	// Handle BSON document comparison
	if expectedBSON, ok := expected.(bson.M); ok {
		if actualBSON, ok := actual.(bson.M); ok {
			AssertBSONEqual(t, expectedBSON, actualBSON, msgAndArgs...)
			return
		}
	}

	// Default to deep equal
	AssertEqual(t, expected, actual, msgAndArgs...)
}

// AssertBSONEqual compares two BSON documents for equality
func AssertBSONEqual(t *testing.T, expected, actual bson.M, msgAndArgs ...interface{}) {
	t.Helper()

	if len(expected) != len(actual) {
		t.Errorf("BSON documents have different lengths. Expected: %d, Actual: %d. %v", len(expected), len(actual), msgAndArgs)
		return
	}

	for key, expectedValue := range expected {
		actualValue, exists := actual[key]
		if !exists {
			t.Errorf("Key %q missing in actual BSON document. %v", key, msgAndArgs)
			continue
		}

		AssertDocumentEqual(t, expectedValue, actualValue, msgAndArgs...)
	}
}

// AssertTimestampRecent checks if a timestamp is recent (within last minute)
func AssertTimestampRecent(t *testing.T, timestamp time.Time, msgAndArgs ...interface{}) {
	t.Helper()
	now := time.Now().UTC()
	diff := now.Sub(timestamp)
	if diff < 0 || diff > time.Minute {
		t.Errorf("Expected timestamp to be recent (within 1 minute), but got %v (diff: %v). %v", timestamp, diff, msgAndArgs)
	}
}

// AssertObjectIDValid checks if an ObjectID is valid (not zero)
func AssertObjectIDValid(t *testing.T, id primitive.ObjectID, msgAndArgs ...interface{}) {
	t.Helper()
	if id.IsZero() {
		t.Errorf("Expected valid ObjectID, but got zero ObjectID. %v", msgAndArgs)
	}
}

// AssertSliceNotEmpty checks if a slice is not empty
func AssertSliceNotEmpty(t *testing.T, slice interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	sliceValue := reflect.ValueOf(slice)
	if sliceValue.Kind() != reflect.Slice {
		t.Errorf("Expected slice, but got %T. %v", slice, msgAndArgs)
		return
	}

	if sliceValue.Len() == 0 {
		t.Errorf("Expected non-empty slice, but got empty slice. %v", msgAndArgs)
	}
}

// AssertSliceEmpty checks if a slice is empty
func AssertSliceEmpty(t *testing.T, slice interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	sliceValue := reflect.ValueOf(slice)
	if sliceValue.Kind() != reflect.Slice {
		t.Errorf("Expected slice, but got %T. %v", slice, msgAndArgs)
		return
	}

	if sliceValue.Len() != 0 {
		t.Errorf("Expected empty slice, but got slice with %d elements. %v", sliceValue.Len(), msgAndArgs)
	}
}
