package paymentintent

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"ledger/internal/platform/render"
	"log"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_ = r.Body.Close()
	}()

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %s", err)
		render.ErrorJSON(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		render.ErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	requestHash, err := hashRequestBody(req)
	if err != nil {
		log.Printf("Error hashing request body: %s", err)
		render.ErrorJSON(w, "Unable to hash request body", http.StatusInternalServerError)
		return
	}

	paymentIntent, replayed, err := h.service.Create(
		r.Context(),
		r.Header.Get("X-Idempotency-Key"),
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
		renderErr(err, w)
		return
	}

	if replayed {
		w.Header().Set("X-Idempotent-Replayed", "true")
	}

	paymentIntent.Id = ""
	render.JSON(w, http.StatusCreated, "Payment Intent created successfully", paymentIntent)
}

func (h *Handler) Capture(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_ = r.Body.Close()
	}()

	var req CaptureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %s", err)
		render.ErrorJSON(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		render.ErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	requestHash, err := hashRequestBody(req)
	if err != nil {
		log.Printf("Error hashing request body: %s", err)
		render.ErrorJSON(w, "Unable to hash request body", http.StatusInternalServerError)
		return
	}

	paymentIntent, replayed, err := h.service.Capture(
		r.Context(), r.Header.Get("X-Idempotency-Key"),
		requestHash, req.PaymentRef, req.Amount,
	)
	if err != nil {
		renderErr(err, w)
		return
	}

	if replayed {
		w.Header().Set("X-Idempotent-Replayed", "true")
	}

	paymentIntent.Id = ""
	render.JSON(w, http.StatusCreated, "Payment captured successfully", paymentIntent)
}

func (h *Handler) Refund(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_ = r.Body.Close()
	}()

	var req RefundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %s", err)
		render.ErrorJSON(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		render.ErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	requestHash, err := hashRequestBody(req)
	if err != nil {
		log.Printf("Error hashing request body: %s", err)
		render.ErrorJSON(w, "Unable to hash request body", http.StatusInternalServerError)
		return
	}

	paymentIntent, replayed, err := h.service.Refund(
		r.Context(), r.Header.Get("X-Idempotency-Key"),
		requestHash, req.PaymentRef, req.Amount,
	)
	if err != nil {
		renderErr(err, w)
		return
	}

	if replayed {
		w.Header().Set("X-Idempotent-Replayed", "true")
	}

	paymentIntent.Id = ""
	render.JSON(w, http.StatusCreated, "Payment refunded successfully", paymentIntent)
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_ = r.Body.Close()
	}()

	var req CencelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %s", err)
		render.ErrorJSON(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		render.ErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	requestHash, err := hashRequestBody(req)
	if err != nil {
		log.Printf("Error hashing request body: %s", err)
		render.ErrorJSON(w, "Unable to hash request body", http.StatusInternalServerError)
		return
	}

	paymentIntent, replayed, err := h.service.Cancel(
		r.Context(),
		r.Header.Get("X-Idempotency-Key"), requestHash,
		req.PaymentRef,
	)
	if err != nil {
		renderErr(err, w)
		return
	}

	if replayed {
		w.Header().Set("X-Idempotent-Replayed", "true")
	}

	paymentIntent.Id = ""
	render.JSON(w, http.StatusCreated, "Payment successfully canceled", paymentIntent)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	paymentRef := r.URL.Query().Get("payment_ref")
	if paymentRef == "" {
		render.ErrorJSON(w, "payment_ref is required", http.StatusBadRequest)
		return
	}

	paymentIntent, err := h.service.GetByRef(r.Context(), paymentRef)
	if err != nil {
		renderErr(err, w)
		return
	}

	paymentIntent.Id = ""
	render.JSON(w, http.StatusOK, "Payment intent retrieved successfully", paymentIntent)
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

func renderErr(err error, w http.ResponseWriter) {
	switch {
	case errors.Is(err, ErrIdempotencyKeyReuse):
		render.ErrorJSON(w, err.Error(), http.StatusConflict)

	case errors.Is(err, ErrBankDeclined):
		render.ErrorJSON(w, err.Error(), http.StatusBadRequest)

	case errors.Is(err, ErrInconsistentState):
		render.ErrorJSON(w, err.Error(), http.StatusInternalServerError)

	case errors.Is(err, ErrInternal):
		render.ErrorJSON(w, err.Error(), http.StatusInternalServerError)

	case errors.Is(err, ErrOperationNotAllowed):
		render.ErrorJSON(w, err.Error(), http.StatusForbidden)

	case errors.Is(err, ErrNotFound):
		render.ErrorJSON(w, err.Error(), http.StatusNotFound)

	default:
		render.ErrorJSON(w, "Unexpected error", http.StatusInternalServerError)
	}
}
