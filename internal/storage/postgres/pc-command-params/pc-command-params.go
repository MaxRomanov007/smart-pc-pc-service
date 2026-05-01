package pcCommandParams

import (
	"context"
	"smart-pc-pc-service/internal/storage/postgres/dbqueries"

	"github.com/MaxRomanov007/smart-pc-go-lib/domain/models"
	"github.com/google/uuid"
)

type Storage struct {
	queries *dbqueries.Queries
}

func New(queries *dbqueries.Queries) *Storage {
	return &Storage{queries: queries}
}

func (s *Storage) ListCommandParameters(
	ctx context.Context,
	userID, _, commandID uuid.UUID,
) ([]models.CommandParameter, error) {
	rows, err := s.queries.UserCommandParameters(
		ctx,
		dbqueries.UserCommandParametersParams{UserID: userID, CommandID: commandID},
	)
	if err != nil {
		return nil, err
	}

	return mapStorageCommandParams(rows), nil
}

func mapStorageCommandParam(command dbqueries.CommandParameter) models.CommandParameter {
	return models.CommandParameter{
		ID:          command.ID,
		CommandID:   command.CommandID,
		Name:        command.Name,
		Description: command.Description,
		Type:        command.Type,
	}
}

func mapStorageCommandParams(commands []dbqueries.CommandParameter) []models.CommandParameter {
	cmds := make([]models.CommandParameter, len(commands))
	for i, r := range commands {
		cmds[i] = mapStorageCommandParam(r)
	}

	return cmds
}
