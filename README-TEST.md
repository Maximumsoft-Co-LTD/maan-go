# maan-go Testing Guide

คู่มือการทดสอบสำหรับไลบรารี maan-go ที่ครอบคลุมการทดสอบแบบต่างๆ ตั้งแต่ unit tests, integration tests, property-based tests, benchmark tests และ testing utilities

## สถิติสรุป

| หมวด | จำนวน |
|------|-------|
| Test files | 44 ไฟล์ |
| Test functions | 205 รายการ |
| Benchmarks | 17 functions + 4 load tests |
| `go build ./...` | PASS |
| `go vet ./...` | PASS |
| `go test -race ./...` | PASS |

## โครงสร้างการทดสอบ

### ประเภทของการทดสอบ

1. **Unit Tests** - ทดสอบฟังก์ชันแต่ละส่วนแยกกัน
2. **Integration Tests** - ทดสอบการทำงานร่วมกับ MongoDB จริง
3. **Property-Based Tests** - ทดสอบคุณสมบัติของระบบด้วยข้อมูลสุ่ม
4. **Benchmark Tests** - วัด performance ของ operations หลักๆ
5. **Load Tests** - ทดสอบ concurrency ภายใต้ workload สูง
6. **Fake Client Tests** - ทดสอบโดยไม่ต้องใช้ MongoDB จริง

### โครงสร้างไฟล์ทดสอบ

```
(root)/
├── db_test.go               # DB[T] reflection tests
├── db_integration_test.go   # DB[T] integration tests (ต้องการ MongoDB)
internal/mongo/
├── *_test.go                    # Unit tests ควบคู่กับ implementation (67 tests)
├── client_integration_test.go   # Legacy integration test (1 test)
├── integration/                 # Integration test package (83 tests)
│   ├── helpers_test.go
│   ├── crud_integration_test.go
│   ├── upsert_integration_test.go
│   ├── atomic_ops_integration_test.go
│   ├── distinct_count_integration_test.go
│   ├── text_regex_integration_test.go
│   ├── index_integration_test.go
│   ├── transaction_integration_test.go
│   ├── change_stream_integration_test.go
│   ├── extended_collection_integration_test.go
│   ├── aggregation_integration_test.go
│   └── model_defaults_integration_test.go
├── bench/                       # Benchmark package (17 benchmarks + 4 load tests)
│   ├── helpers_test.go
│   ├── crud_bench_test.go
│   ├── aggregate_bench_test.go
│   ├── extended_collection_bench_test.go
│   └── load_test.go             # build tag: load_test
├── property_tests/              # Property-based tests (49 tests)
│   ├── base_test.go
│   └── *_test.go
└── testutil/                    # Testing utilities
    ├── assertions.go
    ├── fixtures.go
    └── generators.go
```

### ตาราง Coverage ต่อ Package

| Package | Files | Tests | ประเภท | Feature |
|---------|-------|-------|--------|---------|
| root | 2 | 5 | Unit + Integration | DB[T] reflection |
| internal/mongo | 11 | 68 | Unit (+ 1 legacy integration) | CRUD, aggregation, change streams, tx |
| property_tests | 14 | 49 | Property (≥100 iter) | Invariants, thread safety |
| integration | 12 | 83 | Integration | All features w/ real MongoDB |
| bench | 5 | 17B+4L | Benchmark + Load | Performance |
| **รวม** | **44** | **205 + 21B** | | |

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

#### ติดตั้ง MongoDB Replica Set ด้วย Docker

Integration tests ทั้งหมด (รวม change streams) ต้องการ MongoDB replica set:

```bash
# เริ่ม MongoDB replica set ด้วย Docker
docker run -d --name mongo-rs \
  -p 27017:27017 \
  mongo:7.0 mongod --replSet rs0

# รอ ~3 วินาทีแล้ว initiate replica set
sleep 3
docker exec mongo-rs mongosh --eval \
  'rs.initiate({_id:"rs0",members:[{_id:0,host:"127.0.0.1:27017"}]})'

# ตรวจสอบ replica set status
docker exec mongo-rs mongosh --eval 'rs.status().ok'
```

