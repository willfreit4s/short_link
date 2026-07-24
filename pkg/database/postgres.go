// Package database provides functionality for initializing and managing a PostgreSQL database connection pool.
package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/willfreit4s/short_link/configs"
)

func InitDatabase(cfg *configs.Config, log *slog.Logger) (*pgxpool.Pool, error) {
	log.Info("initializing database pool")
	log.Info("Postgres config", "max_conn ", cfg.MaxConn, ", min_conn ", cfg.MinConn)

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser,
		cfg.DBPass,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	config.MaxConns = int32(cfg.MaxConn)
	config.MinConns = int32(cfg.MinConn)
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Error("database pool init failed", "err", err)
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		log.Error("database ping failed", "err", err)
		pool.Close()
		return nil, err
	}

	log.Info("database pool initialized")

	return pool, nil
}
