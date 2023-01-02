package collector

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func loop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Nanosecond)
	defer ticker.Stop()

	i := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			i = i + 1
		}
	}
}

func TestNewProcessExtCollector(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go loop(ctx)

	collector := NewProcessExtCollector(CollectorOption{ReportErrors: true})
	metrics := make(chan prometheus.Metric)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				collector.Collect(metrics)
			}
		}
	}()

	go func() {
		for metric := range metrics {
			t.Log(metric.Desc())
			m := &dto.Metric{}
			if err := metric.Write(m); err != nil {
				t.Error(err)
			}
			if m.Counter != nil {
				t.Log(m.Counter)
			}
			if m.Gauge != nil {
				t.Log(m.Gauge)
			}
			if m.Summary != nil {
				t.Log(m.Summary)
			}
			if m.Histogram != nil {
				t.Log(m.Histogram)
			}
		}
	}()

	time.Sleep(5 * time.Second)
}
