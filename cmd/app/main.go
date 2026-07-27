package main

import (
	"ledger/internal/domain/bank"
	"ledger/internal/domain/idempotency"
	"ledger/internal/domain/paymentevent"
	"ledger/internal/domain/paymentintent"
	"ledger/internal/jobs/queue"
	"ledger/internal/platform/config"
	"ledger/internal/platform/database"
	"ledger/internal/platform/middleware"
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

	pgsql, err := database.InitPostgres()
	if err != nil {
		log.Fatal(err)
	}

	redis, err := database.InitRedis()
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	ficmartBankRepo := bank.NewFicMartBankRepo(client)

	idempotencyRepo := idempotency.NewRedisRepo(redis)
	idempotencyService := idempotency.NewService(idempotencyRepo)

	paymentEventRepo := paymentevent.NewPostgresRepo(pgsql)

	uow := paymentintent.NewPgsqlUoW(pgsql)

	queueClient, err := queue.NewClient(config.Env.Redis.Addr)
	if err != nil {
		log.Fatal(err)
	}

	notifier := paymentintent.NewRedisNotifier(redis)

	paymentIntentRepo := paymentintent.NewPostgresRepo(pgsql)
	paymentIntentService := paymentintent.NewService(
		paymentIntentRepo,
		paymentEventRepo,
		ficmartBankRepo,
		uow,
		idempotencyService,
		queueClient,
		notifier,
	)
	paymentIntentHandler := paymentintent.NewHandler(paymentIntentService)

	router := chi.NewRouter()

	router.Use(chiMiddleware.Logger, chiMiddleware.Recoverer)

	appMiddleware := middleware.NewMiddleware()
	router.Use(appMiddleware.EnsureIndempotencyKey)

	router.Post("/payment/intent", paymentIntentHandler.Create)
	router.Post("/payment/intent/capture", paymentIntentHandler.Capture)
	router.Post("/payment/intent/refund", paymentIntentHandler.Refund)
	router.Post("/payment/intent/cancel", paymentIntentHandler.Cancel)

	s := &http.Server{
		Addr:           ":" + config.Env.App.Port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}
