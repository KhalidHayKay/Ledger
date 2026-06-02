package main

import (
	"context"
	"ledger/app/config"
	"ledger/app/storage"
	"log"
	"time"

	"github.com/khalidhaykay/cmdforge"
)

var migrations = []cmdforge.Migration{
	{
		Name: "000001_create_payment_intents_table",
		Up: `CREATE TABLE payment_intents (
				id 				  		BIGSERIAL PRIMARY KEY,
				payment_reference 		TEXT UNIQUE,
				bank_authorization_id	TEXT UNIQUE,
				order_id          		TEXT NOT NULL,
				customer_id       		TEXT NOT NULL,
				amount 			  		BIGINT NOT NULL,
				currency          		CHAR(3) NOT NULL,
				status            		TEXT NOT NULL,
				created_at        		TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				updated_at        		TIMESTAMPTZ NOT NULL DEFAULT NOW()
			);`,
		Down: `DROP TABLE IF EXISTS payment_intents CASCADE;`,
	},
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := config.LoadEnv(); err != nil {
		log.Fatal(err)
	}

	pgsql, err := storage.InitPostgres()
	if err != nil {
		log.Fatal(err)
	}

	cli := cmdforge.New(pgsql, migrations)

	cli.Start(ctx)
}
