package kafka

import (
	"time"
)

type options struct {
	enableSASL bool
	username   string
	password   string

	maxMessageBytes int

	enableCompression bool

	flushFrequency time.Duration
}

var (
	defaultOptions = options{
		maxMessageBytes: 1000000,
		flushFrequency:  time.Duration(100) * time.Millisecond,
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

// WithSASL
// Whether to use SASL authentication when connecting to the broker (defaults to false).
func WithSASL(username, password string) Option {
	return newFuncOption(func(o *options) {
		o.enableSASL = true
		o.username = username
		o.password = password
	})
}

// WithMaxMessageBytes
// The maximum permitted size of a message (defaults to 1000000).
// Should be set equal to or smaller than the broker's `message.max.bytes`.
func WithMaxMessageBytes(maxMessageBytes int) Option {
	return newFuncOption(func(o *options) {
		o.maxMessageBytes = maxMessageBytes
	})
}

// WithCompression
// The type of compression to use on messages (defaults to no compression).
func WithCompression() Option {
	return newFuncOption(func(o *options) {
		o.enableCompression = true
	})
}

// WithFlushFrequency
// The best-effort frequency of flushes.
func WithFlushFrequency(flushFrequency time.Duration) Option {
	return newFuncOption(func(o *options) {
		o.flushFrequency = flushFrequency
	})
}
