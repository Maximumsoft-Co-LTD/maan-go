package property_tests

import (
	"context"
	"testing"

	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo"
	"github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo/testutil"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"go.mongodb.org/mongo-driver/bson"
)

// TestQueryFilterCompositionCorrectness tests Property 5: Query filter composition correctness
// **Feature: unit-testing, Property 5: Query filter composition correctness**
// **Validates: Requirements 2.2**
func TestQueryFilterCompositionCorrectness(t *testing.T) {
	runner := NewPropertyTestRunner()

	// Property: For any set of valid filters, composing them using query builders
	// should produce logically correct combined filters
	property := prop.ForAll(
		func(fieldName1, fieldName2 string, value1, value2 interface{}, complexFilter bson.M) bool {
			// Skip test if field names are the same (last value should win)
			if fieldName1 == fieldName2 {
				return true
			}
			// Create test client and collection
			client, err := CreateTestClient()
			if err != nil {
				t.Logf("Failed to create test client: %v", err)
				return false
			}
			defer client.Close()

			ctx := context.Background()
			collection := mongo.NewCollection[testutil.TestDoc](ctx, client, testutil.TestCollectionName)
			extendedCollection := collection.Build(ctx)

			// Test 1: Sequential By() calls should accumulate fields
			step1 := extendedCollection.By(fieldName1, value1)

			step2 := step1.By(fieldName2, value2)
			filter2 := step2.GetFilter().(bson.M)

			// Verify first field is preserved and second is added
			expectedField1 := getExpectedBSONField(fieldName1)
			expectedField2 := getExpectedBSONField(fieldName2)

			if filter2[expectedField1] != value1 {
				t.Logf("First field not preserved: expected %v, got %v", value1, filter2[expectedField1])
				return false
			}

			if filter2[expectedField2] != value2 {
				t.Logf("Second field not added: expected %v, got %v", value2, filter2[expectedField2])
				return false
			}

			// Test 2: Where() should merge with existing By() filters
			step3 := step2.Where(complexFilter)
			filter3 := step3.GetFilter().(bson.M)

			// Verify By() fields are preserved
			if filter3[expectedField1] != value1 {
				t.Logf("By() field 1 not preserved after Where(): expected %v, got %v", value1, filter3[expectedField1])
				return false
			}

			if filter3[expectedField2] != value2 {
				t.Logf("By() field 2 not preserved after Where(): expected %v, got %v", value2, filter3[expectedField2])
				return false
			}

			// Verify Where() fields are added
			for key, value := range complexFilter {
				if !compareValues(filter3[key], value) {
					t.Logf("Where() field not added: key %s, expected %v, got %v", key, value, filter3[key])
					return false
				}
			}

			// Test 3: Filter composition should be commutative for non-conflicting fields
			// Build same filter in different order
			alt1 := extendedCollection.Where(complexFilter).By(fieldName1, value1).By(fieldName2, value2)
			altFilter1 := alt1.GetFilter().(bson.M)

			alt2 := extendedCollection.By(fieldName2, value2).Where(complexFilter).By(fieldName1, value1)
			altFilter2 := alt2.GetFilter().(bson.M)

			// Both should contain the same fields (order doesn't matter for maps)
			if len(altFilter1) != len(altFilter2) {
				t.Logf("Different composition orders produce different filter lengths: %d vs %d", len(altFilter1), len(altFilter2))
				return false
			}

			for key := range altFilter1 {
				if !compareValues(altFilter1[key], altFilter2[key]) {
					t.Logf("Different composition orders produce different values for key %s: %v vs %v", key, altFilter1[key], altFilter2[key])
					return false
				}
			}

			// Test 4: Empty operations should not affect filter
			emptyWhere := step2.Where(bson.M{})
			emptyFilter := emptyWhere.GetFilter().(bson.M)

			if len(emptyFilter) != len(filter2) {
				t.Logf("Empty Where() changed filter length: %d vs %d", len(filter2), len(emptyFilter))
				return false
			}

			for key := range filter2 {
				if !compareValues(emptyFilter[key], filter2[key]) {
					t.Logf("Empty Where() changed filter value for key %s: %v vs %v", key, filter2[key], emptyFilter[key])
					return false
				}
			}

			return true
		},
		genValidFieldName(),
		genValidFieldName(),
		genValidFieldValue(),
		genValidFieldValue(),
		genValidComplexFilter(),
	)

	runner.RunProperty(t, "Query filter composition correctness", property)
}

