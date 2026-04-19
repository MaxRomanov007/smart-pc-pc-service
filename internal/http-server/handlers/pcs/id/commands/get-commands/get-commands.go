package getCommands

import (
	"context"
	"log/slog"
	"net/http"
	"smart-pc-pc-service/internal/http-server/middlewares/uuidmw/pcs"

	"smart-pc-pc-service/internal/domain/models"
	"smart-pc-pc-service/internal/http-server/middlewares/auth"
	"smart-pc-pc-service/internal/lib/api/response"
	"smart-pc-pc-service/internal/lib/logger/sl"

	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type CommandsGetter interface {
	ListPCCommands(ctx context.Context, userID, pcID uuid.UUID) ([]models.Command, error)
}

func New(log *slog.Logger, getter CommandsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.pcs.id.commands.get-commands"
		log := log.With(sl.Op(op), sl.ReqID(r))

		userID := auth.MustUID(r)
		pcID := pcs.MustPcID(r)

		cmds, err := getter.ListPCCommands(r.Context(), userID, pcID)
		if err != nil {
			log.Error("failed to get commands", sl.Err(err))
			render.JSON(w, r, response.InternalError())
			return
		}

		log.Debug("got commands", slog.Any("commands", cmds))
		render.JSON(w, r, response.OK(&cmds))
		return
	}
}
