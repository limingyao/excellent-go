package prometheus

import (
	"net/http"

	"github.com/limingyao/excellent-go/metrics/prometheus/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	defaultRegister = prometheus.NewRegistry()
)

func MustRegister(cs ...prometheus.Collector) {
	defaultRegister.MustRegister(cs...)
}

func RegisterDefault() {
	MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	MustRegister(collector.NewProcessExtCollector(collector.ProcessExtCollectorOpts{}))
	MustRegister(collectors.NewGoCollector())
}

func InstrumentMetricHandler() http.Handler {
	return promhttp.InstrumentMetricHandler(
		defaultRegister, promhttp.HandlerFor(defaultRegister, promhttp.HandlerOpts{}),
	)
}

func Handler() http.Handler {
	return promhttp.HandlerFor(defaultRegister, promhttp.HandlerOpts{})
}
