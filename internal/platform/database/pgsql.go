package database

import (
	"context"
	"ledger/internal/platform/config"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitPostgres() (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pg, err := pgxpool.New(ctx, config.Env.DB.URL())
	if err != nil {
		return nil, err
	}

	err = pg.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return pg, nil
}
