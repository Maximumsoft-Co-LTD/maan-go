# maan-go Testing Guide

คู่มือการทดสอบสำหรับไลบรารี maan-go ที่ครอบคลุมการทดสอบแบบต่างๆ ตั้งแต่ unit tests, integration tests, property-based tests และ testing utilities

## โครงสร้างการทดสอบ

### ประเภทของการทดสอบ

1. **Unit Tests** - ทดสอบฟังก์ชันแต่ละส่วนแยกกัน
2. **Integration Tests** - ทดสอบการทำงานร่วมกับ MongoDB จริง
3. **Property-Based Tests** - ทดสอบคุณสมบัติของระบบด้วยข้อมูลสุ่ม
4. **Fake Client Tests** - ทดสอบโดยไม่ต้องใช้ MongoDB จริง

### โครงสร้างไฟล์ทดสอบ

```
internal/mongo/
├── *_test.go                    # Unit tests ควบคู่กับ implementation
├── client_integration_test.go   # Integration tests ต้องการ MongoDB จริง
├── fake_client_test.go         # Tests สำหรับ fake client
├── property_tests/             # Property-based tests
│   ├── base_test.go           # Base utilities สำหรับ property tests
│   ├── *_test.go              # Property tests แต่ละหัวข้อ
└── testutil/                   # Testing utilities
    ├── assertions.go          # Custom assertion functions
    ├── fixtures.go           # Test fixtures และ scenarios
    └── generators.go         # Property test generators
```

## การรันเทสต์

### Unit Tests

```bash
# รันทุก unit tests
go test ./...

# รันด้วย verbose output
go test -v ./...

# รันเฉพาะ package เดียว
go test ./internal/mongo

# รันเฉพาะ test function
go test -run TestCollectionCreate ./internal/mongo
```

### Integration Tests

```bash
# ต้องตั้ง environment variable สำหรับ MongoDB URI
export MONGO_INTEGRATION_URI="mongodb://localhost:27017"

# รัน integration tests
go test ./internal/mongo -run ClientRoundTrip

# รันด้วย verbose output
go test -v ./internal/mongo -run Integration
```

### Property-Based Tests

```bash
# รัน property tests ทั้งหมด
go test ./internal/mongo/property_tests

# รันด้วยการตั้งค่า iterations
go test ./internal/mongo/property_tests -args -test.count=200

# รันเฉพาะ property test เดียว
go test -run TestCRUDConsistency ./internal/mongo/property_tests
```

### Race Condition Tests

```bash
# ทดสอบ race conditions
go test -race ./...

# ทดสอบ thread safety
go test -race ./internal/mongo/property_tests -run ThreadSafety
```

### Coverage

```bash
# สร้าง coverage report
go test -coverprofile=coverage.out ./...

# ดู coverage ใน HTML
go tool cover -html=coverage.out

# ดู coverage percentage
go tool cover -func=coverage.out
```

## Testing Utilities

### Custom Assertions (`testutil/assertions.go`)

```go
import "github.com/Maximumsoft-Co-LTD/maan-go/internal/mongo/testutil"

func TestExample(t *testing.T) {
    // Basic assertions
    testutil.AssertEqual(t, expected, actual, "values should be equal")
    testutil.AssertNotNil(t, value, "value should not be nil")
    testutil.AssertNoError(t, err, "operation should succeed")
    
    // MongoDB-specific assertions
    testutil.AssertObjectIDValid(t, id, "ID should be valid")
    testutil.AssertTimestampRecent(t, timestamp, "timestamp should be recent")
    testutil.AssertDocumentEqual(t, expectedDoc, actualDoc, "documents should match")
    
    // Collection assertions
    testutil.AssertLen(t, slice, 5, "slice should have 5 elements")
    testutil.AssertContains(t, slice, element, "slice should contain element")
    testutil.AssertSliceNotEmpty(t, results, "results should not be empty")
}
```

### Test Fixtures (`testutil/fixtures.go`)

```go
// สร้าง test documents
doc := testutil.FixtureTestDoc()
defaultDoc := testutil.FixtureDefaultTestDoc()
complexDoc := testutil.FixtureComplexTestDoc()

// สร้าง bulk documents
docs := testutil.FixtureTestDocs(10) // สร้าง 10 documents

// สร้าง contexts สำหรับทดสอบ
ctx := testutil.FixtureContextWithTimeout(5 * time.Second)
cancelledCtx := testutil.FixtureCancelledContext()
```

