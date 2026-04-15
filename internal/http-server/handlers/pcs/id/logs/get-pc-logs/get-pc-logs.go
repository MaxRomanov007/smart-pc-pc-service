package getPcLogs

import (
	"context"
	"log/slog"
	"net/http"
	"smart-pc-pc-service/internal/domain/models"
	"smart-pc-pc-service/internal/http-server/middlewares/auth"
	"smart-pc-pc-service/internal/http-server/middlewares/pcs"
	"smart-pc-pc-service/internal/lib/api/response"
	"smart-pc-pc-service/internal/lib/api/response/pagination"
	"smart-pc-pc-service/internal/lib/logger/sl"
	"strconv"

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

		userID := auth.GetUserUUID(r)
		pcID := pcs.GetPcUUID(r)

		order := r.URL.Query().Get("order")
		if order == "" {
			order = "asc"
		}
		if order != "asc" && order != "desc" {
			log.Warn("order does not match \"asc\" or \"desc\"")
			render.JSON(w, r, response.BadRequest("order does not match \"asc\" or \"desc\""))
			return
		}

		limitStr := r.URL.Query().Get("limit")
		limit := int32(20)
		if limitStr != "" {
			l, err := strconv.ParseInt(limitStr, 10, 32)
			if err != nil || l <= 0 {
				log.Warn("failed to parse limit", sl.Err(err))
				render.JSON(w, r, response.BadRequest("limit must be non negative integer number"))
				return
			}
			limit = int32(l)
		}

		before := r.URL.Query().Get("before")
		after := r.URL.Query().Get("after")
		if before != "" && after != "" {
			log.Warn("before and after query parameters provided in one query")
			render.JSON(
				w,
				r,
				response.BadRequest(
					"request can not contain \"before\" and \"after\" in one query",
				),
			)
			return
		}

		queryLimit := limit + 1

		var logs []models.PcLog
		var cursorID uuid.UUID
		var err error

		switch {
		case after != "":
			cursorID, err = uuid.Parse(after)
			if err != nil {
				log.Warn("failed to parse after cursor", sl.Err(err))
				render.JSON(w, r, response.BadRequest("invalid 'after' cursor"))
				return
			}
			logs, err = getter.ListPCLogsAfterCursor(
				r.Context(),
				userID,
				pcID,
				cursorID,
				order,
				queryLimit,
			)

		case before != "":
			cursorID, err = uuid.Parse(before)
			if err != nil {
				log.Warn("failed to parse before cursor", sl.Err(err))
				render.JSON(w, r, response.BadRequest("invalid 'before' cursor"))
				return
			}
			logs, err = getter.ListPCLogsBeforeCursor(
				r.Context(),
				userID,
				pcID,
				cursorID,
				order,
				queryLimit,
			)

		default:
			logs, err = getter.ListPCLogsFirstPage(r.Context(), userID, pcID, order, queryLimit)
		}

		if err != nil {
			log.Error("failed to fetch logs", sl.Err(err))
			render.JSON(w, r, response.InternalError())
			return
		}

		hasExtra := len(logs) > int(limit)
		if hasExtra {
			logs = logs[:limit]
		}

		var hasPrev, hasNext bool
		switch {
		case after != "":
			hasPrev = true
			hasNext = hasExtra
		case before != "":
			hasPrev = hasExtra
			hasNext = true
		default:
			hasPrev = false
			hasNext = hasExtra
		}

		total, err := getter.CountPCLogs(r.Context(), userID, pcID)
		if err != nil {
			log.Error("failed to count logs", sl.Err(err))
			render.JSON(w, r, response.InternalError())
			return
		}

		pag := &pagination.Pagination{
			Total:   total,
			HasPrev: hasPrev,
			HasNext: hasNext,
		}

		if len(logs) > 0 {
			first := logs[0].ID.String()
			last := logs[len(logs)-1].ID.String()

			if hasPrev {
				pag.PrevCursor = first
			}
			if hasNext {
				pag.NextCursor = last
			}
		}

		render.JSON(w, r, response.OK(&logs).WithPagination(pag))
	}
}
