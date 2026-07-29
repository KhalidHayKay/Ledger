package main

import (
	"ledger/internal/domain/bank"
	"ledger/internal/domain/paymentintent"
	"ledger/internal/jobs/tasks"
	"ledger/internal/jobs/worker"
	"ledger/internal/platform/config"
	"ledger/internal/platform/database"
	"ledger/internal/platform/notifier"
	"log"
	"net/http"
	"time"

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

	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     config.Env.Redis.Addr,
			Password: config.Env.Redis.Password,
			DB:       1,
		},

		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 10,
			// Optionally specify multiple queues with different priority.
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			// See the godoc for other configuration options
		},
	)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	bankRepo := bank.NewFicMartBankRepo(client)
	uow := paymentintent.NewPgsqlUoW(pgsql)
	notifier := notifier.NewRedisNotifier(redis)

	taskHandler := worker.NewPaymentWorker(bankRepo, uow, notifier)

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeCreatePayment, taskHandler.HandleCreate)

	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
