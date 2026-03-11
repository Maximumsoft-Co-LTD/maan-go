# maan-go

`maan-go` (package `maango`) เป็น Go library สำหรับ MongoDB ที่ให้ API แบบ fluent/chainable และ type-safe ผ่าน Go generics รองรับ read/write separation, transaction, aggregation pipeline และ real-time change streams

## คุณสมบัติหลัก

- **Type-safe generics** — `Collection[T]`, `ManyResult[T]`, `Aggregate[T]`, `ChangeStream[T]`
- **Read/Write Separation** — ส่ง `WithWriteURI` + `WithReadURI` แยกกัน, library route อัตโนมัติ
- **Fluent builder** — chain methods ได้ทุกขั้นตอน เช่น `.Find(f).Sort(s).Limit(10).All()`
- **Auto model defaults** — inject ObjectID, created_at, updated_at อัตโนมัติ
- **Transaction** — ทั้งแบบ auto (`WithTx`) และ manual (`StartTx`)
- **Aggregation pipeline** — รองรับทุก stage พร้อม streaming
- **Change Streams** — real-time event watching พร้อม operation filter ครบ
- **FakeClient** สำหรับ unit test ไม่ต้องใช้ MongoDB จริง

## ความต้องการระบบ

- Go 1.21+
- MongoDB Driver v1.x (`go.mongodb.org/mongo-driver`)

## ติดตั้ง

```bash
go get github.com/Maximumsoft-Co-LTD/maan-go
```

---

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    maango "github.com/Maximumsoft-Co-LTD/maan-go"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
    ID    primitive.ObjectID `bson:"_id,omitempty"`
    Name  string             `bson:"name"`
    Email string             `bson:"email"`
}

