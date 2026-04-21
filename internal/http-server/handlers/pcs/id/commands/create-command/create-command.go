package createCommand

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"smart-pc-pc-service/internal/domain/models"
	"smart-pc-pc-service/internal/http-server/middlewares/auth"
	"smart-pc-pc-service/internal/http-server/middlewares/request"
	"smart-pc-pc-service/internal/http-server/middlewares/uuidmw/pcs"
	"smart-pc-pc-service/internal/lib/api/response"
	"smart-pc-pc-service/internal/lib/logger/sl"
	"smart-pc-pc-service/internal/storage"

	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type RequestParameter struct {
	Name        string `json:"name"                  validate:"required,max=255"`
	Description string `json:"description,omitempty" validate:"omitempty,max=1024"`
	Type        int16  `json:"type"                  validate:"required,min=1,max=3"`
}

type Request struct {
	Name        string             `json:"name"                 validate:"omitempty,max=255"`
	Description string             `json:"description"          validate:"omitempty,max=1024"`
	Script      string             `json:"script"               validate:"omitempty,max=8192"`
	Parameters  []RequestParameter `json:"parameters,omitempty" validate:"omitempty,max=10,unique=Name,dive"`
}

type CommandCreator interface {
	CreateUserPcCommand(
		ctx context.Context,
		uid uuid.UUID,
		command models.Command,
	) (models.Command, error)
}

func New(
	log *slog.Logger,
	creator CommandCreator,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.pcs.id.commands.create-command"
		log := log.With(sl.Op(op), sl.ReqID(r))

		userID := auth.MustUID(r)
		pcID := pcs.MustPcID(r)
		req := request.MustGet[Request](r)

		parameters := make([]models.CommandParameter, len(req.Parameters))
		for i, p := range req.Parameters {
			parameters[i] = models.CommandParameter{
				Name:        p.Name,
				Description: p.Description,
				Type:        p.Type,
			}
		}

		command, err := creator.CreateUserPcCommand(r.Context(), userID, models.Command{
			Name:        req.Name,
			Description: req.Description,
			Parameters:  parameters,
			PcID:        pcID,
		})
		if errors.Is(err, storage.ErrNotFound) {
			log.Warn("pc for command not found")
			render.JSON(w, r, response.NotFound("pc for command not found"))
			return
		}
		if err != nil {
			log.Error("failed to create command", sl.Err(err))
			render.JSON(w, r, response.InternalError())
			return
		}

		log.Debug("created command", slog.Any("command", command))
		render.JSON(w, r, command)
		return
	}
}
