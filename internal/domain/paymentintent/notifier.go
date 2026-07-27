package paymentintent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisNotifier struct {
	rdb *redis.Client
}

func NewRedisNotifier(rdb *redis.Client) *RedisNotifier {
	return &RedisNotifier{rdb}
}

func channelKey(paymentRef string) string {
	return fmt.Sprintf("payment:result:%s", paymentRef)
}

func (n *RedisNotifier) Subscribe(paymentRef string) *RedisSubscription {
	sub := n.rdb.Subscribe(context.Background(), channelKey(paymentRef))
	return &RedisSubscription{sub: sub}
}

func (n *RedisNotifier) Publish(ctx context.Context, paymentRef string, result PaymentIntent) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return n.rdb.Publish(ctx, channelKey(paymentRef), data).Err()
}

type RedisSubscription struct {
	sub *redis.PubSub
}

func (s *RedisSubscription) Wait(ctx context.Context) (PaymentIntent, error) {
	ch := s.sub.Channel()

	select {
	case msg := <-ch:
		var pi PaymentIntent
		if err := json.Unmarshal([]byte(msg.Payload), &pi); err != nil {
			return PaymentIntent{}, err
		}
		return pi, nil
	case <-ctx.Done():
		return PaymentIntent{}, ctx.Err()
	}
}

func (s *RedisSubscription) Close() {
	_ = s.sub.Close()
}
