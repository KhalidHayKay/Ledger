package paymentintent

import (
	"context"
	"ledger/internal/bank"
	"log"
)

type Service struct {
	repo     Repository
	bankRepo bank.Repository
}

func NewService(repo Repository, bankRepo bank.Repository) *Service {
	return &Service{repo, bankRepo}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (PaymentIntent, error) {
	body, err := s.bankRepo.Authorize(
		ctx,
		bank.AuthorizeInput{
			Card: bank.Card{
				Number: input.Card.Number,
				CVV:    input.Card.CVV,
				Expiry: bank.CardExpiry{
					Month: input.Card.Expiry.Month,
					Year:  input.Card.Expiry.Year,
				},
			},
			Amount: bank.Amount{
				Figure:   input.Amount.Figure,
				Currency: input.Amount.Currency,
			},
			IdempotencyKey: input.IdempotencyKey,
		},
	)
	if err != nil {
		return PaymentIntent{}, err
	}

	log.Println(body)

	return s.repo.Create(ctx, input.Amount.Figure, input.Amount.Currency, input.OrderId, input.CustomerId), nil
}
