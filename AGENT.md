# Agent Guide — airport-app-backend

This file tells an AI agent exactly how to work in this repository.
Read this before touching any code.

---

## Environment Setup (Already Done in Sandbox)

The sandbox Stage 1 runs these automatically:
```bash
make swagger   # generates docs/ (required — imported by server/routes.go)
make mock      # generates mocks/ (required — imported by all controller tests)
```

If either command fails, no tests will run. Fix generation before anything else.

**Check generation worked:**
```bash
ls /repo/docs/docs.go     # must exist
ls /repo/mocks/           # must contain mock_*_repository.go files
```

**Run tests (controllers only — server/ package requires DB):**
```bash
/usr/local/go/bin/go test ./controllers/... ./middleware/... ./models/... -v
```

**Do NOT run `go test ./...`** — the `server/` package imports `docs/` and requires a database. It will fail in the sandbox. Run tests package by package.

---

## How to Add a New Feature

Follow this pattern for every new endpoint (example: adding DELETE /aircraft/:id).

### Step 1 — Add the handler in `controllers/`

```go
// controllers/aircraft_controller.go
// @Summary Delete aircraft by Id
// @Router /aircraft/{id} [delete]
// @Description Delete aircraft by ID
// @ID delete-aircraft-by-id
// @Tags aircraft
// @Param id path string true "Aircraft Id"
// @Success 200 "ok"
// @Failure 400 "Aircraft not found"
func (ac *AircraftController) HandleDeleteAircraft(ctx *gin.Context) {
    aircraftId := ctx.Param("id")
    err := ac.repository.DeleteAircraft(aircraftId)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"Error": "Incorrect aircraft id: " + aircraftId})
        return
    }
    ctx.JSON(http.StatusOK, "Deleted the aircraft successfully")
}
```

**Every handler must have Swagger annotations** (@Summary, @Router, @Description, @ID, @Tags, @Param, @Success, @Failure).

### Step 2 — Add the interface method in `repositories/`

```go
// repositories/aircraft_repository.go
type IAircraftRepository interface {
    GetAllAircrafts(page int) ([]models.Aircraft, error)
    GetAircraft(id string) (*models.Aircraft, error)
    CreateNewAircraft(aircraft *models.Aircraft) error
    UpdateAircraft(aircraft *models.Aircraft, id string) error
    DeleteAircraft(id string) error  // ← add this
}
```

### Step 3 — Implement in the service repository

```go
// repositories/aircraft_repository.go (implementation)
func (r *AircraftRepository) DeleteAircraft(id string) error {
    result := r.db.Delete(&models.Aircraft{}, "id = ?", id)
    if result.Error != nil || result.RowsAffected == 0 {
        return errors.New("aircraft not found")
    }
    return nil
}
```

### Step 4 — Register the route in `server/`

```go
// server/aircraft_router.go
func (srv *AppServer) AircraftRouter(DB *gorm.DB) {
    repository := repositories.NewAircraftRepository(DB)
    controller := controllers.NewAircraftController(repository)
    srv.router.DELETE("/aircraft/:id", controller.HandleDeleteAircraft)
    // ... existing routes
}
```

### Step 5 — Regenerate the mock (IMPORTANT)

After changing the interface:
```bash
cd /repo && make mock
```

Or run the specific mockgen command:
```bash
mockgen -destination=mocks/mock_aircraft_repository.go \
        -package=mocks \
        airport-app-backend/repositories \
        IAircraftRepository
```

### Step 6 — Write the tests

```go
// controllers/aircraft_controller_test.go
func TestHandleDeleteAircraft(t *testing.T) {
    beforeEachAircraftTest(t)
    aircraftId := "123"
    mockAircraftRepository.EXPECT().DeleteAircraft(aircraftId).Return(nil)
    aircraftContext.Request, _ = http.NewRequest(http.MethodDelete, AIRCRAFT, nil)
    aircraftContext.AddParam("id", aircraftId)

    aircraftController.HandleDeleteAircraft(aircraftContext)

    response := aircraftResponseRecorder.Result()
    assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestHandleDeleteAircraftWhenNotFound(t *testing.T) {
    beforeEachAircraftTest(t)
    nonExistentId := "nonexistent"
    mockAircraftRepository.EXPECT().DeleteAircraft(nonExistentId).Return(errors.New("not found"))
    aircraftContext.Request, _ = http.NewRequest(http.MethodDelete, AIRCRAFT, nil)
    aircraftContext.AddParam("id", nonExistentId)

    aircraftController.HandleDeleteAircraft(aircraftContext)

    response := aircraftResponseRecorder.Result()
    assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}
```

---

## How to Fix a Bug

### Step 1 — Read the failing test (if one exists)

If there's an existing failing test, read it first. It tells you exactly what's expected.

