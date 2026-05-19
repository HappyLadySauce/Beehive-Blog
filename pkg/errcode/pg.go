package errcode

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// IsUniqueViolation reports whether err wraps a PostgreSQL unique-constraint violation (SQLSTATE 23505).
// IsUniqueViolation 判断错误是否包含 PostgreSQL 唯一约束冲突（SQLSTATE 23505）。
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
