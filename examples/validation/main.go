package main

import (
	"fmt"
	"log"
	"time"

	"aktuell/pkg/client"

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
	fmt.Println("🧪 Testing subscription validation...")

	// Test valid subscriptions
	fmt.Println("\n--- Testing VALID subscriptions ---")

	validTests := []struct {
		database   string
		collection string
	}{
		{"InventoryDB", "Products"},
		{"LogsDB", "SystemLogs"},
	}

	for _, test := range validTests {
		fmt.Printf("📋 Subscribing to %s.%s... ", test.database, test.collection)
		if err := c.Subscribe(test.database, test.collection); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		} else {
			fmt.Printf("✅ Success\n")
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Test invalid subscriptions
	fmt.Println("\n--- Testing INVALID subscriptions ---")

	invalidTests := []struct {
		database   string
		collection string
	}{
		{"UnknownDB", "SomeCollection"},
	}

	for _, test := range invalidTests {
		fmt.Printf("📋 Subscribing to %s.%s... ", test.database, test.collection)
		if err := c.Subscribe(test.database, test.collection); err != nil {
			fmt.Printf("✅ Expected error: %v\n", err)
		} else {
			fmt.Printf("❌ Unexpected success\n")
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("\n🏁 Validation test completed!")
}
