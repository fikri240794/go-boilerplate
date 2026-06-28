# Go-Boilerplate: Complete Coding Patterns & Conventions

> **Audience:** AI assistants and human developers working on this codebase.
> **Purpose:** Define every pattern, convention, and structure so that any AI can write, refactor, and test code correctly without guessing.

## About This Boilerplate

This project is a **base template** (boilerplate) designed so that all future development follows the **same code style standards**. Every layer has been carefully structured with consistent patterns — from variable declarations and error handling to repository helpers and transport handlers.

### Replacing the Sample Entity

The existing `Guest` entity is a sample. When you need to replace it with your own domain entity, follow these steps to keep things simple while respecting the boilerplate conventions:

| Step | What to Do | Files to Touch |
|---|---|---|
| 1 | Create your new entity with proper struct tags (`table`, `db`, `primary_key`, `db_type`) | `internal/models/entities/your_entity.go` |
| 2 | Create DTOs (request/response) with `Validate()` and conversion methods | `internal/models/dtos/your_dto.go` |
| 3 | Add database field name constants matching your entity's `db` tags | `internal/models/entities/your_entity.go` |
| 4 | Create/update repository interfaces (`IYourRepository`) | `internal/repositories/your_repository.go` |
| 5 | Create concrete repository struct (embeds `BoilerplateDatabaseRepository[T]`) | `internal/repositories/your_repository.go` |
| 6 | Create service interface and struct using all helpers (`withTransaction`, `tryDeleteEntityCaches`, `publishEvent`, etc.) | `internal/services/your_service.go` |
| 7 | Create transport handlers (HTTP, gRPC, event consumer) + VMs | `transports/{http,grpc,event_consumer}/` |
| 8 | Update Wire providers (`provider.go`) in all layers | `internal/repositories/`, `internal/services/`, `transports/*/handlers/` |
| 9 | Update config with your entity's cache keys, event topics, etc. | `configs/config.go` + `.env` |
| 10 | Delete old `Guest`-specific files and regenerate mocks | Run `make generate && make test` |

