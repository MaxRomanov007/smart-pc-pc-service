package pcLogs

import (
	"context"
	"fmt"
	"log/slog"
	"smart-pc-pc-service/internal/domain/models"
	"strings"
	"time"

	"smart-pc-pc-service/internal/lib/logger/sl"
	"smart-pc-pc-service/internal/lib/mqtt/message"

	"github.com/eclipse/paho.golang/paho"
	"github.com/google/uuid"
)

type Message struct {
	Command     string    `json:"command"`
	ReceivedAt  time.Time `json:"receivedAt"`
	CompletedAt time.Time `json:"completedAt"`
	Status      string    `json:"status"`
	Error       string    `json:"error,omitempty"`
}

type PcLogCreator interface {
	CreatePcLog(context.Context, models.PcLog) error
}

func New(ctx context.Context, log *slog.Logger, creator PcLogCreator) paho.MessageHandler {
	return func(p *paho.Publish) {
		const op = "mqtt.handlers.pc-logs"

		log := log.With(sl.Op(op), sl.PubID(p))

		payload, err := message.Decode[Message](p)
		if err != nil {
			log.Error("failed to decode payload", sl.Err(err))
			return
		}
		log.Debug("got payload", slog.Any("payload", payload))

		pcID, err := getPcID(p)
		if err != nil {
			log.Error("failed to get pcID", sl.Err(err))
			return
		}
		log.Debug("got pcID", slog.String("pcID", pcID.String()))

		if err := creator.CreatePcLog(ctx, models.PcLog{
			PcID:        &pcID,
			CommandID:   payload.Data.Command,
			ReceivedAt:  payload.Data.ReceivedAt,
			CompletedAt: payload.Data.CompletedAt,
			Status:      payload.Data.Status,
			Error:       payload.Data.Error,
		}); err != nil {
			log.Error("failed to create pc log", sl.Err(err))
			return
		}
	}
}

func getPcID(p *paho.Publish) (uuid.UUID, error) {
	const op = "mqtt.handlers.pc-logs.getPcID"

	topicParts := strings.Split(p.Topic, "/")
	if len(topicParts) < 4 {
		return uuid.Nil, fmt.Errorf("%s: invalid topic", op)
	}

	pcID, err := uuid.Parse(topicParts[3])
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: failed to parse topic part to uuid: %w", op, err)
	}

	return pcID, nil
}
