package pcs

import (
	"context"
	"errors"
	"fmt"

	"smart-pc-pc-service/internal/config"
	"smart-pc-pc-service/internal/domain/models"
	"smart-pc-pc-service/internal/lib/storage/postgres"
	"smart-pc-pc-service/internal/lib/strings/slug/slugger"
	"smart-pc-pc-service/internal/storage"
	"smart-pc-pc-service/internal/storage/postgres/dbqueries"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Storage struct {
	queries *dbqueries.Queries
	slugCfg config.Slug
}

func New(queries *dbqueries.Queries, slugCfg config.Slug) *Storage {
	return &Storage{queries: queries, slugCfg: slugCfg}
}

func (s *Storage) PcsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Pc, error) {
	const op = "storage.postgres.pcs.PcsByUserID"

	pcs, err := s.queries.UserPCs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get user pcs: %w", op, err)
	}

	result := make([]models.Pc, 0, len(pcs))
	for _, pc := range pcs {
		result = append(result, parseStoragePc(pc))
	}

	return result, nil
}

func (s *Storage) PcBySlug(
	ctx context.Context,
	userID uuid.UUID,
	slug string,
) (models.Pc, error) {
	const op = "storage.postgres.pcs.PcBySlug"

	pc, err := s.queries.UserPCBySlug(ctx, dbqueries.UserPCBySlugParams{
		UserID: userID,
		Slug:   slug,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Pc{}, storage.ErrNotFound
	}
	if err != nil {
		return models.Pc{}, fmt.Errorf("%s: failed to get user pc by slug: %w", op, err)
	}

	return parseStoragePc(pc), nil
}

func (s *Storage) PcByID(ctx context.Context, userID, pcID uuid.UUID) (models.Pc, error) {
	const op = "storage.postgres.pcs.PcByID"

	pc, err := s.queries.UserPCByID(ctx, dbqueries.UserPCByIDParams{
		UserID: userID,
		ID:     pcID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Pc{}, storage.ErrNotFound
	}
	if err != nil {
		return models.Pc{}, fmt.Errorf("%s: failed to get user pc by id: %w", op, err)
	}

	return parseStoragePc(pc), nil
}

func (s *Storage) UpdatePc(
	ctx context.Context,
	userID, pcID uuid.UUID,
	name, description *string,
	canPowerOn *bool,
) (models.Pc, error) {
	const op = "storage.postgres.pcs.UpdatePc"

	params := dbqueries.UpdateUserPCParams{
		ID:          pcID,
		Name:        name,
		Slug:        nil,
		Description: description,
		CanPowerOn:  canPowerOn,
		UserID:      userID,
	}

	if name == nil {
		pc, err := s.queries.UpdateUserPC(ctx, params)
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Pc{}, storage.ErrNotFound
		}
		if err != nil {
			return models.Pc{}, fmt.Errorf("%s: failed to update user pc: %w", op, err)
		}

		return parseStoragePc(pc), nil
	}

	pcSlugger := slugger.New(s.slugCfg.Retries, func(slug string) (dbqueries.Pc, error, bool) {
		params.Slug = &slug
		pc, err := s.queries.UpdateUserPC(ctx, params)
		return pc, err, postgres.IsUniqueViolation(err)
	})
	pc, err := pcSlugger.DoCtx(ctx, *params.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Pc{}, storage.ErrNotFound
	}
	if err != nil {
		return models.Pc{}, fmt.Errorf("%s: failed to update user pc: %w", op, err)
	}

	return parseStoragePc(pc), nil
}

func (s *Storage) CreatePc(ctx context.Context, pc models.Pc) (models.Pc, error) {
	const op = "storage.postgres.pcs.CreatePc"

	params := dbqueries.CreateUserPCParams{
		UserID:      pc.UserID,
		Name:        pc.Name,
		Description: pc.Description,
		CanPowerOn:  pc.CanPowerOn,
	}

	if params.Name == "" {
		pc, err := s.queries.CreateUserPC(ctx, params)
		if err != nil {
			return models.Pc{}, fmt.Errorf("%s: failed to create user pc: %w", op, err)
		}

		return parseStoragePc(pc), nil
	}

	pcSlugger := slugger.New(s.slugCfg.Retries, func(slug string) (dbqueries.Pc, error, bool) {
		params.Slug = slug
		pc, err := s.queries.CreateUserPC(ctx, params)
		return pc, err, postgres.IsUniqueViolation(err)
	})
	created, err := pcSlugger.DoCtx(ctx, params.Name)
	if err != nil {
		return models.Pc{}, fmt.Errorf("%s: failed to create user pc: %w", op, err)
	}

	return parseStoragePc(created), nil
}

func (s *Storage) DeleteUserPc(ctx context.Context, uid, id uuid.UUID) (models.Pc, error) {
	const op = "storage.postgres.pcs.DeleteUserPc"

	deleted, err := s.queries.DeleteUserPC(ctx, dbqueries.DeleteUserPCParams{
		UserID: uid,
		ID:     id,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Pc{}, storage.ErrNotFound
	}
	if err != nil {
		return models.Pc{}, fmt.Errorf("%s: failed to delete user pc: %w", op, err)
	}

	return parseStoragePc(deleted), nil
}

func parseStoragePc(pc dbqueries.Pc) models.Pc {
	return models.Pc{
		ID:          pc.ID,
		UserID:      pc.UserID,
		Slug:        pc.Slug,
		Name:        pc.Name,
		Description: pc.Description,
		CanPowerOn:  pc.CanPowerOn,
	}
}
