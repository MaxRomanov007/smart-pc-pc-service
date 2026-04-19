package pcs

import (
	"log/slog"
	"net/http"
	"smart-pc-pc-service/internal/http-server/middlewares/uuidmw"

	"github.com/google/uuid"
)

const PcIDURLParam = "pc_id"

func NewParsePcIDMiddleware(log *slog.Logger) func(next http.Handler) http.Handler {
	return uuidmw.NewUUIDMiddleware(log, PcIDURLParam)
}

func MustPcID(r *http.Request) uuid.UUID {
	return uuidmw.MustFromContext(r.Context(), PcIDURLParam)
}
