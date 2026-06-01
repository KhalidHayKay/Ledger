package idempotency

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	redis *redis.Client
}

func NewRedisRepo(redis *redis.Client) *RedisRepo {
	return &RedisRepo{redis}
}

func (r *RedisRepo) SaveKey(ctx context.Context, idempotencyKey string, paymentIntentEncode string) error {
	log.Println("key from repo SET: ", idempotencyKey)
	return r.redis.Set(ctx, idempotencyKey, paymentIntentEncode, 24*time.Hour).Err()
}

func (r *RedisRepo) GetByKey(ctx context.Context, idempotencyKey string) (string, error) {

	log.Println("key from repo GET: ", idempotencyKey)
	return r.redis.Get(ctx, idempotencyKey).Result()
}
