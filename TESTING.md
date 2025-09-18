# Testing Guide for Aktuell

This document outlines the testing strategy and best practices for the Aktuell project.

## Testing Structure

### 1. Unit Tests
Located in each package alongside the source code (`*_test.go` files):

- `pkg/models/models_test.go` - Tests for data structures and validation
- `pkg/server/websocket_test.go` - Tests for WebSocket server components
- `pkg/sync/manager_test.go` - Tests for synchronization logic

### 2. Integration Tests
Located in `tests/integration_test.go`:

- Full stack tests with real MongoDB and WebSocket connections
- End-to-end workflow testing
- Performance and load testing

## Running Tests

### Unit Tests Only
```bash
# Run all unit tests
make test

# Run unit tests for specific package
go test ./pkg/models
go test ./pkg/server
go test ./pkg/sync

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests
```bash
# Set up MongoDB (using Docker)
docker run -d --name mongo-test -p 27017:27017 mongo:latest

# Run integration tests
INTEGRATION_TESTS=1 go test ./tests/...

# Run with custom MongoDB URI
MONGODB_URI=mongodb://localhost:27017 INTEGRATION_TESTS=1 go test ./tests/...

# Clean up
docker stop mongo-test && docker rm mongo-test
```

### All Tests
```bash
# Run everything (requires MongoDB)
make test-all
```

## Test Categories

### 1. Unit Tests
- **Models**: Data validation, serialization/deserialization, constants
- **Server**: WebSocket handling, client management, message routing
- **Sync**: Change stream processing, subscription validation, database management

### 2. Integration Tests
- **WebSocket Communication**: Connection establishment, message exchange, error handling
- **Change Stream Integration**: Real-time change detection and broadcasting
- **Snapshot Streaming**: Initial data loading with pagination
- **Multi-client**: Concurrent connections and broadcasting
- **Performance**: Load testing and performance benchmarks

### 3. End-to-End Tests
- **Full Workflow**: Client connection → subscription → data changes → notifications
- **Error Scenarios**: Invalid subscriptions, network failures, database issues
- **Configuration**: Multi-database, multi-collection scenarios

## Testing Best Practices

### 1. Test Structure (Table-Driven Tests)
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected ExpectedType
        wantErr  bool
    }{
        {
            name:     "valid case",
            input:    validInput,
            expected: expectedOutput,
            wantErr:  false,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := FunctionUnderTest(tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}
```

### 2. Mocking External Dependencies
```go
// Use testify/mock for interfaces
type MockValidator struct {
    mock.Mock
}

func (m *MockValidator) IsValidSubscription(db, coll string) bool {
    args := m.Called(db, coll)
    return args.Bool(0)
}

// In tests:
mockValidator := &MockValidator{}
mockValidator.On("IsValidSubscription", "testdb", "users").Return(true)
```

### 3. Test Fixtures and Setup
```go
func TestMain(m *testing.M) {
    // Setup before all tests
    setup()
    
    // Run tests
    code := m.Run()
    
    // Cleanup after all tests
    cleanup()
    
    os.Exit(code)
}
```

### 4. Benchmarking
```go
func BenchmarkFunction(b *testing.B) {
    // Setup
    input := prepareInput()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        FunctionUnderTest(input)
    }
}
```

## Test Environment Setup

### Local Development
1. Install MongoDB locally or use Docker:
   ```bash
   docker run -d --name mongo-dev -p 27017:27017 mongo:latest
   ```

2. Install test dependencies:
   ```bash
   go mod download
   ```

3. Run tests:
   ```bash
   make test
   ```

### CI/CD Environment
1. Use MongoDB service container
2. Set environment variables:
   - `MONGODB_URI`
   - `INTEGRATION_TESTS=1`
   - `GO_TEST_TIMEOUT=300s`

### Test Data Management
- Use separate test database (`aktuell_test`)
- Clean up after each test
- Use predictable test data
- Avoid external dependencies in unit tests

## Coverage Goals
- **Unit Tests**: >80% line coverage
- **Integration Tests**: Critical paths covered
- **End-to-End Tests**: Main user workflows covered

## Performance Testing
- Benchmark critical functions
- Test with realistic data volumes
- Monitor memory usage and goroutine leaks
- Test concurrent client scenarios

## Debugging Tests
```bash
# Run specific test with detailed output
go test -v -run TestSpecificFunction ./pkg/models

# Run with race detection
go test -race ./...

# Run with memory sanitizer
go test -msan ./...

# Debug failing test
dlv test ./pkg/models -- -test.run TestSpecificFunction
```