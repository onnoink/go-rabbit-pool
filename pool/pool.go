package pool

import (
	"context"
	"github.com/google/uuid"
	"github.com/onnoink/go-rabbit-pool/log"
	"github.com/streadway/amqp"
	syslog "log"
	"sync"
	"time"
)

type ConnectionStats struct {
	TotalCount   int
	CreatedCount int64
}

type Connection struct {
	*amqp.Connection
	id        uuid.UUID
	lastUseAt int64
}


type Pool struct {
	opts    *options
	synchro struct {
		sync.Mutex
		Connections chan *Connection
		stats       ConnectionStats
	}

	factory func() (*Connection, error)
}

func NewPool(url string, opts ...Option) *Pool {

	// default options
	options := &options{
		ctx:              context.Background(),
		url:              url,
		minAlive:         4,
		maxIdle:          8,
		maxAlive:         16,
		idleCloseTimeOut: 600,
		logger:           log.NewStdLogger(syslog.Writer()),
	}

	// custom options
	for _, o := range opts {
		o(options)
	}

	// optimize options
	options.Optimize()

	// new pool
	pool := &Pool{
		opts: options,
		factory: func() (*Connection, error) {
			conn, err := amqp.Dial(url)
			if err != nil {
				return nil, err
			}

			uid, _ := uuid.NewUUID()

			connection := &Connection{
				conn,
				uid,
				time.Now().Unix(),
			}

			return connection, nil
		},
	}

	pool.synchro.Connections = make(chan *Connection, pool.opts.maxAlive)

	if pool.opts.minAlive > 0 {
		pool.initConnection()
	}

	go pool.timerConnectionFactory()

	return pool
}

// Stats return stats of the pool
func (pool *Pool) Stats() *ConnectionStats {
	stats := new(ConnectionStats)
	*stats = pool.synchro.stats
	return stats
}

func (pool *Pool) Acquire() (*Connection, error) {

	var c *Connection
	var err error

	for {
		c, err = pool.getOrCreate()
		if err != nil {
			return nil, err
		}

		if c.IsClosed() {
			if err = pool.Destroy(c); err != nil {
				return nil, err
			}
		} else {
			return c, nil
		}
	}
}

func (pool *Pool) getOrCreate() (*Connection, error) {
	select {
	case connection := <-pool.synchro.Connections:
		connection.lastUseAt = time.Now().Unix()
		return connection, nil
	default:

	}

	pool.synchro.Lock()
	if pool.synchro.stats.TotalCount >= pool.opts.maxAlive {
		pool.synchro.Unlock()
		connection := <-pool.synchro.Connections
		connection.lastUseAt = time.Now().Unix()
		return connection, nil
	}

	connection, err := pool.factory()
	if err != nil {
		pool.synchro.Unlock()
		return nil, err
	}
	pool.synchro.stats.TotalCount++
	pool.synchro.stats.CreatedCount++

	pool.synchro.Unlock()

	return connection, nil
}

func (pool *Pool) Release(con *Connection) error {
	pool.synchro.Lock()
	pool.synchro.Connections <- con
	pool.synchro.Unlock()
	return nil
}

func (pool *Pool) Destroy(con *Connection) error {
	pool.synchro.Lock()
	pool.synchro.stats.TotalCount--
	pool.synchro.Unlock()
	return nil
}

func (pool *Pool) initConnection() {
	pool.synchro.Lock()
	defer pool.synchro.Unlock()
	for i := 0; i < pool.opts.minAlive; i++ {
		con, err := pool.factory()
		if err != nil {
			pool.opts.logger.Log(log.LevelWarn, "fail to init minimum alive connection", err.Error())
			continue
		}
		pool.synchro.Connections <- con
		pool.synchro.stats.TotalCount++
		pool.synchro.stats.CreatedCount++
	}
}

// timerConnectionFactory timing factory to create connection
func (pool *Pool) timerConnectionFactory() {

	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			pool.synchro.Lock()
			// 释放超时等待连接
			if pool.synchro.stats.TotalCount > pool.opts.maxIdle {
				select {
				case connection := <-pool.synchro.Connections:
					if time.Now().Unix()-connection.lastUseAt > pool.opts.idleCloseTimeOut {
						pool.synchro.stats.TotalCount--
					} else {
						pool.synchro.Connections <- connection
					}
				default:

				}
			}

			pool.synchro.Unlock()


		case <-pool.opts.ctx.Done():
			return
		}
	}
}
