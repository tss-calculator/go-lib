package mysql

import (
	"context"
	"errors"
	"sync"
)

type UnitOfWorkFactory interface {
	UnitOfWork(ctx context.Context) (UnitOfWork, error)
}

type UnitOfWork interface {
	Complete(err error) error
	ClientContext() ClientContext
}

type UnitOfWorkCompleteCallback func(ctx context.Context, err error)

func NewUnitOfWorkFactory(
	connectionPool ConnectionPool,
	unitOfWorkCompleteCallback UnitOfWorkCompleteCallback,
) UnitOfWorkFactory {
	return &unitOfWorkFactory{
		connectionPool:             connectionPool,
		unitOfWorkCompleteCallback: unitOfWorkCompleteCallback,
		transactionPool:            make(map[context.Context]*sharedTransaction),
	}
}

type unitOfWorkFactory struct {
	connectionPool             ConnectionPool
	unitOfWorkCompleteCallback UnitOfWorkCompleteCallback

	mu              sync.Mutex
	transactionPool map[context.Context]*sharedTransaction
}

func (factory *unitOfWorkFactory) UnitOfWork(ctx context.Context) (uow UnitOfWork, err error) {
	factory.mu.Lock()
	defer factory.mu.Unlock()

	stx, ok := factory.transactionPool[ctx]
	if ok {
		stx.count++
		uow = &unitOfWork{
			ctx:              ctx,
			tx:               stx,
			completeCallback: factory.unitOfWorkCompleteCallback,
		}
	}
	if uow == nil {
		conn, err := factory.connectionPool.TransactionalConnection(ctx)
		if err != nil {
			return nil, err
		}
		tx, err := conn.BeginTransaction(ctx, nil)
		if err != nil {
			return nil, errors.Join(err, conn.Close())
		}
		stx = &sharedTransaction{
			Transaction:      tx,
			ctx:              ctx,
			count:            1,
			conn:             conn,
			commitCallback:   factory.releaseWithCommit,
			rollbackCallback: factory.releaseWithRollback,
		}
		factory.transactionPool[ctx] = stx
		uow = &unitOfWork{
			ctx:              ctx,
			tx:               stx,
			completeCallback: factory.unitOfWorkCompleteCallback,
		}
	}
	return uow, nil
}

func (factory *unitOfWorkFactory) releaseWithCommit(ctx context.Context) error {
	return factory.releaseWithCallback(ctx, func(stx *sharedTransaction) error {
		return stx.Transaction.Commit()
	})
}

func (factory *unitOfWorkFactory) releaseWithRollback(ctx context.Context) error {
	return factory.releaseWithCallback(ctx, func(stx *sharedTransaction) error {
		return stx.Transaction.Rollback()
	})
}

func (factory *unitOfWorkFactory) releaseWithCallback(ctx context.Context, f func(stx *sharedTransaction) error) error {
	factory.mu.Lock()
	defer factory.mu.Unlock()

	stx, ok := factory.transactionPool[ctx]
	if !ok {
		return nil
	}
	if stx.count == 1 {
		err := f(stx)
		err = errors.Join(err, stx.conn.Close())
		delete(factory.transactionPool, ctx)
		return err
	}
	stx.count--
	return nil
}

type unitOfWork struct {
	ctx              context.Context
	tx               Transaction
	completeCallback UnitOfWorkCompleteCallback
}

func (u *unitOfWork) Complete(err error) (resultErr error) {
	resultErr = err
	defer func() {
		if u.completeCallback != nil {
			u.completeCallback(u.ctx, resultErr)
		}
	}()

	if resultErr != nil {
		rollbackErr := u.tx.Rollback()
		if rollbackErr != nil {
			resultErr = errors.Join(resultErr, rollbackErr)
			return resultErr
		}
		return resultErr
	}
	resultErr = u.tx.Commit()
	return resultErr
}

func (u *unitOfWork) ClientContext() ClientContext {
	return u.tx
}

type sharedTransaction struct {
	Transaction
	ctx              context.Context
	count            int
	conn             TransactionalConnection
	commitCallback   func(ctx context.Context) error
	rollbackCallback func(ctx context.Context) error
}

func (tx *sharedTransaction) Commit() error {
	return tx.commitCallback(tx.ctx)
}

func (tx *sharedTransaction) Rollback() error {
	return tx.rollbackCallback(tx.ctx)
}
