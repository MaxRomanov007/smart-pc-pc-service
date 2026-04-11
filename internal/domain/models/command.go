package models

import "github.com/google/uuid"

type Command struct {
	ID          uuid.UUID `json:"id"`
	PcID        uuid.UUID `json:"pcId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`

	Pc         *Pc                `json:"pc,omitempty"`
	Parameters []CommandParameter `json:"parameters,omitempty"`
	Logs       []PcLog            `json:"logs,omitempty"`
}
