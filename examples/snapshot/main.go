package main

import (
	"fmt"
	"log"
	"sync"
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

	// Connect to server
	if err := c.Connect(); err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer c.Disconnect()

	fmt.Println("🔗 Connected to Aktuell server")
	fmt.Println("📸 Testing snapshot functionality...")

	// Use WaitGroup to wait for snapshot completion
	var wg sync.WaitGroup
	wg.Add(1)

	// Configure snapshot options
	snapOpts := &models.SnapshotOptions{
		IncludeSnapshot: true,
		SnapshotLimit:   50,                               // Limit to 50 documents for testing
		BatchSize:       10,                               // Send in batches of 10
		SnapshotFilter:  nil,                              // No additional filter
		SnapshotSort:    map[string]interface{}{"_id": 1}, // Sort by _id ascending
	}

	// Subscribe with snapshot support
	err := c.SubscribeWithOptions(
		"aktuell",
		"users",
		snapOpts,
		// Change handler
		func(change *models.ChangeEvent) {
			fmt.Printf("🔄 Live Change: %s in %s.%s\n",
				change.OperationType,
				change.Database,
				change.Collection,
			)
		},
		// Snapshot handler
		func(documents []map[string]interface{}, batchNum int, remaining int) {
			fmt.Printf("📦 Snapshot Batch %d: %d documents, %d remaining\n",
				batchNum, len(documents), remaining)

			// Show first document in each batch as example
			if len(documents) > 0 {
				fmt.Printf("   📄 Sample document: %v\n", documents[0])
			}
		},
		// Snapshot complete handler
		func() {
			fmt.Println("✅ Snapshot streaming completed!")
			wg.Done()
		},
		// Error handler
		func(err error) {
			fmt.Printf("❌ Error: %v\n", err)
			wg.Done()
		},
	)

	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	fmt.Println("⏳ Waiting for snapshot to complete...")

	// Wait for snapshot to complete or timeout after 30 seconds
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("🏁 Snapshot test completed!")
	case <-time.After(30 * time.Second):
		fmt.Println("⏰ Timeout waiting for snapshot completion")
	}

	fmt.Println("🎯 Now listening for live changes for 10 seconds...")
	time.Sleep(10 * time.Second)

	fmt.Println("👋 Test complete!")
}
