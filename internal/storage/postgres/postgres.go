package postgres

import (
	"context"
	"fmt"

	"smart-pc-pc-service/internal/config"
	"smart-pc-pc-service/internal/storage/postgres/dbqueries"

	"github.com/jackc/pgx/v5"
)

type Storage struct {
	db      *pgx.Conn
	queries *dbqueries.Queries
}

func New(ctx context.Context, cfg config.PostgresStorage) (*Storage, error) {
	const op = "storage.postgres.New"

	conn, err := pgx.Connect(ctx, cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("%s: failed to connect to postgres storage: %w", op, err)
	}

	return &Storage{
		db:      conn,
		queries: dbqueries.New(conn),
	}, nil
}
