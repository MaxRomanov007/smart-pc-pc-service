package models

import (
	"time"

	"github.com/google/uuid"
)

type CommandLog struct {
	ID          uuid.UUID `json:"id"`
	CommandID   uuid.UUID `json:"commandId"`
	ReceivedAt  time.Time `json:"receivedAt"`
	CompletedAt time.Time `json:"completedAt"`
	Status      string    `json:"status"`
	Error       *string   `json:"error,omitempty"`

	Command *Command `json:"command,omitempty"`
}
