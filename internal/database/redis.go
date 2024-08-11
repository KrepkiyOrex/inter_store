package database

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

var (
	Rdb *redis.Client
	ctx = context.Background()
)

// инициализация Redis
func InitRedis() {
	err := godotenv.Load("config/config.env")
	if err != nil {
		log.Printf("Error loading config.env file")
	}

    addr := os.Getenv("REDIS_ADDR")
    password := os.Getenv("REDIS_PASSWORD")
    db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
    if err != nil {
        log.Fatalf("Invalid REDIS_DB value: %v", err)
    }

	Rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// проверка подключения
	_, err = Rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
}
