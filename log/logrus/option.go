package logrus

import "time"

type options struct {
	level         string        // 日志级别
	ctxFields     []string      // 打印 context 中的字段
	disableCaller bool          // 关闭打印行号
	disableStdout bool          // 关闭终端打印
	logDir        string        // 日志目录
	fileName      string        // 日志名
	maxAge        time.Duration // 日志保留时长
}

var (
	defaultOptions = options{
		level:  "trace",
		maxAge: 7 * 24 * time.Hour,
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

// WithLevel 设置日志级别: trace,debug,info,warn,error,fatal,panic
func WithLevel(level string) Option {
	return newFuncOption(func(o *options) {
		o.level = level
	})
}

// WithContextFields logrus.WithContext() 打印 ctx 中的字段
func WithContextFields(fields ...string) Option {
	return newFuncOption(func(o *options) {
		o.ctxFields = fields
	})
}

// WithDisableCaller 关闭日志打印调用行号
func WithDisableCaller() Option {
	return newFuncOption(func(o *options) {
		o.disableCaller = true
	})
}

// WithDisableStdout 关闭日志终端打印
func WithDisableStdout() Option {
	return newFuncOption(func(o *options) {
		o.disableStdout = true
	})
}

// WithFileLog 设置文件日志信息
func WithFileLog(logDir, fileName string) Option {
	return newFuncOption(func(o *options) {
		o.logDir = logDir
		o.fileName = fileName
	})
}

// WithMaxAge 设置日志最大保留时长
func WithMaxAge(maxAge time.Duration) Option {
	return newFuncOption(func(o *options) {
		o.maxAge = maxAge
	})
}
