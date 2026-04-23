package uuidmw

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"smart-pc-pc-service/internal/lib/api/response"
	"smart-pc-pc-service/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type ctxKey string

const keyPrefix = "uuid_"

func newKey(key string) ctxKey {
	return ctxKey(keyPrefix + key)
}

func NewUUIDMiddleware(
	log *slog.Logger,
	param string,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "middlewares.uuidmw"

			log := log.With(sl.Op(op), sl.ReqID(r), slog.String("param", param))

			raw := chi.URLParam(r, param)
			if raw == "" {
				log.Warn("param is missing")
				render.JSON(
					w,
					r,
					response.BadRequest(fmt.Sprintf("url param %s is missing", param)),
				)
				return
			}
			log.Debug("got param", slog.String("value", raw))

			id, err := uuid.Parse(raw)
			if err != nil {
				log.Warn("invalid param", sl.Err(err))
				render.JSON(
					w,
					r,
					response.BadRequest(fmt.Sprintf("url param %s is invalid", param)),
				)
				return
			}
			log.Debug("param parsed")

			ctx := context.WithValue(r.Context(), newKey(param), id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func FromContext(ctx context.Context, param string) (uuid.UUID, bool) {
	val := ctx.Value(newKey(param))
	id, ok := val.(uuid.UUID)
	return id, ok
}

func MustFromContext(ctx context.Context, param string) uuid.UUID {
	const op = "middlewares.uuidmw.MustFromContext"

	u, ok := FromContext(ctx, param)
	if !ok {
		panic(
			fmt.Errorf(
				"%s: can not get param %s from context, looks like you forgot to use middleware",
				op,
				param,
			),
		)
	}
	return u
}
