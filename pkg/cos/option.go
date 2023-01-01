package cos

type options struct {
	maxIdleConns int
	pathStyle    bool
	disableSSL   bool
}

var (
	defaultOptions = options{
		maxIdleConns: 200,
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

func WithPathStyle() Option {
	return newFuncOption(func(o *options) {
		o.pathStyle = true
	})
}

func WithDisableSSL() Option {
	return newFuncOption(func(o *options) {
		o.disableSSL = true
	})
}
