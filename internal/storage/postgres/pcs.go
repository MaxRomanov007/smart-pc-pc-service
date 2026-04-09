package postgres

import (
	"context"
	"errors"
	"fmt"

	"smart-pc-pc-service/internal/domain/models"
	"smart-pc-pc-service/internal/storage"
	"smart-pc-pc-service/internal/storage/postgres/dbqueries"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s Storage) PcsByUserID(ctx context.Context, userID string) ([]models.Pc, error) {
	const op = "storage.postgres.pcs.PcsByUserID"

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse user id: %w", op, err)
	}

	pcs, err := s.queries.GetPCsByUserID(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get pcs: %w", op, err)
	}

	result := make([]models.Pc, 0, len(pcs))
	for _, pc := range pcs {
		result = append(result, models.Pc{
			ID:          pc.ID,
			UserID:      pc.UserID,
			Slug:        pc.Slug,
			Name:        pc.Name,
			Description: pc.Name,
			CanPowerOn:  pc.CanPowerOn,
		})
	}

	return result, nil
}

func (s Storage) PcBySlug(
	ctx context.Context,
	userID string,
	slug string,
) (models.Pc, error) {
	const op = "storage.postgres.pcs.PcBySlug"

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return models.Pc{}, fmt.Errorf("%s: failed to parse user id: %w", op, err)
	}

	pc, err := s.queries.GetPCBySlug(ctx, &dbqueries.GetPCBySlugParams{
		UserID: userUUID,
		Slug:   slug,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Pc{}, storage.ErrNotFound
	}
	if err != nil {
		return models.Pc{}, fmt.Errorf("%s: failed to get pcs: %w", op, err)
	}

	return models.Pc{
		ID:          pc.ID,
		UserID:      pc.UserID,
		Slug:        pc.Slug,
		Name:        pc.Name,
		Description: pc.Name,
		CanPowerOn:  pc.CanPowerOn,
	}, nil
}