### Step 2 — Write the test BEFORE fixing the code

```go
func TestHandleGetAllAirlinesWhenPageIsNegative(t *testing.T) {
    beforeEachAirlineTest(t)
    // No mock expectation — request rejected before hitting repository
    airlineContext.Request, _ = http.NewRequest(http.MethodGet, AIRLINES+"?page=-1", nil)

    airlineController.HandleGetAllAirlines(airlineContext)

    response := airlineResponseRecorder.Result()
    assert.Equal(t, http.StatusBadRequest, response.StatusCode)
    responseBody, _ := io.ReadAll(response.Body)
    assert.Equal(t, "{\"msg\":\"Page number must be greater than 0\"}", string(responseBody))
}
```

Run it — it MUST FAIL first. If it passes, the test is wrong.

### Step 3 — Fix the implementation

```go
func (ac *AirlineController) HandleGetAllAirlines(ctx *gin.Context) {
    page, _ := strconv.Atoi(ctx.Query("page"))
    if page < 0 {
        ctx.JSON(400, gin.H{"msg": "Page number must be greater than 0"})
        return
    }
    // ...
}
```

### Step 4 — Run tests to confirm GREEN

```bash
/usr/local/go/bin/go test ./controllers/... -v -run TestHandleGetAllAirlines
```

---

## Test Patterns

### Test setup (call beforeEach at start of every test)

```go
var mockAirlineRepository *mocks.MockIAirlineRepository
var airlineController *AirlineController
var airlineContext *gin.Context
var airlineResponseRecorder *httptest.ResponseRecorder

func beforeEachAirlineTest(t *testing.T) {
    gomockController := gomock.NewController(t)
    defer gomockController.Finish()

    mockAirlineRepository = mocks.NewMockIAirlineRepository(gomockController)
    airlineController = NewAirlineController(mockAirlineRepository)
    airlineResponseRecorder = httptest.NewRecorder()
    airlineContext, _ = gin.CreateTestContext(airlineResponseRecorder)
}
```

### Required test scenarios for every handler

| Scenario | Mock setup | Expected status |
|---|---|---|
| Success | `EXPECT().Method().Return(data, nil)` | 200 or 201 |
| Repository error | `EXPECT().Method().Return(nil, errors.New("..."))` | 500 or 400 |
| Invalid input body | No mock (rejected before repo call) | 400 |
| Not found / invalid ID | `EXPECT().Method().Return(nil, errors.New("..."))` | 400 |
| Validation error (e.g. negative page) | No mock (rejected before repo call) | 400 |

### Test data — always use factory functions

```go
// DO THIS:
airline := factory.ConstructAirline()

// NOT THIS:
airline := models.Airline{Id: "abc123", Name: "Test Airline"}
```

Factory functions are in `models/factory/`. They generate valid random test data using `go-faker`.

### Reading response body

```go
response := airlineResponseRecorder.Result()
assert.Equal(t, http.StatusOK, response.StatusCode)

responseBody, _ := io.ReadAll(response.Body)
var result models.Airline
json.Unmarshal([]byte(responseBody), &result)
assert.Equal(t, expected, result)
```

---

## Common Pitfalls

### 1. Missing mocks/ after `make mock` fails

If `make mock` fails, the test file compiles but imports a package that doesn't exist.
**All tests will fail to compile — you see 0 passed, 0 failed.**

Fix: ensure `make mock` succeeds. Check if mockgen is in PATH:
```bash
which mockgen || ls /usr/local/bin/mockgen
```

### 2. Missing docs/ after `make swagger` fails

`server/routes.go` imports `_ "airport-app-backend/docs"`.
If docs/ doesn't exist, `go build ./...` fails.

Fix: run `make swagger` and check for errors.

### 3. Running `go test ./...` when server/ can't compile

`server/` imports `docs/`. Even if `docs/` exists, `server/` needs a database to run.
**Run tests on specific packages only:**
```bash
go test ./controllers/... ./middleware/... ./models/... -v
```

### 4. Mock not updated after interface change

If you add a method to an interface but don't regenerate the mock, compilation fails.
Always run `make mock` after changing any `I*Repository` interface.

### 5. Swagger annotation format

Every handler needs ALL of these annotations or `swag init` may warn:
```go
// @Summary     One-line description
// @Router      /path/{param} [method]
// @Description Longer description
// @ID          unique-id-for-this-endpoint
// @Tags        entity-name
// @Param       param path string true "Description"
// @Success     200 "ok"
// @Failure     400 "Bad request"
```

---

## Running Locally

```bash
# Start dependencies
docker-compose -f docker-compose.yaml up -d

# Generate code
make swagger
make mock

# Run
go run main.go

# Test
go test ./controllers/... ./middleware/... -v
```
