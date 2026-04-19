package auth

import (
	"log/slog"
	"net/http"

	"smart-pc-pc-service/internal/lib/api/response"
	"smart-pc-pc-service/internal/lib/logger/sl"

	"github.com/go-chi/render"
)

func NewStrictUIDMiddleware(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "middlewares.strict-uid"

			log := log.With(sl.Op(op), sl.ReqID(r))

			uid, _ := GetUserInfo(r)
			requestedUID := MustUID(r).String()

			if uid != requestedUID {
				log.Warn(
					"token uid not match requested uid",
					slog.String("requested_uid", requestedUID),
					slog.String("token_uid", uid),
				)
				render.JSON(w, r, response.Forbidden("this resource is not allowed"))
				return
			}

			log.Debug("user id verified")

			next.ServeHTTP(w, r)
		})
	}
}
