package server

import (
	"encoding/json"
	"testing"

	"aktuell/pkg/models"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestHub_ClientCount tests the ClientCount method of Hub
func TestHub_ClientCount(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	server := NewWebSocketServer("localhost:8080", logger)
	hub := server.hub

	// Initially, no clients
	assert.Equal(t, 0, hub.ClientCount())

	// Add a client
	client1 := &Client{ID: "c1", hub: hub, send: make(chan *models.ServerMessage, 1), subscriptions: make(map[string]*models.Subscription)}
	hub.mu.Lock()
	hub.clients[client1] = true
	hub.mu.Unlock()
	assert.Equal(t, 1, hub.ClientCount())

	// Add another client
	client2 := &Client{ID: "c2", hub: hub, send: make(chan *models.ServerMessage, 1), subscriptions: make(map[string]*models.Subscription)}
	hub.mu.Lock()
	hub.clients[client2] = true
	hub.mu.Unlock()
	assert.Equal(t, 2, hub.ClientCount())

	// Remove a client
	hub.mu.Lock()
	delete(hub.clients, client1)
	hub.mu.Unlock()
	assert.Equal(t, 1, hub.ClientCount())

	// Remove the last client
	hub.mu.Lock()
	delete(hub.clients, client2)
	hub.mu.Unlock()
	assert.Equal(t, 0, hub.ClientCount())
}

// MockValidator implements the SubscriptionValidator interface for testing
type MockValidator struct {
	mock.Mock
}

func (m *MockValidator) IsValidSubscription(database, collection string) bool {
	args := m.Called(database, collection)
	return args.Bool(0)
}

func (m *MockValidator) GetConfiguredDatabases() []models.DatabaseConfig {
	args := m.Called()
	return args.Get(0).([]models.DatabaseConfig)
}

// MockSnapshotStreamer implements the SnapshotStreamer interface for testing
type MockSnapshotStreamer struct {
	mock.Mock
}

func (m *MockSnapshotStreamer) StreamSnapshot(database, collection string, snapOpts *models.SnapshotOptions, callback func([]map[string]interface{}, int, int, error)) {
	m.Called(database, collection, snapOpts, callback)

	// Simulate streaming snapshot data
	if snapOpts != nil && snapOpts.IncludeSnapshot {
		testData := []map[string]interface{}{
			{"_id": "1", "name": "John", "age": 30},
			{"_id": "2", "name": "Jane", "age": 25},
		}

		// Call the callback with test data
		callback(testData, 1, 2, nil)
		callback(nil, 1, 2, nil) // Signal completion
	}
}

func TestWebSocketServer_Creation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during testing

	server := NewWebSocketServer("localhost:8080", logger)

	assert.NotNil(t, server)
	assert.NotNil(t, server.hub)
	assert.NotNil(t, server.server)
	assert.Equal(t, "localhost:8080", server.server.Addr)
}

func TestWebSocketServer_SetValidator(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	server := NewWebSocketServer("localhost:8080", logger)
	validator := &MockValidator{}

	server.SetValidator(validator)
	assert.Equal(t, validator, server.validator)
}

func TestWebSocketServer_SetSnapshotStreamer(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	server := NewWebSocketServer("localhost:8080", logger)
	streamer := &MockSnapshotStreamer{}

	server.SetSnapshotStreamer(streamer)
	assert.Equal(t, streamer, server.snapshotStreamer)
}

func TestHub_Creation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	server := NewWebSocketServer("localhost:8080", logger)
	hub := server.hub

	assert.NotNil(t, hub)
	assert.NotNil(t, hub.clients)
	assert.NotNil(t, hub.broadcast)
	assert.NotNil(t, hub.register)
	assert.NotNil(t, hub.unregister)
}

