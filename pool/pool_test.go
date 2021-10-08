package pool

import (
	"flag"
	"testing"
)

var url = flag.String("url", "amqp://guest:guest@127.0.0.1:5672", "url")

var (
	minAlive = 4
	maxAlive = 16
	maxIdle  = 8
)

func TestNewPool(t *testing.T) {
	pool := NewPool(*url, MaxAlive(maxAlive), MaxIdle(maxIdle), MinAlive(minAlive))
	c, err := pool.Acquire()
	if err != nil {
		t.Errorf("acquire connection err %s\n", err.Error())
		return
	}

	pool.Release(c)

	if c.Connection == nil {
		t.Errorf("acquire connection err %s\n", "nil connection")
		return
	}

	if pool.Stats().TotalCount != 4 {
		t.Errorf("acquire connection err %s\n", "wrong minium connection")
		return
	}
}
