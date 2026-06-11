# Agent Harness Rules — airport-app-backend

## Code Standards
- All HTTP handlers must have Swagger annotations (@Summary, @Router, @Description, @Tags, @Param, @Success, @Failure)
- Use dependency injection — never instantiate repositories directly inside controllers
- All error responses must use gin.H{"Error": "message"} format
- All success responses must return an appropriate HTTP status code (200, 201, etc.)

## Architecture
- Models go in models/ — one file per entity
- Repository interfaces go in repositories/ — one file per entity
- HTTP handlers go in controllers/ — one file per entity with a matching _test.go
- Route registration goes in server/ — one *_router.go file per entity
- Test data factories go in models/factory/

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
