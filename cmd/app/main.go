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
	"ledger/internal/platform/notifier"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/hibiken/asynq"
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

	appMiddleware := middleware.NewMiddleware()

	ficmartBankRepo := bank.NewFicMartBankRepo(client)

	idempotencyRepo := idempotency.NewRedisRepo(redis)
	idempotencyService := idempotency.NewService(idempotencyRepo)

	paymentEventRepo := paymentevent.NewPostgresRepo(pgsql)

	uow := paymentintent.NewPgsqlUoW(pgsql)

	notifier := notifier.NewRedisNotifier(redis)

	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     config.Env.Redis.Addr,
		Password: config.Env.Redis.Password,
		DB:       1,
	})

	queueClient := queue.NewClient(asynqClient)

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

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))

	})

	router.Get("/payment/intent", paymentIntentHandler.Get)

	router.Group(func(r chi.Router) {
		r.Use(appMiddleware.EnsureIndempotencyKey)

		r.Post("/payment/intent", paymentIntentHandler.Create)
		r.Post("/payment/intent/capture", paymentIntentHandler.Capture)
		r.Post("/payment/intent/refund", paymentIntentHandler.Refund)
		r.Post("/payment/intent/cancel", paymentIntentHandler.Cancel)
	})

	s := &http.Server{
		Addr:           ":" + config.Env.App.Port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   40 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}
