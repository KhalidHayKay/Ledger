package paymentintent

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	idempotencyKey := r.Header.Get("Idempotency-Key")

	var req CreatePaymentIntentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.service.Create(
		r.Context(),
		CreateInput{
			Card: Card{
				Number: req.Card.Number,
				CVV:    req.Card.CVV,
				Expiry: CardExpiry{
					Month: req.Card.Expiry.Month,
					Year:  req.Card.Expiry.Year,
				},
			},
			Amount:  Amount{req.Amount, req.Currency},
			OrderId: req.OrderId, CustomerId: req.CustomerId,
			IdempotencyKey: idempotencyKey,
		},
	)

	w.Write([]byte("Ok!"))
}