> All the generic helpers (`getEntityMeta`, `prepareQueryStatement`, `logSlowQuery`, `withTransaction`, `buildActiveEntityFilterByIDs`, `tryDeleteEntityCaches`, `publishEvent`, etc.) are **entity-agnostic** — they work with any `TEntity` without modification. This means most of the heavy lifting is already done for you.

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Variable Declaration Style](#2-variable-declaration-style)
3. [Logging Patterns](#3-logging-patterns)
4. [Tracer (OpenTelemetry)](#4-tracer-opentelemetry)
5. [Error Handling](#5-error-handling)
6. [Config Layer](#6-config-layer)
7. [Datasources Layer](#7-datasources-layer)
8. [Entity Layer (Models)](#8-entity-layer-models)
9. [DTO Layer](#9-dto-layer)
10. [Repository Layer](#10-repository-layer)
11. [Service Layer](#11-service-layer)
12. [Transport Layer — HTTP](#12-transport-layer--http)
13. [Transport Layer — gRPC](#13-transport-layer--grpc)
14. [Transport Layer — Event Consumer (NSQ)](#14-transport-layer--event-consumer-nsq)
15. [CMD Layer (Application Entry Points)](#15-cmd-layer-application-entry-points)
16. [Tooling & Code Generation](#16-tooling--code-generation)
17. [Dependency Injection (Wire)](#17-dependency-injection-wire)
18. [Testing Conventions](#18-testing-conventions)
19. [Key Do's and Don'ts](#19-key-dos-and-donts)

---

## 1. Architecture Overview

```
configs/                        → Viper-based configuration (.env)
datasources/
  boilerplate_database/         → PostgreSQL (sqlx) with Master/Slave replication
  event_producer/               → NSQ message producer
  in_memory_database/           → Redis cache (go-redis/redis/v9)
  webhook_site_http_client/     → HTTP client for webhooks
internal/
  models/
    dtos/                       → Request/Response DTOs with validation
    entities/                   → Domain entities + event entities
  repositories/                 → Generic + concrete data access (database, cache, producer)
  services/                     → Business logic
pkg/
  constants/                    → Shared constants
  context/                      → Custom context utilities
  grpc_error/                   → gRPC error helpers
  grpc_metadata/                → gRPC metadata helpers
  logger/                       → Zerolog logger setup
  protobuf_boilerplate/         → Generated protobuf + gRPC stubs
  tracer/                       → OpenTelemetry tracer setup
  uuid/                         → UUID generation (gofrs/uuid/v5)
  validator/                    → go-playground/validator setup
transports/
  http/                         → Fiber HTTP server + handlers + VMs + swagger docs
  grpc/                         → gRPC server + handlers + VMs
  event_consumer/               → NSQ consumer + handlers
cmd/                            → Cobra CLI entry points
```

**Key technologies:**
| Technology | Usage |
|---|---|
| Go 1.25 | Generics (`TEntity interface{}`), `time` package |
| Fiber v2 | HTTP framework (`fiber.Ctx`, `fiber.Router`, `fiber.App`) |
| gRPC + protobuf | RPC framework, `.proto` → `.pb.go` code generation |
| NSQ | Message queue (producer + consumer) |
| sqlx + PostgreSQL | SQL database with Master/Slave replication |
| go-redis/redis/v9 | Redis client (`redis.Client`, `redis.Nil`) |
| Google Wire | Dependency injection (`wire_gen.go` auto-generated) |
| goqube | SQL query builder (dialect-aware: MySQL, PostgreSQL, SQLite, SQL Server) |
| gotask | Goroutine task runner with error handling (`ErrorTask`) |
| mockery v3 | Mock generation from interfaces |
| swaggo/swag | Swagger/OpenAPI doc generation |
| cobra | CLI framework (`cobra.Command`) |
| viper | Configuration loader (`.env` + `mapstructure` tags) |
| zerolog | Structured logging (`log.Info()`, `log.Err()`, `log.Warn()`) |
| go-playground/validator | Struct validation |
| OpenTelemetry | Distributed tracing (`trace.Span`, `tracer.Start()`) |
| gocerr | Typed errors with HTTP status codes (`gocerr.New(code, msg)`) |
| guregu/null | Nullable types (`null.String`, `null.Int64`) |
| gofrs/uuid/v5 | UUID generation (`uuid.UUID`, `uuid.Must(uuid.NewV4())`) |

---

## 2. Variable Declaration Style (CRITICAL)

### 2.1 C-Style Variable Blocks

All variables MUST be declared at function start in a `var (...)` block. The only exception is `for i := range` loop indices.

```go
func (s *GuestService) Create(ctx context.Context, requestDTO *dtos.CreateGuestRequestDTO) (*dtos.GuestResponseDTO, error) {
    var (
        span        trace.Span
        logFields   map[string]interface{}
        entity      *entities.GuestEntity
        responseDTO *dtos.GuestResponseDTO
        err         error
    )
    // body uses = not :=
}
```

### 2.2 `if`-Scope Variables

Even variables used only inside `if` blocks must be declared at the top:

```go
// ✅ Correct
var (
    dbType    string
    pkTag     string
    logLevel  zerolog.Level
)

dbType = field.Tag.Get("db_type")
if dbType != "" { ... }

// ❌ Wrong
if dbType := field.Tag.Get("db_type"); dbType != "" { ... }    // NO
if err := doSomething(); err != nil { ... }                      // NO
```

### 2.3 Loop Variables Exception

Only `for i := range` is allowed as `:=` exception:

```go
for i := range entities {       // ✅ Acceptable
    // body uses pre-declared vars with =
}
```

### 2.4 `recover()` Idiom Exception

`if r := recover(); r != nil { ... }` is allowed (standard Go idiom for panic recovery in goroutines).

---

## 3. Logging Patterns

### 3.1 Logger Import

```go
import (
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)
```

### 3.2 Log Field Initialization

Initialize `logFields` immediately after nil-check and span start:

```go
logFields = map[string]interface{}{
    "requestDTO": requestDTO,
}
```

Add fields incrementally as they become available:
```go
logFields["entity"] = entity
logFields["filter"] = filter
logFields["responseDTO"] = responseDTO
```

### 3.3 Log Message Format

```
[PackageName][FunctionName][StepName] human-readable description in lowercase
```

Examples:
- `[GuestService][Create][Validate] failed to validate dto`
- `[BoilerplateDatabaseRepository][exec][Exec] failed to exec statement`
- `[GuestHandler][BulkCreate][BodyParser] failed to parse request body`
- `[InMemoryDatabaseRepository][Get][Get] failed to get entity`
- `[GuestHandler][HandleCreated] message received`

### 3.4 Log Levels by Error Type

| Error Type | Log Level | Function |
|---|---|---|
| Validation errors | `log.Warn().Err(err)` | DTO validation |
| Client errors (4xx) | `log.WithLevel(zerolog.WarnLevel)` | Service/Handler errors |
| Server errors (5xx) | `log.WithLevel(zerolog.ErrorLevel)` | Service/Handler errors |
| Unexpected errors | `log.Err(err)` | Repository/internal errors |
| Informational | `log.Info()` | Event consumer "message received" |

Standard pattern for dynamic log level:
```go
logLevel = zerolog.WarnLevel
if gocerr.GetErrorCode(err) >= http.StatusInternalServerError {
    logLevel = zerolog.ErrorLevel
}
log.WithLevel(logLevel).Ctx(ctx).Err(err).Fields(logFields).Msg("...")
```

### 3.5 Log Output Format

Always chain in this order: `.Err(err).Ctx(ctx).Fields(logFields).Msg("...")`

```go
log.Err(err).
    Ctx(ctx).
    Fields(logFields).
    Msg("[GuestService][Create][exec] failed to create entity")
```

For `Warn` and `WithLevel`, use the same chain:
```go
log.Warn().Ctx(ctx).Err(err).Fields(logFields).Msg("...")
log.WithLevel(logLevel).Ctx(ctx).Err(err).Fields(logFields).Msg("...")
```

### 3.6 Slow Query Warning

```go
log.Warn().
    Ctx(ctx).
    Fields(logFields).
    Msg(fmt.Sprintf("[BoilerplateDatabaseRepository][%s] slow query", fnName))
```

---

## 4. Tracer (OpenTelemetry)

### 4.1 Span Requirement (MANDATORY)

> ⚠️ **Every function that accepts `context.Context` as a parameter MUST start a tracer span.** This includes public methods, private helpers, handler methods, consumer methods, and repository methods. There are NO exceptions.

The only functions that do NOT need spans are:
- Constructors (`NewXxx`)
- Utility functions without `ctx` parameter
- Methods on value types that don't accept `ctx`

### 4.2 Span Creation

```go
ctx, span = tracer.Start(ctx, "[PackageName][FunctionName]")
defer span.End()
```

Span name format: `[PackageName][FunctionName]` — no step suffix.

### 4.3 Span Coverage (ALL functions with context.Context)

> ⚠️ This list is NOT exhaustive. Any function that accepts `context.Context` as a parameter MUST have a span.

| Layer | Functions Requiring Spans |
|---|---|
| **Service (public)** | `Create`, `DeleteByID`, `UpdateByID`, `FindByID`, `FindAll`, `BulkCreate`, `BulkUpdate`, `BulkDelete`, `ProcessEvent` |
| **Service (private)** | `findEntityByID`, `findListEntity`, `countEntities`, `deleteEntityCaches`, `getListEntityCache`, `setListEntityCache`, `getCountEntitiesCache`, `setEntitiesCountCache`, `getEntityByIDCache`, `setEntityByIDCache` |
| **Repository (statement)** | `Exec`, `Get`, `Select` |
| **Repository (transaction)** | `Commit`, `Rollback`, `Prepare` |
| **Repository (database)** | `exec`, `Count`, `FindAll`, `FindOne`, `Create`, `Update`, `Delete`, `BulkCreate`, `BulkUpdate`, `BeginTransaction` |
| **Repository (cache)** | `Get`, `GetList`, `GetCount`, `Set`, `SetList`, `SetCount`, `Keys`, `Delete`, `Lock`, `Unlock` |
| **Repository (producer)** | `Publish`, `PublishWithDelay`, `PublishBulk`, `PublishBulkWithDelay` |
| **Repository (webhook)** | `SendWebhook` |
| **HTTP Handlers** | `Create`, `FindAll`, `FindByID`, `UpdateByID`, `DeleteByID`, `BulkCreate`, `BulkUpdate`, `BulkDelete` |
| **gRPC Handlers** | `Create`, `FindAll`, `FindByID`, `UpdateByID`, `DeleteByID`, `BulkCreateGuests`, `BulkUpdateGuests`, `BulkDeleteGuests` |
| **Event Consumer** | `HandleCreated`, `HandleDeleted`, `HandleUpdated`, `HandleBulkCreated`, `HandleBulkUpdated`, `HandleBulkDeleted` |
| **Middleware** | All gRPC interceptors, all Fiber middleware that accept `ctx` |

### 4.4 Tracer Propagation for Events

Events carry trace context via `InjectTracerPropagator`:

```go
// In event producer repository — MUST be in ALL 4 publish methods:
ctx, message = message.InjectTracerPropagator(ctx)
```

This is required for: `Publish`, `PublishWithDelay`, `PublishBulk`, `PublishBulkWithDelay`.

---

## 5. Error Handling

### 5.1 Typed Errors

Use `github.com/fikri240794/gocerr` for errors with HTTP status codes:

```go
// Repository layer (SQL database — generic, hides internal details for security):
err = gocerr.New(http.StatusInternalServerError, "error")
```

> ⚠️ **IMPORTANT:** The generic `"error"` message is used ONLY for **SQL database execution errors** (`Exec`, `Get`, `Select`, `Commit`, `Rollback`, `BeginTransaction`). This prevents leaking SQL syntax, table names, or constraint details to API consumers. All other layers (cache, producer, webhook, service, handler) MUST use descriptive error messages with `err.Error()`.

```go
// Cache/Producer/Webhook errors — descriptive (no SQL exposed):
err = gocerr.New(http.StatusInternalServerError, err.Error())

// Service/Handler layer — descriptive:
err = gocerr.New(http.StatusBadRequest, "requestDTO is nil")
err = gocerr.New(http.StatusNotFound, "entity not found")

// With dynamic error message (for build errors):
err = gocerr.New(http.StatusInternalServerError, err.Error())
```

### 5.2 Error Code Checking

```go
errorCode := gocerr.GetErrorCode(err)
if errorCode >= http.StatusInternalServerError {
    // Server error
} else {
    // Client error
}
```

### 5.3 Common Error Responses

| Scenario | Code | Message |
|---|---|---|
| Nil request | 400 | `"requestDTO is nil"` |
| Validation failure | depends on DTO | returned from `Validate()` |
| Entity not found (single) | 404 | `"entity not found"` |
| Entities not found (bulk) | 404 | `"entities not found"` |
| DB query build error | 500 | `err.Error()` (descriptive) |
| DB execution error | 500 | `"error"` (generic, security) |
| Cache miss (redis.Nil) | 404 | `"entity not found"` |
| Cache internal error | 500 | `err.Error()` |

### 5.4 Transaction Error Consistency

Both `Commit()` and `Rollback()` wrap errors with `gocerr`:

```go
// Commit:
err = gocerr.New(http.StatusInternalServerError, "error")

// Rollback:
err = gocerr.New(http.StatusInternalServerError, "error")
```

### 5.5 gRPC Error Conversion

```go
return nil, gocerr.New(http.StatusBadRequest, "message").ToGRPCError()
```

---

## 6. Config Layer

### 6.1 File Location

`configs/config.go` — loaded from `.env` file.

### 6.2 Config Reading

```go
const DefaultConfigPath string = "./configs/.env"

func Read(cfgpath string) *Config
```

Uses `viper` to read `.env`, `mapstructure` tags to map to struct.

### 6.3 Config Struct Hierarchy

```go
type Config struct {
    Environment string              `mapstructure:"ENVIRONMENT"`
    Server      ServerConfig        `mapstructure:"SERVER"`
    Datasource  DatasourceConfig    `mapstructure:"DATASOURCE"`
    Guest       GuestConfig         `mapstructure:"GUEST"`
}
```

**ServerConfig:**
- `Server.Name` — Service name
- `Server.LogLevel` — Zerolog level (int8)
- `Server.HTTP.Port`, `Prefork`, `PrintRoutes`, `RequestTimeout`, `GracefullyShutdownDuration`
- `Server.HTTP.CORS.AllowOrigins`, `AllowMethods`
- `Server.HTTP.Docs.Swagger.Enable`, `FilePath`, `Path`, `Title`
- `Server.GRPC.Port`, `RequestTimeout`
- `Server.EventConsumer.DataSourceName` — NSQ lookupd address
- `Server.Tracer.ServiceName`, `ExporterGRPCAddress` — OpenTelemetry

**DatasourceConfig:**
- `BoilerplateDatabase.Master` — `DriverName`, `DataSourceName`, pool settings, `MaximumQueryDurationWarning`
- `BoilerplateDatabase.Slave` — Same fields as Master
- `InMemoryDatabase.DataSourceName` — Redis address
- `EventProducer.DataSourceName` — NSQ address
- `WebhookSiteHTTPClient.BaseURL`, `Endpoint.Webhook` — external webhook URL

**GuestConfig:**
- `Guest.Cache.Enable`, `Keyf` (format string, e.g. `"guest:%s"`), `Duration`
- `Guest.Event.Created.Enable`, `Topic`
- `Guest.Event.Deleted.Enable`, `Topic`
- `Guest.Event.Updated.Enable`, `Topic`
- `Guest.Event.BulkCreated.Enable`, `Topic`
- `Guest.Event.BulkUpdated.Enable`, `Topic`
- `Guest.Event.BulkDeleted.Enable`, `Topic`

### 6.4 Event Config Pattern

All 6 event configs share the same shape (inline anonymous structs):
```go
Event struct {
    Created     struct { Enable bool; Topic string } `mapstructure:"CREATED"`
    Deleted     struct { Enable bool; Topic string } `mapstructure:"DELETED"`
    Updated     struct { Enable bool; Topic string } `mapstructure:"UPDATED"`
    BulkCreated struct { Enable bool; Topic string } `mapstructure:"BULK_CREATED"`
    BulkUpdated struct { Enable bool; Topic string } `mapstructure:"BULK_UPDATED"`
    BulkDeleted struct { Enable bool; Topic string } `mapstructure:"BULK_DELETED"`
}
```

---

## 7. Datasources Layer

### 7.1 Database (PostgreSQL)

**File:** `datasources/boilerplate_database/datasource.go`

```go
type BoilerplateDatabase struct {
    Master                        *sqlx.DB
    MasterMaxQueryDurationWarning time.Duration
    Slave                         *sqlx.DB
    SlaveMaxQueryDurationWarning  time.Duration
}

func NewBoilerplateDatabase(cfg *configs.Config) *BoilerplateDatabase
```

**Connection pool settings:** MaximumOpenConnections, MaximumIdleConnections, ConnectionMaxIdleTime, ConnectionMaxLifetime.

**Master/Slave pattern:** All writes go to Master, reads can go to Slave. The `useMaster` boolean parameter selects the source for read operations.

### 7.2 Redis Cache

**File:** `datasources/in_memory_database/datasource.go`

```go
type InMemoryDatabase struct {
    RedisClient *redis.Client
    RedisTracer *redis.Client  // Instrumented with OpenTelemetry
}

func NewInMemoryDatabase(cfg *configs.Config) *InMemoryDatabase
```

### 7.3 NSQ Producer

**File:** `datasources/event_producer/datasource.go`

```go
type EventProducer struct {
    NSQProducer *nsq.Producer
}

func NewEventProducer(cfg *configs.Config) *EventProducer
```

Provides:
- `NSQProducer.Publish(topic, body)` — synchronous publish
- `NSQProducer.DeferredPublish(topic, delay, body)` — deferred publish

### 7.4 HTTP Client

**File:** `datasources/webhook_site_http_client/datasource.go`

```go
type WebhookSiteHTTPClient struct {
    config *configs.Config
    client *http.Client
}

func NewWebhookSiteHTTPClient(cfg *configs.Config) *WebhookSiteHTTPClient
```

Used for outbound webhook calls to external services.

---

## 8. Entity Layer (Models)

### 8.1 Location

`internal/models/entities/`

### 8.2 GuestEntity — Domain Entity

```go
type GuestEntity struct {
    Table     string      `table:"guests" db:"-" json:"-"`
    ID        uuid.UUID   `db:"id" json:"id" primary_key:"true" db_type:"uuid"`
    Name      string      `db:"name" json:"name" db_type:"text"`
    Address   null.String `db:"address" json:"address" db_type:"text"`
    CreatedAt int64       `db:"created_at" json:"created_at" db_type:"bigint"`
    CreatedBy string      `db:"created_by" json:"created_by" db_type:"text"`
    UpdatedAt null.Int64  `db:"updated_at" json:"updated_at" db_type:"bigint"`
    UpdatedBy null.String `db:"updated_by" json:"updated_by" db_type:"text"`
    DeletedAt null.Int64  `db:"deleted_at" json:"deleted_at" db_type:"bigint"`
    DeletedBy null.String `db:"deleted_by" json:"deleted_by" db_type:"text"`
}
```

### 8.3 Tag Conventions

| Tag | Purpose | Example Values |
|---|---|---|
| `table` | Table name for the entity | `"guests"` |
| `db` | Database column name | `"id"`, `"name"`, `"-"` (skip field) |
| `primary_key` | Marks the primary key field | `"true"` |
| `db_type` | SQL type name for type casting | `"uuid"`, `"text"`, `"bigint"` |
| `json` | JSON serialization name | `"id"`, `"address"`, `"-"` (skip) |

### 8.4 Database Field Name Constants

```go
const (
    GuestEntityDatabaseFieldID        string = "id"
    GuestEntityDatabaseFieldName      string = "name"
    GuestEntityDatabaseFieldAddress   string = "address"
    GuestEntityDatabaseFieldCreatedAt string = "created_at"
    GuestEntityDatabaseFieldCreatedBy string = "created_by"
    GuestEntityDatabaseFieldUpdatedAt string = "updated_at"
    GuestEntityDatabaseFieldUpdatedBy string = "updated_by"
    GuestEntityDatabaseFieldDeletedAt string = "deleted_at"
    GuestEntityDatabaseFieldDeletedBy string = "deleted_by"
)
```

### 8.5 Entity Methods

```go
func (entity *GuestEntity) MarkAsDeleted(deletedBy string) *GuestEntity
    // Sets DeletedAt = now, DeletedBy = deletedBy.
    // Returns self for method chaining.
```

### 8.6 Event Entities

```go
type GuestEventEntity struct {
    ID        uuid.UUID   `json:"id"`
    Name      string      `json:"name"`
    Address   null.String `json:"address"`
    CreatedAt int64       `json:"created_at"`
    CreatedBy string      `json:"created_by"`
    UpdatedAt null.Int64  `json:"updated_at"`
    UpdatedBy null.String `json:"updated_by"`
}

type EventEntity[T any] struct {
    ID          uuid.UUID             `json:"id"`
    CreatedAt   int64                 `json:"created_at"`
    Publisher   string                `json:"publisher"`
    Propagator  map[string]string     `json:"propagator"`
    Message     T                     `json:"message"`
}
```

### 8.7 Event Entity Constructors

```go
// Converts domain entity to event entity
func NewGuestEventEntity(entity *GuestEntity) *GuestEventEntity

// Creates a generic event envelope with OpenTelemetry propagation support
func NewEventEntity[T any](topic string, message T) *EventEntity[T]
```

### 8.8 Tracer Propagation on EventEntity

```go
func (e *EventEntity[T]) InjectTracerPropagator(ctx context.Context) (context.Context, *EventEntity[T])
    // Injects OpenTelemetry trace context into the event's Propagator map.
    // Returns the modified context and event.
```

---

## 9. DTO Layer

### 9.1 Location

`internal/models/dtos/guest_dto.go`

### 9.2 Request DTOs

Every request DTO has:
- A `Validate() error` method
- Conversion methods (`.ToEntity()`, `.ToDTO()`, `.ToIDs()`, etc.)

| DTO | Validate() validates | Conversion methods |
|---|---|---|
| `CreateGuestRequestDTO` | Name required, CreatedBy required | `ToEntity() *GuestEntity` |
| `DeleteGuestByIDRequestDTO` | ID is valid UUID | — |
| `FindGuestByIDRequestDTO` | ID is valid UUID | — |
| `FindAllGuestRequestDTO` | — (has defaults) | `ToFilterAndSorts() (filter, sorts, err)` |
| `UpdateGuestByIDRequestDTO` | Name required, UpdatedBy required | `ToExistingEntity(existing) *GuestEntity` (merges fields) |
| `BulkCreateGuestsRequestDTO` | All items valid | `ToEntities() []GuestEntity` |
| `BulkUpdateGuestsRequestDTO` | All items valid | `ToIDs() []string` |
| `BulkDeleteGuestsRequestDTO` | All IDs valid UUIDs | `ToIDs() []string` |
| `GuestEventRequestDTO` | — | `ToEntity() *GuestEventEntity` |

### 9.3 Validation Pattern

```go
func (dto *XxxRequestDTO) Validate() error {
    var err error

    if dto.Name == "" {
        return gocerr.New(http.StatusBadRequest, "name is required")
    }

    // UUID validation
    _, err = uuid.FromString(dto.ID)
    if err != nil {
        return gocerr.New(http.StatusBadRequest, err.Error())
    }

    return nil
}
```

### 9.4 Response DTO Constructors

```go
// Single operations:
func NewGuestResponseDTO(entity *entities.GuestEntity) *GuestResponseDTO

// Bulk operations:
func NewBulkCreateGuestsResponseDTO(entities []entities.GuestEntity) *BulkCreateGuestsResponseDTO
func NewBulkUpdateGuestsResponseDTO(entities []entities.GuestEntity) *BulkUpdateGuestsResponseDTO

// Paginated:
func NewFindAllGuestResponseDTO(entities []entities.GuestEntity, count uint64) *FindAllGuestResponseDTO

// Events:
func NewGuestEventResponseDTO(entity *entities.GuestEventEntity) *GuestEventResponseDTO
```

### 9.5 Generic HTTP Response Wrapper

Uses `github.com/fikri240794/gores`:

```go
type ResponseVM[T comparable] struct {
    Code  int       `json:"code"`
    Data  *T        `json:"data,omitempty"`
    Error *ErrorVM  `json:"error,omitempty"`
}
```

Since `T` has `comparable` constraint (cannot be a slice), use pointer-to-slice:
```go
// ✅ Correct for slice data:
gores.NewResponseVM[*[]vms.GuestResponseVM]()

// ❌ Wrong:
gores.NewResponseVM[[]vms.GuestResponseVM]()  // compile error: []T does not implement comparable
```

---

## 10. Repository Layer

### 10.1 Location

`internal/repositories/`

### 10.2 Generic Repository Interfaces

```go
type IBoilerplateDatabaseRepository[TEntity interface{}] interface {
    BeginTransaction(ctx context.Context) (IBoilerplateDatabaseTransaction, error)
    BulkCreate(ctx context.Context, entities []TEntity) error
    BulkUpdate(ctx context.Context, entities []TEntity) error
    Count(ctx context.Context, filter *goqube.Filter, useMaster bool) (uint64, error)
    Create(ctx context.Context, entity *TEntity) error
    Delete(ctx context.Context, filter *goqube.Filter) error
    FindAll(ctx context.Context, filter *goqube.Filter, sorts []goqube.Sort, take uint64, skip uint64, useMaster bool) ([]TEntity, error)
    FindOne(ctx context.Context, filter *goqube.Filter, sorts []goqube.Sort, useMaster bool) (*TEntity, error)
    Update(ctx context.Context, entity *TEntity, filter *goqube.Filter) error
}

type IBoilerplateDatabaseTransaction interface {
    Commit() error
    DriverName() string
    Prepare(ctx context.Context, query string) (IBoilerplateDatabaseStatement, error)
    Rollback() error
}

type IBoilerplateDatabaseStatement interface {
    Exec(ctx context.Context, args ...interface{}) error
    Get(ctx context.Context, dest interface{}, args ...interface{}) error
    Select(ctx context.Context, dest interface{}, args ...interface{}) error
    Close() error
}

type IEventProducerRepository[TEntity interface{}] interface {
    Publish(ctx context.Context, topic string, message *entities.EventEntity[TEntity]) error
    PublishWithDelay(ctx context.Context, topic string, delay time.Duration, message *entities.EventEntity[TEntity]) error
    PublishBulk(ctx context.Context, topic string, message *entities.EventEntity[[]TEntity]) error
    PublishBulkWithDelay(ctx context.Context, topic string, delay time.Duration, message *entities.EventEntity[[]TEntity]) error
}

type IInMemoryDatabaseRepository[TEntity interface{}] interface {
    Delete(ctx context.Context, keys ...string) error
    Get(ctx context.Context, key string) (*TEntity, error)
    GetList(ctx context.Context, key string) ([]TEntity, error)
    GetCount(ctx context.Context, key string) (uint64, error)
    Keys(ctx context.Context, pattern string) ([]string, error)
    Lock(ctx context.Context, key string, expiration time.Duration, retry time.Duration, delay time.Duration) error
    Set(ctx context.Context, key string, value *TEntity, expiration time.Duration) error
    SetList(ctx context.Context, key string, values []TEntity, expiration time.Duration) error
    SetCount(ctx context.Context, key string, count uint64, expiration time.Duration) error
    Unlock(ctx context.Context, key string) error
}
```

### 10.3 Concrete Repository Interfaces

```go
type IGuestRepository interface {
    IBoilerplateDatabaseRepository[entities.GuestEntity]
    WithTransaction(tx IBoilerplateDatabaseTransaction) IGuestRepository
}

type IGuestCacheRepository interface {
    IInMemoryDatabaseRepository[entities.GuestEntity]
}

type IGuestEventProducerRepository interface {
    IEventProducerRepository[entities.GuestEventEntity]
}

type IWebhookSiteRepository interface {
    SendWebhook(ctx context.Context, entity *entities.EventEntity[entities.GuestEventEntity]) error
}
```

### 10.4 Concrete Repository Structs

```go
type GuestRepository struct {
    BoilerplateDatabaseRepository[entities.GuestEntity]
}

type GuestCacheRepository struct {
    InMemoryDatabaseRepository[entities.GuestEntity]
}

type GuestEventProducerRepository struct {
    EventProducerRepository[entities.GuestEventEntity]
}

type WebhookSiteRepository struct {
    config                     *configs.Config
    webhookSiteHTTPClient      *webhook_site_http_client.WebhookSiteHTTPClient
}
```

### 10.5 WithTransaction Pattern (Decorator)

```go
type guestRepositoryWithTransaction struct {
    GuestRepository
    tx repositories.IBoilerplateDatabaseTransaction
}

func (r *GuestRepository) WithTransaction(tx repositories.IBoilerplateDatabaseTransaction) IGuestRepository {
    return &guestRepositoryWithTransaction{
        GuestRepository: GuestRepository{ /* copy fields */ },
        tx:              tx,
    }
}
```

Usage in services:
```go
err = s.guestRepository.WithTransaction(tx).Create(ctx, entity)
```

### 10.6 entityMeta — Reflection Metadata

```go
type entityMeta struct {
    TableName     string
    PrimaryKey    string
    FieldTypeMap  map[string]string       // db_field -> db_type
    FieldValueMap map[string]interface{}  // db_field -> value
}

func (r *BoilerplateDatabaseRepository[TEntity]) getEntityMeta(entity *TEntity) entityMeta
```

**Tag processing order** (must be in this exact sequence):
1. `table` tag → set `TableName`, `continue` (skip from value map)
2. `db` tag → if empty or `"-"`, skip entire field; else use as map key
3. Extract field value: `FieldValueMap[dbTag] = reflect.ValueOf(entity).Elem().Field(i).Interface()`
4. `db_type` tag → if non-empty, `FieldTypeMap[dbTag] = tagValue`
5. `primary_key` tag → if `"true"`, `PrimaryKey = dbTag`

### 10.7 getTableNameAndFields

```go
func (r *BoilerplateDatabaseRepository[TEntity]) getTableNameAndFields() (string, []string)
```

Returns table name and list of `db` tag values for SELECT queries. Fields with `table` tag are skipped (via `continue`), same as `getEntityMeta`.

### 10.8 Statement Preparation Helper

```go
func (r *BoilerplateDatabaseRepository[TEntity]) prepareQueryStatement(
    ctx context.Context,
    logFields map[string]interface{},
    query string,
    useMaster bool,
    fnName string,
) (IBoilerplateDatabaseStatement, error)
```

**Routing logic:**
- `useMaster=true, r.tx!=nil` → `r.tx.Prepare(ctx, query)`
- `useMaster=true, r.tx==nil` → `r.db.Master.PreparexContext(ctx, query)`, wrap in `IBoilerplateDatabaseStatement`
- `useMaster=false` → `r.db.Slave.PreparexContext(ctx, query)`, wrap in `IBoilerplateDatabaseStatement`

**Used by:** `Count`, `FindAll`, `FindOne`, `exec`.

### 10.9 Slow Query Logging Helper

```go
func (r *BoilerplateDatabaseRepository[TEntity]) logSlowQuery(
    logFields map[string]interface{},
    duration time.Duration,
    threshold time.Duration,
    fnName string,
)
```

Adds `duration` to `logFields` and logs warning if `duration > threshold`.

### 10.10 Read Methods and DB Source

| Method | DB Source | useMaster parameter |
|---|---|---|
| `Count` | Slave (default) or Master | ✅ Yes |
| `FindAll` | Slave (default) or Master | ✅ Yes |
| `FindOne` | Slave (default) or Master | ✅ Yes |
| `Create` | Master (via `exec`) | ❌ Always Master |
| `Update` | Master (via `exec`) | ❌ Always Master |
| `Delete` | Master (via `exec`) | ❌ Always Master |
| `BulkCreate` | Master (via `exec`) | ❌ Always Master |
| `BulkUpdate` | Master (via `exec`) | ❌ Always Master |

### 10.11 Error Messages by Method

| Method | Build Error | Execution Error | Not Found |
|---|---|---|---|
| `Count` | `gocerr.New(500, err.Error())` | `gocerr.New(500, "error")` | `gocerr.New(404, "entity not found")` |
| `FindAll` | `gocerr.New(500, err.Error())` | `gocerr.New(500, "error")` | — |
| `FindOne` | `gocerr.New(500, err.Error())` | `gocerr.New(500, "error")` | `gocerr.New(404, "entity not found")` |
| `Create/Update/Delete` | `gocerr.New(500, err.Error())` | Via `exec`: `gocerr.New(500, "error")` | — |
| `BulkCreate/BulkUpdate` | `gocerr.New(500, err.Error())` | Via `exec`: `gocerr.New(500, "error")` | — |
| Cache (in-memory) | — | Dynamic (404 for `redis.Nil`, 500 otherwise) | `gocerr.New(404, "entity not found")` |
| Event producer | — | `gocerr.New(500, err.Error())` | — |
| Webhook | — | `gocerr.New(500, err.Error())` | — |

### 10.12 In-Memory Cache Error Handling

```go
// Get, GetList, GetCount:
if err != nil {
    errorCode = http.StatusNotFound
    if err != redis.Nil {
        errorCode = http.StatusInternalServerError
        log.Err(err).Ctx(ctx).Fields(logFields).Msg("...")
    }
    return nil, gocerr.New(errorCode, err.Error())
}
```

### 10.13 Webhook HTTP Client Pattern

```go
func (r *WebhookSiteRepository) SendWebhook(ctx context.Context, entity *entities.EventEntity[entities.GuestEventEntity]) error {
    var (
        span     trace.Span
        body     []byte
        req      *http.Request
        resp     *http.Response
        logFields map[string]interface{}
        err      error
    )
    ctx, span = tracer.Start(ctx, "[WebhookSiteRepository][SendWebhook]")
    defer span.End()

    body, err = json.Marshal(entity)
    // ... build request, execute, check response status
}
```

---

## 11. Service Layer

### 11.1 Interface

```go
type IGuestService interface {
    BulkCreate(ctx context.Context, requestDTO *dtos.BulkCreateGuestsRequestDTO) (*dtos.BulkCreateGuestsResponseDTO, error)
    BulkDelete(ctx context.Context, requestDTO *dtos.BulkDeleteGuestsRequestDTO) error
    BulkUpdate(ctx context.Context, requestDTO *dtos.BulkUpdateGuestsRequestDTO) (*dtos.BulkUpdateGuestsResponseDTO, error)
    Create(ctx context.Context, requestDTO *dtos.CreateGuestRequestDTO) (*dtos.GuestResponseDTO, error)
    DeleteByID(ctx context.Context, requestDTO *dtos.DeleteGuestByIDRequestDTO) error
    FindAll(ctx context.Context, requestDTO *dtos.FindAllGuestRequestDTO) (*dtos.FindAllGuestResponseDTO, error)
    FindByID(ctx context.Context, requestDTO *dtos.FindGuestByIDRequestDTO) (*dtos.GuestResponseDTO, error)
    UpdateByID(ctx context.Context, requestDTO *dtos.UpdateGuestByIDRequestDTO) (*dtos.GuestResponseDTO, error)
    ProcessEvent(ctx context.Context, requestDTO *dtos.GuestEventRequestDTO) (*dtos.GuestEventResponseDTO, error)
}
```

### 11.2 GuestService Struct

```go
type GuestService struct {
    cfg                          *configs.Config
    guestRepository              repositories.IGuestRepository
    guestCacheRepository         repositories.IGuestCacheRepository
    guestEventProducerRepository repositories.IGuestEventProducerRepository
    webhookSiteRepository        repositories.IWebhookSiteRepository
}
```

### 11.3 Function Structure Template

Every business function follows this exact sequence:

```go
func (s *GuestService) Xxx(ctx context.Context, requestDTO *dtos.XxxRequestDTO) (*dtos.XxxResponseDTO, error) {
    // ── 1. All variables declared at top ──
    var (
        span        trace.Span
        logFields   map[string]interface{}
        entity      *entities.GuestEntity
        responseDTO *dtos.GuestResponseDTO
        err         error
    )

    // ── 2. Start tracer span ──
    ctx, span = tracer.Start(ctx, "[GuestService][Xxx]")
    defer span.End()

    // ── 3. Nil-check requestDTO ──
    if requestDTO == nil {
        return nil, gocerr.New(http.StatusBadRequest, "requestDTO is nil")
    }

    // ── 4. Initialize log fields ──
    logFields = map[string]interface{}{
        "requestDTO": requestDTO,
    }

    // ── 5. Validate DTO ──
    err = requestDTO.Validate()
    if err != nil {
        log.Warn().Ctx(ctx).Err(err).Fields(logFields).Msg("[GuestService][Xxx][Validate] failed to validate dto")
        return nil, err
    }

    // ── 6. Business logic ──
    entity = requestDTO.ToEntity()
    logFields["entity"] = entity

    // ── 7. Transaction (only for mutations) ──
    err = s.withTransaction(ctx, logFields, "Xxx", func(tx repositories.IBoilerplateDatabaseTransaction) error {
        return s.guestRepository.WithTransaction(tx).Create(ctx, entity)
    })
    if err != nil { return nil, err }

    // ── 8. Build response DTO ──
    responseDTO = dtos.NewGuestResponseDTO(entity)
    logFields["responseDTO"] = responseDTO

    // ── 9. Cache invalidation (mutations only) ──
    s.tryDeleteEntityCaches(ctx, logFields, "Xxx")

    // ── 10. Event publishing (mutations only, if enabled) ──
    s.publishEvent(ctx, logFields, s.cfg.Guest.Event.Xxx.Enable, s.cfg.Guest.Event.Xxx.Topic, "Xxx", *entity)

    return responseDTO, nil
}
```

### 11.4 Helper Function Reference

| Helper | Signature | Used In |
|---|---|---|
| `withTransaction` | `(ctx, logFields, fnName, fn func(tx) error) error` | Create, DeleteByID, UpdateByID, BulkCreate, BulkUpdate, BulkDelete |
| `buildActiveEntityFilterByIDs` | `(ids ...string) *goqube.Filter` | Single ID (OperatorEqual) or multiple IDs (OperatorIn) |
| `findEntityByID` | `(ctx, cacheKey, filter) (*GuestEntity, error)` | FindByID |
| `findListEntity` | `(ctx, cacheKey, filter, sorts, take, skip) ([]GuestEntity, error)` | FindAll |
| `countEntities` | `(ctx, cacheKey, filter) (uint64, error)` | FindAll |
| `getEntityByIDCache` | `(ctx, cacheKey) (*GuestEntity, error)` | findEntityByID |
| `setEntityByIDCache` | `(ctx, cacheKey, entity) error` | findEntityByID |
| `getListEntityCache` | `(ctx, cacheKey) ([]GuestEntity, error)` | findListEntity |
| `setListEntityCache` | `(ctx, cacheKey, list) error` | findListEntity |
| `getCountEntitiesCache` | `(ctx, cacheKey) (uint64, error)` | countEntities |
| `setEntitiesCountCache` | `(ctx, cacheKey, count) error` | countEntities |
| `deleteEntityCaches` | `(ctx) error` | tryDeleteEntityCaches |
| `tryDeleteEntityCaches` | `(ctx, logFields, fnName)` | All 6 mutation functions |
| `publishEvent` | `(ctx, logFields, enable, topic, fnName, entities ...GuestEntity)` | Single via `*entity`, Bulk via `entities...` |

### 11.5 Cache-Aside Pattern

Used in `findEntityByID`, `findListEntity`, `countEntities`:

```go
// Step 1: Try cache (if enabled)
if s.cfg.Guest.Cache.Enable {
    data, err = s.getXxxCache(ctx, key)
    if err != nil && gocerr.GetErrorCode(err) >= http.StatusInternalServerError {
        log.Err(err)...  // Log cache errors only for server errors
    }
    if err == nil && data != nil {
        return data, nil  // Cache hit
    }
}

// Step 2: Fall back to database
data, err = s.guestRepository.FindAll(ctx, filter, ...)
if err != nil {
    if gocerr.GetErrorCode(err) >= http.StatusInternalServerError {
        log.Err(err)...
    }
    return nil, err
}

// Step 3: Populate cache (if enabled and data exists)
if s.cfg.Guest.Cache.Enable && len(data) > 0 {
    err = s.setXxxCache(ctx, key, data)
    if err != nil {
        log.Err(err)...  // Non-fatal: log but don't return
        err = nil
    }
}
```

### 11.6 Parallel Execution (gotask)

Used in `FindAll`:

```go
errTask, errTaskCtx = gotask.NewErrorTask(ctx, 2)  // 2 goroutines

errTask.Go(func() error {
    var errRoutine error
    listEntity, errRoutine = s.findListEntity(errTaskCtx, ...)
    return errRoutine
})

errTask.Go(func() error {
    var errRoutine error
    entitiesCount, errRoutine = s.countEntities(errTaskCtx, ...)
    return errRoutine
})

err = errTask.Wait()
if err != nil {
    log.Err(err).Ctx(ctx).Fields(logFields).Msg("[GuestService][FindAll][Wait] failed to find or count entities")
    return nil, err
}
```

---

## 12. Transport Layer — HTTP

### 12.1 Server Build

`transports/http/server.go`:
```go
func BuildHTTPServer(cfg *configs.Config) *Server
```

Returns a Fiber app wrapped in a `Server` struct that provides `ServeHTTP()` and `Close()`.

### 12.2 Handler Struct

```go
type GuestHandler struct {
    guestService services.IGuestService
}

func NewGuestHandler(guestService services.IGuestService) *GuestHandler
```

### 12.3 Handler Method Pattern

```go
func (h *GuestHandler) Xxx(c *fiber.Ctx) error {
    var (
        ctx         context.Context
        span        trace.Span
        logFields   map[string]interface{}
        requestVM   *vms.XxxRequestVM
        requestDTO  *dtos.XxxRequestDTO
        responseDTO *dtos.XxxResponseDTO
        logLevel    zerolog.Level
        responseVM  *gores.ResponseVM[...]
        err         error
    )

    ctx = c.UserContext()
    ctx, span = tracer.Start(ctx, "[GuestHandler][Xxx]")
    defer span.End()

    logFields = map[string]interface{}{}

    // Parse request body
    requestVM = &vms.XxxRequestVM{}
    err = c.BodyParser(requestVM)  // or c.ParamsParser, c.QueryParser
    if err != nil {
        log.Warn().Ctx(ctx).Err(err).Fields(logFields).Msg("[GuestHandler][Xxx][BodyParser] failed to parse request body")
        responseVM = gores.NewResponseVM[bool]().SetErrorFromError(err)
        return c.Status(responseVM.Code).JSON(responseVM)
    }
    logFields["requestVM"] = requestVM

    // Convert VM -> DTO
    requestDTO = requestVM.ToDTO()
    logFields["requestDTO"] = requestDTO

    // Call service
    responseDTO, err = h.guestService.Xxx(ctx, requestDTO)
    if err != nil {
        logLevel = zerolog.WarnLevel
        if gocerr.GetErrorCode(err) >= fiber.StatusInternalServerError {
            logLevel = zerolog.ErrorLevel
        }
        log.WithLevel(logLevel).Ctx(ctx).Err(err).Fields(logFields).Msg("[GuestHandler][Xxx][Xxx] ...")
        responseVM = gores.NewResponseVM[...]().SetErrorFromError(err)
        return c.Status(responseVM.Code).JSON(responseVM)
    }

    // Success response
    responseVM = gores.NewResponseVM[...]().SetCode(fiber.StatusOK).SetData(...)
    return c.Status(responseVM.Code).JSON(responseVM)
}
```

### 12.4 Route Registration

```go
func (h *GuestHandler) SetupRoutes(api fiber.Router) {
    guests := api.Group("/guests")
    guests.Post("/", h.Create)
    guests.Get("/", h.FindAll)
    guests.Get("/:id", h.FindByID)
    guests.Put("/:id", h.UpdateByID)
    guests.Delete("/:id", h.DeleteByID)
    guests.Post("/bulk", h.BulkCreate)
    guests.Put("/bulk", h.BulkUpdate)
    guests.Delete("/bulk", h.BulkDelete)
}
```

### 12.5 HTTP Middleware

**Location:** `transports/http/middlewares/`

Middleware are implemented as **structs with constructor functions** (not bare functions) to support dependency injection via Google Wire. Each middleware method accepts `*fiber.Ctx` and returns an error:

```go
// Struct with Wire constructor
func NewXxxMiddleware() *XxxMiddleware {
    return &XxxMiddleware{}
}

func (mw *XxxMiddleware) Xxx(c *fiber.Ctx) error {
    // Pre-processing
    err := c.Next()
    // Post-processing
    return err
}
```

Middleware that require config inject it via the constructor:
```go
func NewTimeoutMiddleware(cfg *configs.Config) *TimeoutMiddleware {
    return &TimeoutMiddleware{cfg: cfg}
}
```

**Middleware struct (`middlewares.go`):**
```go
type Middlewares struct {
    Recover   *RecoverMiddleware
    Tracer    *TracerMiddleware
    RequestID *RequestIDMiddleware
    Log       *LogMiddleware
    Timeout   *TimeoutMiddleware
}
```

**Wire provider (`provider.go`):**
```go
var Provider wire.ProviderSet = wire.NewSet(
    NewRecoverMiddleware,
    NewTracerMiddleware,
    NewRequestIDMiddleware,
    NewLogMiddleware,
    NewTimeoutMiddleware,
)
```

Injected into `HTTPServer` via `builder.go`:
```go
wire.Struct(new(middlewares.Middlewares), "*"),
```

The `Middlewares` struct is used directly as a field of `HTTPServer`:
```go
type HTTPServer struct {
    middlewares *middlewares.Middlewares
    ...
}
```

**Registration order** (`setupGlobalMiddlewares()` in `transports/http/http_server.go`):
```go
s.server.Use(
    s.middlewares.Recover.Recover,    // 1. Panic recovery (catch panics from all downstream)
    s.middlewares.Tracer.Start,       // 2. Extract trace context from request headers
    s.middlewares.RequestID.Generate, // 3. Generate/propagate request ID
    s.middlewares.Log.Log,            // 4. Request/response logging
    cors.New(cors.Config{...}),       // 5. CORS (config-driven)
    etag.New(),                       // 6. ETag caching
    favicon.New(),                    // 7. Favicon
)

// Swagger (conditional, only when enabled in config)
if s.cfg.Server.HTTP.Docs.Swagger.Enable {
    s.server.Use(swagger.New(swagger.Config{...}))
}

// Timeout is registered LAST (after routes are set up in handlers.SetupRoutes)
s.server.Use(s.middlewares.Timeout.Timeout)
```

**Available middleware details:**

| Middleware | Method | Config | Description |
|---|---|---|---|
| **Recover** | `Recover(c *fiber.Ctx) error` | — | Catches panics via `defer recover()`, returns `500` with `gores.ResponseVM[string]`, logs the panic stack trace |
| **Tracer** | `Start(c *fiber.Ctx) error` | — | Extracts OpenTelemetry trace context from incoming request headers (`otel.GetTextMapPropagator().Extract`) and sets it on `c.UserContext()` |
| **RequestID** | `Generate(c *fiber.Ctx) error` | — | Reads `X-Request-ID` from request header; if empty generates a new UUIDv7 via `uuid.NewV7()`. Stores in context and sets response header |
| **Log** | `Log(c *fiber.Ctx) error` | — | Measures latency, logs path/method/status/headers/body. Uses `zerolog.InfoLevel` (success), `WarnLevel` (4xx), `ErrorLevel` (5xx) |
| **Timeout** | `Timeout(c *fiber.Ctx) error` | `Server.HTTP.RequestTimeout` | Wraps context with `context.WithTimeout`. Registered **after** route setup so static file serving is not interrupted |
| **CORS** | (fiber built-in) | `Server.HTTP.CORS.AllowOrigins/Methods` | Configured via `cors.New(cors.Config{...})` in `setupGlobalMiddlewares()` |

> All custom middleware methods that accept `ctx` must have `tracer.Start(ctx, "[MiddlewareName][Method]")` + `defer span.End()` for distributed tracing.

### 12.6 ViewModel (VM) Pattern

**Location:** `transports/http/models/vms/guest_vm.go`

Request VMs have JSON tags and `ToDTO()`:
```go
type CreateGuestRequestVM struct {
    Name      string `json:"name" validate:"required"`
    Address   string `json:"address"`
    CreatedBy string `json:"created_by" validate:"required"`
}

func (vm *CreateGuestRequestVM) ToDTO() *dtos.CreateGuestRequestDTO
```

Response VMs have constructors:
```go
type GuestResponseVM struct {
    ID        string      `json:"id"`
    Name      string      `json:"name"`
    Address   null.String `json:"address"`
    CreatedAt int64       `json:"created_at"`
    CreatedBy string      `json:"created_by"`
    UpdatedAt null.Int64  `json:"updated_at"`
    UpdatedBy null.String `json:"updated_by"`
}

func NewGuestResponseVM(responseDTO *dtos.GuestResponseDTO) GuestResponseVM
```

### 12.6 Error Response

```go
responseVM = gores.NewResponseVM[bool]().
    SetErrorFromError(err)
return c.Status(responseVM.Code).JSON(responseVM)
```

### 12.7 Success Response Patterns

```go
// Single entity:
responseVM = gores.NewResponseVM[*vms.GuestResponseVM]().
    SetCode(fiber.StatusOK).
    SetData(vms.NewGuestResponseVM(responseDTO))

// Bulk (slice via pointer):
responseVM = gores.NewResponseVM[*[]vms.GuestResponseVM]().
    SetCode(fiber.StatusOK).
    SetData(vms.NewBulkCreateGuestsResponseVM(responseDTO))

// Boolean (for delete):
responseVM = gores.NewResponseVM[bool]().
    SetCode(fiber.StatusOK).
    SetData(true)

// Paginated list:
responseVM = gores.NewResponseVM[*vms.FindAllGuestResponseVM]().
    SetCode(fiber.StatusOK).
    SetData(vms.NewFindAllGuestResponseVM(responseDTO))
```

### 12.8 Parser Types

| Handler | Parser |
|---|---|
| `Create` | `c.BodyParser` |
| `FindAll` | `c.QueryParser` |
| `FindByID` | `c.ParamsParser` |
| `UpdateByID` | `c.ParamsParser` + `c.BodyParser` |
| `DeleteByID` | `c.ParamsParser` (no body) |
| `BulkCreate` | `c.BodyParser` |
| `BulkUpdate` | `c.BodyParser` |
| `BulkDelete` | `c.BodyParser` |

### 12.9 Swagger Annotations

Every HTTP handler MUST have swagger annotations. Generated by `swaggo/swag` to `transports/http/docs/swagger/` (swagger.json, swagger.yaml, docs.go).

**Complete annotation template:**

```go
// @Summary Create a guest
// @Description Create a new guest record
// @Tags Guest
// @Accept json
// @Produce json
// @Param body body vms.CreateGuestRequestVM true "Create guest request"
// @Success 200 {object} gores.ResponseVM[vms.GuestResponseVM]
// @Failure 400 {object} gores.ResponseErrorVM
// @Failure 500 {object} gores.ResponseErrorVM
// @Router /guests [post]
func (h *GuestHandler) Create(c *fiber.Ctx) error { ... }
```

**Annotation patterns by HTTP method:**

| Method | Route | Success Type | Parser | Notes |
|---|---|---|---|---|
| `POST` | `/guests` | `gores.ResponseVM[vms.GuestResponseVM]` | BodyParser | Create single |
| `GET` | `/guests` | `gores.ResponseVM[*vms.FindAllGuestResponseVM]` | QueryParser | Paginated list |
| `GET` | `/guests/:id` | `gores.ResponseVM[vms.GuestResponseVM]` | ParamsParser | Find by ID |
| `PUT` | `/guests/:id` | `gores.ResponseVM[vms.GuestResponseVM]` | ParamsParser + BodyParser | Update by ID |
| `DELETE` | `/guests/:id` | `gores.ResponseVM[bool]` | ParamsParser | Delete by ID (no body) |
| `POST` | `/guests/bulk` | `gores.ResponseVM[*[]vms.GuestResponseVM]` | BodyParser | Bulk create |
| `PUT` | `/guests/bulk` | `gores.ResponseVM[*[]vms.GuestResponseVM]` | BodyParser | Bulk update |
| `DELETE` | `/guests/bulk` | `gores.ResponseVM[bool]` | BodyParser | Bulk delete |

**Swagger generation command:** `go generate ./...` (runs swag init automatically).

---

## 13. Transport Layer — gRPC

### 13.1 Protobuf Definition

**Location:** `pkg/protobuf_boilerplate/boilerplate.proto`

```protobuf
syntax = "proto3";
package protobuf_boilerplate;

service Boilerplate {
    rpc CreateGuest(CreateGuestRequestVM) returns (GuestResponseVM);
    rpc FindAllGuest(FindAllGuestRequestVM) returns (FindAllGuestResponseVM);
    rpc FindGuestByID(FindGuestByIDRequestVM) returns (GuestResponseVM);
    rpc UpdateGuestByID(UpdateGuestByIDRequestVM) returns (GuestResponseVM);
    rpc DeleteGuestByID(DeleteGuestByIDRequestVM) returns (google.protobuf.Empty);
    rpc BulkCreateGuests(BulkCreateGuestsRequestVM) returns (BulkCreateGuestsResponseVM);
    rpc BulkUpdateGuests(BulkUpdateGuestsRequestVM) returns (BulkUpdateGuestsResponseVM);
    rpc BulkDeleteGuests(BulkDeleteGuestsRequestVM) returns (google.protobuf.Empty);
}
```

**Protobuf generation:** See [§16.4 Protobuf Generation](#164-protobuf-generation) for details on how to generate `.pb.go` files.

### 13.2 Server Build

`transports/grpc/server.go`:
```go
func BuildGRPCServer(cfg *configs.Config) *Server
```

### 13.3 gRPC Middleware (Interceptors)

**Location:** `transports/grpc/middlewares/`

gRPC interceptors follow the same **struct + Wire DI** pattern as HTTP middleware. Each interceptor implements `grpc.UnaryServerInterceptor`:

```go
type XxxMiddleware struct{}

func NewXxxMiddleware() *XxxMiddleware {
    return &XxxMiddleware{}
}

func (mw *XxxMiddleware) Xxx(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (interface{}, error) {
    // Pre-processing
    resp, err := handler(ctx, req)
    // Post-processing
    return resp, err
}
```

**Middleware struct (`middlewares.go`):**
```go
type Middlewares struct {
    Recover   *RecoverMiddleware
    Tracer    *TracerMiddleware
    RequestID *RequestIDMiddleware
    Log       *LogMiddleware
    Timeout   *TimeoutMiddleware
}
```

**Convenience method** for collecting all interceptors in registration order:
```go
func (mw *Middlewares) GetUnaryServerInterceptors() []grpc.UnaryServerInterceptor {
    return []grpc.UnaryServerInterceptor{
        mw.Recover.Recover,    // 1. Panic recovery
        mw.Tracer.Start,       // 2. Extract trace context from gRPC metadata
        mw.RequestID.Generate, // 3. Generate/propagate request ID
        mw.Log.Log,            // 4. Request/response logging
        mw.Timeout.Timeout,    // 5. Context timeout
    }
}
```

**Registration** in `NewGRPCServer()` (`transports/grpc/grpc_server.go`):
```go
var grpcServer *grpc.Server = grpc.NewServer(
    grpc.ChainUnaryInterceptor(mw.GetUnaryServerInterceptors()...),
)
```

**Wire provider (`provider.go`):**
```go
var Provider wire.ProviderSet = wire.NewSet(
    NewRecoverMiddleware,
    NewTracerMiddleware,
    NewRequestIDMiddleware,
    NewLogMiddleware,
    NewTimeoutMiddleware,
)
```

Injected into `GRPCServer` via `builder.go`:
```go
wire.Struct(new(middlewares.Middlewares), "*"),
```

**Available interceptors details:**

| Interceptor | Method | Config | Description |
|---|---|---|---|
| **Recover** | `Recover(ctx, req, info, handler)` | — | Catches panics via `defer recover()`, returns `gocerr` converted to gRPC status via `grpc_error.FromError()`, logs stack trace |
| **Tracer** | `Start(ctx, req, info, handler)` | — | Extracts OpenTelemetry trace context from gRPC incoming metadata via `metadata.FromIncomingContext` → `otel.GetTextMapPropagator().Extract` |
| **RequestID** | `Generate(ctx, req, info, handler)` | — | Reads `X-Request-ID` from gRPC metadata; generates UUIDv7 if empty. Sets response header via `grpc.SetHeader()` |
| **Log** | `Log(ctx, req, info, handler)` | — | Measures latency, logs req/info/res/err. Uses `zerolog.InfoLevel` (success), `ErrorLevel` (5xx-equivalent gRPC status via `gostacode.HTTPStatusCodeFromGRPCCode`). Includes `fmt.Sprintf("%.3f ms", ...)` latency |
| **Timeout** | `Timeout(ctx, req, info, handler)` | `Server.GRPC.RequestTimeout` | Runs handler in a goroutine with `context.WithTimeout`. Returns `codes.DeadlineExceeded` on timeout via `select { case <-ctxTimeout.Done(): }` |

> All custom interceptor methods that accept `ctx` must have `tracer.Start(ctx, "[MiddlewareName][Method]")` + `defer span.End()` for distributed tracing. The **Log** interceptor also uses the **response** from the handler for logging — not just the error.

### 13.4 gRPC Handler Pattern

```go
func (h *ImplementedBoilerplateServer) Xxx(ctx context.Context, requestVM *protobuf_boilerplate.XxxRequestVM) (*protobuf_boilerplate.XxxResponseVM, error) {
    var (
        span        trace.Span
        logFields   map[string]interface{}
        requestDTO  *dtos.XxxRequestDTO
        responseDTO *dtos.XxxResponseDTO
        err         error
    )

    ctx, span = tracer.Start(ctx, "[ImplementedBoilerplateServer][Xxx]")
    defer span.End()

    logFields = map[string]interface{}{}

    if requestVM == nil {
        return nil, gocerr.New(http.StatusBadRequest, "requestVM is nil").ToGRPCError()
    }

    requestDTO = vms.XxxRequestVMToDTO(requestVM)
    logFields["requestDTO"] = requestDTO

    responseDTO, err = h.guestService.Xxx(ctx, requestDTO)
    if err != nil {
        logLevel = zerolog.WarnLevel
        if gocerr.GetErrorCode(err) >= http.StatusInternalServerError {
            logLevel = zerolog.ErrorLevel
        }
        log.WithLevel(logLevel).Ctx(ctx).Err(err).Fields(logFields).Msg("[ImplementedBoilerplateServer][Xxx][Xxx] ...")
        return nil, err
    }

    responseVM = protobuf_boilerplate.NewXxxResponseVM(responseDTO)
    return responseVM, nil
}
```

### 13.3 gRPC VM Converters

**Location:** `transports/grpc/models/vms/guest_vm.go`

```go
func XxxRequestVMToDTO(vm *protobuf_boilerplate.XxxRequestVM, ...) *dtos.XxxRequestDTO
func NewXxxResponseVM(responseDTO *dtos.XxxResponseDTO) *protobuf_boilerplate.XxxResponseVM
```

---

## 14. Transport Layer — Event Consumer (NSQ)

### 14.1 Producer-Consumer Flow

```
Service (publish event)
    → EventProducerRepository.Publish (single entity) / PublishBulk (multiple entities)
        → Marshal to JSON (EventEntity[T])
        → InjectTracerPropagator (carry trace context)
        → NSQProducer.Publish(topic, body)
            → NSQ Queue
                → NSQConsumer on same topic
                    → json.Unmarshal into EventRequestVM[T]
                    → guestService.ProcessEvent(ctx, dto)
                        → WebhookSiteRepository.SendWebhook
```

**Event topics** are configured per event type in `Guest.Event.Xxx.Topic` config. Each topic maps to an NSQ channel equal to the consumer name.

**6 event types:** Created, Deleted, Updated, BulkCreated, BulkUpdated, BulkDeleted. Each has its own NSQ topic and consumer.

### 14.2 Consumer Struct

`transports/event_consumer/consumers/guest_consumer.go`:
```go
type GuestConsumer struct {
    bulkCreatedGuestConsumer *nsq.Consumer
    bulkUpdatedGuestConsumer *nsq.Consumer
    bulkDeletedGuestConsumer *nsq.Consumer
    createdGuestConsumer     *nsq.Consumer
    deletedGuestConsumer     *nsq.Consumer
    updatedGuestConsumer     *nsq.Consumer
}
```

### 14.2 Consumer Registration Pattern

```go
if cfg.Guest.Event.Created.Enable {
    consumer.createdGuestConsumer, _ = nsq.NewConsumer(
        cfg.Guest.Event.Created.Topic,
        cfg.Server.EventConsumer.DataSourceName, // channel
        nsq.NewConfig(),
    )
    consumer.createdGuestConsumer.AddHandler(
        handlers.NewMessageHandler(handler.HandleCreated),
    )
}
```

### 14.3 Handler Method Pattern

```go
func (h *GuestHandler) HandleXxx(ctx context.Context, m *nsq.Message) error {
    var (
        span       trace.Span
        logFields  map[string]interface{}
        requestVM  *vms.EventRequestVM[vms.GuestEventRequestVM]
        requestDTO *dtos.GuestEventRequestDTO
        err        error
    )

    ctx, span = tracer.Start(ctx, "[GuestHandler][HandleXxx]")
    defer span.End()

    logFields = map[string]interface{}{
        "messageBody": string(m.Body),
    }

    log.Info().Ctx(ctx).Fields(logFields).Msg("[GuestHandler][HandleXxx] message received")

    requestVM = &vms.EventRequestVM[vms.GuestEventRequestVM]{}
    err = json.Unmarshal(m.Body, requestVM)
    if err != nil {
        log.Err(err).Ctx(ctx).Fields(logFields).Msg("[GuestHandler][HandleXxx][Unmarshal] failed to parse message body")
        return err
    }
    logFields["requestVM"] = requestVM

    if requestVM.Message == nil {
        err = gocerr.New(http.StatusInternalServerError, "message is nil")
        log.Err(err).Ctx(ctx).Fields(logFields).Msg("[GuestHandler][HandleXxx] message is nil")
        return err
    }

    requestDTO = requestVM.Message.ToDTO()
    logFields["requestDTO"] = requestDTO

    _, err = h.guestService.ProcessEvent(ctx, requestDTO)
    if err != nil {
        log.Err(err).Ctx(ctx).Fields(logFields).Msg("[GuestHandler][HandleXxx][ProcessEvent] failed to process event")
        return err
    }

    return nil
}
```

### 14.4 Handler Methods

There are 6 handler methods:
- `HandleCreated` — processes created events
- `HandleDeleted` — processes deleted events
- `HandleUpdated` — processes updated events
- `HandleBulkCreated` — processes bulk created events
- `HandleBulkUpdated` — processes bulk updated events
- `HandleBulkDeleted` — processes bulk deleted events

All follow the identical pattern above. All MUST have a tracer span.

### 14.5 Event Consumer Builder

`transports/event_consumer/builder.go` — `BuildEventConsumer(cfg)` creates all consumers and wires them together.

### 14.6 Event Request VM

```go
type EventRequestVM[T any] struct {
    Message     *T              `json:"message"`
    ID          string          `json:"id"`
    CreatedAt   int64           `json:"created_at"`
    Publisher   string          `json:"publisher"`
    Propagator  map[string]string `json:"propagator"`
}

type GuestEventRequestVM struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Address   string `json:"address"`
    CreatedBy string `json:"created_by"`
}
```

---

## 15. CMD Layer (Application Entry Points)

### 15.1 Command Structure

Uses `github.com/spf13/cobra`. All commands defined in `cmd/` package.

### 15.2 Available Commands

| Command | File | Purpose |
|---|---|---|
| `app` | `cmd/app.go` | HTTP + gRPC + Event Consumer (all 3 in parallel) |
| `http` | `cmd/http.go` | HTTP server only |
| `grpc` | `cmd/grpc.go` | gRPC server only |
| `event-consumer` | `cmd/event_consumer.go` | Event consumer only |
| `database-migration` | `cmd/database_migration.go` | Run SQL schema migrations on Master DB (no servers started) |

### 15.3 App Command (Multi-Server)

```go
func initApp() {
    appCmd = &cobra.Command{
        Use:   "app",
        PreRun: func(cmd *cobra.Command, args []string) {
            // 1. Panic recovery wrapper
            defer func() { if r := recover(); r != nil { log.Printf("[ERROR] Panic recovered: %v", r) } }()

            // 2. Read config + init tracer
            cfg = configs.Read(cfgPath)
            tracer.NewTracer(...)

            // 3. Build all servers in parallel (3 goroutines)
            var task gotask.Task = gotask.NewTask(3)
            task.Go(func() { httpServer = http.BuildHTTPServer(cfg) })
            task.Go(func() { grpcServer = grpc.BuildGRPCServer(cfg) })
            task.Go(func() { eventConsumer = event_consumer.BuildEventConsumer(cfg) })
            task.Wait()
        },
        RunE: func(cmd *cobra.Command, args []string) error {
            // 4. Serve all servers in parallel (3 goroutines)
            errTask, _ = gotask.NewErrorTask(context.Background(), 3)
            errTask.Go(func() error { return httpServer.ServeHTTP() })
            errTask.Go(func() error { return grpcServer.ServeGRPC() })
            errTask.Go(func() error { return eventConsumer.Consume() })
            return errTask.Wait()
        },
    }
}
```

### 15.4 Shared Global Variables

```go
var (
    cfgPath       string                    // --config flag
    cfg           *configs.Config
    httpServer    *http.Server
    grpcServer    *grpc.Server
    eventConsumer *event_consumer.Server
)
```

### 15.5 Database Migration

**File:** `cmd/database_migration.go`

The `database-migration` command runs SQL schema migrations against the Master database.

**Flow:**
```
configs.Read(cfgPath)
    → datasources/boilerplate_database/datasource.go
        → sqlx.Connect (Master DB only)
            → Execute migration SQL files / statements
```

**Implementation pattern:**

```go
func initDatabaseMigration() {
    databaseMigrationCmd = &cobra.Command{
        Use: "database-migration",
        RunE: func(cmd *cobra.Command, args []string) error {
            var (
                cfg *configs.Config
                db  *sqlx.DB
                err error
            )

            cfg = configs.Read(cfgPath)

            // Connect to Master database (read/write)
            db, err = sqlx.Connect(
                cfg.Datasource.BoilerplateDatabase.Master.DriverName,
                cfg.Datasource.BoilerplateDatabase.Master.DataSourceName,
            )
            if err != nil {
                return err
            }
            defer db.Close()

            // Run migrations (embedded SQL or migration files)
            err = runMigrations(db)
            if err != nil {
                return err
            }

            return nil
        },
    }
}
```

**Migration files location:** Embedded in `datasources/boilerplate_database/migrations/` or inline SQL strings.

**Key points:**
- Only connects to **Master** database (not Slave)
- Uses `sqlx.Connect` directly (not via `BoilerplateDatabase` wrapper)
- Does NOT start any server (no HTTP, no gRPC, no consumer)
- Migration files are standard SQL statements
- Each migration should be idempotent (use `IF NOT EXISTS`, `IF EXISTS`)

### 15.6 Root Command

```go
func init() {
    rootCmd.AddCommand(httpCmd, grpcCmd, appCmd, eventConsumerCmd, databaseMigrationCmd)
}
```

---

## 16. Tooling & Code Generation

### 16.1 Makefile

| Command | Action |
|---|---|
| `make clean` | Delete all generated files (mocks, wire_gen, swagger) |
| `make generate` | `go generate ./...` (mockery + swagger + wire) + protobuf generation via Docker (`protobuf-generator`) |
| `make test` | `clean` + `generate` + `go test -v -cover -covermode=atomic ./...` |
| `make app` | `clean` + `generate` + `go run main.go app` |
| `make http` | `clean` + `generate` + `go run main.go http` |
| `make grpc` | `clean` + `generate` + `go run main.go grpc` |
| `make build` | `clean` + `generate` + `go build` |

### 16.2 Mockery v3

Mock generation via `//mockery:generate` annotations above interfaces:

```go
//mockery:generate: true
//mockery:structname: BoilerplateDatabaseRepositoryMock
//mockery:filename: boilerplate_database_repository_mock.go
//mockery:output: internal/repositories/mocks/
type IBoilerplateDatabaseRepository[TEntity interface{}] interface { ... }
```

**Standard annotation parameters:**
| Parameter | Description | Example |
|---|---|---|
| `generate` | Always true | `true` |
| `structname` | Mock struct name | `GuestServiceMock` |
| `filename` | Output file name | `guest_service_mock.go` |
| `output` | Output directory | `internal/services/mocks/` |

**Mock output directories:**
- `internal/repositories/mocks/`
- `internal/services/mocks/`

### 16.3 Swagger Generation

Uses `github.com/swaggo/swag`. Annotations in HTTP handler files. Generated to:
`transports/http/docs/swagger/` (swagger.json, swagger.yaml, docs.go)

### 16.4 Protobuf Generation

**Tool:** `github.com/fikri240794/protobuf-generator` (Docker-based, no host installation needed)

This tool uses Docker to compile `.proto` files into `.pb.go` and `_grpc.pb.go`. It is **not mandatory but highly recommended** — it ensures consistency without installing `protoc`, Go plugins, or any host dependencies.

#### Build the Docker Image

```bash
git clone https://github.com/fikri240794/protobuf-generator.git
cd protobuf-generator
docker build -t protobuf-generator .
```

> The Docker image only needs to be built once. It uses a persistent named volume (`protobuf-generator`) to cache `protoc` and language binaries for subsequent fast runs.

#### Generate Protobuf Files

Run from the root of this project:

```bash
MSYS_NO_PATHCONV=1 docker run --rm \
    -v protobuf-generator:/root \
    -v .:/home/src \
    -e TARGET_LANG=go \
    -e PROTO_FILE_PATH=pkg/protobuf_boilerplate \
    -e PROTO_OUT_PATH=pkg/protobuf_boilerplate \
    protobuf-generator
```

**Configuration variables:**

| Variable | Description | Example |
|---|---|---|
| `TARGET_LANG` | Target language (required) | `go` |
| `PROTO_FILE_PATH` | Relative path to directory with `.proto` files | `pkg/protobuf_boilerplate` |
| `PROTO_OUT_PATH` | Relative path for generated output | `pkg/protobuf_boilerplate` |
| `PROTO_FILE_NAME` | Pattern to match proto files (default: `*.proto`) | `*.proto` |
| `GO_VERSION` | Go version (default: latest) | `1.25.0` |
| `PROTOC_GEN_GO_VERSION` | protoc-gen-go version (default: latest) | `latest` |
| `PROTOC_GEN_GO_GRPC_VERSION` | protoc-gen-go-grpc version (default: latest) | `latest` |

> **Windows (Git Bash) users:** Prefix the command with `MSYS_NO_PATHCONV=1` to prevent path conversion issues. CMD/PowerShell users can omit the prefix.
> **macOS / Linux users:** Omit the `MSYS_NO_PATHCONV=1` prefix.

**Generated files** (checked into version control):
- `pkg/protobuf_boilerplate/boilerplate.pb.go`
- `pkg/protobuf_boilerplate/boilerplate_grpc.pb.go`

### 16.5 Wire DI Generation

Google Wire generates `wire_gen.go` files. These are checked in. Never edit manually.

---

## 17. Dependency Injection (Wire)

### 17.1 Wire Provider Files

```
internal/repositories/provider.go
internal/services/provider.go
transports/http/handlers/provider.go
transports/grpc/handlers/provider.go
transports/event_consumer/handlers/provider.go
```

### 17.2 Provider Pattern

```go
// internal/repositories/provider.go
var GuestRepositorySet = wire.NewSet(
    wire.Struct(new(GuestRepository), "*"),
    wire.Bind(new(IGuestRepository), new(*GuestRepository)),
)
```

Each `wire.NewSet` provides:
1. The concrete struct constructor (`wire.Struct(new(Xxx), "*")`)
2. The interface binding (`wire.Bind(new(IXxx), new(*Xxx))`)

### 17.3 Wire Sets Composition

```go
// internal/services/provider.go
var GuestServiceSet = wire.NewSet(
    wire.Struct(new(GuestService), "*"),
    wire.Bind(new(IGuestService), new(*GuestService)),
)
```

Dependencies are resolved automatically by Wire based on constructor parameter types.

---

## 18. Testing Conventions

### 18.1 General

- **Framework:** `testing` standard library
- **Assertions:** `github.com/stretchr/testify/assert`
- **DB mocking:** `github.com/DATA-DOG/go-sqlmock`
- **Mocking:** `mockery` v3 generated mocks
- **Coverage target:** 100% with `-covermode=atomic`
- **Pattern:** Table-driven tests

### 18.2 Table-Driven Test Template

```go
func Test_FunctionName(t *testing.T) {
    tests := []struct {
        name        string
        setupRepo   func() (*RepoType, sqlmock.Sqlmock)  // or mocks
        ctx         context.Context
        args        ...
        expectError bool
        validate    func(t *testing.T, ...)
    }{
        {
            name: "success case",
            setupRepo: func() (...) {
                mockDB, mock, _ := sqlmock.New()
                // ... setup
                return repo, mock
            },
            ...
            expectError: false,
            validate:    func(t *testing.T, ...) { assert.NoError(t, err) },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            repo, mock := tt.setupRepo()
            // Setup mock expectations based on test name
            switch tt.name {
            case "success case":
                mock.ExpectPrepare("...").WillBeClosed()
                mock.ExpectExec("...").WillReturnResult(sqlmock.NewResult(1, 1))
            case "error case":
                mock.ExpectPrepare("...").WillReturnError(errors.New("error"))
            }
            // Call function
            err := repo.Function(tt.ctx, tt.args...)
            // Assert
            if tt.expectError { assert.NotNil(t, err) }
            if !tt.expectError { assert.NoError(t, err) }
            if tt.validate != nil { tt.validate(t, err) }
            // Verify mock expectations
            assert.NoError(t, mock.ExpectationsWereMet())
        })
    }
}
```

### 18.3 Mock Expectation Patterns

```go
// Query expectations:
mock.ExpectPrepare("SELECT (.+)").WillBeClosed()
rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "John")
mock.ExpectQuery("SELECT (.+)").WillReturnRows(rows)

// Exec expectations:
mock.ExpectPrepare("INSERT INTO (.+)").WillBeClosed()
mock.ExpectExec("INSERT INTO (.+)").WillReturnResult(sqlmock.NewResult(0, 1))

// Error expectations:
mock.ExpectPrepare("UPDATE (.+)").WillReturnError(errors.New("prepare error"))
mock.ExpectExec("DELETE FROM (.+)").WillReturnError(errors.New("exec error"))
```

---

## 19. Key Do's and Don'ts

| Do | Don't |
|---|---|
| Declare all vars in `var (...)` at function start | Use `:=` except in `for i := range` |
| Use `s.withTransaction`, `s.tryDeleteEntityCaches`, `s.publishEvent` | Duplicate transaction/cache/event boilerplate |
| Use `r.prepareQueryStatement`, `r.logSlowQuery` | Duplicate prepare/timing code in Count/FindAll/FindOne |
| Use `s.buildActiveEntityFilterByIDs` | Manually build `{id=? AND deleted_at IS NULL}` filter |
| Start tracer span in every public method | Forget `defer span.End()` |
| Call `message.InjectTracerPropagator(ctx)` in ALL publish methods | Skip tracer propagation in bulk methods |
| Swallow cache/event errors (`err = nil` after log) | Return cache/event errors to caller |
| Use `log.Warn()` for validation errors | Use `log.Err()` for expected validation failures |
| Use `log.WithLevel(logLevel)` for dynamic error levels | Always use `log.Err()` for service errors |
| Use `log.Info()` for event consumer "received" messages | Log everything as errors |
| Add `mapstructure` tags to all config fields | Use field names that don't match `.env` keys |
| Add `db`, `table`, `primary_key`, `db_type` tags to entity fields | Use reflection-incompatible tag formats |
| Regenerate mocks after interface changes | Manually edit generated mock files |
| Run `make generate` before building | Edit `wire_gen.go` manually |
| Check `gocerr.GetErrorCode(err)` before setting log level | Assume all errors are server errors (500) |
| Use `goqube` dialect-aware builders | Write raw SQL strings for complex queries |
| Use `null.String` / `null.Int64` for nullable fields | Use pointers or zero values for nullable columns |
| Use `*[]T` (pointer to slice) with `gores.ResponseVM[T]` | Use `[]T` directly (doesn't satisfy `comparable`) |
| Call `tx.Rollback()` in both operation-failure and commit-failure paths | Forget rollback on commit failure |
| Use `continue` after processing `table` tag in `getEntityMeta`/`getTableNameAndFields` | Process `db` tag on a field that already has `table` tag |
