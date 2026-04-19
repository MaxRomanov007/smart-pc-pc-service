package auth

import (
	"log/slog"
	"net/http"
	"smart-pc-pc-service/internal/http-server/middlewares/uuidmw"

	"github.com/google/uuid"
)

const UIDURLParam = "uid"

func NewParseUIDMiddleware(log *slog.Logger) func(next http.Handler) http.Handler {
	return uuidmw.NewUUIDMiddleware(log, UIDURLParam)
}

func MustUID(r *http.Request) uuid.UUID {
	return uuidmw.MustFromContext(r.Context(), UIDURLParam)
}
