package config

import (
	"log"

	"github.com/delordemm1/qplayground/internal/platform"
	"github.com/redis/go-redis/v9"
)

// InitRedis initializes and returns a Redis client
func InitRedis() *redis.Client {
	opt, err := redis.ParseURL(platform.ENV_REDIS_URL)
	if err != nil {
		log.Fatal("Failed to parse Redis URL:", err)
	}

	client := redis.NewClient(opt)
	return client
}