func main() {
    ctx := context.Background()

    client, err := maango.NewClient(ctx,
        maango.WithWriteURI("mongodb://localhost:27017"),
        maango.WithDatabase("myapp"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    users := maango.NewColl[User](ctx, client, "users")

    // สร้าง document
    u := &User{Name: "Alice", Email: "alice@example.com"}
    if err := users.Create(u); err != nil {
        log.Fatal(err)
    }

    // ค้นหา document เดียว
    var result User
    err = users.FindOne(bson.M{"name": "Alice"}).Result(&result)
    fmt.Printf("found: %+v\n", result)

    // ค้นหาหลาย documents
    docs, err := users.Find(nil).Sort(bson.M{"name": 1}).All()
    _ = docs

    // ลบ
    _ = users.Del(bson.M{"name": "Alice"})
}
```

---

## API Reference

### 1. Client — การเชื่อมต่อ

#### `NewClient`

```go
func NewClient(ctx context.Context, opts ...Option) (Client, error)
```

สร้าง MongoDB client พร้อม read/write separation

#### `NewFakeClient`

```go
func NewFakeClient(opts ...FakeClientOption) (Client, error)
```

สร้าง fake client สำหรับ unit test (ไม่ต่อ MongoDB จริง)

#### Option functions

| Function | คำอธิบาย |
|---|---|
| `WithWriteURI(uri string)` | URI สำหรับ write (required) |
| `WithReadURI(uri string)` | URI สำหรับ read (ถ้าไม่ set จะใช้ write URI) |
| `WithDatabase(name string)` | ชื่อ database (required) |
| `WithTimeout(d time.Duration)` | connection timeout (default 60s) |
| `WithReadPreference(rp)` | read preference ของ read client |
| `WithWriteConcern(wc)` | write concern ของ write client |
| `WithClientOptions(fn)` | mutate `*options.ClientOptions` ก่อน dial |

#### FakeClient option functions

| Function | คำอธิบาย |
|---|---|
| `WithFakeDatabase(name string)` | ชื่อ database สำหรับ fake client (default: "testdb") |
| `WithFakeURI(uri string)` | URI ที่เก็บใน fake client (ไม่ dial จริง) |

#### Client interface methods

| Method | คำอธิบาย |
|---|---|
| `Write() *mongo.Client` | คืน client ที่ใช้ write |
| `Read() *mongo.Client` | คืน client ที่ใช้ read |
| `DbName() string` | ชื่อ database |
| `Close() error` | ปิด connection ทั้งคู่ |

---

### 2. Collection[T] — CRUD หลัก

#### สร้าง Collection

```go
// สร้างจาก Client (แนะนำ)
users := maango.NewColl[User](ctx, client, "users")

// สร้าง extended collection จาก Client
users := maango.NewExCollFromClient[User](ctx, client, "users")
```

#### Context และ Build

| Method | คำอธิบาย |
|---|---|
| `.Ctx(ctx)` | คืน shallow copy ที่ bind context ใหม่ |
| `.Build(ctx)` | คืน `ExtendedCollection[T]` bound กับ ctx |
| `.Name()` | คืนชื่อ collection |

#### Create

```go
// สร้าง document เดียว
err := coll.Create(&doc)
err := coll.Create(&doc, options.InsertOne().SetComment("x"))

// สร้างหลาย documents
err := coll.CreateMany(&docs)
```

`Create`/`CreateMany` เรียก model default hooks อัตโนมัติ (`DefaultId`, `DefaultCreatedAt`, `DefaultUpdatedAt`)

#### Find

```go
// ค้นหา document เดียว → คืน SingleResult[T]
result := coll.FindOne(bson.M{"_id": id})

// ค้นหาหลาย documents → คืน ManyResult[T]
result := coll.Find(bson.M{"active": true})
result := coll.FindMany(bson.M{"active": true})
// Find และ FindMany เหมือนกัน
```

#### Save (Upsert)

```go
// Save = UpdateOne + Upsert:true — ถ้าไม่เจอจะ insert
err := coll.Save(filter, update)
// SaveMany = UpdateMany + Upsert:true
err := coll.SaveMany(filter, update)
```

#### Upd (Update only)

```go
// Upd = UpdateOne ไม่ insert ถ้าไม่เจอ
err := coll.Upd(filter, update)
// UpdMany = UpdateMany ไม่ insert
err := coll.UpdMany(filter, update)
```

#### Del

```go
// ลบ document แรกที่ match
err := coll.Del(filter)
```

#### Aggregation

```go
// คืน Aggregate[T]
agg := coll.Agg(mongo.Pipeline{
    {{"$match", bson.M{"active": true}}},
    {{"$group", bson.M{"_id": "$category", "count": bson.M{"$sum": 1}}}},
})
```

#### Text / Regex Search

```go
// Full-text search (ต้องมี text index)
results, err := coll.TxtFind("golang mongodb")

// Regex search ข้ามหลาย fields (case-insensitive)
results, err := coll.RegexFields("alice", "name", "email")
```

#### Transaction

```go
// Auto commit/rollback
err := coll.WithTx(func(ctx context.Context) error {
    if err := coll.Ctx(ctx).Create(&doc); err != nil {
        return err
    }
    return coll.Ctx(ctx).Del(bson.M{"_id": oldID})
})

// Manual transaction
tx, err := coll.StartTx()
if err != nil { return err }
defer tx.Close(&err)
ctx := tx.Ctx()
err = coll.Ctx(ctx).Create(&doc)
```

---

### 3. SingleResult[T]

```go
var doc User
err := coll.FindOne(bson.M{"_id": id}).
    Proj(bson.M{"name": 1, "email": 1}).
    Sort(bson.M{"name": 1}).
    Hint(bson.M{"name": 1}).
    Opts(options.FindOne().SetMaxTime(5*time.Second)).
    Result(&doc)
```

| Method | คำอธิบาย |
|---|---|
| `.Proj(p any)` | กำหนด projection |
| `.Sort(s any)` | กำหนด sort |
| `.Hint(h any)` | กำหนด index hint |
| `.Opts(fo *options.FindOneOptions)` | merge raw options (override ทับ builder) |
| `.Result(out *T) error` | execute และ decode ผลลัพธ์ คืน `mongo.ErrNoDocuments` ถ้าไม่เจอ |

---

### 4. ManyResult[T]

```go
var docs []User
err := coll.Find(bson.M{"active": true}).
    Sort(bson.M{"name": 1}).
    Limit(20).
    Skip(40).
    Proj(bson.M{"name": 1}).
    Bsz(100).
    Result(&docs)

// All() — คืน slice โดยตรง
docs, err := coll.Find(nil).All()

// Streaming (ทีละ document — ประหยัด memory)
err := coll.Find(nil).Stream(func(ctx context.Context, doc User) error {
    fmt.Println(doc.Name)
    return nil
})

// Count (ไม่สนใจ Limit/Skip)
count, err := coll.Find(bson.M{"active": true}).Cnt()
```

| Method | คำอธิบาย |
|---|---|
| `.Proj(p any)` | projection |
| `.Sort(s any)` | sort |
| `.Hint(h any)` | index hint |
| `.Limit(n int64)` | จำกัดจำนวน |
| `.Skip(n int64)` | ข้ามต้น n ตัว |
| `.Bsz(n int32)` | cursor batch size |
| `.Opts(fo *options.FindOptions)` | merge raw options |
| `.All() ([]T, error)` | execute คืน slice |
| `.Result(out *[]T) error` | execute decode ลง out |
| `.Stream(fn) error` | streaming ทีละตัว |
| `.Each(fn) error` | alias ของ Stream |
| `.Cnt() (int64, error)` | นับ documents |

---

### 5. Aggregate[T]

```go
type Summary struct {
    ID    string `bson:"_id"`
    Total int    `bson:"total"`
}

pipeline := mongo.Pipeline{
    {{"$match", bson.M{"active": true}}},
    {{"$group", bson.M{"_id": "$category", "total": bson.M{"$sum": 1}}}},
}

// คืน typed results
results, err := coll.Agg(pipeline).All()

// คืน raw bson.M
raw, err := coll.Agg(pipeline).Raw()

// Streaming typed
err := coll.Agg(pipeline).
    Disk(true).
    Bsz(500).
    Each(func(ctx context.Context, doc Summary) error {
        fmt.Println(doc.ID, doc.Total)
        return nil
    })

// Streaming raw
err := coll.Agg(pipeline).EachRaw(func(ctx context.Context, doc bson.M) error {
    fmt.Println(doc)
    return nil
})
```

| Method | คำอธิบาย |
|---|---|
| `.Disk(b bool)` | เปิด/ปิด allowDiskUse |
| `.Bsz(n int32)` | cursor batch size |
| `.Opts(ao *options.AggregateOptions)` | merge raw options |
| `.All() ([]T, error)` | execute คืน typed slice |
| `.Result(out *[]T) error` | execute decode ลง out |
| `.Raw() ([]bson.M, error)` | execute คืน raw slice |
| `.Stream(fn) error` | streaming typed |
| `.Each(fn) error` | alias ของ Stream |
| `.EachRaw(fn) error` | streaming raw bson.M |

---

### 6. ExtendedCollection[T]

`ExtendedCollection` คือ dynamic query builder สำหรับสร้าง filter แบบ chain ใช้ผ่าน `coll.Build(ctx)` หรือ `NewExCollFromClient`

```go
builder := coll.Build(ctx)

// ค้นหา
var u User
err := builder.
    By("Name", "Alice").
    Where(bson.M{"active": true}).
    First(&u)

// ค้นหาหลายตัว
var users []User
err := builder.By("Role", "admin").Many(&users)

// นับ
count, err := builder.By("Active", true).Count()

// ตรวจสอบว่ามี
exists, err := builder.By("Email", "alice@example.com").Exists()

// ดู filter ที่ build มา
filter := builder.By("Name", "Alice").GetFilter()

// อัพเดต
err = builder.By("Name", "Alice").Save(bson.M{"$set": bson.M{"role": "admin"}})
err = builder.By("Active", true).SaveMany(bson.M{"$set": bson.M{"status": "ok"}})

// ลบ
err = builder.By("Name", "OldUser").Del()
err = builder.By("Active", false).DelMany()
```

> **หมายเหตุ**: `By("FieldName", value)` จะ resolve field name ผ่าน bson tag ของ struct ก่อน ถ้าไม่เจอจะ convert เป็น snake_case อัตโนมัติ เช่น `"UserName"` → `"user_name"`

| Method | คำอธิบาย |
|---|---|
| `.By(field, value)` | เพิ่ม equality condition ใน filter |
| `.Where(bson.M)` | merge filter เพิ่มเติม |
| `.First(*T) error` | หา document แรก |
| `.Many(*[]T) error` | หาทุก document |
| `.Count() (int64, error)` | นับ |
| `.Exists() (bool, error)` | ตรวจว่ามี |
| `.GetFilter() any` | คืน filter ที่ build ไว้ |
| `.Save(update, ...opts) error` | UpdateOne |
| `.SaveMany(update, ...opts) error` | UpdateMany |
| `.Del(...opts) error` | DeleteOne |
| `.Delete(...opts) error` | alias ของ Del |
| `.DelMany(...opts) error` | DeleteMany |
| `.DeleteMany(...opts) error` | alias ของ DelMany |

---

### 7. Transaction (TxSession)

#### Auto transaction (แนะนำ)

```go
err := coll.WithTx(func(ctx context.Context) error {
    if err := coll.Ctx(ctx).Create(&order); err != nil {
        return err // rollback อัตโนมัติ
    }
    return coll.Ctx(ctx).Upd(
        bson.M{"_id": stockID},
        bson.M{"$inc": bson.M{"qty": -1}},
    )
}) // commit ถ้าไม่มี error
```

#### Manual transaction

```go
func transferFunds(coll maango.Collection[Account], fromID, toID primitive.ObjectID, amount float64) (err error) {
    tx, err := coll.StartTx()
    if err != nil {
        return err
    }
    defer tx.Close(&err) // commit ถ้า err == nil, abort ถ้า err != nil

    ctx := tx.Ctx()
    if err = coll.Ctx(ctx).Upd(bson.M{"_id": fromID}, bson.M{"$inc": bson.M{"balance": -amount}}); err != nil {
        return
    }
    if err = coll.Ctx(ctx).Upd(bson.M{"_id": toID}, bson.M{"$inc": bson.M{"balance": amount}}); err != nil {
        return
    }
    return
}
```

| Method | คำอธิบาย |
|---|---|
| `coll.WithTx(fn)` | auto commit/rollback |
| `coll.StartTx()` | เริ่ม manual transaction คืน `TxSession` |
| `tx.Ctx()` | context ที่ต้องส่งให้ collection operations |
| `tx.Close(&err)` | commit ถ้า `*err == nil`, abort ถ้า `*err != nil` |

---

### 8. Change Streams

Change Streams ต้องการ MongoDB replica set หรือ Atlas cluster

#### ตัวอย่างพื้นฐาน

```go
ctx := context.Background()

err := coll.Watch(ctx).
    Stream(func(st maango.CsEvt[Order]) error {
        evt := st.ChangeEvent
        fmt.Printf("op=%s id=%v\n", evt.OperationType, evt.DocumentKey)
        return nil
    })
```

#### Operation filter

```go
coll.Watch(ctx).OnIst()          // insert เท่านั้น
coll.Watch(ctx).OnUpd()          // update เท่านั้น
coll.Watch(ctx).OnDel()          // delete เท่านั้น
coll.Watch(ctx).OnRep()          // replace เท่านั้น
coll.Watch(ctx).OnIstAndUpd()    // insert + update (ใช้บ่อยที่สุด)
coll.Watch(ctx).On("insert", "replace") // custom filter
```

#### Full document lookup

```go
// update event จะมี FullDocument (ต้องการ extra read)
coll.Watch(ctx).OnIstAndUpd().UpdLookup().
    Stream(func(st maango.CsEvt[Order]) error {
        if doc := st.ChangeEvent.FullDocument; doc != nil {
            fmt.Printf("full doc: %+v\n", *doc)
        }
        return nil
    })

// Error ถ้า full document ไม่ available
coll.Watch(ctx).FullDocRequired().Stream(handler)

// ตั้งค่า fullDocument เอง
coll.Watch(ctx).FullDoc("whenAvailable").Stream(handler)
```

#### Resume หลัง interruption

```go
var lastToken bson.M

// บันทึก resume token ระหว่าง stream
coll.Watch(ctx).Stream(func(st maango.CsEvt[Order]) error {
    lastToken = st.ChangeEvent.ResumeToken
    return nil
})

// resume จาก token ที่บันทึกไว้
coll.Watch(ctx).ResumeAfter(lastToken).Stream(handler)

// StartAfter — resume ได้แม้หลัง invalidate event
coll.Watch(ctx).StartAfter(lastToken).Stream(handler)
```

#### Advanced options

| Method | คำอธิบาย |
|---|---|
| `.Bsz(n int32)` | cursor batch size |
| `.Collation(c)` | collation สำหรับ stream |
| `.Comment(s string)` | comment ใน profiler/logs |
| `.FullDocBefore(opt)` | `fullDocumentBeforeChange`: "off", "whenAvailable", "required" |
| `.MaxAwait(d time.Duration)` | เวลา max ที่ server รอข้อมูลใหม่ |
| `.ShowExpanded(b bool)` | expanded events (MongoDB 6.0+) |
| `.StartAtTime(t *primitive.Timestamp)` | เริ่มจาก operation time ที่กำหนด |
| `.Custom(m bson.M)` | custom document เพิ่มเติมใน command |
| `.CustomPipeline(m bson.M)` | custom document ใน aggregation pipeline |
| `.Opts(o *options.ChangeStreamOptions)` | raw options override |

#### Types

**`ChangeEvent[T]`** — event แต่ละตัวที่ได้จาก stream

| Field | Type | คำอธิบาย |
|---|---|---|
| `ResumeToken` | `bson.M` | token สำหรับ resume |
| `OperationType` | `string` | "insert", "update", "replace", "delete", ... |
| `FullDocument` | `*T` | typed document (nil สำหรับ delete; nil สำหรับ update ถ้าไม่ใช้ UpdLookup) |
| `DocumentKey` | `bson.M` | `_id` (และ shard key ถ้ามี) ของ document ที่เปลี่ยน |
| `Namespace` | `ChangeEventNamespace` | `{DB, Coll}` ที่เกิด event |
| `UpdateDesc` | `*ChangeUpdateDesc` | มีเฉพาะ update event |

**`ChangeUpdateDesc`** — รายละเอียดการ update

| Field | Type | คำอธิบาย |
|---|---|---|
| `UpdatedFields` | `bson.M` | map ของ field ที่เปลี่ยน |
| `RemovedFields` | `[]string` | รายการ field ที่ถูก unset |

**`CsEvt[T]`** — argument ที่ส่งให้ callback

| Field/Method | คำอธิบาย |
|---|---|
| `.ChangeEvent` | ข้อมูล event ทั้งหมด |
| `.Ctx()` | context ที่ควบคุม lifetime ของ stream |

**ข้อมูลที่มีตาม operation type:**

| Operation | `FullDocument` | `UpdateDesc` | `DocumentKey` |
|---|---|---|---|
| insert | มีเสมอ | ไม่มี | มี |
| update | nil (ยกเว้น UpdLookup) | มี | มี |
| replace | มีเสมอ | ไม่มี | มี |
| delete | nil เสมอ | ไม่มี | มี |

---

### 9. Model Defaults

เพื่อให้ library auto-populate `_id`, `created_at`, `updated_at` บน `Create`/`CreateMany`, ให้ struct implement interface `defaultable`:

```go
type Order struct {
    ID        primitive.ObjectID `bson:"_id"`
    CreatedAt time.Time          `bson:"created_at"`
    UpdatedAt time.Time          `bson:"updated_at"`
    Total     float64            `bson:"total"`
}

func (o *Order) DefaultId() primitive.ObjectID {
    if o.ID.IsZero() {
        o.ID = primitive.NewObjectID()
    }
    return o.ID
}

func (o *Order) DefaultCreatedAt() time.Time {
    if o.CreatedAt.IsZero() {
        o.CreatedAt = time.Now().UTC()
    }
    return o.CreatedAt
}

func (o *Order) DefaultUpdatedAt() time.Time {
    o.UpdatedAt = time.Now().UTC()
    return o.UpdatedAt
}
```

> หาก struct ไม่ implement interface นี้ library จะ skip auto-populate

---

### 10. Auto DB Initialization ด้วย `DB[T]`

`DB[T]` ใช้ reflection อ่าน struct tag `collection_name` และ initialize `Coll[T]` / `ExColl[T]` อัตโนมัติ

```go
type MyDatabase struct {
    Users    maango.Coll[User]      `collection_name:"users"`
    Orders   maango.Coll[Order]     `collection_name:"orders"`
    Products maango.ExColl[Product] `collection_name:"products"`
}

db := maango.DB[MyDatabase](ctx, client)

// ใช้งานได้ทันที
var user User
_ = db.Users.FindOne(bson.M{"_id": id}).Result(&user)

var orders []Order
_ = db.Orders.Find(bson.M{"user_id": id}).Sort(bson.M{"created_at": -1}).Result(&orders)

// ExColl ใช้ dynamic query
count, _ := db.Products.
    By("Category", "electronics").
    Where(bson.M{"active": true}).
    Count()
```

| Type | คำอธิบาย |
|---|---|
| `Coll[T]` | wraps `Collection[T]` รองรับ reflection |
| `ExColl[T]` | wraps `ExtendedCollection[T]` รองรับ reflection |
| `DB[T](ctx, client)` | สร้าง `*T` และ initialize ทุก field ที่มี tag `collection_name` |

---

### 11. Testing ด้วย FakeClient

```go
func TestCreateUser(t *testing.T) {
    client, err := maango.NewFakeClient(
        maango.WithFakeDatabase("testdb"),
    )
    if err != nil {
        t.Fatal(err)
    }
    defer client.Close()

    users := maango.NewColl[User](context.Background(), client, "users")

    // ทดสอบ builder logic ที่ไม่ต้อง execute จริง
    filter := users.Build(context.Background()).
        By("Name", "test").
        GetFilter()
    _ = filter
}
```

> **หมายเหตุ**: FakeClient สร้าง `*mongo.Client` ที่ยังไม่ได้ connect จริง เหมาะสำหรับ test ที่ต้องการ wire collection กับ struct fields ผ่าน `DB[T]` หรือ test builder logic สำหรับ test ที่ต้องการ query result จริงให้ใช้ integration test กับ MongoDB จริง

---

## Package Structure

```
.
├── main.go                     # Public API entry point (package maango)
├── db.go                       # DB[T], Coll[T], ExColl[T]
├── go.mod / go.sum
└── internal/mongo/
    ├── api.go                  # Interface definitions ทั้งหมด
    ├── client.go               # MongoDB client + Option functions
    ├── coll.go                 # collection[T] implements Collection[T]
    ├── more-coll.go            # extendedCollection[T] implements ExtendedCollection[T]
    ├── sub-qry.go              # single[T], many[T] builders
    ├── aggregate.go            # agg[T] implements Aggregate[T]
    ├── transaction-session.go  # transactionSession implements TxSession
    ├── change-stream.go        # changeStream[T] implements ChangeStream[T]
    ├── model-defaults.go       # defaultable interface + applyModelDefaults
    └── fake_client.go          # FakeClient สำหรับ testing
```

## Common Commands

```bash
# Build
go build ./...

# Unit tests
go test ./...

# Integration tests (ต้องการ MongoDB replica set)
MONGO_INTEGRATION_URI="mongodb://localhost:27017" go test ./internal/mongo -run ClientRoundTrip

# Race detection
go test -race ./...

# Format / vet
go fmt ./...
go vet ./...
```
