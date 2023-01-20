package collector

import "github.com/prometheus/client_golang/prometheus"

type goSimpleCollector struct {
}

type GoSimpleCollectorOpts struct {
	PidFn        func() (int, error)
	ReportErrors bool
}

func NewGoSimpleCollector(opts GoSimpleCollectorOpts) prometheus.Collector {
	c := &goSimpleCollector{}

	return c
}

func (c *goSimpleCollector) Describe(ch chan<- *prometheus.Desc) {
	// TODO implement me
}

func (c *goSimpleCollector) Collect(ch chan<- prometheus.Metric) {
	// TODO implement me
}
