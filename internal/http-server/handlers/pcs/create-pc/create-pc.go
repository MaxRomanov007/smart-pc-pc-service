package createPc

import (
	"context"
	"log/slog"
	"net/http"
	"smart-pc-pc-service/internal/http-server/middlewares/auth"

	"github.com/MaxRomanov007/smart-pc-go-lib/api/response"
	"github.com/MaxRomanov007/smart-pc-go-lib/domain/models"
	"github.com/MaxRomanov007/smart-pc-go-lib/logger/sl"
	"github.com/MaxRomanov007/smart-pc-go-lib/middlewares/reqmw"
	"github.com/go-chi/render"
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

		userID := auth.MustUID(r)
		req := reqmw.MustGet[Request](r)

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
