package paymentintent

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"ledger/app/render"
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
		render.ErrorJSON(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	requestHash, err := hashRequestBody(req)
	if err != nil {
		render.ErrorJSON(w, "Unable to hash request body", http.StatusInternalServerError)
		return
	}

	paymentIntent, replayed, err := h.service.Create(
		r.Context(),
		idempotencyKey,
		requestHash,
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
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrIdempotencyKeyReuse):
			render.ErrorJSON(w, err.Error(), http.StatusConflict)

		case errors.Is(err, ErrBankDeclined):
			render.ErrorJSON(w, err.Error(), http.StatusUnprocessableEntity)

		case errors.Is(err, ErrInconsistentState):
			render.ErrorJSON(w, err.Error(), http.StatusInternalServerError)

		case errors.Is(err, ErrInternal):
			render.ErrorJSON(w, err.Error(), http.StatusInternalServerError)

		default:
			render.ErrorJSON(w, "Unexpected error", http.StatusInternalServerError)
		}

		return
	}

	if replayed {
		w.Header().Set("X-Idempotent-Replayed", "true")
	}

	render.JSON(w, http.StatusCreated, "Payment Intent created successfully", paymentIntent)
}

func hashRequestBody(body any) (string, error) {
	canonical, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(canonical)
	hash := hex.EncodeToString(sum[:])

	return hash, nil
}
