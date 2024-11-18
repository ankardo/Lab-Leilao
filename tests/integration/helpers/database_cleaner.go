package helpers

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CleanDatabase(uri string, databaseName string) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer client.Disconnect(context.TODO())

	collections, err := client.Database(databaseName).ListCollectionNames(context.TODO(), map[string]interface{}{})
	if err != nil {
		log.Fatalf("Failed to list collections: %v", err)
	}

	for _, collection := range collections {
		if err := client.Database(databaseName).Collection(collection).Drop(context.TODO()); err != nil {
			log.Printf("Failed to drop collection %s: %v", collection, err)
		}
	}
	log.Println("Database cleaned successfully.")
}
