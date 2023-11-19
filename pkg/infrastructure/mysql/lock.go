package mysql

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrLockTimeout   = errors.New("lock timed out")
	ErrLockNotLocked = errors.New("lock not locked")
	ErrLockNotFound  = errors.New("lock not found")
)

type LockFactory interface {
	NewLock(ctx context.Context, lockName string, timeout time.Duration) (Lock, error)
}

type Lock interface {
	Unlock() error
}

func NewLockFactory(connectionPool ConnectionPool) LockFactory {
	return &lockFactory{connectionPool: connectionPool}
}

type lockFactory struct {
	connectionPool ConnectionPool
}

func (factory *lockFactory) NewLock(ctx context.Context, lockName string, timeout time.Duration) (Lock, error) {
	conn, err := factory.connectionPool.TransactionalConnection(ctx)
	if err != nil {
		return nil, err
	}

	lock := lockImpl{
		ctx:      ctx,
		lockName: lockName,
		timeout:  timeout,
		conn:     conn,
	}

	err = lock.Lock()
	if err != nil {
		err = errors.Join(err, conn.Close())
		return nil, err
	}

	return &lock, err
}

type lockImpl struct {
	ctx      context.Context
	lockName string
	timeout  time.Duration
	conn     TransactionalConnection
}

func (l *lockImpl) Lock() error {
	const sqlQuery = "SELECT GET_LOCK(SUBSTRING(CONCAT(?, '.', DATABASE()), 1, 64), ?)"
	var result int
	err := l.conn.GetContext(l.ctx, &result, sqlQuery, l.lockName, int(l.timeout.Seconds()))
	if result == 0 && err == nil {
		return ErrLockTimeout
	}
	return err
}

func (l *lockImpl) Unlock() (err error) {
	defer func() {
		freeErr := l.conn.Close()
		err = errors.Join(err, freeErr)
	}()

	const sqlQuery = "SELECT RELEASE_LOCK(SUBSTRING(CONCAT(?, '.', DATABASE()), 1, 64))"
	var result sql.NullInt32
	err = l.conn.GetContext(l.ctx, &result, sqlQuery, l.lockName)
	if err == nil {
		if !result.Valid {
			err = ErrLockNotFound
			return err
		}
		if result.Int32 == 0 {
			err = ErrLockNotLocked
			return err
		}
	}
	return err
}
