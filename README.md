# maan-go

ไลบรารี Go สำหรับทำงานกับ MongoDB แบบแยกเส้นทางอ่าน/เขียน รองรับการทำงานกับ struct ที่มี type ชัดเจน พร้อม fluent API สำหรับ CRUD, aggregation และ transaction

## คุณสมบัติหลัก

- **แยกเส้นทางอ่าน/เขียน**: รองรับ MongoDB URI แยกกันสำหรับการอ่านและเขียนข้อมูล
- **Type Safety**: Collection แบบ strongly typed ด้วย Go generics (`Collection[T]`)
- **Fluent API**: เมทอดแบบ chainable สำหรับ query, aggregation และ operations
- **Transaction Support**: รองรับทั้งแบบอัตโนมัติ (`WithTx`) และแบบควบคุมเอง (`StartTx`)
- **Model Defaults**: เติมค่าเริ่มต้น (ID, timestamps) อัตโนมัติผ่าน interface methods
- **Testing Support**: มี fake client สำหรับ unit testing โดยไม่ต้องใช้ MongoDB จริง

## ความต้องการระบบ

- **Go 1.22.5+** (ต้องการ generics support)
- **go.mongodb.org/mongo-driver v1.17.4**

## ติดตั้ง

```bash
go get github.com/Maximumsoft-Co-LTD/maan-go
```

## Quick Start

```go
package main

import (
    "context"

    maango "github.com/Maximumsoft-Co-LTD/maan-go"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type BankConfig struct {
    ID   primitive.ObjectID `bson:"_id"`
    Name string             `bson:"name"`
}

func main() {
    ctx := context.Background()
    client, err := maango.NewClient(
        ctx,
        maango.WithWriteURI("mongodb://localhost:27017"),
        maango.WithDatabase("example"),
    )
    if err != nil {
        panic(err)
    }
    defer client.Close()

    bankColl := maango.NewColl[BankConfig](ctx, client, "bank_config")

    cfg := BankConfig{Name: "demo"}
    if err := bankColl.Create(&cfg); err != nil {
        panic(err)
    }

    var stored BankConfig
    if err := bankColl.FindOne(bson.M{"_id": cfg.ID}).Result(&stored); err != nil {
        panic(err)
    }
}
```

## API Overview
- `Client` – เชื่อมต่อ MongoDB และแยก client สำหรับอ่าน/เขียน
- `Coll[T]` – CRUD, aggregation, transaction helper สำหรับ struct ชนิด `T`
- `ExColl[T]` – query builder แบบ chainable (`By`, `Where`, `Count`, `Exists`)
- `SingleResult[T]` / `ManyResult[T]` – fluent query สำหรับ `FindOne` / `FindMany`
- `Aggregate[T]` – ทำ aggregation pipeline พร้อม helper สำหรับ stream และ raw result
- `Session` – จัดการทรานแซกชันแบบควบคุมเองด้วย pattern `defer tx.Close(&err)`
- `ChangeStream[T]` – รับ real-time change events ผ่าน MongoDB Change Streams (`Watch`)
- `CsEvt[T]` – callback argument ของ Stream/Each รวม event data และ context ไว้ในที่เดียว

## การสร้าง Client และ Options

```go
client, err := maango.NewClient(
    ctx,
    maango.WithWriteURI("mongodb://writer:27017"),
    maango.WithReadURI("mongodb://reader:27017"), // ไม่ระบุจะใช้ URI เขียน
    maango.WithDatabase("example"),
    maango.WithTimeout(30*time.Second),
    maango.WithReadPreference(readpref.SecondaryPreferred()),
    maango.WithWriteConcern(writeconcern.New(writeconcern.WMajority())),
    maango.WithClientOptions(func(opts *options.ClientOptions) {
        opts.SetAppName("potja-demo")
    }),
)
```

ตัวเลือกที่รองรับ:
- `WithWriteURI` *(จำเป็น)* – URI สำหรับ client ฝั่งเขียน
- `WithReadURI` – URI สำหรับ client ฝั่งอ่าน
- `WithDatabase` *(จำเป็น)* – ตั้งชื่อฐานข้อมูลที่จะใช้
- `WithTimeout` – timeout ขณะเชื่อมต่อ (ค่าเริ่มต้น 60s)
- `WithReadPreference` – ตั้ง read preference ของฝั่งอ่าน
- `WithWriteConcern` – ตั้ง write concern ของฝั่งเขียน
- `WithClientOptions` – mutator สำหรับ `mongo/options.ClientOptions` (สามารถใส่หลายครั้งได้)

