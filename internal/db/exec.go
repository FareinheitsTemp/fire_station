package db

import "database/sql"

// Exec виконує INSERT/UPDATE/DELETE з параметрами.
// Потрібен CRUD-ендпоінтам веб-API (cmd/webapi).
func (s *Store) Exec(query string, args ...any) (sql.Result, error) {
	return s.db.Exec(query, args...)
}
