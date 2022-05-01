package config

type options struct {
	codecName    string
	providerName string
}

var (
	defaultOptions = options{
		codecName:    "yaml",
		providerName: "file",
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

func WithCodec(name string) Option {
	return newFuncOption(func(o *options) {
		o.codecName = name
	})
}

func WithProvider(name string) Option {
	return newFuncOption(func(o *options) {
		o.providerName = name
	})
}
