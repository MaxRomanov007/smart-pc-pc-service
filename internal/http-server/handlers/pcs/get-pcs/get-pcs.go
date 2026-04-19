package getPcs

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

	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type PcGetter interface {
	PcsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Pc, error)
	PcBySlug(ctx context.Context, userID uuid.UUID, slug string) (models.Pc, error)
}

func New(log *slog.Logger, getter PcGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.pcs.get-pcs"

		log := log.With(sl.Op(op), sl.ReqID(r))

		userID := auth.MustUID(r)

		if slug := r.URL.Query().Get("slug"); slug != "" {
			log.Info("got slug", slog.String("slug", slug))

			pc, err := getter.PcBySlug(r.Context(), userID, slug)
			if errors.Is(err, storage.ErrNotFound) {
				log.Warn("pc not found")
				render.JSON(w, r, response.NotFound("pc not found"))
				return
			}
			if err != nil {
				log.Error("failed to get pc by slug", sl.Err(err))
				render.JSON(w, r, response.InternalError())
				return
			}

			log.Debug("got pc by slug", slog.Any("pc", pc))
			render.JSON(w, r, response.OK(&pc))
			return
		}

		pcs, err := getter.PcsByUserID(r.Context(), userID)
		if err != nil {
			log.Error("failed to get pcs by userID", sl.Err(err))
			render.JSON(w, r, response.InternalError())
			return
		}

		log.Debug("got pcs by user id", slog.Any("pcs", pcs))
		render.JSON(w, r, response.OK(&pcs))
		return
	}
}
