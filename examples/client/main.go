package main

import (
	"fmt"
	"log"
	"time"

	"aktuell/pkg/client"
	"aktuell/pkg/models"

	"github.com/sirupsen/logrus"
)

func main() {
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create client
	c := client.NewClient("ws://localhost:8080/ws", &client.ClientOptions{
		Logger: logger,
	})

	// Set up change handler
	c.OnChange(func(change *models.ChangeEvent) {
		fmt.Printf("=== Change Event ===\n")
		fmt.Printf("Operation: %s\n", change.OperationType)
		fmt.Printf("Database: %s\n", change.Database)
		fmt.Printf("Collection: %s\n", change.Collection)
		fmt.Printf("Document Key: %v\n", change.DocumentKey)

		if change.FullDocument != nil {
			fmt.Printf("Full Document: %v\n", change.FullDocument)
		}

		if change.UpdatedFields != nil {
			fmt.Printf("Updated Fields: %v\n", change.UpdatedFields)
		}

		if len(change.RemovedFields) > 0 {
			fmt.Printf("Removed Fields: %v\n", change.RemovedFields)
		}

		fmt.Printf("Timestamp: %s\n", change.ClientTimestamp.Format(time.RFC3339))
		fmt.Println("===================")
	})

	// Connect to server
	if err := c.Connect(); err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer c.Disconnect()

	fmt.Println("Connected to Aktuell server")

	// Enable auto-reconnect
	c.EnableAutoReconnect(5 * time.Second)

	// Subscribe to multiple databases and collections
	fmt.Println("Subscribing to multiple databases...")

	// Subscribe to InventoryDB database collections
	if err := c.Subscribe("InventoryDB", "Products"); err != nil {
		log.Fatalf("Failed to subscribe to InventoryDB.Products: %v", err)
	}

	if err := c.Subscribe("InventoryDB", "Orders"); err != nil {
		log.Fatalf("Failed to subscribe to InventoryDB.Orders: %v", err)
	}

	// Subscribe to LogsDB database collections
	if err := c.Subscribe("LogsDB", "SystemLogs"); err != nil {
		log.Fatalf("Failed to subscribe to LogsDB.SystemLogs: %v", err)
	}

	fmt.Println("✅ Subscribed to changes across multiple databases:")
	fmt.Println("  - InventoryDB: Products, Orders")
	fmt.Println("  - LogsDB: SystemLogs")
	fmt.Println("\nListening for changes... (Press Ctrl+C to exit)")

	// Keep the client running
	select {}
}
