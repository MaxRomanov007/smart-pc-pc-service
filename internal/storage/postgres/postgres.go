package postgres

import (
	"context"
	"fmt"
	"log/slog"
	pcCommandParams "smart-pc-pc-service/internal/storage/postgres/pc-command-params"
	pcCommands "smart-pc-pc-service/internal/storage/postgres/pc-commands"
	pcLogs "smart-pc-pc-service/internal/storage/postgres/pc-logs"

	"smart-pc-pc-service/internal/config"
	"smart-pc-pc-service/internal/storage/postgres/dbqueries"
	"smart-pc-pc-service/internal/storage/postgres/pcs"

	"github.com/jackc/pgx/v5"
)

type Storage struct {
	Pcs                 *pcs.Storage
	PcLogs              *pcLogs.Storage
	PcCommands          *pcCommands.Storage
	PcCommandParameters *pcCommandParams.Storage
}

func New(ctx context.Context, log *slog.Logger, cfg config.Config) (*Storage, error) {
	const op = "storage.postgres.New"

	conn, err := pgx.Connect(ctx, cfg.Storage.Postgres.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("%s: failed to connect to postgres storage: %w", op, err)
	}

	queries := dbqueries.New(conn)

	return &Storage{
		Pcs:                 pcs.New(queries, cfg.Slug),
		PcLogs:              pcLogs.New(ctx, log, queries, cfg.Batch),
		PcCommands:          pcCommands.New(queries),
		PcCommandParameters: pcCommandParams.New(queries),
	}, nil
}

func (s *Storage) Done() <-chan struct{} {
	return s.PcLogs.Done()
}
