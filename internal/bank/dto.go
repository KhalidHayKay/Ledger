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
	return Payment{
		AuthorizationId: r.AuthorizationID,
		Amount:          r.Amount,
		Currency:        r.Currency,
		Status:          r.Status,
		CreatedAt:       &r.CreatedAt,
		ExpiresAt:       &r.ExpiresAt,
	}
}

type ficMartCaptureResponse struct {
	Amount          int       `json:"amount"`
	AuthorizationID string    `json:"authorization_id"`
	CaptureId       string    `json:"capture_id"`
	CapturedAt      time.Time `json:"captured_at"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
}

func (r ficMartCaptureResponse) ToPayment() Payment {
	return Payment{
		Amount:          r.Amount,
		AuthorizationId: r.AuthorizationID,
		CaptureId:       r.CaptureId,
		CapturedAt:      &r.CapturedAt,
		Currency:        r.Currency,
		Status:          r.Status,
	}
}

type ficMartVoidResponse struct {
	AuthorizationID string    `json:"authorization_id"`
	Status          string    `json:"status"`
	VoidId          string    `json:"void_id"`
	VoidedAt        time.Time `json:"voided_at"`
}

func (r ficMartVoidResponse) ToPayment() Payment {
	return Payment{
		AuthorizationId: r.AuthorizationID,
		Status:          r.Status,
		VoidId:          r.VoidId,
		VoidedAt:        &r.VoidedAt,
	}
}

type ficMartRefundResponse struct {
	Amount    int       `json:"amount"`
	CaptureId string    `json:"capture_id"`
	Currency  string    `json:"currency"`
	RefundId  string    `json:"refund_id"`
	RefundedAt time.Time `json:"refunded_at"`
	Status    string    `json:"status"`
}

func (r ficMartRefundResponse) ToPayment() Payment {
	return Payment{
		Amount:     r.Amount,
		CaptureId:  r.CaptureId,
		Currency:   r.Currency,
		RefundId:   r.RefundId,
		RefundedAt: &r.RefundedAt,
		Status:     r.Status,
	}
}
