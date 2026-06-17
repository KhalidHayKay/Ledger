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
		Up: `
			CREATE TABLE payment_intents (
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
			);
		`,
		Down: `DROP TABLE IF EXISTS payment_intents CASCADE;`,
	},
	{
		Name: "000002_create_payment_process_table",
		Up: `
			CREATE TABLE payment_events (
				id                	BIGSERIAL PRIMARY KEY,
				payment_intent_id 	BIGINT NOT NULL,
				state             	TEXT NOT NULL,
				external_state_id   TEXT NOT NULL,
				metadata          	BYTEA,
				created_at        	TIMESTAMPTZ NOT NULL DEFAULT NOW(),

				CONSTRAINT fk_payment_intents
					FOREIGN KEY (payment_intent_id)
					REFERENCES payment_intents(id)
					ON DELETE CASCADE
			);

			CREATE INDEX idx_payment_events_intent_id
			ON payment_events(payment_intent_id);

			CREATE INDEX idx_payment_events_external_state_id
			ON payment_events(external_state_id);
		`,
		Down: `
			DROP INDEX IF EXISTS idx_payment_events_external_state_id;
			DROP INDEX IF EXISTS idx_payment_events_intent_id;
			
			DROP TABLE IF EXISTS payment_events;
		`,
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
