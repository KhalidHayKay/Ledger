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

func (r *FicMartBankRepo) Authorize(ctx context.Context, input AuthorizeInput) (Payment, error) {
	url := config.Env.BankAPIBaseURL + "/api/v1/authorizations"
	data := map[string]any{
		"amount":       input.Amount.Figure,
		"card_number":  input.Card.Number,
		"cvv":          input.Card.CVV,
		"expiry_month": input.Card.Expiry.Month,
		"expiry_year":  input.Card.Expiry.Year,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error encoding data: %s", err)
		return Payment{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return Payment{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", input.IdempotencyKey)

	res, err := r.client.Do(req)
	if err != nil {
		log.Printf("Client request error: %s", err)
		return Payment{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error reading response body, %s", err)
		return Payment{}, err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		log.Printf("API error: status=%d body=%s", res.StatusCode, string(body))
		return Payment{}, fmt.Errorf("bank API error: %s", string(body))
	}

	var payment Payment
	if err := json.Unmarshal(body, &payment); err != nil {
		return Payment{}, err
	}

	return payment, nil
}