## การใช้ Collection

### CRUD Operations

```go
coll := maango.NewColl[BankConfig](ctx, client, "bank_config")

// 1. CREATE - สร้างเอกสารใหม่
if err := coll.Create(&BankConfig{Name: "first"}); err != nil {
    panic(err)
}

// 2. SAVE (Upsert) - อัปเดตถ้ามี หรือสร้างใหม่ถ้าไม่มี
if err := coll.Save(
    bson.M{"name": "first"}, 
    bson.M{"$set": bson.M{"status": "active"}},
); err != nil {
    panic(err)
}

// 3. UPDATE - อัปเดตเฉพาะเอกสารที่มีอยู่แล้ว (ไม่สร้างใหม่)
if err := coll.Upd(
    bson.M{"name": "first"}, 
    bson.M{"$set": bson.M{"last_updated": time.Now()}},
); err != nil {
    panic(err)
}

// 4. DELETE - ลบเอกสาร
if err := coll.Del(bson.M{"name": "first"}); err != nil {
    panic(err)
}
```

**ความแตกต่างสำคัญ:**
- **`Create`** - สร้างเอกสารใหม่เท่านั้น (จะ error ถ้ามี duplicate key)
- **`Save`/`SaveMany`** - **Upsert** = อัปเดตถ้ามี, สร้างใหม่ถ้าไม่มี
- **`Upd`/`UpdMany`** - **Update Only** = อัปเดตเฉพาะที่มีอยู่แล้ว, ไม่สร้างใหม่

### Query Operations

// Find แบบ fluent
var cfg BankConfig
if err := coll.
    FindOne(bson.M{"name": "first"}).
    Proj(bson.M{"_id": 1, "name": 1}).
    Result(&cfg); err != nil {
    panic(err)
}

// Find หลายเรคคอร์ด พร้อม sort/limit
items, err := coll.
    FindMany(bson.M{"is_active": true}).
    Sort(bson.M{"created_at": -1}).
    Limit(10).
    All()
```

### Query Builder แบบ chainable

```go
var result []BankConfig
err := coll.
    Build(ctx).
    By("Code", "KTB").
    Where(bson.M{"status": "active"}).
    Many(&result)
```

### Example: unit test แบบไม่ต้องต่อ Mongo จริง

```go
func TestFindDefault(t *testing.T) {
    client, err := maango.NewFakeClient()
    if err != nil {
        t.Fatalf("fake client: %v", err)
    }
    defer client.Close()

    coll := maango.NewColl[BankConfig](context.Background(), client, "bank_config")
    // ทดสอบ builder/filter logic ได้โดยไม่ต้องมี MongoDB จริง
    _ = coll.Build(context.Background()).By("Code", "KTB")
}
```

### Aggregation

```go
pipeline := []bson.M{
    {"$match": bson.M{"is_active": true}},
    {"$group": bson.M{"_id": "$status", "count": bson.M{"$sum": 1}}},
}

raw, err := coll.
    Agg(pipeline).
    Disk(true).
    Raw()
if err != nil {
    panic(err)
}
for _, doc := range raw {
    // ต้อง import "fmt"
    fmt.Println(doc["_id"], doc["count"])
}
```

### Change Streams (Real-time Events)

Change Streams ช่วยให้ application รับ real-time events เมื่อมีการ insert/update/delete ใน collection โดยไม่ต้อง polling — ต้องการ MongoDB replica set หรือ sharded cluster

#### โครงสร้าง Callback

```go
// CsEvt[T] คือ argument เดียวของ callback ทุก Stream/Each
// รวม event data และ context ไว้ในที่เดียว
type CsEvt[T any] struct {
    ChangeEvent ChangeEvent[T] // ข้อมูล event ทั้งหมด
}
func (s CsEvt[T]) Ctx() context.Context // context ที่ควบคุม lifetime ของ stream
```

| Field / Method | ความหมาย |
|---|---|
| `st.ChangeEvent.OperationType` | "insert" / "update" / "replace" / "delete" / ... |
| `st.ChangeEvent.FullDocument` | `*T` — typed document (nil สำหรับ delete) |
| `st.ChangeEvent.DocumentKey` | `bson.M` ที่มี `_id` ของ document ที่เปลี่ยน |
| `st.ChangeEvent.UpdateDesc` | fields ที่เปลี่ยน (เฉพาะ update) |
| `st.ChangeEvent.ResumeToken` | token สำหรับ resume stream |
| `st.Ctx()` | context ของ stream |

#### ตัวอย่าง: Watch ทุก event

```go
coll := maango.NewColl[Order](ctx, client, "orders")