#### Legacy Integration Test (`client_integration_test.go`)

```bash
export MONGO_INTEGRATION_URI="mongodb://localhost:27017/?replicaSet=rs0"

go test ./internal/mongo -run ClientRoundTrip
go test -v ./internal/mongo -run Integration
```

#### Integration Package (`internal/mongo/integration/`)

package แยก (83 tests) ครอบคลุมทุก feature ด้วย MongoDB จริง:

```bash
export MONGO_INTEGRATION_URI="mongodb://localhost:27017/?replicaSet=rs0"

# รันทุก integration tests
go test ./internal/mongo/integration/...

# รันด้วย race detection
go test -race ./internal/mongo/integration/...

# รันเฉพาะ feature
go test -run TestCRUD ./internal/mongo/integration/...
go test -run TestTransaction ./internal/mongo/integration/...
go test -run TestChangeStream ./internal/mongo/integration/...
go test -run TestIndex ./internal/mongo/integration/...
```

หาก `MONGO_INTEGRATION_URI` ไม่ได้ตั้งค่า tests จะ skip โดยอัตโนมัติ:

```go
// helpers_test.go — skip guard pattern
func connectTestClient(t *testing.T) mongo.Client {
    t.Helper()
    uri := os.Getenv("MONGO_INTEGRATION_URI")
    if uri == "" {
        t.Skip("MONGO_INTEGRATION_URI not set; skipping integration test")
    }
    // ...
}
```

#### Integration Test Pattern (`uniqueColl`)

ทุก test ใช้ collection ชื่อ unique เพื่อหลีกเลี่ยง test isolation issues:

```go
func TestCRUD_CreateAndFind(t *testing.T) {
    client := connectTestClient(t)
    ctx := context.Background()

    collName := uniqueColl("crud_create")   // e.g. "crud_create_507f1f77bcf86cd799439011"
    coll := mongo.NewColl[integDoc](ctx, client, collName)
    dropColl(t, client, collName)           // cleanup via t.Cleanup

    doc := &integDoc{Name: "test", Value: 42}
    require.NoError(t, coll.Create(doc))
    // ...
}
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

## Benchmark Tests

### รัน Benchmarks

```bash
# รันทุก benchmarks (ต้องการ MONGO_INTEGRATION_URI)
export MONGO_INTEGRATION_URI="mongodb://localhost:27017/?replicaSet=rs0"

go test -bench=. -benchmem -benchtime=5s ./internal/mongo/bench/...

# รันเฉพาะ CRUD benchmarks
go test -bench=BenchmarkCreate -benchmem ./internal/mongo/bench/...

