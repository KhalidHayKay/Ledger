package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"ledger/app/config"
)

func GeneratePaymentRef(id string) string {
	mac := hmac.New(sha256.New, []byte(config.Env.App.Key))
	mac.Write([]byte(fmt.Sprintf("%v", id)))

	return hex.EncodeToString(mac.Sum(nil))[:24]
}
