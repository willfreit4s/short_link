package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	redisclient "github.com/redis/go-redis/v9"
	"github.com/willfreit4s/short_link/configs"
)

func InitRedis(cfg *configs.Config, log *slog.Logger) (*redisclient.Client, error) {
	log.Info("initializing redis client")

	client := redisclient.NewClient(&redisclient.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Error("redis ping failed", "err", err)
		return client, err
	}

	log.Info("redis client initialized")

	return client, nil
}