err := coll.Watch(ctx).
    Stream(func(st maango.CsEvt[Order]) error {
        fmt.Println(st.ChangeEvent.OperationType, st.ChangeEvent.DocumentKey)
        return nil
    })
```

#### ตัวอย่าง: Watch เฉพาะ insert + update พร้อม full document

```go
err := coll.Watch(ctx).
    OnIstAndUpd().   // กรอง insert และ update เท่านั้น
    UpdLookup().     // ดึง full document สำหรับ update events ด้วย
    Stream(func(st maango.CsEvt[Order]) error {
        ev := st.ChangeEvent
        fmt.Printf("op=%s doc=%+v\n", ev.OperationType, ev.FullDocument)
        return nil
    })
```

#### ตัวอย่าง: Resume หลังจาก interruption

```go
var lastToken bson.M

// รอบแรก — เก็บ resume token ทุก event
coll.Watch(ctx).Stream(func(st maango.CsEvt[Order]) error {
    lastToken = st.ChangeEvent.ResumeToken
    return nil
})

// restart ครั้งถัดไป — ต่อจากจุดที่หยุด
coll.Watch(ctx).ResumeAfter(lastToken).Stream(handler)
```

#### Shortcut methods ทั้งหมด

| Method | เทียบเท่า |
|---|---|
| `.OnIst()` | `.On("insert")` |
| `.OnUpd()` | `.On("update")` |
| `.OnDel()` | `.On("delete")` |
| `.OnRep()` | `.On("replace")` |
| `.OnIstAndUpd()` | `.On("insert", "update")` |
| `.UpdLookup()` | `.FullDoc("updateLookup")` |
| `.FullDocRequired()` | `.FullDoc("required")` |

### Transaction

```go
// แบบจัดการให้ครบ
if err := coll.WithTx(func(txCtx context.Context) error {
    return coll.Ctx(txCtx).Create(&BankConfig{Name: "tx"})
}); err != nil {
    panic(err)
}

// แบบควบคุมเองด้วย defer tx.Close(&err)
tx, err := coll.StartTx()
if err != nil {
    panic(err)
}
var txErr error
defer tx.Close(&txErr)

txCtx := tx.Ctx()
if err := coll.Ctx(txCtx).Create(&BankConfig{Name: "manual"}); err != nil {
    txErr = err
    panic(err)
}
```

## Default Values สำหรับโมเดล

หาก struct implement เมทอดต่อไปนี้ `Collection.Create` และ `CreateMany` จะเติมค่าให้อัตโนมัติ

```go
import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type AuditFields struct {
    ID        primitive.ObjectID `bson:"_id"`
    CreatedAt time.Time          `bson:"created_at"`
    UpdatedAt time.Time          `bson:"updated_at"`
}

func (a *AuditFields) DefaultId() primitive.ObjectID {
    if a.ID.IsZero() {
        a.ID = primitive.NewObjectID()
    }
    return a.ID
}

func (a *AuditFields) DefaultCreatedAt() time.Time {
    if a.CreatedAt.IsZero() {
        a.CreatedAt = time.Now().UTC()
    }
    return a.CreatedAt
}

func (a *AuditFields) DefaultUpdatedAt() time.Time {
    if a.UpdatedAt.IsZero() {
        a.UpdatedAt = time.Now().UTC()
    }
    return a.UpdatedAt
}
```

ตัวอย่างเต็มดูได้ที่ `internal/mongo/client_integration_test.go`

## ใช้ร่วมกับ Repository Pattern

```go
type BankRepository struct {
    collection maango.Coll[BankConfig]
}

func NewBankRepository(ctx context.Context, client maango.Client) *BankRepository {
    return &BankRepository{
        collection: maango.NewColl[BankConfig](ctx, client, "bank_config"),
    }
}

func (r *BankRepository) FindDefault(ctx context.Context) (BankConfig, error) {
    var cfg BankConfig
    err := r.collection.Ctx(ctx).FindOne(bson.M{"is_default": true}).Result(&cfg)
    return cfg, err
}
```

## Development

### การรันเทสต์

```bash
# Unit tests
go test ./...

# Integration tests (ต้องการ MongoDB)
MONGO_INTEGRATION_URI="mongodb://localhost:27017" go test ./internal/mongo -run ClientRoundTrip

# ทดสอบพร้อม race detection
go test -race ./...
```

### Code Quality

```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Build library
go build ./...
```

## License

MIT License

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.
