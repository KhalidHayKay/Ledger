package idempotency

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
		return err
	}

	log.Println(string(entryEncode))

	return s.repo.SaveKey(ctx, fmt.Sprintf("idempotency-key:%s", idempotencyKey), string(entryEncode))
}

func (s *Service) GetKeyReserve(ctx context.Context, idempotencyKey string) (*Entry, error) {
	log.Println("key from service GET: ", idempotencyKey)
	result, err := s.repo.GetByKey(ctx, fmt.Sprintf("idempotency-key:%s", idempotencyKey))
	if err != nil {
		return nil, err
	}

	log.Println(result)

	var entry Entry

	err = json.Unmarshal([]byte(result), &entry)
	if err != nil {
		return nil, err
	}

	return &entry, nil
}
