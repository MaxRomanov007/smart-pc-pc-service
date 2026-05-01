package getParameters

import (
	"context"
	"log/slog"
	"net/http"
	"smart-pc-pc-service/internal/http-server/middlewares/auth"
	"smart-pc-pc-service/internal/http-server/middlewares/uuidmw/commands"
	"smart-pc-pc-service/internal/http-server/middlewares/uuidmw/pcs"

	"github.com/MaxRomanov007/smart-pc-go-lib/api/response"
	"github.com/MaxRomanov007/smart-pc-go-lib/domain/models"
	"github.com/MaxRomanov007/smart-pc-go-lib/logger/sl"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type ParamsGetter interface {
	ListCommandParameters(
		ctx context.Context,
		userID, pcID, commandID uuid.UUID,
	) ([]models.CommandParameter, error)
}

func New(log *slog.Logger, getter ParamsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.pcs.id.commands.id.get-parameters"
		log := log.With(sl.Op(op), sl.ReqID(r))

		userID := auth.MustUID(r)
		pcID := pcs.MustPcID(r)
		commandID := commands.MustPcCommandID(r)

		params, err := getter.ListCommandParameters(r.Context(), userID, pcID, commandID)
		if err != nil {
			log.Error("failed to get parameters", sl.Err(err))
			render.JSON(w, r, response.InternalError())
			return
		}

		log.Debug("got parameters", slog.Any("parameters", params))
		render.JSON(w, r, response.OK(&params))
		return
	}
}