# รัน load tests (build tag: load_test)
go test -tags=load_test -race -timeout 120s ./internal/mongo/bench/...
```

### รายชื่อ Benchmarks

**`crud_bench_test.go`** (11 benchmarks)

| Benchmark | วัดอะไร |
|-----------|---------|
| `BenchmarkCreate` | Insert single document |
| `BenchmarkCreateMany_10` | Bulk insert 10 documents |
| `BenchmarkCreateMany_100` | Bulk insert 100 documents |
| `BenchmarkFindOne` | Single document lookup |
| `BenchmarkFindMany_All` | Scan all documents |
| `BenchmarkFindMany_WithFilter` | Filtered query |
| `BenchmarkSave` | Upsert (insert path) |
| `BenchmarkUpd` | Update single document |
| `BenchmarkDel` | Delete single document |
| `BenchmarkFindOneAndUpd` | FindOneAndUpdate |
| `BenchmarkFindOneAndDel` | FindOneAndDelete |

**`aggregate_bench_test.go`** (3 benchmarks)

| Benchmark | วัดอะไร |
|-----------|---------|
| `BenchmarkAgg_SimpleMatch` | Single `$match` stage |
| `BenchmarkAgg_GroupCount` | `$group` + `$count` |
| `BenchmarkAgg_Stream` | Streaming aggregation results |

**`extended_collection_bench_test.go`** (3 benchmarks)

| Benchmark | วัดอะไร |
|-----------|---------|
| `BenchmarkExtColl_ByFirst` | `.By(...).First()` fluent query |
| `BenchmarkExtColl_WhereMany` | `.Where(...).Many()` |
| `BenchmarkExtColl_Count` | `.Where(...).Count()` |

### Load Tests (`load_test.go`, build tag: `load_test`)

| Test | วัดอะไร |
|------|---------|
| `TestLoad_ConcurrentCreate_100` | 100 goroutines concurrent insert |
| `TestLoad_ConcurrentFind_100` | 100 goroutines concurrent find |
| `TestLoad_ConcurrentMixedOps` | Mixed read/write concurrent ops |
| `TestLoad_TransactionContention` | Transaction contention under load |

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
            filter:   bson.M{"_id": bson.NewObjectID()},
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
- ใช้ `uniqueColl()` helper ใน integration tests เพื่อหลีกเลี่ยง test collision
- Cleanup test data ผ่าน `t.Cleanup()` แทน `defer` เพื่อรองรับ parallel tests

### 3. Error Testing

- ทดสอบทั้ง happy path และ error cases
- ใช้ cancelled context เพื่อทดสอบ cancellation
- ทดสอบ timeout scenarios

### 4. Performance Testing

- ใช้ `go test -bench` สำหรับ benchmark tests
- ทดสอบ memory usage ด้วย `-benchmem`
- รัน load tests ด้วย `-tags=load_test` เพื่อทดสอบ concurrency
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
        options: >-
          --health-cmd "mongosh --eval 'db.adminCommand(\"ping\")'"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Build
        run: go build ./...

      - name: Vet
        run: go vet ./...

      - name: Init replica set
        run: |
          sleep 3
          docker exec $(docker ps -qf "ancestor=mongo:7.0") \
            mongosh --eval \
            'rs.initiate({_id:"rs0",members:[{_id:0,host:"127.0.0.1:27017"}]})'

      - name: Unit + Property Tests
        run: go test -race ./...

      - name: Integration Tests
        run: go test -race ./internal/mongo/integration/...
        env:
          MONGO_INTEGRATION_URI: mongodb://localhost:27017/?replicaSet=rs0

      - name: Legacy Integration Test
        run: go test ./internal/mongo -run ClientRoundTrip
        env:
          MONGO_INTEGRATION_URI: mongodb://localhost:27017/?replicaSet=rs0

      - name: Benchmarks (smoke)
        run: go test -bench=. -benchtime=1x ./internal/mongo/bench/...
        env:
          MONGO_INTEGRATION_URI: mongodb://localhost:27017/?replicaSet=rs0
```

## สรุป

การทดสอบใน maan-go ครอบคลุม 44 ไฟล์ / 205 test functions / 21 benchmarks:

| Layer | ประเภท | สถานะ |
|-------|--------|--------|
| **Unit Tests** | FakeClient, ไม่ต้องการ MongoDB | `go test ./...` |
| **Property Tests** | ≥100 iterations per property | `go test ./internal/mongo/property_tests` |
| **Integration Tests** | MongoDB replica set จำเป็น | `MONGO_INTEGRATION_URI=... go test ./internal/mongo/integration/...` |
| **Benchmarks** | วัด throughput + memory | `go test -bench=. -benchmem ./internal/mongo/bench/...` |
| **Load Tests** | Concurrency stress | `go test -tags=load_test -race ./internal/mongo/bench/...` |
| **Race Detection** | ทุก layer | `go test -race ./...` |

ใช้ testing utilities ที่มีให้เพื่อเขียน tests ที่มีคุณภาพและง่ายต่อการบำรุงรักษา
