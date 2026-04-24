package pcCommands

import (
	"context"
	"errors"
	"fmt"
	"smart-pc-pc-service/internal/domain/models"
	"smart-pc-pc-service/internal/storage"
	"smart-pc-pc-service/internal/storage/postgres/dbqueries"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Storage {
	return &Storage{db: db}
}

func (s *Storage) ListPCCommands(
	ctx context.Context,
	userID, pcID uuid.UUID,
) ([]models.Command, error) {
	const op = "storage.postgres.pc-commands.ListPCCommands"

	queries := dbqueries.New(s.db)

	rows, err := queries.UserPCCommands(
		ctx,
		dbqueries.UserPCCommandsParams{UserID: userID, PcID: pcID},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get user pc commands: %w", op, err)
	}

	return mapStorageCommands(rows), nil
}

func (s *Storage) CreateUserPcCommand(
	ctx context.Context,
	uid uuid.UUID,
	command models.Command,
) (created models.Command, err error) {
	const op = "storage.postgres.pc-commands.CreatePcCommand"

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.Command{}, fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}

	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback(ctx)
			if rollbackErr != nil {
				err = fmt.Errorf(
					"%s: failed to rollback (error: %w), after operation failed (error: %w)",
					op,
					rollbackErr,
					err,
				)
			}
			return
		}

		commitErr := tx.Commit(ctx)
		if commitErr != nil {
			err = fmt.Errorf("%s: failed to commit transaction: %w", op, commitErr)
		}
	}()

	queries := dbqueries.New(tx)

	dbCommand, err := queries.CreateUserPcCommand(ctx, dbqueries.CreateUserPcCommandParams{
		Name:        command.Name,
		Description: command.Description,
		PcID:        *command.PcID,
		UserID:      uid,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Command{}, storage.ErrNotFound
	}
	if err != nil {
		return models.Command{}, fmt.Errorf("%s: failed to create user pc command: %w", op, err)
	}

	if command.Parameters == nil {
		return mapStorageCommand(dbCommand), nil
	}

	params := make([]dbqueries.CreateUserPcCommandParametersParams, len(command.Parameters))
	for i, param := range command.Parameters {
		params[i] = dbqueries.CreateUserPcCommandParametersParams{
			Name:        param.Name,
			Description: param.Description,
			Type:        param.Type,
			CommandID:   dbCommand.ID,
		}
	}
	if _, err := queries.CreateUserPcCommandParameters(ctx, params); err != nil {
		return models.Command{}, fmt.Errorf(
			"%s: failed to create user pc command parameters: %w",
			op,
			err,
		)
	}

	return mapStorageCommand(dbCommand), nil
}

func (s *Storage) DeleteUserPcCommand(
	ctx context.Context,
	uid, pcID, commandID uuid.UUID,
) (models.Command, error) {
	const op = "storage.postgres.pc-commands.DeleteUserPcCommand"

	queries := dbqueries.New(s.db)

	deleted, err := queries.DeleteUserPcCommand(ctx, dbqueries.DeleteUserPcCommandParams{
		ID:     commandID,
		PcID:   pcID,
		UserID: uid,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Command{}, storage.ErrNotFound
	}
	if err != nil {
		return models.Command{}, fmt.Errorf("%s: failed to delete user pc command: %w", op, err)
	}

	return mapStorageCommand(deleted), nil
}

func (s *Storage) UpdateUserPcCommand(
	ctx context.Context,
	uid uuid.UUID,
	command models.Command,
) (updated models.Command, err error) {
	const op = "storage.postgres.pc-commands.UpdateUserPcCommand"

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.Command{}, fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}

	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback(ctx)
			if rollbackErr != nil {
				err = fmt.Errorf(
					"%s: failed to rollback (error: %w), after operation failed (error: %w)",
					op,
					rollbackErr,
					err,
				)
			}
			return
		}

		commitErr := tx.Commit(ctx)
		if commitErr != nil {
			err = fmt.Errorf("%s: failed to commit transaction: %w", op, commitErr)
		}
	}()

	queries := dbqueries.New(tx)

	updatedCommand, err := queries.UpdateUserPcCommand(ctx, dbqueries.UpdateUserPcCommandParams{
		Name:        &command.Name,
		Description: &command.Description,
		ID:          *command.ID,
		UserID:      uid,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Command{}, fmt.Errorf("%s: %w", op, storage.ErrNotFound)
	}
	if err != nil {
		return models.Command{}, fmt.Errorf("%s: failed to update user pc command: %w", op, err)
	}

	result := mapStorageCommand(updatedCommand)

	if command.Parameters == nil {
		return result, nil
	}

	ids := make([]uuid.UUID, len(command.Parameters))
	names := make([]string, len(command.Parameters))
	descriptions := make([]string, len(command.Parameters))
	types := make([]int16, len(command.Parameters))
	for i, param := range command.Parameters {
		if param.ID != uuid.Nil {
			ids[i] = param.ID
		} else {
			id, err := uuid.NewV7()
			if err != nil {
				return models.Command{}, fmt.Errorf("%s: failed to create uuid: %w", op, err)
			}
			ids[i] = id
		}
		names[i] = param.Name
		descriptions[i] = param.Description
		types[i] = param.Type
	}

	params, err := queries.SyncUserPcCommandParameters(
		ctx,
		dbqueries.SyncUserPcCommandParametersParams{
			CommandID:    *result.ID,
			UserID:       uid,
			Ids:          ids,
			Names:        names,
			Descriptions: descriptions,
			Types:        types,
		},
	)
	if err != nil {
		return models.Command{}, fmt.Errorf(
			"%s: failed to sync user pc command parameters: %w",
			op,
			err,
		)
	}

	result.Parameters = mapStorageCommandParams(params)

	return result, nil
}

func mapStorageCommand(command dbqueries.Command) models.Command {
	return models.Command{
		ID:          &command.ID,
		PcID:        &command.PcID,
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

func mapStorageCommandParam(
	command dbqueries.SyncUserPcCommandParametersRow,
) models.CommandParameter {
	return models.CommandParameter{
		ID:          command.ID,
		CommandID:   command.CommandID,
		Name:        command.Name,
		Description: command.Description,
		Type:        command.Type,
	}
}

func mapStorageCommandParams(
	commands []dbqueries.SyncUserPcCommandParametersRow,
) []models.CommandParameter {
	cmds := make([]models.CommandParameter, len(commands))
	for i, r := range commands {
		cmds[i] = mapStorageCommandParam(r)
	}

	return cmds
}
