package pcCommands

import (
	"context"
	"smart-pc-pc-service/internal/domain/models"
	"smart-pc-pc-service/internal/storage/postgres/dbqueries"

	"github.com/google/uuid"
)

type Storage struct {
	queries *dbqueries.Queries
}

func New(queries *dbqueries.Queries) *Storage {
	return &Storage{queries: queries}
}

func (s *Storage) ListPCCommands(
	ctx context.Context,
	userID, pcID uuid.UUID,
) ([]models.Command, error) {
	rows, err := s.queries.UserPCCommands(
		ctx,
		dbqueries.UserPCCommandsParams{UserID: userID, PcID: pcID},
	)
	if err != nil {
		return nil, err
	}

	return mapStorageCommands(rows), nil
}

func mapStorageCommand(command dbqueries.Command) models.Command {
	return models.Command{
		ID:          command.ID,
		PcID:        command.PcID,
		Name:        command.Name,
		Description: command.Description,
	}
}

func mapStorageCommands(commands []dbqueries.Command) []models.Command {
	cmds := make([]models.Command, len(commands))
	for i, r := range commands {
		cmds[i] = mapStorageCommand(r)
	}

	return cmds
}
