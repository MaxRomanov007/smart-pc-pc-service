package postgres

import (
	"context"
	"fmt"

	"smart-pc-pc-service/internal/config"
	"smart-pc-pc-service/internal/storage/postgres/dbqueries"
	"smart-pc-pc-service/internal/storage/postgres/pcs"

	"github.com/jackc/pgx/v5"
)

type Storage struct {
	Pcs *pcs.Storage
}

func New(ctx context.Context, pgCfg config.PostgresStorage, slugCfg config.Slug) (*Storage, error) {
	const op = "storage.postgres.New"

	conn, err := pgx.Connect(ctx, pgCfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("%s: failed to connect to postgres storage: %w", op, err)
	}

	queries := dbqueries.New(conn)

	return &Storage{
		Pcs: pcs.New(queries, slugCfg),
	}, nil
}
