// Package configs provides configuration loading and management for the application.
package configs

import (
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type config struct {
	// Database configuration
	DBHost  string
	DBPort  string
	DBUser  string
	DBPass  string
	DBName  string
	MaxConn int
	MinConn int

	// Redis configuration
	RedisHost string
	RedisPort string
	RedisPass string
	RedisDB   int
	RedisTTL  time.Duration

	// Server configuration
	ServerPort  int
	ServiceName string
	Environment string
}

type Config = config

var Instance *Config

func LoadConfig() *config {
	_ = godotenv.Load()

	v := viper.New()

	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", "5432")
	v.SetDefault("DB_USER", "postgres")
	v.SetDefault("DB_PASS", "postgres")
	v.SetDefault("DB_NAME", "short_link")
	v.SetDefault("MAX_CONN", 30)
	v.SetDefault("MIN_CONN", 10)
	v.SetDefault("REDIS_HOST", "localhost")
	v.SetDefault("REDIS_PORT", "6379")
	v.SetDefault("REDIS_PASS", "wsbmrc8PH18AIjID")
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("REDIS_TTL", "24h")
	v.SetDefault("SERVER_PORT", 8080)
	v.SetDefault("SERVICE_NAME", "short_link_service")
	v.SetDefault("ENVIRONMENT", "local")

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	v.SetConfigFile(".env")
	v.SetConfigType("env")

	Instance = &Config{
		DBHost:      v.GetString("DB_HOST"),
		DBPort:      v.GetString("DB_PORT"),
		DBUser:      v.GetString("DB_USER"),
		DBPass:      v.GetString("DB_PASS"),
		DBName:      v.GetString("DB_NAME"),
		MaxConn:     v.GetInt("MAX_CONN"),
		MinConn:     v.GetInt("MIN_CONN"),
		RedisHost:   v.GetString("REDIS_HOST"),
		RedisPort:   v.GetString("REDIS_PORT"),
		RedisPass:   v.GetString("REDIS_PASS"),
		RedisDB:     v.GetInt("REDIS_DB"),
		RedisTTL:    v.GetDuration("REDIS_TTL"),
		ServerPort:  v.GetInt("SERVER_PORT"),
		ServiceName: v.GetString("SERVICE_NAME"),
		Environment: v.GetString("ENVIRONMENT"),
	}

	return Instance
}

func (c *config) GetEnv() string {
	return c.Environment
}
