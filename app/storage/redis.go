package storage

import (
	"context"
	"ledger/app/config"
	"time"

	"github.com/redis/go-redis/v9"
)

func InitRedis() (*redis.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	url := config.Env.Redis.Host + ":" + config.Env.Redis.Port

	redis := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: config.Env.Redis.Password,
		DB:       0,
	})

	_, err := redis.Ping(ctx).Result()
	if err != nil {
		err := redis.Close()
		if err != nil {
			return nil, err
		}

		return nil, err
	}

	return redis, nil
}