// genValidFieldName generates valid field names for TestDoc
func genValidFieldName() gopter.Gen {
	return gen.OneConstOf("Name", "Value", "Active", "ID")
}

// genValidFieldValue generates valid field values
func genValidFieldValue() gopter.Gen {
	return gen.OneGenOf(
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
		gen.IntRange(-1000, 1000),
		gen.Bool(),
		testutil.GenObjectID(),
	)
}

// genValidComplexFilter generates valid complex BSON filters that don't conflict with TestDoc fields
func genValidComplexFilter() gopter.Gen {
	return gen.OneGenOf(
		// Simple field filter (using non-conflicting field names)
		gopter.CombineGens(
			gen.OneConstOf("extra_field", "metadata", "tags", "custom_field"),
			genValidFieldValue(),
		).Map(func(values []interface{}) bson.M {
			return bson.M{values[0].(string): values[1]}
		}),

		// Range filter
		gen.IntRange(1, 100).Map(func(n int) bson.M {
			return bson.M{"range_field": bson.M{"$gte": n}}
		}),

		// Logical filter with non-conflicting field names
		gen.OneGenOf(
			gen.Const(bson.M{"$or": []bson.M{{"temp1": "val1"}, {"temp2": "val2"}}}),
			gen.Const(bson.M{"$and": []bson.M{{"temp3": "val3"}, {"temp4": "val4"}}}),
		),

		// Empty filter
		gen.Const(bson.M{}),
	)
}

// getExpectedBSONField maps struct field names to expected BSON field names
func getExpectedBSONField(fieldName string) string {
	switch fieldName {
	case "ID":
		return "_id"
	case "Name":
		return "name"
	case "Value":
		return "value"
	case "Active":
		return "active"
	default:
		// Convert to snake_case for unknown fields
		return toSnakeCase(fieldName)
	}
}

// toSnakeCase converts CamelCase to snake_case
func toSnakeCase(s string) string {
	if len(s) == 0 {
		return s
	}

	result := make([]rune, 0, len(s)*2)
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		if r >= 'A' && r <= 'Z' {
			result = append(result, r-'A'+'a')
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// compareValues safely compares two values, handling slices and maps
func compareValues(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// For simple types, use direct comparison
	switch va := a.(type) {
	case string, int, int32, int64, bool:
		return a == b
	case bson.M:
		vb, ok := b.(bson.M)
		if !ok {
			return false
		}
		return compareBSONM(va, vb)
	case []bson.M:
		vb, ok := b.([]bson.M)
		if !ok {
			return false
		}
		return compareBSONMSlice(va, vb)
	default:
		// For other types, try direct comparison (may panic for uncomparable types)
		defer func() {
			if recover() != nil {
				// If comparison panics, consider them different
			}
		}()
		return a == b
	}
}

// compareBSONM compares two bson.M maps
func compareBSONM(a, b bson.M) bool {
	if len(a) != len(b) {
		return false
	}
	for key, valueA := range a {
		valueB, exists := b[key]
		if !exists {
			return false
		}
		if !compareValues(valueA, valueB) {
			return false
		}
	}
	return true
}

// compareBSONMSlice compares two slices of bson.M
func compareBSONMSlice(a, b []bson.M) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !compareBSONM(a[i], b[i]) {
			return false
		}
	}
	return true
}
