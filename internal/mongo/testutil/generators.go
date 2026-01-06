package testutil

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestDoc represents a basic test document for property testing
type TestDoc struct {
	ID     primitive.ObjectID `bson:"_id"`
	Name   string             `bson:"name"`
	Value  int                `bson:"value"`
	Active bool               `bson:"active"`
}

// DefaultTestDoc represents a test document with default interfaces
type DefaultTestDoc struct {
	ID        primitive.ObjectID `bson:"_id"`
	Name      string             `bson:"name"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

func (d *DefaultTestDoc) DefaultId() primitive.ObjectID {
	if d.ID.IsZero() {
		d.ID = primitive.NewObjectID()
	}
	return d.ID
}

func (d *DefaultTestDoc) DefaultCreatedAt() time.Time {
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now().UTC()
	}
	return d.CreatedAt
}

func (d *DefaultTestDoc) DefaultUpdatedAt() time.Time {
	if d.UpdatedAt.IsZero() {
		d.UpdatedAt = time.Now().UTC()
	}
	return d.UpdatedAt
}

// ComplexTestDoc represents a complex nested document for advanced testing
type ComplexTestDoc struct {
	ID       primitive.ObjectID `bson:"_id"`
	Metadata map[string]any     `bson:"metadata"`
	Tags     []string           `bson:"tags"`
	Nested   NestedDoc          `bson:"nested"`
}

type NestedDoc struct {
	Field1 string `bson:"field1"`
	Field2 int    `bson:"field2"`
}

// GenTestDoc generates random TestDoc instances
func GenTestDoc() gopter.Gen {
	return gopter.CombineGens(
		GenObjectID(),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
		gen.IntRange(-1000, 1000),
		gen.Bool(),
	).Map(func(values []interface{}) *TestDoc {
		return &TestDoc{
			ID:     values[0].(primitive.ObjectID),
			Name:   values[1].(string),
			Value:  values[2].(int),
			Active: values[3].(bool),
		}
	})
}

// GenDefaultTestDoc generates random DefaultTestDoc instances
func GenDefaultTestDoc() gopter.Gen {
	return gopter.CombineGens(
		GenObjectID(),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
		GenTime(),
		GenTime(),
	).Map(func(values []interface{}) *DefaultTestDoc {
		return &DefaultTestDoc{
			ID:        values[0].(primitive.ObjectID),
			Name:      values[1].(string),
			CreatedAt: values[2].(time.Time),
			UpdatedAt: values[3].(time.Time),
		}
	})
}

// GenComplexTestDoc generates random ComplexTestDoc instances
func GenComplexTestDoc() gopter.Gen {
	return gopter.CombineGens(
		GenObjectID(),
		GenStringMap(),
		GenStringSlice(),
		GenNestedDoc(),
	).Map(func(values []interface{}) *ComplexTestDoc {
		return &ComplexTestDoc{
			ID:       values[0].(primitive.ObjectID),
			Metadata: values[1].(map[string]any),
			Tags:     values[2].([]string),
			Nested:   values[3].(NestedDoc),
		}
	})
}

// GenNestedDoc generates random NestedDoc instances
func GenNestedDoc() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString(),
		gen.IntRange(0, 100),
	).Map(func(values []interface{}) NestedDoc {
		return NestedDoc{
			Field1: values[0].(string),
			Field2: values[1].(int),
		}
	})
}

// GenObjectID generates random MongoDB ObjectIDs
func GenObjectID() gopter.Gen {
	return gen.UInt64().Map(func(n uint64) primitive.ObjectID {
		return primitive.NewObjectID()
	})
}

// GenTime generates random time values
func GenTime() gopter.Gen {
	return gen.Int64Range(0, time.Now().Unix()).Map(func(timestamp int64) time.Time {
		return time.Unix(timestamp, 0).UTC()
	})
}

// GenStringMap generates random string maps
func GenStringMap() gopter.Gen {
	return gen.SliceOfN(3, gopter.CombineGens(
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 20 }),
		gen.OneGenOf(
			gen.AlphaString(),
			gen.Int(),
			gen.Bool(),
		),
	)).Map(func(pairs [][]interface{}) map[string]any {
		result := make(map[string]any)
		for _, pair := range pairs {
			if len(pair) >= 2 {
				key := pair[0].(string)
				value := pair[1]
				result[key] = value
			}
		}
		return result
	})
}

// GenStringSlice generates random string slices
func GenStringSlice() gopter.Gen {
	return gen.SliceOf(
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 20 }),
	).SuchThat(func(slice []string) bool {
		return len(slice) <= 10 // Keep slices small for testing
	})
}

// GenBSONFilter generates random MongoDB filter expressions
func GenBSONFilter() gopter.Gen {
	return gen.OneGenOf(
		// Simple equality filters
		GenSimpleFilter(),
		// Complex logical filters
		GenLogicalFilter(),
		// Empty filter
		gen.Const(bson.M{}),
	)
}

// GenSimpleFilter generates simple equality filters
func GenSimpleFilter() gopter.Gen {
	return gopter.CombineGens(
		gen.OneConstOf("name", "value", "active", "field1", "field2"),
		gen.OneGenOf(
			gen.AlphaString(),
			gen.Int(),
			gen.Bool(),
		),
	).Map(func(values []interface{}) bson.M {
		return bson.M{values[0].(string): values[1]}
	})
}

// GenLogicalFilter generates logical filters with $and, $or
func GenLogicalFilter() gopter.Gen {
	return gen.OneGenOf(
		// $and filter
		gen.SliceOfN(2, GenSimpleFilter()).Map(func(filters []bson.M) bson.M {
			return bson.M{"$and": filters}
		}),
		// $or filter
		gen.SliceOfN(2, GenSimpleFilter()).Map(func(filters []bson.M) bson.M {
			return bson.M{"$or": filters}
		}),
	)
}

// GenContext generates random contexts with values
func GenContext() gopter.Gen {
	return gen.OneGenOf(
		// Background context
		gen.Const(context.Background()),
		// Context with string value
		gopter.CombineGens(
			gen.AlphaString(),
			gen.AlphaString(),
		).Map(func(values []interface{}) context.Context {
			return context.WithValue(context.Background(), values[0].(string), values[1].(string))
		}),
		// Context with timeout
		gen.IntRange(1, 1000).Map(func(ms int) context.Context {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(ms)*time.Millisecond)
			// Note: In test generators, we intentionally don't call cancel immediately
			// as the context is meant to be used by the test. The test should handle cleanup.
			_ = cancel
			return ctx
		}),
	)
}

// GenCancelledContext generates cancelled contexts for testing cancellation
func GenCancelledContext() gopter.Gen {
	return gen.Const(func() context.Context {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		return ctx
	}())
}

// GenAggregationPipeline generates random aggregation pipelines
func GenAggregationPipeline() gopter.Gen {
	return gen.SliceOfN(1, GenPipelineStage()).Map(func(stages []bson.M) bson.A {
		result := make(bson.A, len(stages))
		for i, stage := range stages {
			result[i] = stage
		}
		return result
	})
}

// GenPipelineStage generates individual aggregation pipeline stages
func GenPipelineStage() gopter.Gen {
	return gen.OneGenOf(
		// $match stage
		GenSimpleFilter().Map(func(filter bson.M) bson.M {
			return bson.M{"$match": filter}
		}),
		// $project stage
		gen.Const(bson.M{"$project": bson.M{"name": 1, "value": 1}}),
		// $sort stage
		gen.OneConstOf(
			bson.M{"$sort": bson.M{"name": 1}},
			bson.M{"$sort": bson.M{"value": -1}},
		),
		// $limit stage
		gen.IntRange(1, 100).Map(func(limit int) bson.M {
			return bson.M{"$limit": limit}
		}),
	)
}

// GenCollectionName generates random collection names
func GenCollectionName() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0 && len(s) < 30
	}).Map(func(s string) string {
		return fmt.Sprintf("test_%s_%d", s, rand.Intn(10000))
	})
}

// GenDatabaseName generates random database names
func GenDatabaseName() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0 && len(s) < 20
	}).Map(func(s string) string {
		return fmt.Sprintf("testdb_%s", s)
	})
}

// GenContextKey generates random context keys for testing
func GenContextKey() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0 && len(s) < 20
	}).Map(func(s string) string {
		return "test_key_" + s
	})
}

// GenContextValue generates random context values for testing
func GenContextValue() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0 && len(s) < 50
	}).Map(func(s string) string {
		return "test_value_" + s
	})
}

// GenTestDocSlice generates random slices of TestDoc pointers
func GenTestDocSlice() gopter.Gen {
	return gen.SliceOfN(5, GenTestDoc()).SuchThat(func(docs []*TestDoc) bool {
		return len(docs) <= 10 // Keep slices small for testing
	})
}

// GenAggregationOptions generates random aggregation options
func GenAggregationOptions() gopter.Gen {
	return gopter.CombineGens(
		gen.Bool(),              // AllowDiskUse
		gen.Int32Range(1, 1000), // BatchSize
	).Map(func(values []interface{}) map[string]interface{} {
		return map[string]interface{}{
			"allowDiskUse": values[0].(bool),
			"batchSize":    values[1].(int32),
		}
	})
}

// GenErrorMessage generates random error messages for testing
func GenErrorMessage() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0 && len(s) < 100
	}).Map(func(s string) string {
		return "test_error_" + s
	})
}

// GenAlphaString generates random alphabetic strings for testing
func GenAlphaString() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0 && len(s) < 50
	})
}
