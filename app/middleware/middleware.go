package middleware

import "net/http"

type Middleware struct{}

func NewMiddleware() *Middleware {
	return &Middleware{}
}

func (m *Middleware) EnsureIndempotencyKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idempotencyKey := r.Header.Get("Idempotency-Key")

		if idempotencyKey == "" {
			http.Error(w, "Idempotency-Key header must be set", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}
