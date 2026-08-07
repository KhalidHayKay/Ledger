package idempotency

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo}
}

func (s *Service) ReserveKey(ctx context.Context, idempotencyKey, requestHash, paymentRef string) error {
	entry := Entry{
		RequestHash: requestHash,
		PaymentRef:  paymentRef,
	}

	entryEncode, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Error encoding idempotency entry: %s", err)
		return err
	}

	err = s.repo.SaveKey(ctx, fmt.Sprintf("idempotency-key:%s", idempotencyKey), string(entryEncode))
	if err != nil {
		log.Printf("Error saving idempotency key: %s", err)
		return err
	}

	return nil
}

func (s *Service) GetKeyReserve(ctx context.Context, idempotencyKey string) (*Entry, error) {
	result, err := s.repo.GetByKey(ctx, fmt.Sprintf("idempotency-key:%s", idempotencyKey))
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Printf("Error fetching idempotency key: %s", err)
		}

		return nil, err
	}

	var entry Entry

	err = json.Unmarshal([]byte(result), &entry)
	if err != nil {
		log.Printf("Error decoding idempotency entry: %s", err)
		return nil, err
	}

	return &entry, nil
}

func (s *Service) RemoveKey(ctx context.Context, idempotencyKey string) error {
	return s.repo.RemoveKey(ctx, fmt.Sprintf("idempotency-key:%s", idempotencyKey))
}
