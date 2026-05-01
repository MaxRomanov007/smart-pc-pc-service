package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func FinishTx(ctx context.Context, tx pgx.Tx, resultErr *error) {
	const op = "lib.storage.postgres.FinishTx"

	if *resultErr != nil {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			*resultErr = fmt.Errorf(
				"%s: failed to rollback (error: %w), after operation failed (error: %w)",
				op,
				rollbackErr,
				*resultErr,
			)
		}
		return
	}

	commitErr := tx.Commit(ctx)
	if commitErr != nil {
		*resultErr = fmt.Errorf("%s: failed to commit transaction: %w", op, commitErr)
	}
}
