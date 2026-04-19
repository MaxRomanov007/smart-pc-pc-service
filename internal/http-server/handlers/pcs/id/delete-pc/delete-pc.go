package deletePc

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"smart-pc-pc-service/internal/domain/models"
	"smart-pc-pc-service/internal/http-server/middlewares/auth"
	"smart-pc-pc-service/internal/http-server/middlewares/pcs"
	"smart-pc-pc-service/internal/lib/api/response"
	"smart-pc-pc-service/internal/lib/logger/sl"
	"smart-pc-pc-service/internal/storage"

	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type PcDeleter interface {
	DeleteUserPc(ctx context.Context, uid, id uuid.UUID) (models.Pc, error)
}

func New(log *slog.Logger, deleter PcDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.pcs.id.delete-pc"

		log := log.With(sl.Op(op), sl.ReqID(r))

		userID := auth.MustGetUID(r)
		pcID := pcs.MustGetPcID(r)

		deleted, err := deleter.DeleteUserPc(r.Context(), userID, pcID)
		if errors.Is(err, storage.ErrNotFound) {
			log.Warn("pc not found")
			render.JSON(w, r, response.NotFound("pc not found"))
			return
		}
		if err != nil {
			log.Error(op, "failed to delete pc")
			render.JSON(w, r, response.InternalError())
			return
		}

		log.Debug("pc deleted", slog.Any("pc", deleted))
		render.JSON(w, r, response.OK(&deleted))
		return
	}
}
