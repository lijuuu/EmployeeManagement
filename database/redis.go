package database

import (
	"context"
	"github.com/lijuuu/EmployeeManagement/config"
	"github.com/redis/go-redis/v9"
)

func InitRedis(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}