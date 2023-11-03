package mysql

import (
	"database/sql"
	stderrors "errors"

	"github.com/pkg/errors"
)

var ErrLockTimeout = stderrors.New("lock timed out")
var ErrLockNotLocked = stderrors.New("lock not locked")
var ErrLockNotFound = stderrors.New("lock not found")

const lockTimeoutSeconds = 5

func NewLock(client Client, lockName string) Lock {
	return Lock{
		client:   client,
		lockName: lockName,
	}
}

type Lock struct {
	client   Client
	lockName string
}

func (l *Lock) Lock() error {
	const sqlQuery = "SELECT GET_LOCK(SUBSTRING(CONCAT(?, '.', DATABASE()), 1, 64), ?)"
	var result int
	err := l.client.Get(&result, sqlQuery, l.lockName, lockTimeoutSeconds)
	if result == 0 && err == nil {
		return errors.WithStack(ErrLockTimeout)
	}
	return errors.WithStack(err)
}

func (l *Lock) Unlock() error {
	const sqlQuery = "SELECT RELEASE_LOCK(SUBSTRING(CONCAT(?, '.', DATABASE()), 1, 64))"
	var result sql.NullInt32
	err := l.client.Get(&result, sqlQuery, l.lockName)
	if err == nil {
		if !result.Valid {
			return errors.WithStack(ErrLockNotFound)
		}
		if result.Int32 == 0 {
			return errors.WithStack(ErrLockNotLocked)
		}
	}
	return errors.WithStack(err)
}
