# maan-go

ไลบรารี Go สำหรับทำงานกับ MongoDB แบบแยกเส้นทางอ่าน/เขียน รองรับการทำงานกับ struct ที่มี type ชัดเจน พร้อม fluent API สำหรับ CRUD, aggregation และ transaction

## ติดตั้ง

```bash
go get maan-go
```

หรือถ้าเก็บซอร์สไว้ในรีโปเดียวกัน สามารถอ้างอิงโมดูลด้วย path `maan-go` ได้ทันที

## Quick Start

```go
package main

import (
    "context"

    maango "maan-go"
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

    bankColl := maango.NewCollection[BankConfig](ctx, client, "bank_config")

    cfg := BankConfig{Name: "demo"}
    if err := bankColl.Create(&cfg); err != nil {
        panic(err)
    }

    var stored BankConfig
    if err := bankColl.FindOne(bson.M{"_id": cfg.ID}).Res(&stored); err != nil {
        panic(err)
    }
}
```

## Public API Snapshot
- `Client` – เชื่อมต่อ MongoDB และแยก client สำหรับอ่าน/เขียน
- `Collection[T]` – CRUD, aggregation, transaction helper สำหรับ struct ชนิด `T`
- `ExtendedCollection[T]` – query builder แบบ chainable (`By`, `Where`, `Count`, `Exists`)
- `SingleResult[T]` / `ManyResult[T]` – fluent query สำหรับ `FindOne` / `FindMany`
- `Aggregate[T]` – ทำ aggregation pipeline พร้อม helper สำหรับ stream และ raw result
- `TxSession` – จัดการทรานแซกชันแบบควบคุมเอง (`Commit`, `Rollback`)

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

```go
coll := maango.NewCollection[BankConfig](ctx, client, "bank_config")

// Insert
if err := coll.Create(&BankConfig{Name: "first"}); err != nil {
    panic(err)
}

// Find แบบ fluent
var cfg BankConfig
if err := coll.
    FindOne(bson.M{"name": "first"}).
    Proj(bson.M{"_id": 1, "name": 1}).
    Res(&cfg); err != nil {
    panic(err)
}

// Find หลายเรคคอร์ด พร้อม sort/limit
items, err := coll.
    FindMany(bson.M{"is_active": true}).
    Sort(bson.M{"created_at": -1}).
    Lim(10).
    All()
```

### Query Builder แบบ chainable

```go
var result []BankConfig
err := coll.
    Build(ctx).
    By("Code", "KTB").          // map field struct -> ฟิลด์ใน Mongo ให้อัตโนมัติ
    Where(bson.M{"status": "active"}).
    Many(&result)
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

### Transaction

```go
// แบบจัดการให้ครบ
if err := coll.WithTx(func(txCtx context.Context) error {
    return coll.Ctx(txCtx).Create(&BankConfig{Name: "tx"})
}); err != nil {
    panic(err)
}

// แบบควบคุมเอง
tx, err := coll.StartTx()
if err != nil {
    panic(err)
}
txCtx := tx.SessionCtx()
if err := coll.Ctx(txCtx).Create(&BankConfig{Name: "manual"}); err != nil {
    tx.Rollback()
    panic(err)
}
if err := tx.Commit(); err != nil {
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

ตัวอย่างเต็มดูได้ที่ `internal/entities/bank-config.go`

## ใช้ร่วมกับ Repository Pattern

```go
type BankRepository struct {
    collection maango.Collection[BankConfig]
}

func NewBankRepository(ctx context.Context, client maango.Client) *BankRepository {
    return &BankRepository{
        collection: maango.NewCollection[BankConfig](ctx, client, "bank_config"),
    }
}

func (r *BankRepository) FindDefault(ctx context.Context) (BankConfig, error) {
    var cfg BankConfig
    err := r.collection.Ctx(ctx).FindOne(bson.M{"is_default": true}).Res(&cfg)
    return cfg, err
}
```

## Integration Test

ไฟล์ `pkg/mongo/client_integration_test.go` จำลองการทำงานจริงกับ MongoDB หากไม่ตั้งค่าตัวแปร `MONGO_INTEGRATION_URI` เทสต์จะถูกข้าม

```bash
MONGO_INTEGRATION_URI="mongodb://localhost:27017" go test ./pkg/mongo -run ClientRoundTrip
```

# maan-go
