package idempotency

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	redis *redis.Client
}

func NewRedisRepo(redis *redis.Client) *RedisRepo {
	return &RedisRepo{redis}
}

func (r *RedisRepo) SaveKey(ctx context.Context, idempotencyKey string, entry string) error {
	return r.redis.Set(ctx, idempotencyKey, entry, 24*time.Hour).Err()
}

func (r *RedisRepo) GetByKey(ctx context.Context, idempotencyKey string) (string, error) {
	return r.redis.Get(ctx, idempotencyKey).Result()
}

func (r *RedisRepo) RemoveKey(ctx context.Context, idempotencyKey string) error {
	return r.redis.Del(ctx, idempotencyKey).Err()
}
