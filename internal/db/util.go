package db

import (
	"context"
	"database/sql"
)

// ExecContext виконує INSERT/UPDATE/DELETE з параметрами та контекстом.
func (s *Store) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return s.db.ExecContext(ctx, query, args...)
}

// QueryRowContext виконує запит одного рядка з контекстом.
func (s *Store) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return s.db.QueryRowContext(ctx, query, args...)
}
