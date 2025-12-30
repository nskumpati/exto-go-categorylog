package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client      *mongo.Client
	Database    *mongo.Database
	Collections struct {
		Categories *mongo.Collection
		Documents  *mongo.Collection
		Formats    *mongo.Collection
	}
)

type Config struct {
	URI          string
	DatabaseName string
	Timeout      time.Duration
}

// Initialize MongoDB connection
func Connect(config Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	// Set client options
	clientOptions := options.Client().ApplyURI(config.URI)

	// Connect to MongoDB
	var err error
	Client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Ping the database to verify connection
	err = Client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	// Initialize database and collections
	Database = Client.Database(config.DatabaseName)
	Collections.Categories = Database.Collection("categories")
	Collections.Documents = Database.Collection("documents")
	Collections.Formats = Database.Collection("formats")

	log.Printf("Successfully connected to MongoDB: %s", config.DatabaseName)
	return nil
}

// Disconnect closes the MongoDB connection
func Disconnect() error {
	if Client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := Client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %v", err)
	}

	log.Println("Disconnected from MongoDB")
	return nil
}

// GetCollection returns a collection by name
func GetCollection(name string) *mongo.Collection {
	if Database == nil {
		log.Fatal("Database not initialized. Call Connect() first")
	}
	return Database.Collection(name)
}

// IsConnected checks if the MongoDB connection is active
func IsConnected() bool {
	if Client == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := Client.Ping(ctx, nil)
	return err == nil
}
