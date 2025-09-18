package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"aktuell/pkg/models"
	"aktuell/pkg/server"
	"aktuell/pkg/sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	MongoDB struct {
		URI       string                  `mapstructure:"uri"`
		Databases []models.DatabaseConfig `mapstructure:"databases"`
		// Legacy support for single database config
		Database    string   `mapstructure:"database"`
		Collections []string `mapstructure:"collections"`
	} `mapstructure:"mongodb"`

	Server struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	} `mapstructure:"server"`

	Logging struct {
		Level string `mapstructure:"level"`
	} `mapstructure:"logging"`
}

// normalizeConfig converts legacy single database config to multi-database format
func normalizeConfig(config *Config) []models.DatabaseConfig {
	var dbConfigs []models.DatabaseConfig

	// If new multi-database configuration is provided, use it
	if len(config.MongoDB.Databases) > 0 {
		return config.MongoDB.Databases
	}

	// Handle legacy single database configuration
	if config.MongoDB.Database != "" {
		dbConfigs = append(dbConfigs, models.DatabaseConfig{
			Name:        config.MongoDB.Database,
			Collections: config.MongoDB.Collections,
		})
	}

	return dbConfigs
}

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	// Set log level
	if level, err := logrus.ParseLevel(config.Logging.Level); err == nil {
		logger.SetLevel(level)
	}

	// Normalize configuration (handle legacy single database format)
	dbConfigs := normalizeConfig(config)

	logger.WithFields(logrus.Fields{
		"mongodb_uri":    config.MongoDB.URI,
		"databases":      len(dbConfigs),
		"server_address": fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port),
	}).Info("Starting Aktuell server")

	// Log database configurations
	for _, dbConfig := range dbConfigs {
		logger.WithFields(logrus.Fields{
			"database":    dbConfig.Name,
			"collections": dbConfig.Collections,
		}).Info("Configured database")
	}

	// Create MongoDB connection
	database, err := sync.NewDatabase(config.MongoDB.URI, "", logger) // Empty database name for multi-db support
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to MongoDB")
	}

	// Create WebSocket server
	serverAddr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
	wsServer := server.NewWebSocketServer(serverAddr, logger)

	// Create sync manager with multiple databases
	syncManager := sync.NewMultiDBManager(database, wsServer, dbConfigs, logger)

	// Set the sync manager as the validator and snapshot streamer for the WebSocket server
	wsServer.SetValidator(syncManager)
	wsServer.SetSnapshotStreamer(syncManager)

	// Start sync manager
	if err := syncManager.Start(); err != nil {
		logger.WithError(err).Fatal("Failed to start sync manager")
	}

	// Start WebSocket server in a goroutine
	go func() {
		if err := wsServer.Start(); err != nil {
			logger.WithError(err).Fatal("WebSocket server failed")
		}
	}()

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("Aktuell server started successfully")
	<-sigCh

	logger.Info("Shutting down Aktuell server...")

	// Graceful shutdown - stop components in reverse order of startup
	logger.Info("Stopping sync manager...")
	if err := syncManager.Stop(); err != nil {
		logger.WithError(err).Error("Error stopping sync manager")
	}

	logger.Info("Stopping WebSocket server...")
	if err := wsServer.Stop(); err != nil {
		logger.WithError(err).Error("Error stopping WebSocket server")
	}

	logger.Info("Closing database connection...")
	database.Close()

	logger.Info("Aktuell server shutdown complete")
}

// loadConfig loads configuration from various sources
func loadConfig() (*Config, error) {
	// Set default values
	viper.SetDefault("mongodb.uri", "mongodb://localhost:27017")
	viper.SetDefault("mongodb.database", "aktuell")
	viper.SetDefault("mongodb.collections", []string{})
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("logging.level", "info")

	// Environment variable configuration
	viper.SetEnvPrefix("AKTUELL")
	viper.AutomaticEnv()

	// Config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Try to read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
