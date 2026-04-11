package updatePc

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

type PcUpdater interface {
	UpdatePc(
		ctx context.Context,
		userID, pcID uuid.UUID,
		name, description *string,
		canPowerOn *bool,
	) (models.Pc, error)
}

type Request struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	CanPowerOn  *bool   `json:"canPowerOn"`
}

func New(log *slog.Logger, updater PcUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.pcs.id.update-pc"

		log := log.With(sl.Op(op), sl.ReqID(r))

		userID := auth.GetUserUUID(r)

		pcID, err := uuid.Parse(chi.URLParam(r, "pc_id"))
		if err != nil {
			log.Error("invalid pc id", sl.Err(err))
			render.JSON(w, r, response.BadRequest("invalid pc id"))
			return
		}

		log.Debug("got pc id", slog.String("pc_id", pcID.String()))

		var req Request
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, response.BadRequest("failed to decode request"))
			return
		}

		pc, err := updater.UpdatePc(
			r.Context(),
			userID,
			pcID,
			req.Name,
			req.Description,
			req.CanPowerOn,
		)
		if errors.Is(err, storage.ErrNotFound) {
			log.Warn("pc not found")
			render.JSON(w, r, response.NotFound("pc not found"))
			return
		}
		if err != nil {
			log.Error("failed to update pc", sl.Err(err))
			render.JSON(w, r, response.InternalError())
			return
		}

		log.Debug("pc updated successfully", slog.Any("pc", pc))
		render.JSON(w, r, response.OK(&pc))
		return
	}
}
