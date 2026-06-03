package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"ledger/app/config"
	"strconv"
)

func GeneratePaymentRef(id int64) string {
	mac := hmac.New(sha256.New, []byte(config.Env.App.Key))

	var buf [20]byte
	b := strconv.AppendInt(buf[:0], id, 10)

	mac.Write(b)

	return hex.EncodeToString(mac.Sum(nil))[:24]
}
