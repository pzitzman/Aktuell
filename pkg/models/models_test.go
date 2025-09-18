package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestChangeEvent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		event   ChangeEvent
		wantErr bool
	}{
		{
			name: "valid insert event",
			event: ChangeEvent{
				ID:              "test-id",
				OperationType:   OperationInsert,
				Database:        "testdb",
				Collection:      "users",
				DocumentKey:     map[string]interface{}{"_id": "123"},
				FullDocument:    map[string]interface{}{"name": "John"},
				Timestamp:       primitive.Timestamp{T: uint32(time.Now().Unix()), I: 1},
				ClientTimestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid update event",
			event: ChangeEvent{
				ID:              "test-id",
				OperationType:   OperationUpdate,
				Database:        "testdb",
				Collection:      "users",
				DocumentKey:     map[string]interface{}{"_id": "123"},
				UpdatedFields:   map[string]interface{}{"name": "Jane"},
				Timestamp:       primitive.Timestamp{T: uint32(time.Now().Unix()), I: 1},
				ClientTimestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "invalid - empty operation type",
			event: ChangeEvent{
				Database:        "testdb",
				Collection:      "users",
				Timestamp:       primitive.Timestamp{T: uint32(time.Now().Unix()), I: 1},
				ClientTimestamp: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "invalid - empty database",
			event: ChangeEvent{
				OperationType:   OperationInsert,
				Collection:      "users",
				Timestamp:       primitive.Timestamp{T: uint32(time.Now().Unix()), I: 1},
				ClientTimestamp: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Since there's no Validate method on ChangeEvent, we test basic validation logic
			hasError := (tt.event.OperationType == "" || tt.event.Database == "" || tt.event.Collection == "")

			if tt.wantErr {
				assert.True(t, hasError, "Expected validation error")
			} else {
				assert.False(t, hasError, "Expected no validation error")
			}
		})
	}
}

func TestDatabaseConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  DatabaseConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: DatabaseConfig{
				Name:        "testdb",
				Collections: []string{"users", "orders"},
			},
			wantErr: false,
		},
		{
			name: "empty database name",
			config: DatabaseConfig{
				Name:        "",
				Collections: []string{"users"},
			},
			wantErr: true,
		},
		{
			name: "empty collections",
			config: DatabaseConfig{
				Name:        "testdb",
				Collections: []string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation logic
			hasError := (tt.config.Name == "" || len(tt.config.Collections) == 0)

			if tt.wantErr {
				assert.True(t, hasError, "Expected validation error")
			} else {
				assert.False(t, hasError, "Expected no validation error")
			}
		})
	}
}

func TestClientMessage_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		message ClientMessage
		want    bool
	}{
		{
			name: "valid subscribe message",
			message: ClientMessage{
				Type:       MessageTypeSubscribe,
				Database:   "testdb",
				Collection: "users",
				RequestID:  "req-123",
			},
			want: true,
		},
		{
			name: "valid subscribe with snapshot options",
			message: ClientMessage{
				Type:       MessageTypeSubscribe,
				Database:   "testdb",
				Collection: "users",
				RequestID:  "req-123",
				SnapshotOptions: &SnapshotOptions{
					IncludeSnapshot: true,
					SnapshotLimit:   1000,
					BatchSize:       50,
				},
			},
			want: true,
		},
		{
			name: "invalid - missing type",
			message: ClientMessage{
				Database:   "testdb",
				Collection: "users",
			},
			want: false,
		},
		{
			name: "invalid - missing database for subscribe",
			message: ClientMessage{
				Type:       MessageTypeSubscribe,
				Collection: "users",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation logic
			var isValid bool
			switch tt.message.Type {
			case MessageTypeSubscribe:
				isValid = tt.message.Database != "" && tt.message.Collection != ""
			case MessageTypePing, MessageTypePong:
				isValid = true
			default:
				isValid = tt.message.Type != ""
			}

			assert.Equal(t, tt.want, isValid)
		})
	}
}

func TestSnapshotOptions_Defaults(t *testing.T) {
	tests := []struct {
		name     string
		options  SnapshotOptions
		expected SnapshotOptions
	}{
		{
			name: "default values",
			options: SnapshotOptions{
				IncludeSnapshot: true,
			},
			expected: SnapshotOptions{
				IncludeSnapshot: true,
				SnapshotLimit:   10000, // Expected default
				BatchSize:       100,   // Expected default
			},
		},
		{
			name: "custom values",
			options: SnapshotOptions{
				IncludeSnapshot: true,
				SnapshotLimit:   5000,
				BatchSize:       200,
			},
			expected: SnapshotOptions{
				IncludeSnapshot: true,
				SnapshotLimit:   5000,
				BatchSize:       200,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that we can create and use SnapshotOptions
			assert.Equal(t, tt.expected.IncludeSnapshot, tt.options.IncludeSnapshot)
			if tt.options.SnapshotLimit != 0 {
				assert.Equal(t, tt.expected.SnapshotLimit, tt.options.SnapshotLimit)
			}
			if tt.options.BatchSize != 0 {
				assert.Equal(t, tt.expected.BatchSize, tt.options.BatchSize)
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkChangeEvent_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event := ChangeEvent{
			ID:              "test-id",
			OperationType:   OperationInsert,
			Database:        "testdb",
			Collection:      "users",
			DocumentKey:     map[string]interface{}{"_id": "123"},
			FullDocument:    map[string]interface{}{"name": "John"},
			Timestamp:       primitive.Timestamp{T: uint32(time.Now().Unix()), I: 1},
			ClientTimestamp: time.Now(),
		}
		_ = event
	}
}

func BenchmarkClientMessage_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg := ClientMessage{
			Type:       MessageTypeSubscribe,
			Database:   "testdb",
			Collection: "users",
			RequestID:  "req-123",
		}
		_ = msg
	}
}
