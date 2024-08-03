package database

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	// Импортируйте свой модуль аутентификации
	// Импортируйте свой модуль утилит
)

// Подключение к MongoDB
var Client *mongo.Client

func InitMongo() {
	var err error
	Client, err = mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
}
