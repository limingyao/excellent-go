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

func HandleDefault(httpMux *http.ServeMux, opts ...Option) {
	httpMux.Handle("/metrics", Handler(opts...))
}

func Handler(opts ...Option) http.Handler {
	defaultOpts := defaultOptions
	for _, o := range opts {
		o.apply(&defaultOpts)
	}

	if !defaultOpts.disableProcessCollector {
		MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	}
	if !defaultOpts.disableProcessExtCollector {
		MustRegister(collector.NewProcessExtCollector(collector.ProcessExtCollectorOpts{}))
	}
	if defaultOpts.enableDiskUsageCollector {
		MustRegister(collector.NewDiskUsageCollector(collector.DiskUsageCollectorOpts{}))
	}
	if defaultOpts.enableGoCollector {
		MustRegister(collectors.NewGoCollector())
	}
	if !defaultOpts.disableGoSimpleCollector {
		MustRegister(collector.NewGoSimpleCollector(collector.GoSimpleCollectorOpts{}))
	}
	if defaultOpts.disableInstrumentMetricHandler {
		return promhttp.HandlerFor(defaultRegister, promhttp.HandlerOpts{})
	}
	return promhttp.InstrumentMetricHandler(
		defaultRegister, promhttp.HandlerFor(defaultRegister, promhttp.HandlerOpts{}),
	)
}
