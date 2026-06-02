package main

import (
	"ledger/app/config"
	"ledger/app/middleware"
	"ledger/app/storage"
	"ledger/internal/bank"
	"ledger/internal/idempotency"
	"ledger/internal/paymentintent"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	if err := config.LoadEnv(); err != nil {
		log.Fatal(err)
	}

	pgsql, err := storage.InitPostgres()
	if err != nil {
		log.Fatal(err)
	}

	redis, err := storage.InitRedis()
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	ficmartBankRepo := bank.NewFicMartBankRepo(client)

	idempotencyRepo := idempotency.NewRedisRepo(redis)
	idempotencyService := idempotency.NewService(idempotencyRepo)

	paymentIntentRepo := paymentintent.NewPostgresRepo(pgsql)
	paymentIntentService := paymentintent.NewService(paymentIntentRepo, ficmartBankRepo, idempotencyService)
	paymentIntentHandler := paymentintent.NewHandler(paymentIntentService)

	router := chi.NewRouter()

	router.Use(chiMiddleware.Logger, chiMiddleware.Recoverer)

	appMiddleware := middleware.NewMiddleware()
	router.Use(appMiddleware.EnsureIndempotencyKey)

	router.Post("/payment/intent", paymentIntentHandler.Create)

	s := &http.Server{
		Addr:           ":" + config.Env.App.Port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}
