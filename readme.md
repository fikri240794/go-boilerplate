# Go Base Project Template or Boilerplate

This is a Golang project starter (also called boilerplate) that uses **Hexagonal Architecture**. It clearly separates concerns into layers and modules, making your code clean, easy to scale, test, and maintain.

---

## 🧱 Design & Structure

### Hexagonal Architecture
<p align="center">
  <img src="./documentations/architecture.jpg" style="width:75%;" />
</p>

This project is built using the Hexagonal Architecture pattern. It organizes the code into clear layers:

* **Datasources** – Responsible for connecting to external systems like databases, APIs (HTTP or gRPC), or message brokers. It acts as a gateway between your app and the outside world.

* **Usecases**

  * **Repositories** – Defines interfaces and how data is accessed using datasources. It separates business logic from technical details.
  * **Services** – Contains the main business logic and rules. This is where decisions and actions based on your application's purpose are made.

* **Transports (Handlers & Middlewares)** – Handles external requests (like HTTP or gRPC), then routes them to the correct service.

### Folder Structure

The structure helps keep things clean and modular:

```
.
├───cmd/                                  # Application entry points (for build & run the migration, http, grpc, etc)
├───configs/                              # Configuration files (.env)
├───datasources/                          # Gateway to external systems
│   ├───database/                         # Database connection, queries, and migrations
│   │   └───migrations/                   # Database schema migration files
│   ├───event_driven_publisher/           # Code to publish messages to event/message brokers
│   ├───http_client/                      # HTTP client integrations with external APIs
│   └───grpc_client/                      # gRPC client implementations
├───internal/                             # Core business logic and domain models
│   ├───models/                           # Data models used internally
│   │   ├───dtos/                         # Data Transfer Objects (for input/output transformations)
│   │   └───entities/                     # Domain entities representing core business concepts
│   ├───repositories/                     # Interfaces and logic to access data via datasources
│   └───services/                         # Core business logic (use cases)
├───pkg/                                  # Reusable helper packages and utilities
│   ├───some_protobuf_models/             # Protocol Buffer (*.proto) & generated protobuf model
│   ├───some_utility/                     # Custom utility functions
│   └───some_helper/                      # General-purpose helpers
└───transports/                           # Entry points for different communication protocols
    ├───event_driven_subscriber/          # Event-Driven subscriber layer
    │   ├───handlers/                     # Handlers for incoming event messages
    │   └───models/
    │       └───vms/                      # ViewModels for message processing
    ├───grpc/                             # gRPC transport layer
    │   ├───handlers/                     # gRPC handlers
    │   ├───middlewares/                  # Middleware (auth, logging, etc.)
    │   └───models/
    │       └───vms/                      # Logic to map from generated protobuf models to dtos
    └───http/                             # HTTP transport layer
        ├───handlers/                     # HTTP route handlers
        ├───middlewares/                  # Middleware (auth, logging, etc.)
        └───models/
            └───vms/                      # ViewModels used in HTTP layer
```

---

## 🧪 Sample Feature: Guest Management

This boilerplate includes a full sample to manage `Guest` data, showing how everything works together:

