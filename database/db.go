package database

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	UsersCollection *mongo.Collection
	PinsCollection  *mongo.Collection
)

func ConnectDB() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	uri := os.Getenv("MONGO_URL")
	if uri == "" {
		log.Fatal("You must set your MONGO_URL environmental variable")
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	// Ping the database
	if err := client.Ping(context.Background(), nil); err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB!")

	db := client.Database("pinterest")
	UsersCollection = db.Collection("users")
	PinsCollection = db.Collection("pins")
}
