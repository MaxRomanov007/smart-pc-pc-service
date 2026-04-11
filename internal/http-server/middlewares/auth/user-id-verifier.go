package auth

import (
	"log/slog"
	"net/http"

	"smart-pc-pc-service/internal/lib/api/response"
	"smart-pc-pc-service/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

func NewUserIdVerifierMiddleware(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "middlewares.user-id-verifier"

			log := log.With(sl.Op(op), sl.ReqID(r))

			userID, _ := GetUserInfo(r)

			if _, err := uuid.Parse(userID); err != nil {
				log.Warn("cannot parse user id to uuid", sl.Err(err))
				render.JSON(w, r, response.BadRequest("invalid user id"))
				return
			}

			requestedUID := chi.URLParam(r, "uid")
			if userID != requestedUID {
				log.Warn(
					"token uid not match requested uid",
					slog.String("requested_uid", requestedUID),
					slog.String("token_uid", userID),
				)
				render.JSON(w, r, response.Forbidden("this resource is not allowed"))
				return
			}

			log.Debug("user id verified")

			next.ServeHTTP(w, r)
		})
	}
}
