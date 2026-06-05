package bank

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"ledger/app/config"
	"log"
	"net/http"
)

type FicMartBankRepo struct {
	client *http.Client
}

func NewFicMartBankRepo(client *http.Client) *FicMartBankRepo {
	return &FicMartBankRepo{client}
}

const (
	EndpointAuthorize = "/api/v1/authorizations"
	EndpointCapture   = "/api/v1/captures"
	EndpointRefund    = "/api/v1/refunds"
	EndpointVoid      = "/api/v1/voids"
)

func (r *FicMartBankRepo) Authorize(ctx context.Context, idempotencyKey string, input AuthorizeInput) (Payment, error) {
	data := map[string]any{
		"amount":       input.Amount.Figure,
		"card_number":  input.Card.Number,
		"cvv":          input.Card.CVV,
		"expiry_month": input.Card.Expiry.Month,
		"expiry_year":  input.Card.Expiry.Year,
	}

	body, err := r.makePostRequest(ctx, EndpointAuthorize, data, idempotencyKey)
	if err != nil {
		return Payment{}, nil
	}

	var resData ficMartAuthorizeResponse
	if err := json.Unmarshal(body, &resData); err != nil {
		return Payment{}, err
	}

	return resData.ToPayment(), nil
}

func (r *FicMartBankRepo) Capture(ctx context.Context, idempotencyKey, authorizationId, amount string) (Payment, error) {
	data := map[string]any{
		"amount":           amount,
		"authorization_id": authorizationId,
	}

	body, err := r.makePostRequest(ctx, EndpointAuthorize, data, idempotencyKey)
	if err != nil {
		return Payment{}, nil
	}

	var resData ficMartAuthorizeResponse
	if err := json.Unmarshal(body, &resData); err != nil {
		return Payment{}, err
	}

	return resData.ToPayment(), nil
}

func (r *FicMartBankRepo) Void(ctx context.Context, idempotencyKey, authorizationId string) (Payment, error) {
	data := map[string]any{
		"authorization_id": authorizationId,
	}

	body, err := r.makePostRequest(ctx, EndpointVoid, data, idempotencyKey)
	if err != nil {
		return Payment{}, nil
	}

	var resData ficMartVoidResponse
	if err := json.Unmarshal(body, &resData); err != nil {
		return Payment{}, err
	}

	return resData.ToPayment(), nil
}

func (r *FicMartBankRepo) Refund(ctx context.Context, idempotencyKey, captureId, amount string) (Payment, error) {
	data := map[string]any{
		"capture_id": captureId,
		"amount":     amount,
	}

	body, err := r.makePostRequest(ctx, EndpointRefund, data, idempotencyKey)
	if err != nil {
		return Payment{}, nil
	}

	var resData ficMartRefundResponse
	if err := json.Unmarshal(body, &resData); err != nil {
		return Payment{}, err
	}

	return resData.ToPayment(), nil
}

func (r *FicMartBankRepo) makePostRequest(ctx context.Context, endpoint string, data any, idempotencyKey string) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error encoding data: %s", err)
		return []byte(""), err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", config.Env.BankAPIBaseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return []byte(""), err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", idempotencyKey)

	res, err := r.client.Do(req)
	if err != nil {
		log.Printf("Client request error: %s", err)
		return []byte(""), err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error reading response body, %s", err)
		return []byte(""), err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		log.Printf("API error: status=%d body=%s", res.StatusCode, string(body))
		return []byte(""), fmt.Errorf("bank API error: %s", string(body))
	}

	return body, nil
}
