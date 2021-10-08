package pool

import (
	"context"
	"github.com/onnoink/go-rabbit-pool/log"
)

type Option func(*options)

type options struct {
	ctx              context.Context
	url              string
	minAlive         int
	maxAlive         int
	maxIdle          int
	idleCloseTimeOut int64
	logger           log.Logger
}

func (o *options) Optimize() {
	if o.minAlive > o.maxAlive {
		o.minAlive = o.maxAlive
	}
	if o.maxIdle > o.maxAlive {
		o.maxIdle = o.maxAlive
	}
	if o.maxIdle <= o.minAlive {
		o.maxIdle = o.minAlive
	}
}

func Context(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

func MinAlive(minAlive int) Option {
	return func(o *options) {
		o.minAlive = minAlive
	}
}

func MaxAlive(maxAlive int) Option {
	return func(o *options) {
		o.maxAlive = maxAlive
	}
}

func MaxIdle(maxIdle int) Option {
	return func(o *options) {
		o.maxIdle = maxIdle
	}
}

func IdleCloseTimeOut(t int64) Option {
	return func(o *options) {
		o.idleCloseTimeOut = t
	}
}

func Logger(logger log.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}
