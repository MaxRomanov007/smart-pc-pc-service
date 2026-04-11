package getPc

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"smart-pc-pc-service/internal/domain/models"
	"smart-pc-pc-service/internal/http-server/middlewares/auth"
	"smart-pc-pc-service/internal/lib/api/response"
	"smart-pc-pc-service/internal/lib/logger/sl"
	"smart-pc-pc-service/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type PcGetter interface {
	PcByID(ctx context.Context, userID, pcID uuid.UUID) (models.Pc, error)
}

func New(log *slog.Logger, getter PcGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.pcs.id.get-pc"

		log := log.With(sl.Op(op), sl.ReqID(r))

		userID := auth.GetUserUUID(r)

		pcID, err := uuid.Parse(chi.URLParam(r, "pc_id"))
		if err != nil {
			log.Error("invalid pc id", sl.Err(err))
			render.JSON(w, r, response.BadRequest("invalid pc id"))
			return
		}

		log.Debug("got pc id", slog.String("pc_id", pcID.String()))

		pc, err := getter.PcByID(r.Context(), userID, pcID)
		if errors.Is(err, storage.ErrNotFound) {
			log.Warn("pc not found")
			render.JSON(w, r, response.NotFound("pc not found"))
			return
		}
		if err != nil {
			log.Error("failed to get pc", sl.Err(err))
			render.JSON(w, r, response.InternalError())
			return
		}

		log.Debug("got pc", slog.Any("pc", pc))
		render.JSON(w, r, response.OK(&pc))
		return
	}
}
