package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	// Get collection
	collection := client.Database("aktuell").Collection("users")

	fmt.Println("MongoDB data generator started")
	fmt.Println("This will insert, update, and delete documents every few seconds")
	fmt.Println("Press Ctrl+C to stop")

	counter := 1

	for {
		// Insert a new document
		user := bson.M{
			"name":      fmt.Sprintf("User %d", counter),
			"email":     fmt.Sprintf("user%d@example.com", counter),
			"age":       20 + (counter % 50),
			"createdAt": time.Now(),
		}

		result, err := collection.InsertOne(context.TODO(), user)
		if err != nil {
			log.Printf("Failed to insert document: %v", err)
		} else {
			fmt.Printf("Inserted document with ID: %v\n", result.InsertedID)
		}

		time.Sleep(8 * time.Second)

		// Update a random document
		filter := bson.M{"name": bson.M{"$regex": "^User"}}
		update := bson.M{
			"$set": bson.M{
				"lastUpdated": time.Now(),
				"status":      "active",
			},
		}

		updateResult, err := collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			log.Printf("Failed to update document: %v", err)
		} else if updateResult.ModifiedCount > 0 {
			fmt.Printf("Updated %d document(s)\n", updateResult.ModifiedCount)
		}

		time.Sleep(2 * time.Second)

		// Delete old documents (older than 30 seconds)
		thirtySecondsAgo := time.Now().Add(-30 * time.Second)
		deleteFilter := bson.M{"createdAt": bson.M{"$lt": thirtySecondsAgo}}

		deleteResult, err := collection.DeleteMany(context.TODO(), deleteFilter)
		if err != nil {
			log.Printf("Failed to delete documents: %v", err)
		} else if deleteResult.DeletedCount > 0 {
			fmt.Printf("Deleted %d document(s)\n", deleteResult.DeletedCount)
		}

		time.Sleep(2 * time.Second)
		counter++
	}
}
