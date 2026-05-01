package pcs

import (
	"log/slog"
	"net/http"

	"github.com/MaxRomanov007/smart-pc-go-lib/middlewares/uuidmw"
	"github.com/google/uuid"
)

const PcIDURLParam = "pc_id"

func NewMiddleware(log *slog.Logger) func(next http.Handler) http.Handler {
	return uuidmw.New(log, PcIDURLParam)
}

func MustPcID(r *http.Request) uuid.UUID {
	return uuidmw.MustFromContext(r.Context(), PcIDURLParam)
}
