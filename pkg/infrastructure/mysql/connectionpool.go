package mysql

import (
	"context"
	"sync"
)

type ConnectionPool interface {
	TransactionalConnection(ctx context.Context) (TransactionalConnection, error)
}

func NewConnectionPool(client TransactionalClient) ConnectionPool {
	return &connectionPool{
		client: client,
		pool:   make(map[context.Context]*sharedConnection),
	}
}

type connectionPool struct {
	client TransactionalClient

	mu   sync.Mutex
	pool map[context.Context]*sharedConnection
}

func (cp *connectionPool) TransactionalConnection(ctx context.Context) (TransactionalConnection, error) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	conn, ok := cp.pool[ctx]
	if ok {
		conn.count++
	}
	if conn == nil {
		c, err := cp.client.Connection(ctx)
		if err != nil {
			return nil, err
		}
		conn = &sharedConnection{
			TransactionalConnection: c,
			ctx:                     ctx,
			count:                   1,
			releaseCallback:         cp.release,
		}
		cp.pool[ctx] = conn
	}
	return conn, nil
}

func (cp *connectionPool) release(ctx context.Context) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	conn, ok := cp.pool[ctx]
	if ok {
		return nil
	}
	if conn.count == 1 {
		err := conn.TransactionalConnection.Close()
		delete(cp.pool, ctx)
		return err
	}
	conn.count--
	return nil
}

type sharedConnection struct {
	TransactionalConnection
	ctx             context.Context
	count           int
	releaseCallback func(ctx context.Context) error
}

func (sc *sharedConnection) Close() error {
	return sc.releaseCallback(sc.ctx)
}