### Property Test Generators (`testutil/generators.go`)

```go
import "github.com/leanovate/gopter"

// สร้าง generators สำหรับ property tests
docGen := testutil.GenTestDoc()
filterGen := testutil.GenBSONFilter()
pipelineGen := testutil.GenAggregationPipeline()

// ใช้ใน property tests
property := prop.ForAll(func(doc *testutil.TestDoc) bool {
    // Test logic here
    return true
}, docGen)
```

## การเขียน Unit Tests

### Basic Unit Test Pattern

```go
func TestCollectionCreate(t *testing.T) {
    // Setup
    client, err := mongo.NewFakeClient()
    testutil.AssertNoError(t, err, "fake client creation should succeed")
    defer client.Close()
    
    coll := mongo.NewColl[testutil.TestDoc](context.Background(), client, "test")
    
    // Test data
    doc := testutil.FixtureTestDoc()
    
    // Execute
    err = coll.Create(doc)
    
    // Assert
    testutil.AssertNoError(t, err, "create should succeed")
    testutil.AssertObjectIDValid(t, doc.ID, "ID should be set")
}
```

### Table-Driven Tests

```go
func TestCollectionFindOne(t *testing.T) {
    tests := []struct {
        name     string
        filter   bson.M
        expected *testutil.TestDoc
        wantErr  bool
    }{
        {
            name:     "find by ID",
            filter:   bson.M{"_id": primitive.NewObjectID()},
            expected: testutil.FixtureTestDoc(),
            wantErr:  false,
        },
        {
            name:    "not found",
            filter:  bson.M{"name": "nonexistent"},
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## การเขียน Integration Tests

### Integration Test Pattern

```go
func TestClientRoundTrip(t *testing.T) {
    uri := os.Getenv("MONGO_INTEGRATION_URI")
    if uri == "" {
        t.Skip("MONGO_INTEGRATION_URI not set, skipping integration test")
    }
    
    ctx := context.Background()
    client, err := mongo.NewClient(
        ctx,
        mongo.WithWriteURI(uri),
        mongo.WithDatabase("test_integration"),
    )
    testutil.AssertNoError(t, err, "client creation should succeed")
    defer client.Close()
    
    // Test with real MongoDB
    coll := mongo.NewColl[testutil.TestDoc](ctx, client, "integration_test")
    
    // Cleanup
    defer func() {
        _ = coll.Drop()
    }()
    
    // Test operations
    doc := testutil.FixtureTestDoc()
    err = coll.Create(doc)
    testutil.AssertNoError(t, err, "create should succeed")
}
```

## การเขียน Property-Based Tests

### Property Test Pattern

```go
func TestCRUDConsistency(t *testing.T) {
    runner := property_tests.NewPropertyTestRunner()
    
    property := prop.ForAll(func(doc *testutil.TestDoc) bool {
        // Setup
        helper, err := property_tests.NewPropertyTestHelper()
        if err != nil {
            return false
        }
        defer helper.Cleanup()
        
        coll := mongo.NewColl[testutil.TestDoc](
            context.Background(), 
            helper.Client(), 
            testutil.TestCollectionName,
        )
        
        // Test: Create then Find should return same document
        if err := coll.Create(doc); err != nil {
            return false
        }
        
        var found testutil.TestDoc
        err = coll.FindOne(bson.M{"_id": doc.ID}).Result(&found)
        if err != nil {
            return false
        }
        
        // Property: Created document should equal found document
        return doc.ID == found.ID && 
               doc.Name == found.Name && 
               doc.Value == found.Value
    }, testutil.GenTestDoc())
    
    runner.RunProperty(t, "CRUD operations should be consistent", property)
}
```

### Custom Property Test Configuration

```go
func TestWithCustomConfig(t *testing.T) {
    config := testutil.PropertyTestConfig{
        Iterations: 500,           // รัน 500 iterations
        MaxShrinks: 200,          // ลด input สูงสุด 200 ครั้ง
        RandomSeed: 12345,        // ใช้ seed คงที่เพื่อ reproducible results
        Timeout:    60 * time.Second, // timeout 60 วินาที
    }
    
    runner := property_tests.NewPropertyTestRunner().WithConfig(config)
    
    // Run property test with custom config
    runner.RunProperty(t, "custom test", property)
}
```

## การทดสอบ Error Handling

### Error Scenarios

```go
func TestErrorHandling(t *testing.T) {
    tests := []struct {
        name        string
        setup       func() (mongo.Client, error)
        expectError bool
        errorType   string
    }{
        {
            name: "cancelled context",
            setup: func() (mongo.Client, error) {
                ctx := testutil.FixtureCancelledContext()
                return mongo.NewClient(ctx, mongo.WithWriteURI("mongodb://localhost:27017"))
            },
            expectError: true,
            errorType:   "context cancelled",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            client, err := tt.setup()
            
            if tt.expectError {
                testutil.AssertError(t, err, "should return error")
            } else {
                testutil.AssertNoError(t, err, "should not return error")
                defer client.Close()
            }
        })
    }
}
```

## การทดสอบ Concurrency

### Thread Safety Tests

```go
func TestThreadSafety(t *testing.T) {
    client, err := mongo.NewFakeClient()
    testutil.AssertNoError(t, err, "fake client creation should succeed")
    defer client.Close()
    
    coll := mongo.NewColl[testutil.TestDoc](context.Background(), client, "thread_test")
    
    const numGoroutines = 10
    const numOperations = 100
    
    var wg sync.WaitGroup
    errors := make(chan error, numGoroutines*numOperations)
    
    // Start multiple goroutines performing operations
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            for j := 0; j < numOperations; j++ {
                doc := &testutil.TestDoc{
                    Name:  fmt.Sprintf("doc_%d_%d", id, j),
                    Value: id*numOperations + j,
                }
                
                if err := coll.Create(doc); err != nil {
                    errors <- err
                }
            }
        }(i)
    }
    
    wg.Wait()
    close(errors)
    
    // Check for errors
    for err := range errors {
        t.Errorf("Concurrent operation failed: %v", err)
    }
}
```

## Best Practices

### 1. Test Organization

- ใช้ table-driven tests สำหรับ test cases หลายๆ แบบ
- แยก unit tests และ integration tests ชัดเจน
- ใช้ property-based tests สำหรับทดสอบคุณสมบัติของระบบ

### 2. Test Data Management

- ใช้ fixtures สำหรับ test data ที่ใช้บ่อย
- ใช้ generators สำหรับ property tests
- Cleanup test data หลังจากทดสอบเสร็จ

### 3. Error Testing

- ทดสอบทั้ง happy path และ error cases
- ใช้ cancelled context เพื่อทดสอบ cancellation
- ทดสอบ timeout scenarios

### 4. Performance Testing

- ใช้ `go test -bench` สำหรับ benchmark tests
- ทดสอบ memory usage ด้วย `-benchmem`
- ทดสอบ race conditions ด้วย `-race`

### 5. Mock และ Fake Objects

- ใช้ `FakeClient` สำหรับ unit tests
- Mock external dependencies
- ใช้ dependency injection เพื่อง่ายต่อการทดสอบ

## การ Debug Tests

### Verbose Output

```bash
# ดู detailed test output
go test -v ./...

