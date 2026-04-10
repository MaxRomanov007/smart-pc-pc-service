package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	PgErrUniqueViolation = "23505"
)

func IsUniqueViolation(err error) bool {
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	return ok && pgErr.Code == PgErrUniqueViolation
}
