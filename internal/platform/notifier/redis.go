package notifier

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

type RedisNotifier struct {
	rdb *redis.Client
}

func NewRedisNotifier(rdb *redis.Client) *RedisNotifier {
	return &RedisNotifier{rdb}
}

func (n *RedisNotifier) Subscribe(channel string) Subscription {
	sub := n.rdb.Subscribe(context.Background(), channel)
	return &RedisSubscription{sub: sub}
}

func (n *RedisNotifier) Publish(ctx context.Context, channel, result string) error {
	return n.rdb.Publish(ctx, channel, result).Err()
}

type RedisSubscription struct {
	sub *redis.PubSub
}

func (s *RedisSubscription) Wait(ctx context.Context) (string, error) {
	ch := s.sub.Channel()

	select {
	case msg := <-ch:
		log.Printf("Received message from channel %s: %s", msg.Channel, msg.Payload)
		return msg.Payload, nil
	case <-ctx.Done():
		log.Printf("Context done while waiting for message: %v", ctx.Err())
		return "", ctx.Err()
	}
}

func (s *RedisSubscription) Close() {
	_ = s.sub.Close()
}
