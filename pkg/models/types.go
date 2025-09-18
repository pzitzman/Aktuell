package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChangeEvent represents a change event from MongoDB change streams
type ChangeEvent struct {
	ID              string                 `json:"id" bson:"_id"`
	OperationType   string                 `json:"operationType" bson:"operationType"`
	Database        string                 `json:"database" bson:"ns.db"`
	Collection      string                 `json:"collection" bson:"ns.coll"`
	DocumentKey     map[string]interface{} `json:"documentKey" bson:"documentKey"`
	FullDocument    map[string]interface{} `json:"fullDocument,omitempty" bson:"fullDocument,omitempty"`
	UpdatedFields   map[string]interface{} `json:"updatedFields,omitempty" bson:"updateDescription.updatedFields,omitempty"`
	RemovedFields   []string               `json:"removedFields,omitempty" bson:"updateDescription.removedFields,omitempty"`
	Timestamp       primitive.Timestamp    `json:"timestamp" bson:"clusterTime"`
	ClientTimestamp time.Time              `json:"clientTimestamp"`
}

// SnapshotOptions configures initial snapshot streaming
type SnapshotOptions struct {
	IncludeSnapshot bool                   `json:"include_snapshot"`          // Whether to stream existing documents
	SnapshotLimit   int                    `json:"snapshot_limit,omitempty"`  // Max documents to stream (default: 10000)
	BatchSize       int                    `json:"batch_size,omitempty"`      // Documents per batch (default: 100)
	SnapshotFilter  map[string]interface{} `json:"snapshot_filter,omitempty"` // Additional filter for snapshot
	SnapshotSort    map[string]interface{} `json:"snapshot_sort,omitempty"`   // Sort order for snapshot
}

// ClientMessage represents a message sent from client to server
type ClientMessage struct {
	Type            string                 `json:"type"`
	Database        string                 `json:"database,omitempty"`
	Collection      string                 `json:"collection,omitempty"`
	Data            map[string]interface{} `json:"data,omitempty"`
	RequestID       string                 `json:"requestId,omitempty"`
	SubscriptionID  string                 `json:"subscriptionId,omitempty"`   // Used for unsubscribe requests
	SnapshotOptions *SnapshotOptions       `json:"snapshot_options,omitempty"` // Options for initial snapshot
}

// ServerMessage represents a message sent from server to client
type ServerMessage struct {
	Type              string                   `json:"type"`
	Change            *ChangeEvent             `json:"change,omitempty"`
	Error             string                   `json:"error,omitempty"`
	ErrorCode         int                      `json:"errorCode,omitempty"`
	RequestID         string                   `json:"requestId,omitempty"`
	Success           bool                     `json:"success,omitempty"`
	Data              interface{}              `json:"data,omitempty"`               // Generic data field for various responses
	SnapshotData      []map[string]interface{} `json:"snapshot_data,omitempty"`      // Batch of snapshot documents
	SnapshotBatch     int                      `json:"snapshot_batch,omitempty"`     // Current batch number
	SnapshotTotal     int                      `json:"snapshot_total,omitempty"`     // Total documents in snapshot
	SnapshotRemaining int                      `json:"snapshot_remaining,omitempty"` // Documents remaining
}

// Subscription represents a client's subscription to changes
type Subscription struct {
	ID              string           `json:"id"`
	ClientID        string           `json:"clientId"`
	Database        string           `json:"database"`
	Collection      string           `json:"collection"`
	CreatedAt       time.Time        `json:"createdAt"`
	SnapshotOptions *SnapshotOptions `json:"snapshot_options,omitempty"`
}

// DatabaseConfig represents configuration for a specific database
type DatabaseConfig struct {
	Name        string   `json:"name" mapstructure:"name"`
	Collections []string `json:"collections" mapstructure:"collections"`
}

// Client represents a connected WebSocket client
type Client struct {
	ID            string          `json:"id"`
	Connection    interface{}     `json:"-"` // WebSocket connection (avoided import cycle)
	Subscriptions []*Subscription `json:"subscriptions"`
	ConnectedAt   time.Time       `json:"connectedAt"`
	LastSeen      time.Time       `json:"lastSeen"`
}

// MessageType constants for different message types
const (
	MessageTypeSubscribe     = "subscribe"
	MessageTypeUnsubscribe   = "unsubscribe"
	MessageTypeChange        = "change"
	MessageTypeError         = "error"
	MessageTypeInsert        = "insert"
	MessageTypeUpdate        = "update"
	MessageTypeDelete        = "delete"
	MessageTypePing          = "ping"
	MessageTypePong          = "pong"
	MessageTypeHealth        = "health"         // Health check endpoint
	MessageTypeSnapshot      = "snapshot"       // Batch of initial documents
	MessageTypeSnapshotStart = "snapshot_start" // Snapshot streaming started
	MessageTypeSnapshotEnd   = "snapshot_end"   // Snapshot streaming completed
)

// Operation types from MongoDB change streams
const (
	OperationInsert  = "insert"
	OperationUpdate  = "update"
	OperationReplace = "replace"
	OperationDelete  = "delete"
	OperationDrop    = "drop"
	OperationRename  = "rename"
)

// SubscriptionValidator interface for validating subscription requests
type SubscriptionValidator interface {
	IsValidSubscription(database, collection string) bool
	GetConfiguredDatabases() []DatabaseConfig
}

// SnapshotStreamer interface for streaming initial collection snapshots
type SnapshotStreamer interface {
	StreamSnapshot(database, collection string, snapOpts *SnapshotOptions, callback func([]map[string]interface{}, int, int, error))
}
