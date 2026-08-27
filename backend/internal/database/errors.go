package database

import (
	"errors"

	"github.com/mattn/go-sqlite3"
)

func IsUniqueConstraintError(err error) bool {
	sqliteErr, ok := errors.AsType[sqlite3.Error](err)
	if ok {
		return sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique
	}

	return false
}
