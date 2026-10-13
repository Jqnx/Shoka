package database

import (
	"errors"

	"github.com/mattn/go-sqlite3"
)

// IsUniqueConstraintError reports whether err is a SQLite unique constraint
// violation, i.e. an insert or update collided with an existing row.
func IsUniqueConstraintError(err error) bool {
	sqliteErr, ok := errors.AsType[sqlite3.Error](err)
	if ok {
		return sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique
	}

	return false
}
