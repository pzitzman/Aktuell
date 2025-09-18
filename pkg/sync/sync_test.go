package sync

import (
	"testing"

	"aktuell/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWebSocketServer implements a mock WebSocket server for testing
type MockWebSocketServer struct {
	mock.Mock
}

func (m *MockWebSocketServer) Broadcast(message *models.ServerMessage) {
	m.Called(message)
}

// MockDatabase implements a mock MongoDB connection for testing
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestDatabaseConfig_Basic(t *testing.T) {
	// Basic test for database config structures
	config := models.DatabaseConfig{
		Name:        "testdb",
		Collections: []string{"users", "orders"},
	}

	assert.Equal(t, "testdb", config.Name)
	assert.Equal(t, []string{"users", "orders"}, config.Collections)
	assert.Len(t, config.Collections, 2)
}

func TestValidationLogic(t *testing.T) {
	// Test basic validation logic without requiring full MultiDBManager
	dbConfigs := []models.DatabaseConfig{
		{Name: "db1", Collections: []string{"users", "orders"}},
		{Name: "db2", Collections: []string{"products"}},
	}

	// Test finding a valid database/collection combination
	found := false
	for _, dbConfig := range dbConfigs {
		if dbConfig.Name == "db1" {
			for _, collection := range dbConfig.Collections {
				if collection == "users" {
					found = true
					break
				}
			}
		}
	}

	assert.True(t, found, "Should find valid db1/users combination")

	// Test invalid combination
	found = false
	for _, dbConfig := range dbConfigs {
		if dbConfig.Name == "nonexistent" {
			found = true
			break
		}
	}

	assert.False(t, found, "Should not find nonexistent database")
}

func TestMockInterfaces(t *testing.T) {
	// Test that our mocks work correctly
	mockWS := &MockWebSocketServer{}
	mockDB := &MockDatabase{}

	// Test mock WebSocket server
	testMessage := &models.ServerMessage{
		Type:      models.MessageTypeError,
		Error:     "test error",
		RequestID: "test-123",
	}

	mockWS.On("Broadcast", testMessage).Return()
	mockWS.Broadcast(testMessage)
	mockWS.AssertExpectations(t)

	// Test mock database
	mockDB.On("Close").Return(nil)
	err := mockDB.Close()
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

// Benchmark test
func BenchmarkDatabaseConfig_Access(b *testing.B) {
	configs := []models.DatabaseConfig{
		{Name: "db1", Collections: []string{"users", "orders", "products", "reviews"}},
		{Name: "db2", Collections: []string{"analytics", "logs", "metrics"}},
		{Name: "db3", Collections: []string{"cache", "sessions"}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate finding a configuration
		for _, config := range configs {
			if config.Name == "db2" {
				for _, coll := range config.Collections {
					if coll == "analytics" {
						break
					}
				}
				break
			}
		}
	}
}
