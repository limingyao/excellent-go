package collector

import "github.com/prometheus/client_golang/prometheus"

type diskUsageCollector struct {
}

type DiskUsageCollectorOpts struct {
}

func NewDiskUsageCollector(opts DiskUsageCollectorOpts) prometheus.Collector {
	c := &diskUsageCollector{}

	return c
}

func (c *diskUsageCollector) Describe(ch chan<- *prometheus.Desc) {
	// TODO implement me
}

func (c *diskUsageCollector) Collect(ch chan<- prometheus.Metric) {
	// TODO implement me
}
