package updateCommand

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"smart-pc-pc-service/internal/domain/models"
	"smart-pc-pc-service/internal/http-server/middlewares/auth"
	"smart-pc-pc-service/internal/http-server/middlewares/request"
	"smart-pc-pc-service/internal/http-server/middlewares/uuidmw/commands"
	"smart-pc-pc-service/internal/http-server/middlewares/uuidmw/pcs"
	"smart-pc-pc-service/internal/lib/api/response"
	"smart-pc-pc-service/internal/lib/logger/sl"
	"smart-pc-pc-service/internal/storage"

	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type RequestParameter struct {
	ID          uuid.UUID `json:"id,omitempty"          validate:"omitempty,uuid"`
	Name        string    `json:"name"                  validate:"required,max=255"`
	Description string    `json:"description,omitempty" validate:"omitempty,max=1024"`
	Type        int16     `json:"type"                  validate:"required,min=1,max=3"`
}

type Request struct {
	Name        string             `json:"name"                 validate:"omitempty,max=255"`
	Description string             `json:"description"          validate:"omitempty,max=1024"`
	Script      string             `json:"script"               validate:"omitempty,max=8192"`
	Parameters  []RequestParameter `json:"parameters,omitempty" validate:"omitempty,max=10,unique=Name,dive"`
}

type CommandUpdater interface {
	UpdateUserPcCommand(
		ctx context.Context,
		uid uuid.UUID,
		command models.Command,
	) (updated models.Command, err error)
}

func New(log *slog.Logger, updater CommandUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.pcs.id.commands.id.update-command"
		log := log.With(sl.Op(op), sl.ReqID(r))

		userID := auth.MustUID(r)
		pcID := pcs.MustPcID(r)
		commandID := commands.MustPcCommandID(r)
		req := request.MustGet[Request](r)

		var parameters []models.CommandParameter
		if len(req.Parameters) > 0 {
			parameters = make([]models.CommandParameter, len(req.Parameters))
		}
		for i, p := range req.Parameters {
			parameters[i] = models.CommandParameter{
				ID:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				Type:        p.Type,
			}
		}

		updated, err := updater.UpdateUserPcCommand(r.Context(), userID, models.Command{
			ID:          commandID,
			PcID:        pcID,
			Name:        req.Name,
			Description: req.Description,
			Parameters:  parameters,
		})
		if errors.Is(err, storage.ErrNotFound) {
			log.Warn("command not found", sl.Err(err))
			render.JSON(w, r, response.NotFound("command not found"))
			return
		}
		if err != nil {
			log.Error("failed to update command", sl.Err(err))
			render.JSON(w, r, response.InternalError())
			return
		}

		log.Debug("command updated", slog.Any("updated", updated))
		render.JSON(w, r, response.OK(&updated))
		return
	}
}
