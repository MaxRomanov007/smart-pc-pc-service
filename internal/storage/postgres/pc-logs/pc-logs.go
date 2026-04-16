package pcLogs

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
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
		return fmt.Errorf("%s: failed to send data to batcher: %w", op, err)
	}

	return nil
}

func (s *Storage) ListPCLogsAfterCursor(
	ctx context.Context,
	userID, pcID, cursor uuid.UUID,
	order string,
	limit int32,
) ([]models.PcLog, error) {
	rows, err := s.queries.ListPCLogsAfterCursor(ctx, dbqueries.ListPCLogsAfterCursorParams{
		UserID: userID, PcID: pcID, Cursor: cursor, Order: order, Limit: limit,
	})
	if err != nil {
		return nil, err
	}
	return mapAfterLogs(rows), nil
}

func (s *Storage) ListPCLogsBeforeCursor(
	ctx context.Context,
	userID, pcID, cursor uuid.UUID,
	order string,
	limit int32,
) ([]models.PcLog, error) {
	rows, err := s.queries.ListPCLogsBeforeCursor(ctx, dbqueries.ListPCLogsBeforeCursorParams{
		UserID: userID, PcID: pcID, Cursor: cursor, Order: order, Limit: limit,
	})
	if err != nil {
		return nil, err
	}
	logs := mapBeforeLogs(rows)
	slices.Reverse(logs)
	return logs, nil
}

func (s *Storage) ListPCLogsFirstPage(
	ctx context.Context,
	userID, pcID uuid.UUID,
	order string,
	limit int32,
) ([]models.PcLog, error) {
	rows, err := s.queries.ListPCLogsFirstPage(ctx, dbqueries.ListPCLogsFirstPageParams{
		UserID: userID, PcID: pcID, Order: order, Limit: limit,
	})
	if err != nil {
		return nil, err
	}
	return mapFirstPageLogs(rows), nil
}

func (s *Storage) CountPCLogs(ctx context.Context, userID, pcID uuid.UUID) (int64, error) {
	return s.queries.CountPCLogs(ctx, dbqueries.CountPCLogsParams{UserID: userID, PcID: pcID})
}

func mapFirstPageLogs(rows []dbqueries.ListPCLogsFirstPageRow) []models.PcLog {
	logs := make([]models.PcLog, len(rows))
	for i, r := range rows {
		logs[i] = models.PcLog{
			ID:          r.ID,
			CommandID:   r.CommandID,
			ReceivedAt:  r.ReceivedAt,
			CompletedAt: r.CompletedAt,
			Status:      r.Status,
			Error:       *r.Error,
		}
		if r.CommandName != nil {
			logs[i].Command = &models.Command{Name: *r.CommandName}
		}
	}
	return logs
}

func mapAfterLogs(rows []dbqueries.ListPCLogsAfterCursorRow) []models.PcLog {
	logs := make([]models.PcLog, len(rows))
	for i, r := range rows {
		logs[i] = models.PcLog{
			ID:          r.ID,
			CommandID:   r.CommandID,
			ReceivedAt:  r.ReceivedAt,
			CompletedAt: r.CompletedAt,
			Status:      r.Status,
			Error:       *r.Error,
		}
		if r.CommandName != nil {
			logs[i].Command = &models.Command{
				Name: *r.CommandName,
			}
		}
	}
	return logs
}

func mapBeforeLogs(rows []dbqueries.ListPCLogsBeforeCursorRow) []models.PcLog {
	logs := make([]models.PcLog, len(rows))
	for i, r := range rows {
		logs[i] = models.PcLog{
			ID:          r.ID,
			CommandID:   r.CommandID,
			ReceivedAt:  r.ReceivedAt,
			CompletedAt: r.CompletedAt,
			Status:      r.Status,
			Error:       *r.Error,
		}
		if r.CommandName != nil {
			logs[i].Command = &models.Command{
				Name: *r.CommandName,
			}
		}
	}
	return logs
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
			pcIDs[i] = *l.PcID
			cmdIDs[i] = l.CommandID
			recvAts[i] = l.ReceivedAt
			compAts[i] = l.CompletedAt
			statuses[i] = l.Status
			errors[i] = l.Error
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
