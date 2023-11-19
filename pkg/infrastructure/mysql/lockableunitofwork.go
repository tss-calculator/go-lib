package mysql

import (
	"context"
	"errors"
	"time"
)

const DefaultLockTimeout = time.Second * 5

type LockableUnitOfWorkFactory interface {
	NewLockableUnitOfWork(ctx context.Context, lockName string, timeout time.Duration) (LockableUnitOfWork, error)
}

type LockableUnitOfWork interface {
	UnitOfWork
}

func NewLockableUnitOfWorkFactory(
	lockFactory LockFactory,
	unitOfWorkFactory UnitOfWorkFactory,
) LockableUnitOfWorkFactory {
	return &lockableUnitOfWorkFactory{
		lockFactory:       lockFactory,
		unitOfWorkFactory: unitOfWorkFactory,
	}
}

type lockableUnitOfWorkFactory struct {
	lockFactory       LockFactory
	unitOfWorkFactory UnitOfWorkFactory
}

func (factory *lockableUnitOfWorkFactory) NewLockableUnitOfWork(ctx context.Context, lockName string, timeout time.Duration) (LockableUnitOfWork, error) {
	var lock Lock

	if lockName != "" {
		var err error
		lock, err = factory.lockFactory.NewLock(ctx, lockName, timeout)
		if err != nil {
			return nil, err
		}
	}

	unitOfWork, err := factory.unitOfWorkFactory.UnitOfWork(ctx)
	if err != nil {
		if lock != nil {
			err = errors.Join(err, lock.Unlock())
		}
		return nil, err
	}

	return &lockableUnitOfWork{
		lock:       lock,
		UnitOfWork: unitOfWork,
	}, nil
}

type lockableUnitOfWork struct {
	UnitOfWork
	lock Lock
}

func (u *lockableUnitOfWork) Complete(err error) error {
	returnErr := u.UnitOfWork.Complete(err)
	if u.lock != nil {
		returnErr = errors.Join(returnErr, u.lock.Unlock())
	}
	return returnErr
}
