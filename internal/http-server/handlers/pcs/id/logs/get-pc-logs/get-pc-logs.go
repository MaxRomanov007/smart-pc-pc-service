package getPcLogs

import (
	"context"
	"log/slog"
	"net/http"
	"smart-pc-pc-service/internal/domain/models"
	"smart-pc-pc-service/internal/http-server/middlewares/auth"
	"smart-pc-pc-service/internal/http-server/middlewares/pcs"
	"smart-pc-pc-service/internal/lib/api/response"
	"smart-pc-pc-service/internal/lib/logger/sl"

	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type PcLogsGetter interface {
	ListPCLogsFirstPage(
		ctx context.Context,
		userID, pcID uuid.UUID,
		order string,
		limit int32,
	) ([]models.PcLog, error)
	ListPCLogsAfterCursor(
		ctx context.Context,
		userID, pcID, cursor uuid.UUID,
		order string,
		limit int32,
	) ([]models.PcLog, error)
	ListPCLogsBeforeCursor(
		ctx context.Context,
		userID, pcID, cursor uuid.UUID,
		order string,
		limit int32,
	) ([]models.PcLog, error)
	CountPCLogs(ctx context.Context, userID, pcID uuid.UUID) (int64, error)
}

func New(log *slog.Logger, getter PcLogsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.pcs.id.logs.get-pc-logs"
		log := log.With(sl.Op(op), sl.ReqID(r))

		userID := auth.MustGetUID(r)
		pcID := pcs.MustGetPcID(r)

		params, err := parseLogsQueryParams(r)
		if err != nil {
			log.Warn("invalid query parameters", sl.Err(err))
			render.JSON(w, r, response.BadRequest("invalid query parameters"))
			return
		}

		logs, err := fetchLogs(r.Context(), getter, params, userID, pcID)
		if err != nil {
			log.Error("failed to fetch logs", sl.Err(err))
			render.JSON(w, r, response.InternalError())
			return
		}

		total, err := getter.CountPCLogs(r.Context(), userID, pcID)
		if err != nil {
			log.Error("failed to count logs", sl.Err(err))
			render.JSON(w, r, response.InternalError())
			return
		}

		result := buildPagination(logs, params, total)
		render.JSON(w, r, response.OK(&result.Logs).WithPagination(result.Pag))
	}
}

func fetchLogs(
	ctx context.Context,
	getter PcLogsGetter,
	params *LogsQueryParams,
	userID, pcID uuid.UUID,
) ([]models.PcLog, error) {
	queryLimit := params.Limit + 1
	switch params.Type {
	case "after":
		return getter.ListPCLogsAfterCursor(
			ctx,
			userID,
			pcID,
			params.Cursor,
			params.Order,
			queryLimit,
		)
	case "before":
		return getter.ListPCLogsBeforeCursor(
			ctx,
			userID,
			pcID,
			params.Cursor,
			params.Order,
			queryLimit,
		)
	default:
		return getter.ListPCLogsFirstPage(ctx, userID, pcID, params.Order, queryLimit)
	}
}
