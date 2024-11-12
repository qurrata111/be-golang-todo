package config

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

var RDB *redis.Client
var CTX = context.Background()

func InitRedis() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	RDB = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	if err := RDB.Ping(CTX).Err(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

}
