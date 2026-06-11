# Airport App Backend — Codebase Context

## What This Application Does

A REST API for managing airport operations. It handles:
- **Airlines** — CRUD for airline records
- **Gates** — Airport gate management
- **Aircraft** — Aircraft fleet management
- **Slots** — Time slot booking for gates
- **Health** — Service health check

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.22 |
| Web framework | Gin (`github.com/gin-gonic/gin`) |
| ORM | GORM (`gorm.io/gorm`) with PostgreSQL driver |
| Tracing | OpenCensus + Jaeger |
| Logging | zerolog (`github.com/rs/zerolog`) |
| API docs | Swagger via swaggo (`github.com/swaggo/swag`) |
| Mocking | gomock (`github.com/golang/mock`) |
| Assertions | testify (`github.com/stretchr/testify`) |
| Test data | go-faker (`github.com/go-faker/faker`) |

## Architecture

```
HTTP Request
    │
    ▼
server/ (route registration)
    │
    ▼
controllers/ (HTTP handlers)
    │   validates input, calls repository, returns JSON
    ▼
repositories/ (interfaces + DB implementations)
    │   IAirlineRepository, IGateRepository, etc.
    ▼
database/connection.go (GORM + PostgreSQL)
```

The key design: **controllers depend on repository INTERFACES, not implementations**.
This makes unit testing clean — mock the interface, never touch the real database.

## Directory Structure

```
airport-app-backend/
├── controllers/        HTTP handlers (one file + one _test.go per entity)
│   ├── airline_controller.go
│   ├── airline_controller_test.go
│   └── ... (gate, aircraft, slot, health)
├── repositories/       Database interfaces + implementations
│   ├── airline_repository.go       (IAirlineRepository interface)
│   ├── service_repository.go       (base struct with *gorm.DB)
│   └── ... (gate, aircraft, slot, health)
├── models/             Go structs for each entity
│   ├── airline.go
│   ├── factory/        Test data factories (ConstructAirline(), etc.)
│   └── ... (gate, aircraft, slot, health)
├── server/             Route registration (one *_router.go per entity)
│   ├── routes.go       Main setup — connects everything
│   ├── airline_router.go
│   └── ... (gate, aircraft, slot, health)
├── middleware/         HTTP middleware (logging, caching, security headers)
├── database/
│   ├── connection.go           GORM connection setup
│   └── db_migration/           Flyway SQL migrations (V001__ ... V009__)
├── mocks/              gomock generated mocks (GITIGNORED — run make mock)
├── docs/               Swagger generated docs (GITIGNORED — run make swagger)
├── config/             App configuration from environment variables
└── certs/              TLS certificate generation
```

## Key Models

```go
// Airline
type Airline struct {
    Id    string `json:"id" gorm:"primarykey"`
    Name  string `json:"name" binding:"required"`
    Count int    `json:"count"`
}

// Gate
type Gate struct {
    Id         string `json:"id" gorm:"primarykey"`
    GateNumber string `json:"gate_number" binding:"required"`
}

// Aircraft
type Aircraft struct {
    Id        string  `json:"id" gorm:"primarykey"`
    Name      string  `json:"name" binding:"required"`
    AirlineId string  `json:"airline_id" binding:"required"`
}

// Slot
type Slot struct {
    Id         string `json:"id" gorm:"primarykey"`
    GateId     string `json:"gate_id" binding:"required"`
    AircraftId string `json:"aircraft_id" binding:"required"`
    StartTime  string `json:"start_time" binding:"required"`
    EndTime    string `json:"end_time" binding:"required"`
}
```

## API Routes

| Method | Path | Handler | Description |
|---|---|---|---|
| GET | /airlines | HandleGetAllAirlines | List all airlines (pagination via ?page=N) |
| GET | /airline/:id | HandleGetAirline | Get airline by ID |
| POST | /airline | HandleCreateNewAirline | Create airline |
| PUT | /airline/:id | HandleUpdateAirline | Update airline |
| DELETE | /airline/:id | HandleDeleteAirline | Delete airline |
| GET | /gates | HandleGetAllGates | List all gates |
| GET | /gate/:id | HandleGate | Get gate by ID |
| POST | /gate | HandleCreateNewGate | Create gate |
| PUT | /gate/:id | HandleUpdateGate | Update gate |
| DELETE | /gate/:id | HandleDeleteGate | Delete gate |
| GET | /aircrafts | HandleGetAllAircrafts | List aircraft |
| GET | /aircraft/:id | HandleGetAircraft | Get aircraft by ID |
| POST | /aircraft | HandleCreateNewAircraft | Create aircraft |
| PUT | /aircraft/:id | HandleUpdateAircraft | Update aircraft |
| DELETE | /aircraft/:id | HandleDeleteAircraft | Delete aircraft |
| GET | /slots | HandleGetAllSlots | List slots |
| GET | /slot/:id | HandleGetSlot | Get slot by ID |
| POST | /slot | HandleCreateNewSlot | Create slot |
| DELETE | /slot/:id | HandleDeleteSlot | Delete slot |
| GET | /health | HandleHealth | Health check |
| GET | /swagger/* | — | Swagger UI |

## Environment Variables

```bash
AIRPORT_HOST=localhost
AIRPORT_POSTGRES_USER=postgres
AIRPORT_POSTGRES_PASSWORD=password
AIRPORT_DB_NAME=airport
AIRPORT_PORT=5432
AIRPORT_SSL_MODE=disable
APP_SRV_ADDRESS=0.0.0.0:8080
APP_ENABLE_TLS=false
```

## Generated Code (must be regenerated after changes)

**Mocks** (`mocks/` — gitignored):
```bash
make mock
# Generates: mocks/mock_airline_repository.go, mock_gate_repository.go, etc.
# Run after: changing any repository interface in repositories/
```

**Swagger docs** (`docs/` — gitignored):
```bash
make swagger
# Generates: docs/docs.go, docs/swagger.json, docs/swagger.yaml
# Run after: changing Swagger annotations on any handler
```

## Database Migrations

Migrations live in `database/db_migration/` and run via Flyway:
```
V001__CreateUUIDExtension.sql
V002__CreateAirlinesTable.sql
V003__CreateTableGates.sql
V004__PopulateWithRandValues.sql
V005__AlterAirlinesTable.sql
V006__CreateAircraftTable.sql
V007__PopulateAircraftsTable.sql
V008__CreateSlotsTable.sql
V009__PopulateSlotsTable.sql
```

Run locally: `docker-compose up -d` (starts Postgres + Flyway migration)
