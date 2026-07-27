package main

import (
	"ledger/internal/domain/bank"
	"ledger/internal/domain/paymentintent"
	"ledger/internal/jobs/tasks"
	"ledger/internal/jobs/worker"
	"ledger/internal/platform/database"
	"log"
	"net/http"
	"time"

	"github.com/hibiken/asynq"
)

const redisAddr = "127.0.0.1:6379"

func main() {
	pgsql, err := database.InitPostgres()
	if err != nil {
		log.Fatal(err)
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
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

	taskHandler := worker.NewPaymentWorker(bankRepo, uow)

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeCreatePayment, taskHandler.HandleCreate)

	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
