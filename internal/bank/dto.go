package bank

import "time"

type ficMartAuthorizeResponse struct {
	Amount          int       `json:"amount"`
	AuthorizationID string    `json:"authorization_id"`
	PaymentId       string    `json:"payment"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	ExpiresAt       time.Time `json:"expires_at"`
}

func (r ficMartAuthorizeResponse) ToPayment() Payment {
	paymentId := r.PaymentId
	if paymentId == "" {
		paymentId = r.AuthorizationID
	}

	return Payment{
		Id:        paymentId,
		Amount:    r.Amount,
		Currency:  r.Currency,
		Status:    r.Status,
		CreatedAt: r.CreatedAt,
		ExpiresAt: r.ExpiresAt,
	}
}
