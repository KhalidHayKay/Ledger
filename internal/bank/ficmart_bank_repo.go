package bank

import (
	"bytes"
	"context"
	"encoding/json"
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

func (r *FicMartBankRepo) Authorize(ctx context.Context, input AuthorizeInput) (string, error) {
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
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", input.IdempotencyKey)

	res, err := r.client.Do(req)

	if err != nil {
		log.Printf("Client request error: %s", err)
		return "", err
	}
	defer res.Body.Close()

	// if res.StatusCode != 200 {
	// 	log.Println("Server response not OK")
	// 	return "", errors.New("server response not OK")
	// }

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	return string(body), nil
}
