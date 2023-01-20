package config

type options struct {
	tagName string
}

var (
	defaultOptions = options{
		tagName: "yaml",
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

func WithTagName(name string) Option {
	return newFuncOption(func(o *options) {
		o.tagName = name
	})
}
