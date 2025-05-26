package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PostgresDSN   string
	RedisAddr     string
	RedisPassword string
	RedisUsername string
	AdminEmail    string
	AdminPassword string
	JWTSecret     string
}

func LoadConfig() (*Config, error) {
	godotenv.Load()

	cfg := &Config{
		PostgresDSN:   os.Getenv("POSTGRES_DSN"),
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisUsername: os.Getenv("REDIS_USERNAME"),
		AdminEmail:    os.Getenv("ADMIN_EMAIL"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
	}

	// validate mandatory fields
	if cfg.PostgresDSN == "" || cfg.JWTSecret == "" {
		return nil, errors.New("required environment variables are missing")
	}

	return cfg, nil
}
