package pcLogs

import (
	"context"
	"fmt"
	"log/slog"
	"smart-pc-pc-service/internal/config"
	"smart-pc-pc-service/internal/domain/models"
	"smart-pc-pc-service/internal/lib/batcher"
	"smart-pc-pc-service/internal/lib/logger/sl"
	"smart-pc-pc-service/internal/storage/postgres/dbqueries"
	"time"

	"github.com/google/uuid"
)

type Storage struct {
	queries           *dbqueries.Queries
	createLogsBatcher *batcher.Batcher[models.PcLog]
}

func New(
	ctx context.Context,
	log *slog.Logger,
	queries *dbqueries.Queries,
	batchCfg config.Batch,
) *Storage {
	createLogsBatcher := batcher.New(ctx, &batcher.Options[models.PcLog]{
		Flush:   NewFlushCreatePcLogsFunc(ctx, queries),
		MaxSize: batchCfg.MaxSize,
		Timeout: batchCfg.Timeout,
		OnFlushError: func(err error) {
			const op = "storage.postgres.pc-logs.FlushCreatePcLogsOnError"
			log.Error("failed to flush create logs batch", sl.Op(op), sl.Err(err))
		},
	})

	return &Storage{queries: queries, createLogsBatcher: createLogsBatcher}
}

func (s *Storage) Done() <-chan struct{} {
	return s.createLogsBatcher.Done()
}

func (s *Storage) CreatePcLog(ctx context.Context, log models.PcLog) error {
	const op = "storage.postgres.pc-logs.CreatePcLog"

	if err := s.createLogsBatcher.Send(ctx, log); err != nil {
		return fmt.Errorf("%s: failed to send command to batcher: %w", op, err)
	}

	return nil
}

func NewFlushCreatePcLogsFunc(
	ctx context.Context,
	queries *dbqueries.Queries,
) batcher.FlushFunc[models.PcLog] {
	return func(logs []models.PcLog) error {
		const op = "storage.postgres.pc-logs.FlushCreatePcLogs"

		if len(logs) == 0 {
			return nil
		}

		n := len(logs)
		pcIDs := make([]uuid.UUID, n)
		cmdIDs := make([]string, n)
		recvAts := make([]time.Time, n)
		compAts := make([]time.Time, n)
		statuses := make([]string, n)
		errors := make([]string, n)

		for i, l := range logs {
			pcIDs[i] = l.PcID
			cmdIDs[i] = l.CommandID
			recvAts[i] = l.ReceivedAt
			compAts[i] = l.CompletedAt
			statuses[i] = l.Status
			if l.Error != nil {
				errors[i] = *l.Error
			}
		}

		err := queries.BatchInsertPCLogs(ctx, dbqueries.BatchInsertPCLogsParams{
			PcIds:        pcIDs,
			CommandIds:   cmdIDs,
			ReceivedAts:  recvAts,
			CompletedAts: compAts,
			Statuses:     statuses,
			Errors:       errors,
		})
		if err != nil {
			return fmt.Errorf("%s: failed to insert pc logs: %w", op, err)
		}
		return nil
	}
}
