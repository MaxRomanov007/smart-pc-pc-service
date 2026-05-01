package commands

import (
	"log/slog"
	"net/http"

	"github.com/MaxRomanov007/smart-pc-go-lib/middlewares/uuidmw"
	"github.com/google/uuid"
)

const PcCommandIDURLParam = "pc_command_id"

func NewMiddleware(log *slog.Logger) func(next http.Handler) http.Handler {
	return uuidmw.New(log, PcCommandIDURLParam)
}

func MustPcCommandID(r *http.Request) uuid.UUID {
	return uuidmw.MustFromContext(r.Context(), PcCommandIDURLParam)
}
