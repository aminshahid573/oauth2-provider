package redis

import (
	"context"

	"github.com/aminshahid573/authexa/internal/config"
	"github.com/redis/go-redis/v9"
)

func NewClient(cfg config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	return rdb, nil
}
