package http

import "time"

type options struct {
	timeout      time.Duration
	maxIdleConns int
}

var (
	defaultOptions = options{
		timeout:      5 * time.Second,
		maxIdleConns: 100,
	}
)

type Option interface {
	apply(*options)
}

type funcOption struct {
	f func(*options)
}

func (fo *funcOption) apply(o *options) {
	fo.f(o)
}

func newFuncOption(f func(*options)) *funcOption {
	return &funcOption{
		f: f,
	}
}

func WithMaxIdleConns(maxIdleConns int) Option {
	return newFuncOption(func(o *options) {
		o.maxIdleConns = maxIdleConns
	})
}
