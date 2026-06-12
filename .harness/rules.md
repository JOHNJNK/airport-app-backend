# Agent Harness Rules — airport-app-backend

## Code Standards
- All HTTP handlers must have Swagger annotations (@Summary, @Router, @Description, @Tags, @Param, @Success, @Failure)
- Use dependency injection — never instantiate repositories directly inside controllers
- All error responses must use gin.H{"Error": "message"} format — UPPERCASE "Error" key, NOT lowercase "error"
- All success responses must return an appropriate HTTP status code (200, 201, etc.)

## Page parameter validation (exact pattern to use — do not deviate)

When adding strconv.Atoi validation for a `page` query parameter, use this exact pattern:

```go
pageStr := ctx.Query("page")
if pageStr != "" {
    var err error
    page, err = strconv.Atoi(pageStr)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid page parameter"})  // gate_controller
        // OR: ctx.JSON(http.StatusBadRequest, gin.H{"Error": "invalid page parameter"})  // airline_controller
        return
    }
}
```

Key: only validate when the string is non-empty. `?page=` (empty) and `?page` (absent) both default to 0 — do NOT return 400 for empty string. Only return 400 for `?page=abc` (non-numeric, non-empty).

Tests must match:
- `?page=abc` → 400
- `?page=` → 200 (empty = default 0)
- `?page` absent → 200 (absent = default 0)
- `?page=-1` → 400

## Error response format (critical — each file has its own pattern, match it exactly)

The two controllers use DIFFERENT error key casing. You MUST match the file you are editing:

**gate_controller.go** → uses lowercase `"error"`:
```go
ctx.JSON(http.StatusBadRequest, gin.H{"error": "your message here"})
```
Tests for gate_controller assert: `{"error":"your message here"}`

**airline_controller.go** → uses uppercase `"Error"`:
```go
ctx.JSON(http.StatusBadRequest, gin.H{"Error": "your message here"})
```
Tests for airline_controller assert: `{"Error":"your message here"}`

Never mix them. If you write `gin.H{"Error": ...}` in gate_controller.go the test will fail.
If you write `gin.H{"error": ...}` in airline_controller.go the test will fail.

## Architecture
- Models go in models/ — one file per entity
- Repository interfaces go in repositories/ — one file per entity
- HTTP handlers go in controllers/ — one file per entity with a matching _test.go
- Route registration goes in server/ — one *_router.go file per entity
- Test data factories go in models/factory/

## File Naming (critical — do not guess)

Mock files use pattern `{entity}_repository_mock.go` (NOT `mock_{entity}_repository.go`):
- mocks/airline_repository_mock.go
- mocks/gate_repository_mock.go
- mocks/aircraft_repository_mock.go
- mocks/slot_repository_mock.go
- mocks/health_repository_mock.go

Source files are in controllers/, models/, repositories/, server/ (NOT internal/).

## Test Standards
- Use gomock for mocking repository interfaces — never use real database in unit tests
- Use gin.CreateTestContext + httptest.NewRecorder for HTTP handler tests
- Use factory functions from models/factory/ for test data — never hardcode UUIDs or names
- Every handler function must have tests for:
    - Success case (mock returns valid data, expect 200/201)
    - Repository error (mock returns error, expect 500 or 400)
    - Invalid request body (missing required fields, expect 400)
    - Not found / invalid ID (expect 400)
- If you add a new repository interface method, regenerate its mock:
    mockgen -source=repositories/<name>_repository.go -destination=mocks/mock_<name>_repository.go

## Coverage
- Minimum 90% coverage on all changes
- 100% coverage on new controller handler functions
