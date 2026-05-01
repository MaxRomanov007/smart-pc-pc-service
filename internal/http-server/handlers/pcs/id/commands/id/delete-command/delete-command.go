package deleteCommand

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"smart-pc-pc-service/internal/http-server/middlewares/auth"
	"smart-pc-pc-service/internal/http-server/middlewares/uuidmw/commands"
	"smart-pc-pc-service/internal/http-server/middlewares/uuidmw/pcs"
	"smart-pc-pc-service/internal/storage"

	"github.com/MaxRomanov007/smart-pc-go-lib/api/response"
	"github.com/MaxRomanov007/smart-pc-go-lib/domain/models"
	"github.com/MaxRomanov007/smart-pc-go-lib/logger/sl"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type CommandDeleter interface {
	DeleteUserPcCommand(ctx context.Context, uid, pcID, commandID uuid.UUID) (models.Command, error)
}

func New(log *slog.Logger, deleter CommandDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.pcs.id.commands.id.get-parameters"
		log := log.With(sl.Op(op), sl.ReqID(r))

		userID := auth.MustUID(r)
		pcID := pcs.MustPcID(r)
		commandID := commands.MustPcCommandID(r)

		deleted, err := deleter.DeleteUserPcCommand(r.Context(), userID, pcID, commandID)
		if errors.Is(err, storage.ErrNotFound) {
			log.Warn("command not found")
			render.JSON(w, r, response.NotFound("command not found"))
			return
		}
		if err != nil {
			log.Error("failed to delete command", sl.Err(err))
			render.JSON(w, r, response.InternalError())
			return
		}

		log.Debug("command deleted", slog.Any("command", deleted))
		render.JSON(w, r, response.OK(&deleted))
		return
	}
}
