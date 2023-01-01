package logrus_test

import (
	"context"
	"testing"

	"github.com/limingyao/excellent-go/log/logrus"
	log "github.com/sirupsen/logrus"
)

func TestReadableFormatter(t *testing.T) {
	logrus.Init(logrus.WithContextFields("key1", "key2"))

	ctx := context.WithValue(context.Background(), "key1", "value1")
	ctx = context.WithValue(ctx, "key2", "value2")
	ctx = context.WithValue(ctx, "key3", "value3")

	log.WithContext(ctx).Debugf("testing log")
}

func BenchmarkReadableFormatter(b *testing.B) {
	logrus.Init(
		logrus.WithContextFields("key1", "key2"),
		logrus.WithFileLog("/tmp", "benchmarking"),
		logrus.WithDisableStdout(),
	)

	ctx := context.WithValue(context.Background(), "key1", "value1")
	ctx = context.WithValue(ctx, "key2", "value2")
	ctx = context.WithValue(ctx, "key3", "value3")

	for i := 0; i < b.N; i++ {
		log.WithContext(ctx).Debugf("benchmarking log")
	}
}