func TestWebSocketMessage_Parsing(t *testing.T) {
	tests := []struct {
		name        string
		messageJSON string
		expected    models.ClientMessage
		wantErr     bool
	}{
		{
			name:        "valid subscribe message",
			messageJSON: `{"type":"subscribe","database":"testdb","collection":"users","requestId":"req-123"}`,
			expected: models.ClientMessage{
				Type:       models.MessageTypeSubscribe,
				Database:   "testdb",
				Collection: "users",
				RequestID:  "req-123",
			},
			wantErr: false,
		},
		{
			name:        "subscribe with snapshot options",
			messageJSON: `{"type":"subscribe","database":"testdb","collection":"users","requestId":"req-123","snapshot_options":{"include_snapshot":true,"batch_size":50}}`,
			expected: models.ClientMessage{
				Type:       models.MessageTypeSubscribe,
				Database:   "testdb",
				Collection: "users",
				RequestID:  "req-123",
				SnapshotOptions: &models.SnapshotOptions{
					IncludeSnapshot: true,
					BatchSize:       50,
				},
			},
			wantErr: false,
		},
		{
			name:        "ping message",
			messageJSON: `{"type":"ping"}`,
			expected: models.ClientMessage{
				Type: models.MessageTypePing,
			},
			wantErr: false,
		},
		{
			name:        "invalid JSON",
			messageJSON: `{"type":"subscribe","database":}`,
			expected:    models.ClientMessage{},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var message models.ClientMessage
			err := json.Unmarshal([]byte(tt.messageJSON), &message)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Type, message.Type)
				assert.Equal(t, tt.expected.Database, message.Database)
				assert.Equal(t, tt.expected.Collection, message.Collection)
				assert.Equal(t, tt.expected.RequestID, message.RequestID)

				if tt.expected.SnapshotOptions != nil {
					require.NotNil(t, message.SnapshotOptions)
					assert.Equal(t, tt.expected.SnapshotOptions.IncludeSnapshot, message.SnapshotOptions.IncludeSnapshot)
					assert.Equal(t, tt.expected.SnapshotOptions.BatchSize, message.SnapshotOptions.BatchSize)
				}
			}
		})
	}
}

func TestServerMessage_Creation(t *testing.T) {
	tests := []struct {
		name    string
		message models.ServerMessage
		wantErr bool
	}{
		{
			name: "change message",
			message: models.ServerMessage{
				Type: models.MessageTypeChange,
				Change: &models.ChangeEvent{
					ID:            "change-123",
					OperationType: models.OperationInsert,
					Database:      "testdb",
					Collection:    "users",
					DocumentKey:   map[string]interface{}{"_id": "user-123"},
				},
				RequestID: "req-123",
			},
			wantErr: false,
		},
		{
			name: "error message",
			message: models.ServerMessage{
				Type:      models.MessageTypeError,
				Error:     "Invalid subscription",
				RequestID: "req-123",
			},
			wantErr: false,
		},
		{
			name: "snapshot message",
			message: models.ServerMessage{
				Type: models.MessageTypeSnapshot,
				SnapshotData: []map[string]interface{}{
					{"_id": "1", "name": "John"},
					{"_id": "2", "name": "Jane"},
				},
				SnapshotBatch:     1,
				SnapshotTotal:     2,
				SnapshotRemaining: 0,
				RequestID:         "req-123",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.message)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, jsonData)

				// Verify we can unmarshal it back
				var unmarshaled models.ServerMessage
				err = json.Unmarshal(jsonData, &unmarshaled)
				assert.NoError(t, err)
				assert.Equal(t, tt.message.Type, unmarshaled.Type)
				assert.Equal(t, tt.message.RequestID, unmarshaled.RequestID)
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkServerMessage_Marshal(b *testing.B) {
	message := models.ServerMessage{
		Type: models.MessageTypeChange,
		Change: &models.ChangeEvent{
			ID:            "change-123",
			OperationType: models.OperationInsert,
			Database:      "testdb",
			Collection:    "users",
			DocumentKey:   map[string]interface{}{"_id": "user-123"},
		},
		RequestID: "req-123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(message)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClientMessage_Unmarshal(b *testing.B) {
	messageJSON := []byte(`{"type":"subscribe","database":"testdb","collection":"users","requestId":"req-123"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var message models.ClientMessage
		err := json.Unmarshal(messageJSON, &message)
		if err != nil {
			b.Fatal(err)
		}
	}
}