# ดู test coverage
go test -v -cover ./...

# ดู benchmark results
go test -v -bench=. ./...
```

### Test-Specific Debugging

```go
func TestDebugExample(t *testing.T) {
    if testing.Verbose() {
        t.Logf("Debug info: %+v", someVariable)
    }
    
    // Use t.Helper() in utility functions
    assertHelper := func(condition bool, msg string) {
        t.Helper()
        if !condition {
            t.Errorf("Assertion failed: %s", msg)
        }
    }
}
```

## Continuous Integration

### GitHub Actions Example

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    services:
      mongodb:
        image: mongo:7.0
        ports:
          - 27017:27017
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.22.5'
      - run: go test ./...
      - run: go test -race ./...
      - name: Integration Tests
        run: go test ./internal/mongo -run ClientRoundTrip
        env:
          MONGO_INTEGRATION_URI: mongodb://localhost:27017
```

## สรุป

การทดสอบใน maan-go ครอบคลุม:

- **Unit Tests** - ทดสอบแต่ละฟังก์ชันแยกกัน
- **Integration Tests** - ทดสอบกับ MongoDB จริง
- **Property Tests** - ทดสอบคุณสมบัติด้วยข้อมูลสุ่ม
- **Concurrency Tests** - ทดสอบ thread safety
- **Error Handling Tests** - ทดสอบการจัดการ error

ใช้ testing utilities ที่มีให้เพื่อเขียน tests ที่มีคุณภาพและง่ายต่อการบำรุงรักษา