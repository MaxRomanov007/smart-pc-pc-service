package update_pc

import (
	"context"
	"log/slog"
	"net/http"

	"smart-pc-pc-service/internal/domain/models"
	"smart-pc-pc-service/internal/http-server/middlewares/auth"
	"smart-pc-pc-service/internal/lib/api/response"
	"smart-pc-pc-service/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type PcUpdater interface {
	UpdatePc(
		ctx context.Context,
		userID, pcID string,
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
		const op = "http-server.handlers.pcs.update-pc"

		log := log.With(sl.Op(op), sl.ReqId(r))

		userID, _ := auth.GetUserInfo(r)
		pcID := chi.URLParam(r, "pcID")

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
