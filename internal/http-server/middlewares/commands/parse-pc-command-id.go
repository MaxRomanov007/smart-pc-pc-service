package commands

import (
	"log/slog"
	"net/http"
	"smart-pc-pc-service/internal/http-server/middlewares/uuidmw"

	"github.com/google/uuid"
)

const PcCommandIDURLParam = "pc_command_id"

func NewParsePcCommandIDMiddleware(log *slog.Logger) func(next http.Handler) http.Handler {
	return uuidmw.NewUUIDMiddleware(log, PcCommandIDURLParam)
}

func MustGetPcCommandID(r *http.Request) uuid.UUID {
	return uuidmw.MustFromContext(r.Context(), PcCommandIDURLParam)
}
