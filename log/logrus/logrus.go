package logrus

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

const defaultTimestampFormat = "2006-01-02 15:04:05.000"

type CallerPrettyfier func(*runtime.Frame) (function string, file string)

type ReadableFormatter struct {
	TimestampFormat  string
	DisableSorting   bool
	CallerPrettyfier CallerPrettyfier
	CtxFields        []string
}

func simplifyFilePath(s string) string {
	lastN := 2
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '/' {
			lastN--
			if lastN == 0 {
				return s[i+1:]
			}
		}
	}
	return s
}

var defaultCallerPrettyfier = func(f *runtime.Frame) (function string, file string) {
	fileVal := fmt.Sprintf("%s:%d", simplifyFilePath(f.File), f.Line)

	funcVal := f.Function
	if i := strings.LastIndex(funcVal, "."); i >= 0 && i+1 < len(funcVal) {
		funcVal = funcVal[i+1:]
	}

	return funcVal, fileVal
}

func (f *ReadableFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(logrus.Fields)
	for k, v := range entry.Data {
		data[k] = v
	}
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}

	fixedKeys := make([]string, 0, 4+len(data))
	fixedKeys = append(fixedKeys, logrus.FieldKeyTime)
	fixedKeys = append(fixedKeys, logrus.FieldKeyLevel)

	var funcVal, fileVal string
	if entry.HasCaller() {
		if f.CallerPrettyfier != nil {
			funcVal, fileVal = f.CallerPrettyfier(entry.Caller)
		} else {
			funcVal = entry.Caller.Function
			fileVal = fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
		}

		if funcVal != "" {
			fixedKeys = append(fixedKeys, logrus.FieldKeyFunc)
		}
		if fileVal != "" {
			fixedKeys = append(fixedKeys, logrus.FieldKeyFile)
		}
	}

	if !f.DisableSorting {
		sort.Strings(keys)
		fixedKeys = append(fixedKeys, keys...)
	} else {
		fixedKeys = append(fixedKeys, keys...)
	}

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = defaultTimestampFormat
	}

	// level
	//f.appendKeyValue(b, logrus.FieldKeyLevel, strings.ToUpper(entry.Level.String())[0:4])
	_, _ = fmt.Fprintf(b, "[%s] ", strings.ToUpper(entry.Level.String())[0:4])

	// time
	//f.appendKeyValue(b, logrus.FieldKeyTime, entry.Time.Format(timestampFormat))
	_, _ = fmt.Fprintf(b, "[%s] ", entry.Time.Format(timestampFormat))

	// caller info
	if entry.HasCaller() {
		//f.appendKeyValue(b, logrus.FieldKeyFunc, funcVal)
		//f.appendKeyValue(b, logrus.FieldKeyFile, fileVal)
		_, _ = fmt.Fprintf(b, "[%s(%s)]", fileVal, funcVal)
	}

	// fields
	if entry.Context != nil && len(f.CtxFields) > 0 {
		for _, key := range f.CtxFields {
			if val := entry.Context.Value(key); val != nil {
				f.appendKeyValue(b, key, val)
			}
		}
	}
	for _, key := range fixedKeys {
		var value interface{}
		switch key {
		case logrus.FieldKeyTime, logrus.FieldKeyLevel,
			logrus.FieldKeyMsg, logrus.FieldKeyLogrusError,
			logrus.FieldKeyFunc, logrus.FieldKeyFile:
			continue
		default:
			value = data[key]
		}
		f.appendKeyValue(b, key, value)
	}

	// message
	f.appendKeyValue(b, "", entry.Message)

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *ReadableFormatter) appendKeyValue(b *bytes.Buffer, key string, value interface{}) {
	if b.Len() > 0 {
		b.WriteByte(' ')
	}
	b.WriteString(key)
	if key != "" {
		b.WriteByte('=')
	}
	f.appendValue(b, value)
}

func (f *ReadableFormatter) appendValue(b *bytes.Buffer, value interface{}) {
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}

	b.WriteString(fmt.Sprintf("%s", stringVal))
}

func init() {
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&ReadableFormatter{CallerPrettyfier: defaultCallerPrettyfier})
}

func Init(opts ...Option) {
	defaultOpts := defaultOptions
	for _, o := range opts {
		o.apply(&defaultOpts)
	}

	// formatter
	logrus.SetFormatter(&ReadableFormatter{
		CallerPrettyfier: defaultCallerPrettyfier,
		CtxFields:        defaultOpts.ctxFields,
	})

	// level
	level, err := logrus.ParseLevel(defaultOpts.level)
	if err != nil {
		logrus.WithError(err).Fatal()
	}
	logrus.SetLevel(level)

	// report caller
	logrus.SetReportCaller(!defaultOpts.disableCaller)

	// output
	var writers []io.Writer
	if !defaultOpts.disableStdout {
		writers = append(writers, os.Stderr)
	}
	if len(defaultOpts.logDir) > 0 && len(defaultOpts.fileName) > 0 {
		writer, err := rotatelogs.New(
			filepath.Join(defaultOpts.logDir, fmt.Sprintf("%s.log.%%Y%%m%%d_%%H", defaultOpts.fileName)),
			rotatelogs.WithLinkName(filepath.Join(defaultOpts.logDir, fmt.Sprintf("%s.log", defaultOpts.fileName))),
			rotatelogs.WithRotationTime(time.Hour),
			rotatelogs.WithMaxAge(defaultOpts.maxAge),
		)
		if err != nil {
			logrus.WithError(err).Fatal()
		}
		writers = append(writers, writer)
	}
	if len(writers) < 1 {
		logrus.Fatal("no logger output")
	}

	logrus.SetOutput(io.MultiWriter(writers...))
}
