package bank

import "time"

type ficMartAuthorizeResponse struct {
	Amount          int       `json:"amount"`
	Currency        string    `json:"currency"`
	AuthorizationId string    `json:"authorization_id"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	ExpiresAt       time.Time `json:"expires_at"`
}

func (r ficMartAuthorizeResponse) ToPayment() Payment {
	return Payment{
		AuthorizationId: r.AuthorizationId,
		Amount:          r.Amount,
		Currency:        r.Currency,
		Status:          r.Status,
		CreatedAt:       &r.CreatedAt,
		ExpiresAt:       &r.ExpiresAt,
	}
}

type ficMartCaptureResponse struct {
	Amount          int       `json:"amount"`
	Currency        string    `json:"currency"`
	AuthorizationId string    `json:"authorization_id"`
	Status          string    `json:"status"`
	CaptureId       string    `json:"capture_id"`
	CapturedAt      time.Time `json:"captured_at"`
}

func (r ficMartCaptureResponse) ToPayment() Payment {
	return Payment{
		Amount:          r.Amount,
		Currency:        r.Currency,
		AuthorizationId: r.AuthorizationId,
		Status:          r.Status,
		CaptureId:       r.CaptureId,
		CapturedAt:      &r.CapturedAt,
	}
}

type ficMartVoidResponse struct {
	AuthorizationId string    `json:"authorization_id"`
	Status          string    `json:"status"`
	VoidId          string    `json:"void_id"`
	VoidedAt        time.Time `json:"voided_at"`
}

func (r ficMartVoidResponse) ToPayment() Payment {
	return Payment{
		AuthorizationId: r.AuthorizationId,
		Status:          r.Status,
		VoidId:          r.VoidId,
		VoidedAt:        &r.VoidedAt,
	}
}

type ficMartRefundResponse struct {
	Amount     int       `json:"amount"`
	Currency   string    `json:"currency"`
	CaptureId  string    `json:"capture_id"`
	Status     string    `json:"status"`
	RefundId   string    `json:"refund_id"`
	RefundedAt time.Time `json:"refunded_at"`
}

func (r ficMartRefundResponse) ToPayment() Payment {
	return Payment{
		Amount:     r.Amount,
		Currency:   r.Currency,
		CaptureId:  r.CaptureId,
		Status:     r.Status,
		RefundId:   r.RefundId,
		RefundedAt: &r.RefundedAt,
	}
}

type ficMartErrorResponse struct {
	Code    string `json:"error"`
	Message string `json:"message"`
}