* Create, update, delete, bulk create, bulk update, bulk delete, and get Guest data via HTTP and gRPC
* Read data from PostgreSQL or Redis cache
* Publish changes (create/update/delete/bulk create/bulk update/bulk delete) to a message broker (NSQ)
* Receive events from message broker and forward them to [Webhook.site](https://webhook.site) using HTTP client
* Open Telemetry Tracer Data

---

### Infrastructure Diagram
<p align="center">
  <img src="./documentations/infrastructure.jpg" style="width:75%;" />
</p>

#### Components:

* **HTTP Server** – Built with [Fiber](https://docs.gofiber.io/), a simple and fast Go web framework
* **gRPC Server** – Built with native Go gRPC libraries
* **NSQ** – Lightweight and fast message broker, ideal for development and testing
* **OpenTelemetry + Jaeger** – Used for tracing and debugging
* **PostgreSQL** – Main relational database
* **Redis** – Used as a cache for faster reads
* **Webhook.site** – External service to see forwarded requests in real-time

---

## 🗄️ Database: `guests` Table

| Column      | Type   | Required | Description                              |
| ----------- | ------ | -------- | ---------------------------------------- |
| id          | UUID   | Yes      | Unique identifier for the guest          |
| name        | text   | Yes      | Guest name                               |
| address     | text   | No       | Address (optional)                       |
| created\_at | bigint | Yes      | When the record was created (epoch time) |
| created\_by | text   | Yes      | Who created the record                   |
| updated\_at | bigint | Yes      | When the record was last updated         |
| updated\_by | text   | Yes      | Who updated the record                   |
| deleted\_at | bigint | Yes      | When the record was soft-deleted         |
| deleted\_by | text   | Yes      | Who soft-deleted the record              |

---

## 🔁 Sequence Diagrams
* **Create Guest**
  <p align="center">
    <img src="./documentations/create-guest.png" style="width:50%;" />
  </p>
- **Update Guest by ID**  
  <p align="center">
    <img src="./documentations/update-guest.png" style="width:50%;" />
  </p>
- **Delete Guest by ID**  
  <p align="center">
    <img src="./documentations/delete-guest.png" style="width:50%;" />
  </p>
- **Find Guests by Filter**  
  <p align="center">
    <img src="./documentations/find-all-guest-by-filter.png" style="width:50%;" />
  </p>
- **Find Guest by ID**  
  <p align="center">
    <img src="./documentations/find-guest-by-id.png" style="width:50%;" />
  </p>
- **Bulk Create Guest**  
  <p align="center">
    <img src="./documentations/bulk-create-guest.png" style="width:50%;" />
  </p>
- **Bulk Update Guest**  
  <p align="center">
    <img src="./documentations/bulk-update-guest.png" style="width:50%;" />
  </p>

---

## 🌐 HTTP API with Swagger

API documentation is available at:

```
{{HTTP_SERVER_URL}}/docs/swagger
```

**Create Guest**
```
Method: POST
URL: {{HTTP_SERVER_URL}}/guests
Request:
  Headers:
    Content-Type: application/json
  Body:
    {
      "address": "123 Main Street, Apt. 4B, New York, NY 10001, USA",
      "name": "John Snow"
    }
Response:
  Headers:
    Content-Type: application/json
  Code: 201
    Body:
      {
        "code": 201,
        "data": {
          "id": "019681d0-c726-72c2-8c41-110cbca4e680",
          "name": "John Snow",
          "address": "123 Main Street, Apt. 4B, New York, NY 10001, USA",
          "created_at": 1745934665510,
          "created_by": "00000000-0000-0000-0000-000000000000"
        }
      }
  Code: >=400
    Body:
      {
        "code": 400,
        "error": {
          "message": "Bad Request",
          "error_fields": [
            {
              "field": "field_name",
              "message": "some message of error validation"
            }
          ]
        }
      }
```
Example cURL:
```bash
curl -X POST '{{HTTP_SERVER_URL}}/guests' \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "John Snow",
    "address": "123 Main Street, New York"
  }'
```

**Update Guest**
```
Method: PUT
URL: {{HTTP_SERVER_URL}}/guests/019681d0-c726-72c2-8c41-110cbca4e680
Request:
  Headers:
    Content-Type: application/json
  Body:
    {
      "address": "123 Main Street, Apt. 4B, New York, NY 10001, USA",
      "name": "John Snow"
    }
Response:
  Headers:
    Content-Type: application/json
  Code: 200
    Body:
      {
        "code": 200,
        "data": {
          "id": "019681d0-c726-72c2-8c41-110cbca4e680",
          "name": "John Snow",
          "address": "123 Main Street, Apt. 4B, New York, NY 10001, USA",
          "created_at": 1745934665510,
          "created_by": "00000000-0000-0000-0000-000000000000",
          "updated_at": 1745935815436,
          "updated_by": "00000000-0000-0000-0000-000000000000"
        }
      }
  Code: >=400
    Body:
      {
        "code": 400,
        "error": {
          "message": "Bad Request",
          "error_fields": [
            {
              "field": "field_name",
              "message": "some message of error validation"
            }
          ]
        }
      }
```
Example cURL:
```bash
curl -X 'PUT' \
  '{{HTTP_SERVER_URL}}/guests/019681d0-c726-72c2-8c41-110cbca4e680' \
  -H 'Content-Type: application/json' \
  -d '{
  "address": "123 Main Street, Apt. 4B, New York, NY 10001, USA",
  "name": "John Snow"
}'
```

**Delete Guest by ID**
```
Method: DELETE
URL: {{HTTP_SERVER_URL}}/guests/019681d0-c726-72c2-8c41-110cbca4e680
Request:
  Headers:
  Body:
Response:
  Headers:
    Content-Type: application/json
  Code: 200
    Body:
      {
        "code": 200,
        "data": true
      }
  Code: >=400
    Body:
      {
        "code": 400,
        "error": {
          "message": "Bad Request",
          "error_fields": [
            {
              "field": "field_name",
              "message": "some message of error validation"
            }
          ]
        }
      }
```
Example cURL:
```bash
curl -X 'DELETE' \
  '{{HTTP_SERVER_URL}}/guests/019681d0-c726-72c2-8c41-110cbca4e680'
```

**Find All Guest by Filter**
```
Method: GET
URL: {{HTTP_SERVER_URL}}/guests?keyword=John&sorts=name&take=1&skip=0
Request:
  Headers:
  Body:
Response:
  Headers:
    Content-Type: application/json
  Code: 200
    Body:
      {
        "code": 200,
        "data": {
          "list": [
            {
              "id": "01960b83-d408-73c3-a200-29d50902604b",
              "name": "John Snow",
              "address": "123 Main Street, Apt. 4B, New York, NY 10001, USA",
              "created_at": 1743949911048,
              "created_by": "00000000-0000-0000-0000-000000000000"
            }
          ],
          "count": 2
        }
      }
  Code: >=400
    Body:
      {
        "code": 400,
        "error": {
          "message": "Bad Request",
          "error_fields": [
            {
              "field": "field_name",
              "message": "some message of error validation"
            }
          ]
        }
      }
```
Example cURL:
```
curl -X 'GET' \
  '{{HTTP_SERVER_URL}}/guests?keyword=John&sorts=name&take=1&skip=0'
```

**Find Guest by ID**
```
Method: GET
URL: {{HTTP_SERVER_URL}}/guests/019681d0-c726-72c2-8c41-110cbca4e680
Request:
  Headers:
  Body:
Response:
  Headers:
    Content-Type: application/json
  Code: 200
    Body:
      {
        "code": 200,
        "data": {
          "id": "019681d0-c726-72c2-8c41-110cbca4e680",
          "name": "John Snow",
          "address": "123 Main Street, Apt. 4B, New York, NY 10001, USA",
          "created_at": 1745934665510,
          "created_by": "00000000-0000-0000-0000-000000000000",
          "updated_at": 1745935815436,
          "updated_by": "00000000-0000-0000-0000-000000000000"
        }
      }
  Code: >=400
    Body:
      {
        "code": 400,
        "error": {
          "message": "Bad Request",
          "error_fields": [
            {
              "field": "field_name",
              "message": "some message of error validation"
            }
          ]
        }
      }
```
Example cURL:
```bash
curl -X 'GET' \
  '{{HTTP_SERVER_URL}}/guests/019681d0-c726-72c2-8c41-110cbca4e680'
```

**Bulk Create Guest**
```
Method: POST
URL: {{HTTP_SERVER_URL}}/guests/bulk
Request:
  Headers:
    Content-Type: application/json
  Body:
    [
      {
        "address": "123 Main Street, Apt. 4B, New York, NY 10001, USA",
        "name": "John Snow"
      },
      {
        "address": "456 Oak Avenue, Suite 12, Los Angeles, CA 90001, USA",
        "name": "Jane Doe"
      }
    ]
Response:
  Headers:
    Content-Type: application/json
  Code: 201
    Body:
      {
        "code": 201,
        "data": [
          {
            "id": "019681d0-c726-72c2-8c41-110cbca4e680",
            "name": "John Snow",
            "address": "123 Main Street, Apt. 4B, New York, NY 10001, USA",
            "created_at": 1745934665510,
            "created_by": "00000000-0000-0000-0000-000000000000"
          },
          {
            "id": "019681d0-c726-72c2-8c41-110cbca4e681",
            "name": "Jane Doe",
            "address": "456 Oak Avenue, Suite 12, Los Angeles, CA 90001, USA",
            "created_at": 1745934665510,
            "created_by": "00000000-0000-0000-0000-000000000000"
          }
        ]
      }
  Code: >=400
    Body:
      {
        "code": 400,
        "error": {
          "message": "Bad Request",
          "error_fields": [
            {
              "field": "field_name",
              "message": "some message of error validation"
            }
          ]
        }
      }
```
Example cURL:
```bash
curl -X POST '{{HTTP_SERVER_URL}}/guests/bulk' \
  -H 'Content-Type: application/json' \
  -d '[
    {
      "name": "John Snow",
      "address": "123 Main Street, New York"
    },
    {
      "name": "Jane Doe",
      "address": "456 Oak Avenue, Los Angeles"
    }
  ]'
```

**Bulk Update Guest**
```
Method: PUT
URL: {{HTTP_SERVER_URL}}/guests/bulk
Request:
  Headers:
    Content-Type: application/json
  Body:
    [
      {
        "id": "019681d0-c726-72c2-8c41-110cbca4e680",
        "address": "789 Pine Road, Floor 5, Chicago, IL 60601, USA",
        "name": "John Snow Updated"
      },
      {
        "id": "019681d0-c726-72c2-8c41-110cbca4e681",
        "address": "321 Maple Drive, Suite 7, Houston, TX 77001, USA",
        "name": "Jane Doe Updated"
      }
    ]
Response:
  Headers:
    Content-Type: application/json
  Code: 200
    Body:
      {
        "code": 200,
        "data": [
          {
            "id": "019681d0-c726-72c2-8c41-110cbca4e680",
            "name": "John Snow Updated",
            "address": "789 Pine Road, Floor 5, Chicago, IL 60601, USA",
            "created_at": 1745934665510,
            "created_by": "00000000-0000-0000-0000-000000000000",
            "updated_at": 1745935815436,
            "updated_by": "00000000-0000-0000-0000-000000000000"
          },
          {
            "id": "019681d0-c726-72c2-8c41-110cbca4e681",
            "name": "Jane Doe Updated",
            "address": "321 Maple Drive, Suite 7, Houston, TX 77001, USA",
            "created_at": 1745934665510,
            "created_by": "00000000-0000-0000-0000-000000000000",
            "updated_at": 1745935815436,
            "updated_by": "00000000-0000-0000-0000-000000000000"
          }
        ]
      }
  Code: >=400
    Body:
      {
        "code": 400,
        "error": {
          "message": "Bad Request",
          "error_fields": [
            {
              "field": "field_name",
              "message": "some message of error validation"
            }
          ]
        }
      }
```
Example cURL:
```bash
curl -X 'PUT' \
  '{{HTTP_SERVER_URL}}/guests/bulk' \
  -H 'Content-Type: application/json' \
  -d '[
    {
      "id": "019681d0-c726-72c2-8c41-110cbca4e680",
      "name": "John Snow Updated",
      "address": "789 Pine Road, Chicago"
    },
    {
      "id": "019681d0-c726-72c2-8c41-110cbca4e681",
      "name": "Jane Doe Updated",
      "address": "321 Maple Drive, Houston"
    }
  ]'
```

**Bulk Delete Guest**
```
Method: DELETE
URL: {{HTTP_SERVER_URL}}/guests/bulk
Request:
  Headers:
    Content-Type: application/json
  Body:
    {
      "ids": [
        "019681d0-c726-72c2-8c41-110cbca4e680",
        "019681d0-c726-72c2-8c41-110cbca4e681"
      ]
    }
Response:
  Headers:
    Content-Type: application/json
  Code: 200
    Body:
      {
        "code": 200,
        "data": true
      }
  Code: >=400
    Body:
      {
        "code": 400,
        "error": {
          "message": "Bad Request",
          "error_fields": [
            {
              "field": "field_name",
              "message": "some message of error validation"
            }
          ]
        }
      }
```
Example cURL:
```bash
curl -X 'DELETE' \
  '{{HTTP_SERVER_URL}}/guests/bulk' \
  -H 'Content-Type: application/json' \
  -d '{
    "ids": [
      "019681d0-c726-72c2-8c41-110cbca4e680",
      "019681d0-c726-72c2-8c41-110cbca4e681"
    ]
  }'
```

---

## 📡 gRPC API

Use the `.proto` files located in:

```
./pkg/protobuf_boilerplate
```

You can open these in Postman or other gRPC clients and call the endpoints following the proto message structure.

---

## ⚙️ Setting Up

### Install Go

Install Go using [GoBrew](https://github.com/kevincobain2000/gobrew) for flexible version control. Check the `go.mod` file for the required version.

### Setup Databases and Dependencies

#### PostgreSQL (via Docker):

```bash
docker run --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=12345678 -d postgres
```

#### Redis (via Docker):

```bash
docker run --name redis -p 6379:6379 -d redis
```

#### NSQ:

Follow [this guide](https://github.com/fikri240794/go-nsq-pubsub) to run NSQ locally with Docker.

#### OpenTelemetry:

Follow [this guide](https://github.com/fikri240794/go-otel-tracer-example) to set up tracing locally.

### Configuration File

This project uses a `.env` file for all configurations. Copy `.env.example` and modify it according to your setup:

```
cp .env.example ./configs/.env
```

OR

you can copy this configuration and modify it according to your setup:

```
ENVIRONMENT=development
SERVER.NAME=boilerplate
SERVER.LOG_LEVEL=1
SERVER.HTTP.PORT=3000
SERVER.HTTP.PREFORK=false
SERVER.HTTP.PRINT_ROUTES=false
SERVER.HTTP.REQUEST_TIMEOUT=1s
SERVER.HTTP.GRACEFULLY_SHUTDOWN_DURATION=3s
SERVER.HTTP.CORS.ALLOW_ORIGINS=*
SERVER.HTTP.CORS.ALLOW_METHODS=GET,POST,HEAD,PUT,DELETE,PATCH
SERVER.HTTP.DOCS.SWAGGER.ENABLE=true
SERVER.HTTP.DOCS.SWAGGER.FILE_PATH=./transports/http/docs/swagger/swagger.json
SERVER.HTTP.DOCS.SWAGGER.PATH=/docs/swagger
SERVER.HTTP.DOCS.SWAGGER.TITLE=Boilerplate API Docs
SERVER.GRPC.PORT=3001
SERVER.GRPC.REQUEST_TIMEOUT=1s
SERVER.EVENT_CONSUMER.DATA_SOURCE_NAME=localhost:4161
SERVER.TRACER.SERVICE_NAME=boilerplate
SERVER.TRACER.EXPORTER_GRPC_ADDRESS=localhost:4317
DATASOURCE.BOILERPLATE_DATABASE.MASTER.DRIVER_NAME=postgres
DATASOURCE.BOILERPLATE_DATABASE.MASTER.DATA_SOURCE_NAME=postgres://postgres:12345678@localhost:5432/boilerplate?sslmode=disable
DATASOURCE.BOILERPLATE_DATABASE.MASTER.MAXIMUM_OPEN_CONNECTIONS=4
DATASOURCE.BOILERPLATE_DATABASE.MASTER.MAXIMUM_IDLE_CONNECTIONS=2
DATASOURCE.BOILERPLATE_DATABASE.MASTER.CONNECTION_MAXIMUM_IDLE_TIME=30s
DATASOURCE.BOILERPLATE_DATABASE.MASTER.CONNECTION_MAXIMUM_LIFE_TIME=1m
DATASOURCE.BOILERPLATE_DATABASE.MASTER.MAXIMUM_QUERY_DURATION_WARNING=500ms
DATASOURCE.BOILERPLATE_DATABASE.SLAVE.DRIVER_NAME=postgres
DATASOURCE.BOILERPLATE_DATABASE.SLAVE.DATA_SOURCE_NAME=postgres://postgres:12345678@localhost:5432/boilerplate?sslmode=disable
DATASOURCE.BOILERPLATE_DATABASE.SLAVE.MAXIMUM_OPEN_CONNECTIONS=4
DATASOURCE.BOILERPLATE_DATABASE.SLAVE.MAXIMUM_IDLE_CONNECTIONS=2
DATASOURCE.BOILERPLATE_DATABASE.SLAVE.CONNECTION_MAXIMUM_IDLE_TIME=30s
DATASOURCE.BOILERPLATE_DATABASE.SLAVE.CONNECTION_MAXIMUM_LIFE_TIME=1m
DATASOURCE.BOILERPLATE_DATABASE.SLAVE.MAXIMUM_QUERY_DURATION_WARNING=500ms
DATASOURCE.IN_MEMORY_DATABASE.DATA_SOURCE_NAME=redis://localhost:6379/0
DATASOURCE.EVENT_PRODUCER.DATA_SOURCE_NAME=localhost:4150
DATASOURCE.WEBHOOK_SITE_HTTP_CLIENT.BASE_URL=https://webhook.site
DATASOURCE.WEBHOOK_SITE_HTTP_CLIENT.ENDPOINT.WEBHOOK=/a46fd97b-b775-428c-890d-9d71851a6c32 ## Please change this value by accessing the https://webhook.site, and copy the /uuid path
GUEST.CACHE.ENABLE=true
GUEST.CACHE.KEYF=caches:entities:guests:%s
GUEST.CACHE.DURATION=5m
GUEST.EVENT.CREATED.ENABLE=true
GUEST.EVENT.CREATED.TOPIC=guest-created
GUEST.EVENT.DELETED.ENABLE=true
GUEST.EVENT.DELETED.TOPIC=guest-deleted
GUEST.EVENT.UPDATED.ENABLE=true
GUEST.EVENT.UPDATED.TOPIC=guest-updated
```

> The file **must be placed inside `./configs`** directory.

---

## 🚀 Running the App

### 1. Install Dependencies

```bash
go mod tidy
```

### 2. Run Migrations

Ensure PostgreSQL is up and database `boilerplate` exists. Then run:

```bash
go run main.go migrate up \
  -t postgres://postgres:12345678@localhost:5432/boilerplate?sslmode=disable \
  -d ./datasources/boilerplate_database/migrations
```

### 3. Start the App

You can choose to run only part of the app or all app together:

```bash
make http            # Run HTTP server
make grpc            # Run gRPC server
make event-consumer  # Run message consumer
make app             # Run everything together
```

---

## ✅ Testing the System

* Open [Swagger UI](http://localhost:3000/docs/swagger) & do some request with it
* Check PostgreSQL and Redis data with tools like **DBeaver** or **Redis Insight**
* View webhook events at [Webhook.site](https://webhook.site)
* View the tracer data at [Jaeger UI](http://localhost:16686)

<p align="center">
  <img src="./documentations/swagger.png" style="width:75%;" />
</p>

---

## 📌 Final Notes

This boilerplate is a great starting point to build scalable, testable, and maintainable Go applications using a clean architecture. Feel free to fork and customize it for your needs.

### Why This Boilerplate?

This project was created as a **base template** so that all future development has a **consistent code style standard**. Every layer (config, datasource, entity, repository, service, transport) follows the same patterns, making it easy for any developer — human or AI — to understand, extend, and maintain the codebase.

### How to Replace the Guest Entity

The sample `Guest` feature is meant to be replaced with your own domain entity. To make this as easy as possible while staying true to the boilerplate patterns:

1. **Replace entity** — Update `internal/models/entities/` with your domain entity and its struct tags (`table`, `db`, `primary_key`, `db_type`).
2. **Update DTOs** — Adjust `internal/models/dtos/` to match your entity fields.
3. **Regenerate mocks** — Run `make generate` after updating interfaces.
4. **Update handlers/VMs** — Adjust `transports/{http,grpc,event_consumer}/` handlers, VMs, and converters.
5. **Update config** — Adjust `configs/` with your own cache keys, event topics, etc.
6. **Remove old files** — Delete `Guest`-specific files from each layer, then verify the build and tests pass.

> 💡 The boilerplate patterns are designed so that swapping an entity is mostly a **rename-and-replace** exercise — the generic repository, service helpers (`withTransaction`, `tryDeleteEntityCaches`, `publishEvent`, etc.), and transport patterns stay the same.

Happy Coding! 🚀
