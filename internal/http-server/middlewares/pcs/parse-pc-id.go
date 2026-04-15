package pcs

import (
	"fmt"
	"log/slog"
	"net/http"
	"smart-pc-pc-service/internal/lib/api/response"
	"smart-pc-pc-service/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

func NewPcIDVerifierMiddleware(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "middlewares.pcs.pc-id-verifier"

			log := log.With(sl.Op(op), sl.ReqID(r))

			pcID, err := uuid.Parse(chi.URLParam(r, "pc_id"))
			if err != nil {
				log.Warn("invalid pc id", sl.Err(err))
				render.JSON(w, r, response.BadRequest("invalid pc id"))
				return
			}

			log.Debug("pc id parsed", slog.String("pc_id", pcID.String()))

			next.ServeHTTP(w, r)
		})
	}
}

func GetPcUUID(r *http.Request) uuid.UUID {
	const op = "middlewares.pcs.GetPcUUID"

	pcUUID, err := uuid.Parse(chi.URLParam(r, "pc_id"))
	if err != nil {
		panic(
			fmt.Errorf(
				"%s: failed to parse pc id to uuid: %w",
				op,
				err,
			),
		)
	}

	return pcUUID
}
