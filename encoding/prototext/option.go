package prototext

type options struct {
	limit int // max string length
}

var (
	defaultOptions = options{
		limit: 0, // 0 unlimited
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

func WithStringLimit(limit int) Option {
	return newFuncOption(func(o *options) {
		o.limit = limit
	})
}
