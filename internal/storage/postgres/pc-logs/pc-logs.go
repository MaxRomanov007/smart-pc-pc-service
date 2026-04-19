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
		Flush:   NewFlushCreatePcLogsFunc(ctx, queries, log),
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
	rows, err := s.queries.UserPCLogsAfterCursor(ctx, dbqueries.UserPCLogsAfterCursorParams{
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
	rows, err := s.queries.UserPCLogsBeforeCursor(ctx, dbqueries.UserPCLogsBeforeCursorParams{
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
	rows, err := s.queries.UserPCLogsFirstPage(ctx, dbqueries.UserPCLogsFirstPageParams{
		UserID: userID, PcID: pcID, Order: order, Limit: limit,
	})
	if err != nil {
		return nil, err
	}
	return mapFirstPageLogs(rows), nil
}

func (s *Storage) CountPCLogs(ctx context.Context, userID, pcID uuid.UUID) (int64, error) {
	return s.queries.UserPCLogsCount(
		ctx,
		dbqueries.UserPCLogsCountParams{UserID: userID, PcID: pcID},
	)
}

func mapFirstPageLogs(rows []dbqueries.UserPCLogsFirstPageRow) []models.PcLog {
	logs := make([]models.PcLog, len(rows))
	for i, r := range rows {
		logs[i] = models.PcLog{
			ID:          r.ID,
			CommandID:   r.CommandID,
			ReceivedAt:  r.ReceivedAt,
			CompletedAt: r.CompletedAt,
			Status:      r.Status,
			Error:       r.Error,
		}
		if r.CommandName != nil {
			logs[i].Command = &models.Command{Name: *r.CommandName}
		}
	}
	return logs
}

func mapAfterLogs(rows []dbqueries.UserPCLogsAfterCursorRow) []models.PcLog {
	logs := make([]models.PcLog, len(rows))
	for i, r := range rows {
		logs[i] = models.PcLog{
			ID:          r.ID,
			CommandID:   r.CommandID,
			ReceivedAt:  r.ReceivedAt,
			CompletedAt: r.CompletedAt,
			Status:      r.Status,
			Error:       r.Error,
		}
		if r.CommandName != nil {
			logs[i].Command = &models.Command{
				Name: *r.CommandName,
			}
		}
	}
	return logs
}

func mapBeforeLogs(rows []dbqueries.UserPCLogsBeforeCursorRow) []models.PcLog {
	logs := make([]models.PcLog, len(rows))
	for i, r := range rows {
		logs[i] = models.PcLog{
			ID:          r.ID,
			CommandID:   r.CommandID,
			ReceivedAt:  r.ReceivedAt,
			CompletedAt: r.CompletedAt,
			Status:      r.Status,
			Error:       r.Error,
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
	log *slog.Logger,
) batcher.FlushFunc[models.PcLog] {
	return func(logs []models.PcLog) error {
		const op = "storage.postgres.pc-logs.FlushCreatePcLogs"

		if len(logs) == 0 {
			return nil
		}

		log.Debug("flushing logs", sl.Op(op), slog.Int("count", len(logs)))

		params := make([]dbqueries.CreatePCLogsParams, len(logs))
		for i, l := range logs {
			params[i] = dbqueries.CreatePCLogsParams{
				CommandID:   l.CommandID,
				ReceivedAt:  l.ReceivedAt,
				CompletedAt: l.CompletedAt,
				Status:      l.Status,
				Error:       l.Error,
			}
			if l.PcID != nil {
				params[i].PcID = *l.PcID
			}
		}

		_, err := queries.CreatePCLogs(ctx, params)
		if err != nil {
			return fmt.Errorf("%s: failed to insert pc logs: %w", op, err)
		}
		return nil
	}
}
