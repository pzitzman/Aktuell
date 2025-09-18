package sync

import (
	"context"
	"fmt"
	"sync"
	"time"

	"aktuell/pkg/models"
	"aktuell/pkg/server"

	"github.com/sirupsen/logrus"
)

// Manager coordinates synchronization between MongoDB change streams and WebSocket clients
type Manager struct {
	database    *Database
	wsServer    *server.WebSocketServer
	logger      *logrus.Logger
	ctx         context.Context
	cancel      context.CancelFunc
	collections []string
	wg          sync.WaitGroup
}

// NewManager creates a new synchronization manager
func NewManager(database *Database, wsServer *server.WebSocketServer, collections []string, logger *logrus.Logger) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		database:    database,
		wsServer:    wsServer,
		logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
		collections: collections,
	}
}

// Start starts the synchronization manager
func (m *Manager) Start() error {
	// Start MongoDB change stream
	if err := m.database.StartChangeStream(m.collections); err != nil {
		return err
	}

	// Start processing change events
	m.wg.Add(1)
	go m.processChangeEvents()

	// Start health monitoring
	m.wg.Add(1)
	go m.healthMonitor()

	m.logger.WithFields(logrus.Fields{
		"collections": m.collections,
	}).Info("Synchronization manager started")

	return nil
}

// Stop stops the synchronization manager
func (m *Manager) Stop() {
	m.cancel()
	m.wg.Wait()
	m.logger.Info("Synchronization manager stopped")
}

// processChangeEvents processes change events from MongoDB and broadcasts them to WebSocket clients
func (m *Manager) processChangeEvents() {
	defer m.wg.Done()

	changesCh := m.database.GetChanges()

	for {
		select {
		case <-m.ctx.Done():
			m.logger.Info("Change event processor stopping")
			return

		case change, ok := <-changesCh:
			if !ok {
				m.logger.Warn("Change events channel closed")
				return
			}

			m.logger.WithFields(logrus.Fields{
				"operation":  change.OperationType,
				"database":   change.Database,
				"collection": change.Collection,
				"doc_id":     change.DocumentKey,
			}).Debug("Processing change event")

			// Broadcast the change to WebSocket clients
			m.wsServer.BroadcastChange(change)
		}
	}
}

// healthMonitor periodically logs health information
func (m *Manager) healthMonitor() {
	defer m.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			hub := m.wsServer.GetHub()
			m.logger.WithFields(logrus.Fields{
				"active_clients": m.getClientCount(hub),
				"collections":    m.collections,
			}).Info("Sync manager health check")
		}
	}
}

// getClientCount safely gets the number of active clients
func (m *Manager) getClientCount(hub *server.Hub) int {
	// We would need to add a method to Hub to get client count safely
	// For now, we'll return 0 as a placeholder
	return 0
}

// SyncStats represents synchronization statistics
type SyncStats struct {
	ActiveClients      int              `json:"activeClients"`
	TotalChangeEvents  int64            `json:"totalChangeEvents"`
	Collections        []string         `json:"collections"`
	LastChangeTime     *time.Time       `json:"lastChangeTime,omitempty"`
	ChangeEventsByType map[string]int64 `json:"changeEventsByType"`
}

// Stats returns current synchronization statistics
func (m *Manager) Stats() *SyncStats {
	return &SyncStats{
		ActiveClients:      0, // Would get from hub
		Collections:        m.collections,
		ChangeEventsByType: make(map[string]int64),
	}
}

// MultiDBManager coordinates synchronization between multiple MongoDB databases and WebSocket clients
type MultiDBManager struct {
	database  *Database
	wsServer  *server.WebSocketServer
	logger    *logrus.Logger
	ctx       context.Context
	cancel    context.CancelFunc
	dbConfigs []models.DatabaseConfig
	managers  map[string]*Manager // Database name -> single-db manager
	wg        sync.WaitGroup
}

// NewMultiDBManager creates a new multi-database synchronization manager
func NewMultiDBManager(database *Database, wsServer *server.WebSocketServer, dbConfigs []models.DatabaseConfig, logger *logrus.Logger) *MultiDBManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &MultiDBManager{
		database:  database,
		wsServer:  wsServer,
		logger:    logger,
		ctx:       ctx,
		cancel:    cancel,
		dbConfigs: dbConfigs,
		managers:  make(map[string]*Manager),
	}
}

// Start starts all database synchronization managers
func (m *MultiDBManager) Start() error {
	for _, dbConfig := range m.dbConfigs {
		// Create a separate Database instance for each database
		dbInstance, err := NewDatabase(m.database.GetConnectionURI(), dbConfig.Name, m.logger)
		if err != nil {
			return fmt.Errorf("failed to create database instance for %s: %w", dbConfig.Name, err)
		}

		// Create a manager for this specific database
		manager := NewManager(dbInstance, m.wsServer, dbConfig.Collections, m.logger)
		m.managers[dbConfig.Name] = manager

		// Start the manager
		if err := manager.Start(); err != nil {
			return fmt.Errorf("failed to start manager for database %s: %w", dbConfig.Name, err)
		}

		m.logger.WithFields(logrus.Fields{
			"database":    dbConfig.Name,
			"collections": dbConfig.Collections,
		}).Info("Started synchronization for database")
	}

	return nil
}

// Stop stops all database synchronization managers
func (m *MultiDBManager) Stop() error {
	m.cancel()

	// Stop all managers - they don't return errors
	for dbName, manager := range m.managers {
		manager.Stop()
		m.logger.WithField("database", dbName).Info("Stopped synchronization for database")
	}

	m.wg.Wait()
	return nil
}

// Stats returns aggregated synchronization statistics across all databases
func (m *MultiDBManager) Stats() map[string]*SyncStats {
	stats := make(map[string]*SyncStats)
	for dbName, manager := range m.managers {
		stats[dbName] = manager.Stats()
	}
	return stats
}

// IsValidSubscription validates if a database/collection combination is configured
func (m *MultiDBManager) IsValidSubscription(database, collection string) bool {
	for _, dbConfig := range m.dbConfigs {
		if dbConfig.Name == database {
			// If no specific collections configured, allow all collections in the database
			if len(dbConfig.Collections) == 0 {
				return true
			}
			// Check if the specific collection is configured
			for _, configuredCollection := range dbConfig.Collections {
				if configuredCollection == collection {
					return true
				}
			}
			// Database found but collection not in the configured list
			return false
		}
	}
	// Database not found in configuration
	return false
}

// GetConfiguredDatabases returns a list of configured databases and their collections
func (m *MultiDBManager) GetConfiguredDatabases() []models.DatabaseConfig {
	return m.dbConfigs
}

// StreamSnapshot streams existing documents from a database collection
func (m *MultiDBManager) StreamSnapshot(database, collection string, snapOpts *models.SnapshotOptions, callback func([]map[string]interface{}, int, int, error)) {
	// Find the manager for this database
	manager, exists := m.managers[database]
	if !exists {
		callback(nil, 0, 0, fmt.Errorf("database '%s' is not configured", database))
		return
	}

	// Use the database instance from the manager to stream snapshot
	manager.database.StreamSnapshot(collection, snapOpts, callback)
}
