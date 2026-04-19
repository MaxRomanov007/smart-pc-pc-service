package createPc

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"smart-pc-pc-service/internal/domain/models"
	"smart-pc-pc-service/internal/http-server/middlewares/auth"
	"smart-pc-pc-service/internal/lib/api/response"
	"smart-pc-pc-service/internal/lib/logger/sl"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Name        string `json:"name,omitempty"        validate:"omitempty,max=255"`
	Description string `json:"description,omitempty" validate:"omitempty,max=1024"`
}

type PcCreator interface {
	CreatePc(ctx context.Context, pc models.Pc) (models.Pc, error)
}

func New(log *slog.Logger, creator PcCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.pcs.create-pc"
		log := log.With(sl.Op(op), sl.ReqID(r))

		userID := auth.MustGetUID(r)

		var req Request
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, response.InternalError())
			return
		}

		if err := validator.New().Struct(req); err != nil {
			if err, ok := errors.AsType[validator.ValidationErrors](err); ok {
				log.Warn("invalid request body", sl.Err(err))
				render.JSON(w, r, response.ValidationError(err))
				return
			}

			log.Error("failed to validate request", sl.Err(err))
			render.JSON(w, r, response.InternalError())
			return
		}

		created, err := creator.CreatePc(r.Context(), models.Pc{
			UserID:      userID,
			Name:        req.Name,
			Description: req.Description,
		})
		if err != nil {
			log.Error("failed to create pc", sl.Err(err))
			render.JSON(w, r, response.InternalError())
			return
		}

		log.Debug("pc created", slog.Any("pc", created))
		render.JSON(w, r, response.OK(&created))
		return
	}
}
