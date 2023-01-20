package prometheus

type options struct {
	disableInstrumentMetricHandler bool
	disableProcessCollector        bool
	disableProcessExtCollector     bool
	enableDiskUsageCollector       bool
	diskUsagePath                  string
	enableGoCollector              bool
	disableGoSimpleCollector       bool
}

var (
	defaultOptions = options{}
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

func WithDisableInstrumentMetricHandler() Option {
	return newFuncOption(func(o *options) {
		o.disableInstrumentMetricHandler = true
	})
}

func WithDisableProcessCollector() Option {
	return newFuncOption(func(o *options) {
		o.disableProcessCollector = true
	})
}

func WithDisableProcessExtCollector() Option {
	return newFuncOption(func(o *options) {
		o.disableProcessExtCollector = true
	})
}

func WithDiskUsageCollector(path string) Option {
	return newFuncOption(func(o *options) {
		o.enableDiskUsageCollector = true
		o.diskUsagePath = path
	})
}

func WithGoCollector() Option {
	return newFuncOption(func(o *options) {
		o.enableGoCollector = true
		o.disableGoSimpleCollector = true
	})
}
