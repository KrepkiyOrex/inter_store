package database

import (
	"context"
	"encoding/json"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConfig struct {
	Database   string `json:"database"`
	Collection string `json:"collection"`
}

var AppConfig MongoConfig

var Client *mongo.Client

func init() {
	configFile, err := os.ReadFile("config/mongo.config.json")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	err = json.Unmarshal(configFile, &AppConfig)
	if err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}
}

// инициализирует клиент MongoDB и подключается к базе данных.
func InitMongoClint() {
	log.Info("Connecting to MongoDB...")

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("Mongo connect error: %v", err)
	}

	log.Info("Pinging MongoDB...")

	// Увеличим таймаут для контекста, чтобы более наглядно показать процесс ожидания
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Mongo ping error: %v", err)
	}

	log.Info("MongoDB connection established successfully.")

	Client = client
}

// return a collection from a MongoDB
func GetCollection() *mongo.Collection {
	return Client.Database(AppConfig.Database).Collection(AppConfig.Collection)
}
